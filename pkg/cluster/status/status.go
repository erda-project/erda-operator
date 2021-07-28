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

package status

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/erda-project/dice-operator/pkg/cluster/check"
	"github.com/erda-project/dice-operator/pkg/cluster/daemonset"
	"github.com/erda-project/dice-operator/pkg/cluster/deployment"
	"github.com/erda-project/dice-operator/pkg/spec"
	statusop "github.com/erda-project/dice-operator/pkg/status"
)

type StatusUpdater struct {
	k8sclient  kubernetes.Interface
	restclient rest.Interface
	target     *spec.DiceCluster
}

func New(k8sclient kubernetes.Interface, restclient rest.Interface, target *spec.DiceCluster) *StatusUpdater {
	return &StatusUpdater{
		k8sclient:  k8sclient,
		restclient: restclient,
		target:     target,
	}
}

func (u *StatusUpdater) Update(crdName string) error {
	dsList, err := getDaemonsetList(u.k8sclient, u.target.Namespace, crdName)
	if err != nil {
		return err
	}
	deployList, err := getDeploymentList(u.k8sclient, u.target.Namespace, crdName)
	if err != nil {
		return err
	}

	//logrus.Printf("status: dsList: %v\n", dsList)
	//logrus.Printf("status: deployList: %s\n", deployList)
	statuses := computeStatus(deployList, dsList)
	if err := statusop.UpdateComponentStatus(u.restclient, u.target.Namespace, u.target.Name, statuses); err != nil {
		return err
	}
	return nil
}

func computeStatus(deployList *appsv1.DeploymentList, dsList *appsv1.DaemonSetList) map[string]spec.ComponentStatus {
	result := map[string]spec.ComponentStatus{}
	for _, deploy := range deployList.Items {
		svcName := deployment.ExtractDiceSvcName(deploy.Name)
		if check.CheckDeploymentAvailable(&deploy) {
			result[svcName] = spec.ComponentStatusReady
		} else {
			result[svcName] = spec.ComponentStatusUnReady
		}
	}
	for _, ds := range dsList.Items {
		svcName := daemonset.ExtractDiceSvcName(ds.Name)
		if check.CheckDaemonsetAvailable(&ds) {
			result[svcName] = spec.ComponentStatusReady
		} else {
			result[svcName] = spec.ComponentStatusUnReady
		}
	}
	return result
}

func getDeploymentList(k8sclient kubernetes.Interface, namespace, crdName string) (*appsv1.DeploymentList, error) {
	return k8sclient.AppsV1().Deployments(namespace).
		List(context.Background(), metav1.ListOptions{LabelSelector: "dice/koperator=true," + fmt.Sprintf("dice/cluster-name=%s", crdName)})
}

func getDaemonsetList(k8sclient kubernetes.Interface, namespace, crdName string) (*appsv1.DaemonSetList, error) {
	return k8sclient.AppsV1().DaemonSets(namespace).
		List(context.Background(), metav1.ListOptions{LabelSelector: "dice/koperator=true," + fmt.Sprintf("dice/cluster-name=%s", crdName)})
}
