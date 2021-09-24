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

package pkg

import (
	"fmt"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func ComposeVolumesFromConfigurations(component *erdav1beta1.Component) []corev1.Volume {
	volumes := []corev1.Volume{}
	for _, config := range component.Configurations {
		volume := corev1.Volume{
			Name: config.Name,
		}
		switch config.Type {
		case erdav1beta1.ConfigurationSecret:
			volume.VolumeSource = corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  config.Name,
					DefaultMode: func(defaultMode int32) *int32 { return &defaultMode }(420),
				},
			}
		case erdav1beta1.ConfigurationConfigMap:
			volume.VolumeSource = corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: config.Name,
					},
				},
			}
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

func ComposeVolumeFromComponentStorage(component *erdav1beta1.Component) []corev1.Volume {
	volumes := []corev1.Volume{}
	for index, v := range component.Storage.Volumes {
		volume := corev1.Volume{
			Name: fmt.Sprintf("volume-%s-%d", component.Name, index),
		}
		if v.StorageClass != "" {
			volume.VolumeSource = corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("pvc-%s-%d", component.Name, index),
					ReadOnly:  v.ReadOnly,
				},
			}
		} else {
			volume.VolumeSource = corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: v.SourcePath,
					Type: func(pathType corev1.HostPathType) *corev1.HostPathType { return &pathType }(corev1.HostPathDirectoryOrCreate),
				},
			}
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

func ComposeVolumes(component *erdav1beta1.Component) []corev1.Volume {
	volumes := ComposeVolumesFromConfigurations(component)
	volumes = append(volumes, ComposeVolumeFromComponentStorage(component)...)

	// For same as the deployment which gets from the Kubernetes
	if len(volumes) == 0 {
		return nil
	}
	return volumes
}

func ComposeComponentStoragesVolumeMount(component *erdav1beta1.Component) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	for index, v := range component.Storage.Volumes {
		volumeMount := corev1.VolumeMount{
			Name:      fmt.Sprintf("volume-%s-%d", component.Name, index),
			ReadOnly:  v.ReadOnly,
			MountPath: v.TargetPath,
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}
	return volumeMounts
}

func ComposeConfigurationsVolumeMount(component *erdav1beta1.Component) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	for _, config := range component.Configurations {
		volumeMount := corev1.VolumeMount{
			Name:      config.Name,
			ReadOnly:  true,
			MountPath: config.TargetPath,
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}

	return volumeMounts
}

func ComposeVolumeMounts(component *erdav1beta1.Component) []corev1.VolumeMount {
	volumeMounts := ComposeConfigurationsVolumeMount(component)
	volumeMounts = append(volumeMounts, ComposeComponentStoragesVolumeMount(component)...)
	// For same as the deployment which gets from the Kubernetes
	if len(volumeMounts) == 0 {
		return nil
	}
	return volumeMounts
}
