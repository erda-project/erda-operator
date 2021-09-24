// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-test/deep"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg"
	"github.com/erda-project/erda-operator/pkg/utils"
)

func (r *ErdaReconciler) ReconcileWorkload(ctx context.Context,
	component erdav1beta1.Component, references []metav1.OwnerReference) (error, bool) {

	workLoadErr, needUpdateStatus := r.CreateOrUpdateWorkLoad(ctx, &component, references)
	if workLoadErr != nil {
		r.Log.Error(workLoadErr, "handle workload error", component.WorkLoad, component.Name, component.Namespace)
		return workLoadErr, needUpdateStatus
	}

	if len(component.Network.ServiceDiscovery) > 0 {
		k8sServiceErr := r.CreateOrUpdateKubernetesService(ctx, &component, references)
		if k8sServiceErr != nil {
			r.Log.Error(k8sServiceErr, "handle component error", component.Name, component.Namespace)
			return k8sServiceErr, needUpdateStatus
		}
		ingressCount := 0
		for _, sd := range component.Network.ServiceDiscovery {
			if sd.Domain != "" {
				ingressCount++
			}
		}
		if ingressCount > 0 {
			ingressErr := r.CreateOrUpdateIngress(ctx, &component, references)
			if ingressErr != nil {
				r.Log.Error(ingressErr, "handle ingress error")
				return ingressErr, needUpdateStatus
			}
		}
	}
	return nil, needUpdateStatus
}

func (r *ErdaReconciler) CreateOrUpdateWorkLoad(ctx context.Context,
	component *erdav1beta1.Component, owners []metav1.OwnerReference) (error, bool) {

	var (
		obj    client.Object
		newObj client.Object
	)

	switch component.WorkLoad {
	case erdav1beta1.PerNode:
		obj = &appsv1.DaemonSet{}
		newObj = pkg.ComposeDaemonSet(component, owners)
	case erdav1beta1.Stateless:
		obj = &appsv1.Deployment{}
		newObj = pkg.ComposeDeployment(component, owners)
	case erdav1beta1.Stateful:
		obj = &appsv1.StatefulSet{}
		newObj = pkg.ComposeStatefulSet(component, owners)
	}

	err := r.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, obj)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err, false
		}
		return r.Create(ctx, newObj), true
	}

	workload, err := r.DiffResource(obj, newObj)
	if err != nil {
		return err, false
	}
	if workload != nil {
		return r.Update(ctx, workload), true
	}
	return nil, false
}

