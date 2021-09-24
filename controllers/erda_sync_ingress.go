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

	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg"
	"github.com/erda-project/erda-operator/pkg/utils"
)

func (r *ErdaReconciler) CreateOrUpdateIngress(ctx context.Context,
	component *erdav1beta1.Component, owners []metav1.OwnerReference) error {
	var ingress client.Object
	var newIngress client.Object
	if r.IngressVersion == "v1" {
		newIngress = pkg.ComposeIngressV1(component, owners)
		ingress = &networkingv1.Ingress{}
	} else {
		newIngress = pkg.ComposeIngressV1Beta1(component, owners)
		ingress = &networkingv1beta1.Ingress{}
	}

	err := r.Get(ctx, client.ObjectKey{
		Name:      component.Name,
		Namespace: component.Namespace,
	}, ingress)
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			return err
		}
		return r.Create(ctx, newIngress)
	}
	updateIngress, err := r.DiffResource(ingress, newIngress)
	if err != nil {
		return err
	}
	if updateIngress != nil {
		return r.Update(ctx, updateIngress)
	}
	return nil
}

func (r *ErdaReconciler) DeleteIngress(key types.NamespacedName) error {
	deleteOptions := client.DeleteOptions{}
	deleteOptions.PropagationPolicy = utils.ConvertDeletePropagationToPoint(metav1.DeletePropagationBackground)

	var ingress client.Object
	if r.IngressVersion == "v1" {
		ingress = &networkingv1.Ingress{}
	} else {
		ingress = &networkingv1beta1.Ingress{}
	}

	getIngressErr := r.Get(context.Background(), key, ingress)
	if getIngressErr != nil {
		return client.IgnoreNotFound(getIngressErr)
	}

	return client.IgnoreNotFound(r.Delete(context.Background(), ingress, &deleteOptions))
}
