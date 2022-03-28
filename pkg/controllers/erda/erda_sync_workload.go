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

package erda

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-test/deep"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/helper"
	"github.com/erda-project/erda-operator/pkg/utils"
)

func (r *ErdaReconciler) ReconcileWorkload(ctx context.Context,
	component erdav1beta1.Component, references []metav1.OwnerReference) (error, bool) {
	// set component.WorkLoad default value Stateless
	if component.WorkLoad == "" {
		component.WorkLoad = erdav1beta1.Stateless
	}

	workLoadErr, needUpdateStatus := r.CreateOrUpdateWorkLoad(ctx, &component, references)
	if workLoadErr != nil {
		r.Log.Error(workLoadErr, "handle workload error", "type", component.WorkLoad,
			"name", component.Name, "namespace", component.Namespace)
		return workLoadErr, needUpdateStatus
	}

	if len(component.Network.ServiceDiscovery) > 0 {
		k8sServiceErr := r.CreateOrUpdateKubernetesService(ctx, &component, references)
		if k8sServiceErr != nil {
			r.Log.Error(k8sServiceErr, "handle component error", "name", component.Name,
				"namespace", component.Namespace)
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
		newObj = helper.ComposeDaemonSet(component, owners)
	case erdav1beta1.Stateless:
		obj = &appsv1.Deployment{}
		newObj = helper.ComposeDeployment(component, owners)
	case erdav1beta1.Stateful:
		obj = &appsv1.StatefulSet{}
		newObj = helper.ComposeStatefulSet(component, owners)
	default:
		return errors.Errorf("unsupported workload type %s", component.WorkLoad), false
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
		deploySpec := helper.ComposeDeploymentSpecFromK8sDeployment(oldDeployment)

		if equal := deep.Equal(deploySpec, newDeployment.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name: %s diff object is %+v", newDeployment.Name, equal))
			return newDeployment, nil
		}
	}
	if oldDaemonSet, ok := oldObj.(*appsv1.DaemonSet); ok {
		newDaemonSet := newObj.(*appsv1.DaemonSet)
		daemonSetSpec := helper.ComposeDaemonSetSpecFromK8sDaemonSet(oldDaemonSet)
		if equal := deep.Equal(daemonSetSpec, newDaemonSet.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name: %s diff object is %+v", newDaemonSet.Name, equal))
			return newDaemonSet, nil
		}
	}
	if oldStatefulSet, ok := oldObj.(*appsv1.StatefulSet); ok {
		newStatefulSet := newObj.(*appsv1.DaemonSet)
		statefulSetSpec := helper.ComposeStatefulSetSpecFromK8sStatefulSet(oldStatefulSet)
		if equal := deep.Equal(statefulSetSpec, newStatefulSet.Spec); equal != nil {
			r.Log.Info(fmt.Sprintf("name: %s diff object is %+v", newStatefulSet.Name, equal))
			return newStatefulSet, nil
		}
	}
	if oldK8sService, ok := oldObj.(*corev1.Service); ok {
		newK8sService := newObj.(*corev1.Service)
		k8sServiceSpec := helper.ComposeKubernetesServiceSpecFromK8sService(oldK8sService)
		if equal := deep.Equal(k8sServiceSpec, newK8sService.Spec); equal != nil {
			newK8sService.ResourceVersion = oldK8sService.ResourceVersion
			newK8sService.Spec.ClusterIP = oldK8sService.Spec.ClusterIP
			r.Log.Info(fmt.Sprintf("name %s diff object is %+v", newK8sService.Name, equal))
			return newK8sService, nil
		}
	}

	if oldIngress, ok := oldObj.(*networkingv1.Ingress); ok {
		newIngress := newObj.(*networkingv1.Ingress)
		ingressSpec := helper.ComposeIngressV1SpecFromK8sIngress(oldIngress)
		// spec deep equal
		specEqual := deep.Equal(ingressSpec, newIngress.Spec)
		annotationEqual := deep.Equal(oldIngress.Annotations, newIngress.Annotations)
		if specEqual != nil || annotationEqual != nil {
			if specEqual != nil && annotationEqual != nil {
				r.Log.Info(fmt.Sprintf("name %s diff annotation object is %+v, spec object is %+v",
					newIngress.Name, annotationEqual, specEqual))
			} else if specEqual != nil {
				r.Log.Info(fmt.Sprintf("name %s diff spec object is %+v", newIngress.Name, specEqual))
			} else if annotationEqual != nil {
				r.Log.Info(fmt.Sprintf("name %s diff annotation object is %+v", newIngress.Name, annotationEqual))
			}
			return newIngress, nil
		}
	}
	return nil, nil
}

func (r *ErdaReconciler) deleteWorkLoad(obj client.Object) error {
	deleteOptions := client.DeleteOptions{}
	deleteOptions.PropagationPolicy = utils.ConvertDeletePropagationToPoint(metav1.DeletePropagationBackground)

	r.Log.Info("workflow resource need to be deleted",
		"name", obj.GetName(), "namespace", obj.GetNamespace())
	err := r.Delete(context.Background(), obj, &deleteOptions)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	objKey := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}

	r.Log.Info("service resource need to be deleted", "name", obj.GetName(), "namespace", obj.GetNamespace())
	deleteServiceErr := r.DeleteKubernetesService(objKey)
	if deleteServiceErr != nil {
		return deleteServiceErr
	}

	r.Log.Info("ingress resource need to be deleted", "name", obj.GetName(), "namespace", obj.GetNamespace())
	deleteIngressErr := r.DeleteIngress(objKey)
	if deleteIngressErr != nil {
		return deleteIngressErr

	}
	return nil
}