func (r *ErdaReconciler) DiffResource(oldObj, newObj client.Object) (client.Object, error) {

	if reflect.DeepEqual(oldObj.GetObjectKind(), newObj.GetObjectKind()) {
		return nil, fmt.Errorf("different kind between oldOjb %s and newObj %s", oldObj.GetObjectKind(), newObj.GetObjectKind())
	}
	if oldDeployment, ok := oldObj.(*appsv1.Deployment); ok {
		newDeployment := newObj.(*appsv1.Deployment)
		deploySpec := pkg.ComposeDeploymentSpecFromK8sDeployment(oldDeployment)

		if equal := deep.Equal(deploySpec, newDeployment.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name: %s diff object is %+v", newDeployment.Name, equal))
			return newDeployment, nil
		}
	}
	if oldDaemonSet, ok := oldObj.(*appsv1.DaemonSet); ok {
		newDaemonSet := newObj.(*appsv1.DaemonSet)
		daemonSetSpec := pkg.ComposeDaemonSetSpecFromK8sStatefulSet(oldDaemonSet)
		if equal := deep.Equal(daemonSetSpec, newDaemonSet.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name: %s diff object is %+v", newDaemonSet.Name, equal))
			return newDaemonSet, nil
		}
	}
	if oldStatefulSet, ok := oldObj.(*appsv1.StatefulSet); ok {
		newStatefulSet := newObj.(*appsv1.DaemonSet)
		statefulSetSpec := pkg.ComposeStatefulSetSpecFromK8sStatefulSet(oldStatefulSet)
		if equal := deep.Equal(statefulSetSpec, newStatefulSet.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name: %s diff object is %+v", newStatefulSet.Name, equal))
			return newStatefulSet, nil
		}
	}
	if oldK8sService, ok := oldObj.(*corev1.Service); ok {
		newK8sService := newObj.(*corev1.Service)
		k8sServiceSpec := pkg.ComposeKubernetesServiceSpecFromK8sService(oldK8sService)
		if equal := deep.Equal(k8sServiceSpec, newK8sService.Spec); equal != nil {
			newK8sService.ResourceVersion = oldK8sService.ResourceVersion
			newK8sService.Spec.ClusterIP = oldK8sService.Spec.ClusterIP
			r.Log.Info(fmt.Sprintf("name %s diff object is %+v", newK8sService.Name, equal))
			return newK8sService, nil
		}
	}
	if oldIngress, ok := oldObj.(*networkingv1beta1.Ingress); ok {
		newIngress := newObj.(*networkingv1beta1.Ingress)
		ingressSpec := pkg.ComposeIngressV1Beta1SpecFromK8sIngress(oldIngress)
		if equal := deep.Equal(ingressSpec, newIngress.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name %s diff object is %+v", newIngress.Name, equal))
			return newIngress, nil
		}
	}
	if oldIngress, ok := oldObj.(*networkingv1.Ingress); ok {
		newIngress := newObj.(*networkingv1.Ingress)
		ingressSpec := pkg.ComposeIngressV1SpecFromK8sIngress(oldIngress)
		if equal := deep.Equal(ingressSpec, newIngress.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name %s diff object is %+v", newIngress.Name, equal))
			return newIngress, nil
		}
	}
	return nil, nil
}

