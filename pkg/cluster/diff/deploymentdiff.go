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

package diff

import (
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/pkg/strutil"
)

type DeploymentListDiff struct {
	currentDeployments map[string]appsv1.Deployment
	targetDeployments  map[string]appsv1.Deployment
}

func NewDeploymentListDiff(current, target []*appsv1.Deployment) *DeploymentListDiff {
	currentDeployments := map[string]appsv1.Deployment{}
	targetDeployments := map[string]appsv1.Deployment{}

	for _, deploy := range current {
		currentDeployments[deploy.Name] = *deploy
	}

	for _, deploy := range target {
		targetDeployments[deploy.Name] = *deploy
	}

	return &DeploymentListDiff{
		currentDeployments: currentDeployments,
		targetDeployments:  targetDeployments,
	}
}

type DeploymentListActions struct {
	AddedDeployments   map[string]appsv1.Deployment
	UpdatedDeployments map[string]appsv1.Deployment
	DeletedDeployments map[string]appsv1.Deployment
}

func (a *DeploymentListActions) String() string {
	add, update, delete := []string{}, []string{}, []string{}
	for _, deploy := range a.AddedDeployments {
		add = append(add, deploy.Name)
	}
	for _, deploy := range a.UpdatedDeployments {
		update = append(update, deploy.Name)
	}
	for _, deploy := range a.DeletedDeployments {
		delete = append(delete, deploy.Name)
	}

	return fmt.Sprintf("DeploymentListActions: ADD: [%s], UPDATE: [%s], DELETE: [%s]",
		strutil.Join(add, ", "), strutil.Join(update, ", "), strutil.Join(delete, ", "))
}

func (d *DeploymentListDiff) GetActions() *DeploymentListActions {
	r := &DeploymentListActions{}
	missingInSet1, missingInSet2, shared := diffDeploySet(d.currentDeployments, d.targetDeployments)
	r.AddedDeployments = missingInSet1
	r.DeletedDeployments = missingInSet2
	r.UpdatedDeployments = getNotEqualDeployments(d.currentDeployments, d.targetDeployments, shared)
	return r
}

func diffDeploySet(set1, set2 map[string]appsv1.Deployment) (
	missingInSet1, missingInSet2, shared map[string]appsv1.Deployment) {
	missingInSet1 = make(map[string]appsv1.Deployment)
	missingInSet2 = make(map[string]appsv1.Deployment)
	shared = make(map[string]appsv1.Deployment)

	for k, v := range set1 {
		if v2, ok := set2[k]; !ok {
			missingInSet2[k] = v
		} else {
			shared[k] = v2
		}
	}

	for k, v := range set2 {
		if _, ok := set1[k]; !ok {
			missingInSet1[k] = v
		}
	}
	return
}

func isDeploymentEqual(deploy1, deploy2 appsv1.Deployment) bool {
	a1, a2 := getDiceAnnotations(deploy1.Annotations, deploy2.Annotations)
	if !isAnnotationsEqual(a1, a2) {
		logrus.Infof("diff annotations, %s/%s: %v -> %v", deploy1.Namespace, deploy2.Name, a1, a2)
		return false
	}

	if !reflect.DeepEqual(deploy1.Spec.Replicas, deploy2.Spec.Replicas) {
		logrus.Infof("diff replicas: %s/%s: %d -> %d",
			deploy1.Namespace, deploy1.Name, *deploy1.Spec.Replicas, *deploy2.Spec.Replicas)
		return false
	}
	if len(deploy1.Spec.Template.Spec.Containers) != len(deploy2.Spec.Template.Spec.Containers) {
		return false
	}
	containerSet1 := map[string]corev1.Container{}
	containerSet2 := map[string]corev1.Container{}

	for _, c := range deploy1.Spec.Template.Spec.Containers {
		containerSet1[c.Name] = c
	}
	for _, c := range deploy2.Spec.Template.Spec.Containers {
		containerSet2[c.Name] = c
		if _, ok := containerSet1[c.Name]; !ok {
			return false
		}
	}
	for name := range containerSet1 {
		if !isContainerEqual(containerSet1[name], containerSet2[name],
			fmt.Sprintf("%s/%s", deploy1.Namespace, deploy1.Name), true) {
			return false
		}
	}
	return true
}

