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

package helper

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	ServiceAccountName = "erda-operator"
)

// ComposePodTemplateSpecByComponent returns a PodTemplateSpec for the given component
func ComposePodTemplateSpecByComponent(component *erdav1beta1.Component) corev1.PodTemplateSpec {
	isHostNetwork := component.Network.Type == erdav1beta1.NetworkKindHost
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: utils.AppendLabels(component.Labels, map[string]string{
				erdav1beta1.ErdaComponentLabel: component.Name,
			}),
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			Containers: []corev1.Container{
				{
					Name:            component.Name,
					Resources:       component.Resources,
					Image:           component.ImageInfo.Image,
					ImagePullPolicy: corev1.PullPolicy(component.ImageInfo.PullPolicy),
					Env:             component.Envs,
					EnvFrom:         component.EnvFrom,
					LivenessProbe:   ComposeLivenessProbe(component),
					ReadinessProbe:  ComposeReadinessProbe(component),
					Command:         ComposeCommand(component.Command),
					VolumeMounts:    ComposeVolumeMounts(component),
					SecurityContext: &corev1.SecurityContext{},
				},
			},
			ImagePullSecrets:   ComposeImagePullSecret(component.ImageInfo.PullSecret),
			ServiceAccountName: ServiceAccountName,
			Volumes:            ComposeVolumes(component),
			Affinity:           ComposeAffinityByService(component),
			HostAliases:        ConvertStringSliceToHostAlias(component.Hosts),
			Tolerations: []corev1.Toleration{
				{
					Key:    "node-role.kubernetes.io/master",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "node-role.kubernetes.io/lb",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			HostNetwork: isHostNetwork,
			DNSPolicy: func(isHostNetwork bool) corev1.DNSPolicy {
				if isHostNetwork {
					return corev1.DNSClusterFirstWithHostNet
				}
				return corev1.DNSClusterFirst
			}(isHostNetwork),
		},
	}

	// snippet with annotation
	// security context snippet
	if component.Annotations[erdav1beta1.AnnotationComponentPrivileged] != "" {
		enabled, _ := strconv.ParseBool(component.Annotations[erdav1beta1.AnnotationComponentPrivileged])
		if enabled {
			podTemplateSpec.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
				Privileged: &enabled,
			}
		}
	}
	// service account snippet
	if component.Annotations[erdav1beta1.AnnotationComponentSA] != "" {
		podTemplateSpec.Spec.ServiceAccountName = component.Annotations[erdav1beta1.AnnotationComponentSA]
	}

	// annotations inject
	if component.Annotations[erdav1beta1.AnnotationComponentAnnotations] != "" {
		annotations := make(map[string]string)
		err := yaml.Unmarshal([]byte(component.Annotations[erdav1beta1.AnnotationComponentAnnotations]),
			&annotations)
		if err != nil {
			// TODO: tips error
			return podTemplateSpec
		}
		podTemplateSpec.Annotations = annotations
	}

	return podTemplateSpec
}

func ComposePodTemplateSpecFromPodTemplate(spec corev1.PodTemplateSpec) corev1.PodTemplateSpec {
	container := spec.Spec.Containers[0]
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      spec.Labels,
			Annotations: spec.Annotations,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: spec.Spec.RestartPolicy,
			Containers: []corev1.Container{
				{
					Name:            container.Name,
					Resources:       container.Resources,
					Image:           container.Image,
					ImagePullPolicy: container.ImagePullPolicy,
					Env:             container.Env,
					EnvFrom:         container.EnvFrom,
					Command:         container.Command,
					LivenessProbe:   container.LivenessProbe,
					ReadinessProbe:  container.ReadinessProbe,
					VolumeMounts:    container.VolumeMounts,
					SecurityContext: container.SecurityContext,
				},
			},
			ServiceAccountName: spec.Spec.ServiceAccountName,
			Volumes:            spec.Spec.Volumes,
			Affinity:           spec.Spec.Affinity,
			HostAliases:        spec.Spec.HostAliases,
			Tolerations:        spec.Spec.Tolerations,
			ImagePullSecrets:   spec.Spec.ImagePullSecrets,
			HostNetwork:        spec.Spec.HostNetwork,
			DNSPolicy:          spec.Spec.DNSPolicy,
		},
	}
	return podTemplateSpec
}

func ComposePodTemplateSpecByJob(job *erdav1beta1.Job) corev1.PodTemplateSpec {
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: job.Annotations,
			Labels:      composeJobPodLabels(job),
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            job.Name,
					Resources:       job.Resources,
					Image:           job.ImageInfo.Image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             job.Envs,
					Command:         job.Command,
				},
			},
			Affinity: ComposeAffinityByJob(job),
			Tolerations: []corev1.Toleration{
				{
					Key:    "node-role.kubernetes.io/master",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "node-role.kubernetes.io/lb",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
	return podTemplateSpec
}

func ComposeCommand(cmds []string) []string {
	var commands []string
	if len(cmds) > 0 {
		commands = append([]string{"/bin/sh", "-c"}, cmds...)
	}
	return commands
}

func ComposeImagePullSecret(imagePullSecret string) []corev1.LocalObjectReference {
	if imagePullSecret != "" {
		return []corev1.LocalObjectReference{
			{
				Name: imagePullSecret,
			},
		}
	}
	return nil
}
