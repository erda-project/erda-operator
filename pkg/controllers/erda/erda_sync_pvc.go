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

	"k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/erda-project/erda-operator/api/v1beta1"
)

func (r *ErdaReconciler) SyncPersistentVolumeClaim(component v1beta1.Component) error {
	for index, v := range component.Storage.Volumes {
		pvc := corev1.PersistentVolumeClaim{}
		pvcName := fmt.Sprintf("pvc-%s-%d", component.Name, index+1)
		err := r.Get(context.Background(), types.NamespacedName{
			Name:      pvcName,
			Namespace: component.Namespace,
		}, &pvc)
		if client.IgnoreNotFound(err) != nil {
			r.Log.Error(err, fmt.Sprintf("get pvc %s error", err))
			return err
		}
		if errors.IsNotFound(err) && v.StorageClass != "" {
			pvc.Name = pvcName
			pvc.Namespace = component.Namespace
			pvc.Spec.StorageClassName = func(s string) *string { return &s }(v.StorageClass)
			pvc.Spec.Resources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceStorage: *v.Size,
				},
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *v.Size,
				},
			}
			createErr := r.Client.Create(context.Background(), &pvc)
			if createErr != nil {
				r.Log.Error(createErr, fmt.Sprintf("create pvc %s in %s error", pvc.Name, pvc.Namespace))
				return createErr
			}
		}
		if v.Size != nil && !pvc.Spec.Resources.Limits.Storage().Equal(*v.Size) {
			pvc.Spec.Resources = corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceStorage: *v.Size,
				},
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *v.Size,
				},
			}
			updateErr := r.Client.Update(context.Background(), &pvc)
			if updateErr != nil {
				r.Log.Error(updateErr, fmt.Sprintf("update pvc %s in %s error", pvc.Name, pvc.Namespace))
				return updateErr
			}
		}
	}
	return nil
}