func isContainerEqual(container1, container2 corev1.Container, location string, deploy bool) bool {
	imageEqual := container1.Image == container2.Image
	cmdEqual := reflect.DeepEqual(container1.Command, container2.Command) &&
		reflect.DeepEqual(container1.Args, container2.Args)
	envEqual := isEnvsEqual(container1.Env, container2.Env,
		fmt.Sprintf("%s/%s", location, container1.Name))
	portsEqual := isPortsEqual(container1.Ports, container2.Ports, deploy)
	resourceEqual := !deploy ||
		container1.Resources.Requests.Cpu().Equal(*container2.Resources.Requests.Cpu()) &&
			container1.Resources.Requests.Memory().Equal(*container2.Resources.Requests.Memory()) &&
			container1.Resources.Limits.Cpu().Equal(*container2.Resources.Limits.Cpu()) &&
			container1.Resources.Limits.Memory().Equal(*container2.Resources.Limits.Memory())
	r := imageEqual && cmdEqual && envEqual && resourceEqual && portsEqual
	if !r {
		if !imageEqual {
			logrus.Infof("diff image: %s/%s: %s -> %s",
				location, container1.Name,
				container1.Image, container2.Image)
		}

		if !cmdEqual {
			logrus.Infof("diff cmd: %s/%s: %v -> %v",
				location, container1.Name,
				append(container1.Command, container1.Args...),
				append(container2.Command, container2.Args...),
			)
		}

		if !portsEqual {
			logrus.Infof("diff ports: %s/%s: %+v -> %+v",
				location, container1.Name, container1.Ports, container2.Ports)
		}

		// env diff logs in `isEnvsEqual`
		if !resourceEqual {
			resource1 := fmt.Sprintf("[req: [cpu: %s, mem: %s], limit: [cpu: %s, mem: %s]]",
				container1.Resources.Requests.Cpu(), container1.Resources.Requests.Memory(),
				container1.Resources.Limits.Cpu(), container1.Resources.Limits.Memory())
			resource2 := fmt.Sprintf("[req: [cpu: %s, mem: %s], limit: [cpu: %s, mem: %s]]",
				container2.Resources.Requests.Cpu(), container2.Resources.Requests.Memory(),
				container2.Resources.Limits.Cpu(), container2.Resources.Limits.Memory())

			logrus.Infof("diff resource: %s/%s: %s -> %s",
				location, container1.Name, resource1, resource2)
		}
	}
	return r
}

func isEnvsEqual(envs1, envs2 []corev1.EnvVar, location string) bool {
	env1map := map[string]corev1.EnvVar{}
	env2map := map[string]corev1.EnvVar{}
	for _, env := range envs1 {
		env1map[env.Name] = env
	}
	for _, env := range envs2 {
		env2map[env.Name] = env
	}
	for _, env2 := range envs2 {
		if env1, ok := env1map[env2.Name]; !ok || env1.Value != env2.Value {
			logrus.Infof("diff env %s:%s  src_env:%s  tag_env:%s", location, env2.Name, env1.Value, env2.Value)
			return false
		}
	}
	for _, env1 := range envs1 {
		if env2, ok := env2map[env1.Name]; !ok || env1.Value != env2.Value {
			logrus.Infof("diff env %s:%s  src_env:%s  target_env:%s", location, env1.Name, env1.Value, env2.Value)
			return false
		}
	}

	return true
}

func getNotEqualDeployments(set1, set2, shared map[string]appsv1.Deployment) map[string]appsv1.Deployment {
	r := map[string]appsv1.Deployment{}
	for k := range shared {
		if !isDeploymentEqual(set1[k], set2[k]) {
			r[k] = set2[k]
		}
	}
	return r
}

func isPortsEqual(ports1, ports2 []corev1.ContainerPort, isDeploy bool) bool {
	if !isDeploy {
		return true
	}

	if len(ports1) != len(ports2) {
		return false
	}

	for _, p1 := range ports1 {

		if portCount(p1.ContainerPort, ports1) != portCount(p1.ContainerPort, ports2) {
			return false
		}

		p2Index := portIndex(p1.ContainerPort, ports2)
		if p2Index < int32(0) {
			return false
		}

		p2 := ports2[p2Index]
		if p1.Protocol != p2.Protocol {
			return false
		}
	}

	return true
}

func portCount(port int32, ports []corev1.ContainerPort) int32 {
	count := int32(0)

	for _, p := range ports {
		if p.ContainerPort == port {
			count++
		}
	}

	return count
}

func portIndex(port int32, ports []corev1.ContainerPort) int32 {
	index := int32(-1)

	for idx, p := range ports {
		if port == p.ContainerPort {
			index = int32(idx)
			break
		}
	}

	return index
}
