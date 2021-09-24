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
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/utils"
)

const (
	JobTTLInterval int32 = 600
)

func ComposeDaemonSet(component *erdav1beta1.Component, references []metav1.OwnerReference) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: utils.ComposeObjectMetadataFromComponent(component, references),
		Spec:       composeDaemonSetSpecFromErdaComponent(component),
		Status:     appsv1.DaemonSetStatus{},
	}
}

func composeDaemonSetSpecFromErdaComponent(component *erdav1beta1.Component) appsv1.DaemonSetSpec {
	return appsv1.DaemonSetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: utils.AppendLabels(component.Labels, map[string]string{
				erdav1beta1.ErdaOperatorLabel: "true",
				erdav1beta1.ErdaOperatorApp:   component.Name,
			}),
		},
		Template:             ComposePodTemplateSpecByComponent(component),
		RevisionHistoryLimit: utils.ConvertInt32ToPointInt32(3),
	}
}

func ComposeDaemonSetSpecFromK8sStatefulSet(daemonSet *appsv1.DaemonSet) appsv1.DaemonSetSpec {
	return appsv1.DaemonSetSpec{
		Selector:             daemonSet.Spec.Selector,
		Template:             ComposePodTemplateSpecFromPodTemplate(daemonSet.Spec.Template),
		RevisionHistoryLimit: daemonSet.Spec.RevisionHistoryLimit,
	}
}

func ComposeStatefulSet(component *erdav1beta1.Component,
	references []metav1.OwnerReference) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: utils.ComposeObjectMetadataFromComponent(component, references),
		Spec:       composeStatefulSetSpecFromErdaComponent(component),
	}
}

func composeStatefulSetSpecFromErdaComponent(component *erdav1beta1.Component) appsv1.StatefulSetSpec {
	return appsv1.StatefulSetSpec{
		Replicas: component.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: utils.AppendLabels(component.Labels, map[string]string{
				erdav1beta1.ErdaOperatorLabel: "true",
				erdav1beta1.ErdaOperatorApp:   component.Name,
			}),
		},
		Template:             ComposePodTemplateSpecByComponent(component),
		ServiceName:          component.Name,
		RevisionHistoryLimit: utils.ConvertInt32ToPointInt32(3),
	}
}

func ComposeStatefulSetSpecFromK8sStatefulSet(statefulSet *appsv1.StatefulSet) appsv1.StatefulSetSpec {
	return appsv1.StatefulSetSpec{
		Replicas:             statefulSet.Spec.Replicas,
		Selector:             statefulSet.Spec.Selector,
		Template:             ComposePodTemplateSpecFromPodTemplate(statefulSet.Spec.Template),
		ServiceName:          statefulSet.Spec.ServiceName,
		RevisionHistoryLimit: statefulSet.Spec.RevisionHistoryLimit,
	}
}

func ComposeDeployment(component *erdav1beta1.Component,
	references []metav1.OwnerReference) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: utils.ComposeObjectMetadataFromComponent(component, references),
		Spec:       composeDeploymentSpecFromErdaComponent(component),
	}
}

func composeDeploymentSpecFromErdaComponent(component *erdav1beta1.Component) appsv1.DeploymentSpec {
	return appsv1.DeploymentSpec{
		Replicas: component.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: utils.AppendLabels(component.Labels, map[string]string{
				erdav1beta1.ErdaOperatorLabel: "true",
				erdav1beta1.ErdaOperatorApp:   component.Name,
			}),
		},
		Template:             ComposePodTemplateSpecByComponent(component),
		RevisionHistoryLimit: utils.ConvertInt32ToPointInt32(3),
	}
}

func ComposeDeploymentSpecFromK8sDeployment(deployment *appsv1.Deployment) appsv1.DeploymentSpec {
	return appsv1.DeploymentSpec{
		Replicas:             deployment.Spec.Replicas,
		Selector:             deployment.Spec.Selector,
		Template:             ComposePodTemplateSpecFromPodTemplate(deployment.Spec.Template),
		RevisionHistoryLimit: deployment.Spec.RevisionHistoryLimit,
	}
}

func ComposeKubernetesJob(job *erdav1beta1.Job, references []metav1.OwnerReference) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: utils.ComposeObjectMetadataFromJob(job, references),
		Spec: batchv1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: utils.AppendLabels(job.Labels, map[string]string{
					erdav1beta1.ErdaOperatorLabel: "true",
					erdav1beta1.ErdaOperatorApp:   job.Name,
				}),
			},
			TTLSecondsAfterFinished: func(in int32) *int32 { return &in }(JobTTLInterval),
			Template:                ComposePodTemplateSpecByJob(job),
		},
	}
}
