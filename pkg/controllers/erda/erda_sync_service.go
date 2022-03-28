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

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/helper"
	"github.com/erda-project/erda-operator/pkg/utils"
)

func (r *ErdaReconciler) CreateOrUpdateKubernetesService(ctx context.Context,
	component *erdav1beta1.Component, owners []metav1.OwnerReference) error {

	newK8sService := helper.ComposeKubernetesService(component, owners)
	k8sService := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Namespace: component.Namespace, Name: component.Name}, k8sService)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		return r.Create(ctx, newK8sService)
	}
	updateK8sService, err := r.DiffResource(k8sService, newK8sService)
	if err != nil {
		return err
	}
	if updateK8sService != nil {
		return r.Update(ctx, updateK8sService)
	}
	return nil
}

func (r *ErdaReconciler) DeleteKubernetesService(key types.NamespacedName) error {
	deleteOptions := client.DeleteOptions{}
	deleteOptions.PropagationPolicy = utils.ConvertDeletePropagationToPoint(metav1.DeletePropagationBackground)

	k8sService := corev1.Service{}
	getServiceErr := r.Get(context.Background(), key, &k8sService)
	if getServiceErr != nil {
		return client.IgnoreNotFound(getServiceErr)
	}
	delServiceErr := r.Delete(context.Background(), &k8sService, &deleteOptions)
	return client.IgnoreNotFound(delServiceErr)
}
