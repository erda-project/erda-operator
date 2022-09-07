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
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/erda-project/dice-operator/pkg/cluster/check"
	"github.com/erda-project/dice-operator/pkg/cluster/daemonset"
	"github.com/erda-project/dice-operator/pkg/cluster/deployment"
	"github.com/erda-project/dice-operator/pkg/cluster/diff"
	"github.com/erda-project/dice-operator/pkg/cluster/hpa"
	"github.com/erda-project/dice-operator/pkg/cluster/launch"
	"github.com/erda-project/dice-operator/pkg/cluster/status"
	"github.com/erda-project/dice-operator/pkg/cluster/vpa"
	"github.com/erda-project/dice-operator/pkg/spec"
	statusop "github.com/erda-project/dice-operator/pkg/status"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type Syncer struct {
	target       *spec.DiceCluster
	k8sclient    kubernetes.Interface
	restclient   rest.Interface
	clientconfig *rest.Config
	clus         *Cluster
}

func NewSyncer(clus *Cluster) *Syncer {
	return &Syncer{
		target:       clus.target,
		k8sclient:    clus.k8sclient,
		restclient:   clus.client,
		clientconfig: clus.clientconfig,
		clus:         clus,
	}
}

func sprintMapKeys(m map[string]*diceyml.Service) string {
	r := []string{}
	for k := range m {
		r = append(r, k)
	}
	return fmt.Sprintf("%v", r)
}

func (c *Syncer) Sync() {
	vpaClientset := vpa_clientset.NewForConfigOrDie(c.clientconfig)
	actions := diff.NewSpecDiff(nil, c.target).GetActions()
	deployments := actions.AddedServices

	daemonsets := actions.AddedDaemonSet

	dsNeedToUpdate, dsNeedToAdd, dsNeedToDelete, err := c.checkDaemonsets(daemonsets)
	if err != nil {
		logrus.Errorf("Failed to Sync: %v", err)
		return
	}
	if len(dsNeedToUpdate)+len(dsNeedToAdd)+len(dsNeedToDelete) > 0 {
		logrus.Infof("sync daemonsets: UPDATE: %s, ADD: %v, DELETE: %v",
			sprintMapKeys(dsNeedToUpdate), sprintMapKeys(dsNeedToAdd), sprintMapKeys(dsNeedToDelete))
	}
	deployNeedToUpdate, deployNeedToAdd, deployNeedToDelete, err := c.checkDeployments(deployments)
	if err != nil {
		logrus.Errorf("Failed to Sync: %v", err)
		return
	}
	if len(deployNeedToUpdate)+len(deployNeedToAdd)+len(deployNeedToDelete) > 0 {
		logrus.Infof("sync deployments: UPDATE: %v, ADD: %v, DELETE: %v",
			sprintMapKeys(deployNeedToUpdate), sprintMapKeys(deployNeedToAdd), sprintMapKeys(deployNeedToDelete))
	}

	needPAs := c.checkPodAutoscalers(deployments, daemonsets, vpaClientset)
	if len(needPAs) > 0 {
		logrus.Infof("sync pod autoscalers: %v", sprintMapKeys(needPAs))
	}

	syncactions := diff.Actions{
		AddedServices:       deployNeedToAdd,
		UpdatedServices:     deployNeedToUpdate,
		DeletedServices:     deployNeedToDelete,
		AddedDaemonSet:      dsNeedToAdd,
		UpdatedDaemonSet:    dsNeedToUpdate,
		DeletedDaemonSet:    dsNeedToDelete,
		UpdatedServicesPA:   needPAs,
		EnableAutoScaleDiff: actions.EnableAutoScaleDiff,
	}

	launcher := launch.NewLauncher(&syncactions,
		c.target, c.clus.ownerRefs, c.k8sclient, vpaClientset, c.restclient, c.target.Status.Phase, c.clus.serviceToPA)
	if err := launcher.Launch(); err != nil {
		logrus.Printf("launch failed when sync, err: %s", err)
	}
	if err := status.New(c.k8sclient, c.restclient, c.target).Update(c.target.Name); err != nil {
		logrus.Errorf("status updater err: %v", err)
	}
}

