// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cluster

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/erda-project/dice-operator/pkg/cluster/diff"
	"github.com/erda-project/dice-operator/pkg/cluster/jobs"
	"github.com/erda-project/dice-operator/pkg/cluster/launch"
	"github.com/erda-project/dice-operator/pkg/cluster/status"
	"github.com/erda-project/dice-operator/pkg/crd"
	"github.com/erda-project/dice-operator/pkg/envs"
	"github.com/erda-project/dice-operator/pkg/spec"
	statusop "github.com/erda-project/dice-operator/pkg/status"
)

type Cluster struct {
	namespace string
	target    *spec.DiceCluster
	client    rest.Interface
	k8sclient kubernetes.Interface

	ownerRefs []metav1.OwnerReference

	stopCh   chan struct{}
	updateCh chan spec.DiceCluster
}

func New(specdc *spec.DiceCluster, client rest.Interface, k8sclient kubernetes.Interface) (*Cluster, error) {
	logrus.Infof("starting dice cluster: %s/%s", specdc.Namespace, specdc.Name)
	c := &Cluster{
		target:    specdc,
		client:    client,
		k8sclient: k8sclient,
		ownerRefs: buildOwnerRefs(specdc),
		stopCh:    make(chan struct{}, 1),
		updateCh:  make(chan spec.DiceCluster, 10),
	}
	// replace all built-in envs
	if err := injectEnvs(k8sclient, specdc); err != nil {
		return nil, err
	}
	if specdc.Status.Phase == spec.ClusterPhaseNone {
		if err := c.create(); err != nil {
			return nil, err
		}
	}
	logrus.Infof("Starting periodicSync: %s/%s", c.target.Namespace, c.target.Name)
	go c.PeriodicSync()
	return c, nil
}

func (c *Cluster) create() error {
	if len(c.target.Spec.MainPlatform) == 0 {
		// deploy center cluster

		// execute init jobs
		statusop.UpdateConditionAndPhase(c.client, c.target, c.target.Namespace, c.target.Name,
			spec.Condition{Reason: "execute init jobs"}, spec.ClusterPhaseInitJobs)

		if err := jobs.CreateAndWait(c.k8sclient, c.target.Spec.InitJobs.Jobs, c.target, c.ownerRefs); err != nil {
			statusop.UpdateConditionAndPhase(c.client, c.target, c.target.Namespace, c.target.Name,
				spec.Condition{Reason: fmt.Sprintf("create init jobs failed: %v", err)}, spec.ClusterPhaseFailed)
			return nil
		}
	}

	// start deploy components
	statusop.UpdateConditionAndPhase(c.client, c.target, c.target.Namespace, c.target.Name,
		spec.Condition{Reason: "starting dice cluster"}, spec.ClusterPhaseCreating)
	actions := diff.NewSpecDiff(nil, c.target).GetActions()
	logrus.Infof("cluster: %s/%s, actions: %+v", c.target.Namespace, c.target.Name, actions)
	// init all componentStatus
	allsvc := map[string]spec.ComponentStatus{}
	for svc := range actions.AddedServices {
		allsvc[svc] = spec.ComponentStatusUnReady
	}
	for svc := range actions.AddedDaemonSet {
		allsvc[svc] = spec.ComponentStatusUnReady
	}
	statusop.UpdateComponentStatus(c.client, c.target.Namespace, c.target.Name, allsvc)

	launcher := launch.NewLauncher(actions, c.target, c.ownerRefs, c.k8sclient, c.client, spec.ClusterPhaseCreating)
	if err := launcher.Launch(); err != nil {
		// error or pending status already updated in Launch()
		return nil
	}
	if err := status.New(c.k8sclient, c.client, c.target).Update(c.target.Name); err != nil {
		return err
	}
	statusop.UpdateConditionAndPhase(c.client, c.target, c.target.Namespace, c.target.Name,
		spec.Condition{Reason: "create dice cluster success"}, spec.ClusterPhaseRunning)
	return nil
}

func (c *Cluster) Update(newspec *spec.DiceCluster) {
	if err := injectEnvs(c.k8sclient, newspec); err != nil {
		logrus.Errorf("Failed to injectEnvs: %v", err)
		return
	}
	actions := diff.NewSpecDiff(c.target, newspec).GetActions()
	if actions.EmptyAction() {
		return
	}
	c.updateCh <- *newspec
	logrus.Infof("Request a update: %s/%s", newspec.Namespace, newspec.Name)
}

func (c *Cluster) update(newspec spec.DiceCluster) {
	c.target = &newspec
}

func (c *Cluster) Delete() {
	// Because 'ownerRefs' has been set, when you delete 'dice' object,
	// all objects belonging to this dice cluster are automatically deleted.
	logrus.Infof("Deleting dice cluster: %s/%s", c.target.Namespace, c.target.Name)
	c.stopCh <- struct{}{}
}

func (c *Cluster) PeriodicSync() {
	for {
		select {
		case <-c.stopCh:
			logrus.Infof("Quit periodicSync: %s/%s", c.target.Namespace, c.target.Name)
			return
		case newspec := <-c.updateCh:
			c.update(newspec)
		case <-time.After(15 * time.Second):
			if c.target.Status.Phase != spec.ClusterPhaseRunning &&
				c.target.Status.Phase != spec.ClusterPhasePending &&
				c.target.Status.Phase != spec.ClusterPhaseCreating {
				// if 'resetStatus==true', reset status to running
				resetStatus(c)
				break
			}
			syncer := NewSyncer(c)
			syncer.Sync()
		}
	}
}

func buildOwnerRefs(clus *spec.DiceCluster) []metav1.OwnerReference {
	blockOwnerDeletion := true
	isController := true
	return []metav1.OwnerReference{{
		APIVersion:         crd.GetCRDGroupVersion(),
		Kind:               crd.CRDKind,
		Name:               clus.Name,
		UID:                clus.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}}
}

// injectEnvs inject envs from addon configmap and predefined values
func injectEnvs(client kubernetes.Interface, cluster *spec.DiceCluster) error {
	addonData, err := spec.GetAddonConfigMap(client, cluster)
	if err != nil {
		logrus.Errorf("failed to get addon ConfigMap %v", err)
	}
	clusterInfo, err := spec.GetClusterInfoConfigMap(client, cluster)
	if err != nil {
		logrus.Errorf("failed to get ClusterInfo ConfigMap %v", err)
		clusterInfo = map[string]string{}
	}
	injectEnvs := envs.GenDiceSvcENVs(cluster, addonData, clusterInfo)
	envs.InjectENVs(clusterInfo, injectEnvs, cluster)
	return nil
}

func resetStatus(c *Cluster) {
	if !c.target.Spec.ResetStatus {
		return
	}
	err := statusop.UpdatePhase(c.client, c.target, c.target.Namespace, c.target.Name, spec.ClusterPhaseRunning)
	if err != nil {
		return
	}
	statusop.RevertResetStatus(c.client, c.target.Namespace, c.target.Name)
}