func (r *ErdaReconciler) DeleteWorkLoads(ctx context.Context, erda erdav1beta1.Erda) error {

	listOptions := client.ListOptions{
		Namespace: erda.Namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set{
			erdav1beta1.ErdaOperatorLabel: "true",
		}),
	}

	workloadTypeList := []client.ObjectList{&appsv1.DeploymentList{}, &appsv1.DaemonSetList{}, &appsv1.DaemonSetList{}}
	for _, objList := range workloadTypeList {
		err := r.List(ctx, objList, &listOptions)
		if err != nil {
			errWrap := errors.Wrap(err, fmt.Sprintf("list %s error:", objList.GetObjectKind()))
			r.Log.Error(errWrap, "delete workload error")
			return errWrap
		}
		for _, app := range erda.Spec.Applications {
			err = r.deleteWorkLoad(objList, app.Components)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ErdaReconciler) deleteWorkLoad(objs client.ObjectList, components []erdav1beta1.Component) error {
	var objList []client.Object
	switch v := objs.(type) {
	case *appsv1.DeploymentList:
		for _, item := range v.Items {
			objList = append(objList, &item)
		}
	case *appsv1.DaemonSetList:
		for _, item := range v.Items {
			objList = append(objList, &item)
		}
	case *appsv1.StatefulSetList:
		for _, item := range v.Items {
			objList = append(objList, &item)
		}
	}
	deleteOptions := client.DeleteOptions{}
	deleteOptions.PropagationPolicy = utils.ConvertDeletePropagationToPoint(metav1.DeletePropagationBackground)
	for _, obj := range objList {
		for _, component := range components {
			if component.Name == obj.GetName() {
				err := r.Delete(context.Background(), obj, &deleteOptions)
				if err != nil {
					return client.IgnoreNotFound(err)
				}
				objKey := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
				deleteServiceErr := r.DeleteKubernetesService(objKey)
				if deleteServiceErr != nil {
					return deleteServiceErr
				}
				deleteIngressErr := r.DeleteIngress(objKey)
				if deleteIngressErr != nil {
					return deleteIngressErr

				}
			}
		}
	}
	return nil
}

func (r *ErdaReconciler) SyncWorkLoadStatus(erda *erdav1beta1.Erda) error {

	workloadTypeList := []client.ObjectList{&appsv1.DeploymentList{}, &appsv1.DaemonSetList{}, &appsv1.StatefulSetList{}}

	for _, objList := range workloadTypeList {
		err := r.List(context.Background(), objList,
			client.InNamespace(erda.Namespace),
			client.MatchingLabels{erdav1beta1.ErdaOperatorLabel: "true"})
		if err != nil {
			errWrap := errors.Wrap(err, fmt.Sprintf("list %s error:", objList.GetObjectKind()))
			r.Log.Error(errWrap, "sync workload status error")
			return errWrap
		}

		objs := map[string]client.Object{}
		switch v := objList.(type) {
		case *appsv1.DeploymentList:
			for _, item := range v.Items {
				newDeployment := item.DeepCopy()
				objs[item.Name] = newDeployment
			}
		case *appsv1.DaemonSetList:
			for _, item := range v.Items {
				newDaemonset := item.DeepCopy()
				objs[item.Name] = newDaemonset
			}
		case *appsv1.StatefulSetList:
			for _, item := range v.Items {
				newStatefulset := item.DeepCopy()
				objs[item.Name] = newStatefulset
			}
		}

		if len(objs) > 0 {
			for index, app := range erda.Spec.Applications {
				allComponentsReady := true
				appStatus := erda.Status.Applications[index]
				for componentIndex, component := range app.Components {
					if obj, ok := objs[component.Name]; ok {
						status := r.getWorkLoadStatus(obj)
						appStatus.Components[componentIndex].Status = status
						if status != erdav1beta1.StatusReady {
							allComponentsReady = false
						}
					}
				}
				erda.Status.Applications[index] = appStatus
				if allComponentsReady {
					erda.Status.Applications[index].Status = erdav1beta1.StatusReady
				} else {
					erda.Status.Applications[index].Status = erdav1beta1.StatusDeploying
				}
			}
		}
	}

	err := r.Status().Update(context.Background(), erda)
	if err != nil {
		return err
	}
	return nil
}

func (r *ErdaReconciler) getWorkLoadStatus(object client.Object) erdav1beta1.StatusType {
	switch v := object.(type) {
	case *appsv1.DaemonSet:
		if v.Status.DesiredNumberScheduled == v.Status.NumberAvailable && v.Status.NumberUnavailable == 0 {
			return erdav1beta1.StatusReady
		}
	case *appsv1.Deployment:
		if v.Status.AvailableReplicas == v.Status.Replicas && v.Status.UnavailableReplicas == 0 {
			return erdav1beta1.StatusReady
		}
	case *appsv1.StatefulSet:
		if v.Status.ReadyReplicas == v.Status.Replicas {
			return erdav1beta1.StatusReady
		}
	}
	return erdav1beta1.StatusDeploying
}

func (r *ErdaReconciler) VerifiedComponentStatus(erda erdav1beta1.Erda) bool {
	for _, appStatus := range erda.Status.Applications {
		if appStatus.Status != erdav1beta1.StatusReady {
			return false
		}
	}
	return true
}

func (r *ErdaReconciler) InitStatus(erda *erdav1beta1.Erda) {
	if erda.Status == nil {
		erda.Status = &erdav1beta1.ErdaStatus{
			Phase:        erdav1beta1.PhaseDeploying,
			Applications: make([]erdav1beta1.ApplicationStatus, 0, 0),
		}
		appStatus := []erdav1beta1.ApplicationStatus{}
		for index, app := range erda.Spec.Applications {
			appStatus = append(appStatus, erdav1beta1.ApplicationStatus{
				Name:       app.Name,
				Status:     erdav1beta1.StatusUnKnown,
				Components: make([]erdav1beta1.ComponentStatus, 0, 0),
			})
			for _, component := range app.Components {
				appStatus[index].Components = append(appStatus[index].Components, erdav1beta1.ComponentStatus{
					Name:   component.Name,
					Status: erdav1beta1.StatusUnKnown,
				})
			}
		}
		erda.Status.Applications = appStatus
	}
}
