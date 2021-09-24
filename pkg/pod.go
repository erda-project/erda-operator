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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/utils"
)

const (
	ServiceAccountName = "erda-operator"
)

func ComposePodTemplateSpecByComponent(component *erdav1beta1.Component) corev1.PodTemplateSpec {
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: utils.AppendLabels(component.Labels, map[string]string{
				erdav1beta1.ErdaOperatorLabel: "true",
				erdav1beta1.ErdaOperatorApp:   component.Name,
			}),
			Annotations: component.Annotations,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			HostNetwork:   component.Network.Type == erdav1beta1.NetworkKindHost,
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
		},
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
				},
			},
			ServiceAccountName: spec.Spec.ServiceAccountName,
			Volumes:            spec.Spec.Volumes,
			Affinity:           spec.Spec.Affinity,
			HostAliases:        spec.Spec.HostAliases,
			Tolerations:        spec.Spec.Tolerations,
			ImagePullSecrets:   spec.Spec.ImagePullSecrets,
			HostNetwork:        spec.Spec.HostNetwork,
		},
	}
	return podTemplateSpec
}

func ComposePodTemplateSpecByJob(job *erdav1beta1.Job) corev1.PodTemplateSpec {
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      job.Labels,
			Annotations: job.Annotations,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            job.Name,
					Resources:       job.Spec.Resources,
					Image:           job.Spec.Image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             job.Spec.Envs,
					Command:         job.Spec.Command,
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