func (r *ErdaReconciler) SyncWorkLoadStatus(ctx context.Context, erda *erdav1beta1.Erda) error {
	workloadTypeList := []client.ObjectList{&appsv1.DeploymentList{}, &appsv1.DaemonSetList{}, &appsv1.StatefulSetList{}}

	isDeploying := false

	// use component name with workflow type to primary key.
	objs := map[string]client.Object{}
	for _, objList := range workloadTypeList {
		err := r.List(context.Background(), objList,
			client.InNamespace(erda.Namespace),
			client.MatchingLabels{erdav1beta1.ErdaOperatorLabel: "true"})
		if err != nil {
			errWrap := errors.Wrap(err, fmt.Sprintf("list %s error:", objList.GetObjectKind()))
			r.Log.Error(errWrap, "sync workload status error")
			return errWrap
		}

		switch v := objList.(type) {
		case *appsv1.DeploymentList:
			for _, item := range v.Items {
				objs[composeObjectName(item.Name, erdav1beta1.Stateless)] = item.DeepCopy()
			}
		case *appsv1.DaemonSetList:
			for _, item := range v.Items {
				objs[composeObjectName(item.Name, erdav1beta1.PerNode)] = item.DeepCopy()
			}
		case *appsv1.StatefulSetList:
			for _, item := range v.Items {
				objs[composeObjectName(item.Name, erdav1beta1.Stateful)] = item.DeepCopy()
			}
		}
	}

	appsStatus := make([]erdav1beta1.ApplicationStatus, 0, len(erda.Spec.Applications))
	for _, app := range erda.Spec.Applications {
		allComponentsReady := true
		compStatus := make([]erdav1beta1.ComponentStatus, 0, len(app.Components))
		for _, component := range app.Components {
			searchName := composeObjectName(component.Name, component.WorkLoad)
			obj, ok := objs[searchName]
			if !ok {
				compStatus = append(compStatus, erdav1beta1.ComponentStatus{
					Name:   component.Name,
					Status: erdav1beta1.StatusUnKnown,
				})
			}

			delete(objs, searchName)

			compStatus = append(compStatus, erdav1beta1.ComponentStatus{
				Name: component.Name,
				Status: func() erdav1beta1.StatusType {
					status := r.getWorkLoadStatus(obj)
					if status != erdav1beta1.StatusReady {
						allComponentsReady = false
						isDeploying = true
					}
					return status
				}(),
			})
		}

		appsStatus = append(appsStatus, erdav1beta1.ApplicationStatus{
			Name:       app.Name,
			Components: compStatus,
			Status: func(allComponentsReady bool) erdav1beta1.StatusType {
				if allComponentsReady {
					return erdav1beta1.StatusReady
				}
				return erdav1beta1.StatusDeploying
			}(allComponentsReady),
		})
	}

	if erda.Status == nil {
		erda.Status = &erdav1beta1.ErdaStatus{}
	}

	erda.Status.Applications = appsStatus
	// objs is not empty, means some workloads need to gc
	if !isDeploying && len(objs) == 0 {
		erda.Status.Phase = erdav1beta1.PhaseReady
	} else {
		erda.Status.Phase = erdav1beta1.PhaseDeploying
	}

	if err := r.Status().Update(ctx, erda); err != nil {
		r.Log.Error(err, "update status error")
		return err
	}

	for _, obj := range objs {
		if err := r.deleteWorkLoad(obj); err != nil {
			r.Log.Error(err, "delete workload error")
			return err
		}
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

func composeObjectName(componentName string, workflowType erdav1beta1.WorkLoadType) string {
	return fmt.Sprintf("%s-%s", componentName, workflowType)
}