func (c *Syncer) checkDeployments(dicesvcs map[string]*diceyml.Service) (
	needToUpdate, needToAdd, needToDelete map[string]*diceyml.Service, err error) {
	needToUpdate = make(map[string]*diceyml.Service)
	needToAdd = make(map[string]*diceyml.Service)
	needToDelete = make(map[string]*diceyml.Service)

	var deploylist *appsv1.DeploymentList
	deploylist, err = c.k8sclient.AppsV1().Deployments(c.target.Namespace).
		List(context.Background(), metav1.ListOptions{LabelSelector: "dice/koperator=true," +
			fmt.Sprintf("dice/cluster-name=%s", c.target.Name)})
	if err != nil {
		return
	}

	currentDeployList := deploylist.Items

	generatedDeployList := []appsv1.Deployment{}
	for name, dicesvc := range dicesvcs {
		var deploy *appsv1.Deployment
		deploy, err = deployment.BuildDeployment(name, dicesvc, c.target, c.clus.ownerRefs)
		if err != nil {
			return
		}
		generatedDeployList = append(generatedDeployList, *deploy)
	}
	actions := diff.NewDeploymentListDiff(currentDeployList, generatedDeployList).GetActions()
	componentStatusMap := map[string]spec.ComponentStatus{}

	for _, deploy := range actions.UpdatedDeployments {
		dicesvcname := deployment.ExtractDiceSvcName(deploy.Name)
		needToUpdate[dicesvcname] = dicesvcs[dicesvcname]
		componentStatusMap[dicesvcname] = spec.ComponentStatusNeedCreateOrUpdate
	}

	for _, deploy := range actions.AddedDeployments {
		dicesvcname := deployment.ExtractDiceSvcName(deploy.Name)
		needToAdd[dicesvcname] = dicesvcs[dicesvcname]
		componentStatusMap[dicesvcname] = spec.ComponentStatusNeedCreateOrUpdate
	}

	for _, deploy := range actions.DeletedDeployments {
		dicesvcname := deployment.ExtractDiceSvcName(deploy.Name)
		needToDelete[dicesvcname] = dicesvcs[dicesvcname]
	}

	for i := range deploylist.Items {
		name := deployment.ExtractDiceSvcName(deploylist.Items[i].Name)
		if !check.CheckDeploymentAvailable(&deploylist.Items[i]) {
			if _, ok := componentStatusMap[name]; !ok {
				componentStatusMap[name] = spec.ComponentStatusUnReady
			}
		} else {
			if _, ok := componentStatusMap[name]; !ok {
				componentStatusMap[name] = spec.ComponentStatusReady
			}
		}
	}

	c.syncComponentStatus(componentStatusMap)
	allready := true
	for _, v := range componentStatusMap {
		if v != spec.ComponentStatusReady {
			allready = false
		}
	}
	if allready {
		statusop.UpdatePhase(c.restclient, c.target, c.target.Namespace, c.target.Name, spec.ClusterPhaseRunning)
	}
	return
}

