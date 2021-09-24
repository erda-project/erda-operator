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

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
)

func (r *ErdaReconciler) SyncConfigurations(component *erdav1beta1.Component) {

	for _, config := range component.Configurations {

		var (
			cfg        client.Object
			newCfg     client.Object
			data       map[string][]byte
			stringData map[string]string
		)
		switch config.Type {
		case erdav1beta1.ConfigurationConfigMap:
			cfg = &corev1.ConfigMap{}
		case erdav1beta1.ConfigurationSecret:
			cfg = &corev1.Secret{}
		}

		err := r.Client.Get(context.Background(), types.NamespacedName{
			Namespace: component.Namespace,
			Name:      config.Name,
		}, cfg)

		switch v := cfg.(type) {
		case *corev1.ConfigMap:
			data = v.BinaryData
			stringData = v.Data
			newCfg = ComposeConfigMap(&config, component.Namespace)
		case *corev1.Secret:
			data = v.Data
			stringData = v.StringData
			newCfg = ComposeSecret(&config, component.Namespace)
		}

		if client.IgnoreNotFound(err) != nil {
			r.Log.Error(err, fmt.Sprintf("get secret %s error", cfg.GetName()))
			continue
		}

		if errors.IsNotFound(err) && (data != nil || stringData != nil) {

			createErr := r.Create(context.Background(), newCfg)
			if createErr != nil {
				r.Log.Error(createErr, fmt.Sprintf("create secret %s error", newCfg.GetName()))
				continue
			}
		}

		if !DiffConfiguration(cfg, newCfg) {
			updateErr := r.Update(context.Background(), newCfg)
			if updateErr != nil {
				r.Log.Error(updateErr, fmt.Sprintf("create secret %s error", newCfg.GetName()))
				continue
			}
		}
	}
	return
}

func ComposeConfigMap(config *erdav1beta1.Configuration, namespace string) *corev1.ConfigMap {
	configMap := corev1.ConfigMap{}
	configMap.Name = config.Name
	configMap.Namespace = namespace
	configMap.Data = config.StringData
	configMap.BinaryData = config.Data
	return &configMap
}

func ComposeSecret(config *erdav1beta1.Configuration, namespace string) *corev1.Secret {
	secret := corev1.Secret{}
	secret.Name = config.Name
	secret.Namespace = namespace
	secret.Data = config.Data
	secret.StringData = config.StringData
	return &secret
}

func DiffConfiguration(oldCfg, newCfg client.Object) bool {
	if oldSecret, ok := oldCfg.(*corev1.Secret); ok {
		newSecret := newCfg.(*corev1.Secret)
		if !reflect.DeepEqual(oldSecret.Data, newSecret.Data) {
			return true
		}
		if !reflect.DeepEqual(oldSecret.StringData, newSecret.StringData) {
			return true
		}
	}
	if oldConfigMap, ok := oldCfg.(*corev1.ConfigMap); ok {
		newConfigMap := newCfg.(*corev1.ConfigMap)
		if !reflect.DeepEqual(oldConfigMap.Data, newConfigMap.Data) {
			return true
		}
		if !reflect.DeepEqual(oldConfigMap.BinaryData, newConfigMap.BinaryData) {
			return true
		}
	}
	return false
}