func (c *Syncer) checkDaemonsets(dicesvcs map[string]*diceyml.Service) (
	needToUpdate, needToAdd, needToDelete map[string]*diceyml.Service, err error) {
	needToUpdate = make(map[string]*diceyml.Service)
	needToAdd = make(map[string]*diceyml.Service)
	needToDelete = make(map[string]*diceyml.Service)

	var dslist *appsv1.DaemonSetList
	dslist, err = c.k8sclient.AppsV1().DaemonSets(c.target.Namespace).
		List(context.Background(), metav1.ListOptions{LabelSelector: "dice/koperator=true," +
			fmt.Sprintf("dice/cluster-name=%s", c.target.Name)})
	if err != nil {
		return
	}
	currentDSList := dslist.Items
	generatedDSList := []appsv1.DaemonSet{}
	for name, dicesvc := range dicesvcs {
		var ds *appsv1.DaemonSet
		ds, err = daemonset.BuildDaemonSet(name, dicesvc, c.target, c.clus.ownerRefs)
		if err != nil {
			return
		}
		generatedDSList = append(generatedDSList, *ds)
	}
	actions := diff.NewDaemonsetListDiff(currentDSList, generatedDSList).GetActions()

	for _, ds := range actions.UpdatedDaemonsets {
		dicesvcname := daemonset.ExtractDiceSvcName(ds.Name)
		needToUpdate[dicesvcname] = dicesvcs[dicesvcname]
	}
	for _, ds := range actions.AddedDaemonsets {
		dicesvcname := daemonset.ExtractDiceSvcName(ds.Name)
		needToAdd[dicesvcname] = dicesvcs[dicesvcname]
	}
	for _, ds := range actions.DeletedDaemonsets {
		dicesvcname := daemonset.ExtractDiceSvcName(ds.Name)
		needToDelete[dicesvcname] = dicesvcs[dicesvcname]
	}

	return
}

func (c *Syncer) checkPodAutoscalers(deployments, daemonsets map[string]*diceyml.Service, vpaClientSet *vpa_clientset.Clientset) (
	needToPA map[string]*diceyml.Service) {

	needToPA = make(map[string]*diceyml.Service)

	vpaList, err := vpa.ListHPAInNamespace(vpaClientSet, c.target)
	if err != nil {
		logrus.Errorf("list vpa in namespace %s error: %v", c.target.Namespace, err)
		return needToPA
	}

	hpaList, err := hpa.ListHPAInNamespace(c.k8sclient, c.target)
	if err != nil {
		logrus.Errorf("list vpa in namespace %s error: %v", c.target.Namespace, err)
		return needToPA
	}

	if c.target.Spec.EnableAutoScale {
		for svcName, diceSvc := range deployments {
			if _, ok := c.clus.serviceToPA[svcName]; ok {
				needToPA[svcName] = diceSvc
			}
		}
		for svcName, diceSvc := range daemonsets {
			if _, ok := c.clus.serviceToPA[svcName]; ok {
				needToPA[svcName] = diceSvc
			}
		}

		for _, hpa := range hpaList.Items {
			svcName := strings.SplitN(hpa.Name, "-", 2)[1]
			if _, ok := needToPA[svcName]; ok {
				delete(needToPA, svcName)
			}
		}

		for _, vpa := range vpaList.Items {
			svcName := strings.SplitN(vpa.Name, "-", 2)[1]
			if _, ok := needToPA[svcName]; ok {
				delete(needToPA, svcName)
			}
		}
	} else {
		for _, hpa := range hpaList.Items {
			svcName := strings.SplitN(hpa.Name, "-", 2)[1]
			if _, ok := deployments[svcName]; ok {
				needToPA[svcName] = deployments[svcName]
				continue
			}
			if _, ok := daemonsets[svcName]; ok {
				needToPA[svcName] = daemonsets[svcName]
				continue
			}
		}

		for _, vpa := range vpaList.Items {
			svcName := strings.SplitN(vpa.Name, "-", 2)[1]
			if _, ok := deployments[svcName]; ok {
				needToPA[svcName] = deployments[svcName]
				continue
			}
			if _, ok := daemonsets[svcName]; ok {
				needToPA[svcName] = daemonsets[svcName]
				continue
			}
		}
	}

	return needToPA
}

func (c *Syncer) syncComponentStatus(componentStatus map[string]spec.ComponentStatus) {
	if err := statusop.UpdateComponentStatus(c.restclient, c.target.Namespace, c.target.Name, componentStatus); err != nil {
		logrus.Errorf("Failed to update ComponentStatus, err: %v", err)
	}
}
