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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"fmt"
)

var (
	defaultBackOffLimit int32 = 6
)

func ComposeKubernetesJob(erdaName string, job *erdav1beta1.Job, references []metav1.OwnerReference) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-%s-%s", erdaName, strings.ToLower(string(job.Type)), job.Name),
			Namespace:       job.Namespace,
			Labels:          composeJobPodLabels(job),
			Annotations:     job.Annotations,
			OwnerReferences: references,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: func(in *int32) *int32 {
				if in == nil {
					return &defaultBackOffLimit
				}
				return in
			}(job.Retries),
			TTLSecondsAfterFinished: func(in int32) *int32 { return &in }(JobTTLInterval),
			Template:                ComposePodTemplateSpecByJob(job),
		},
	}
}

func composeJobPodLabels(job *erdav1beta1.Job) map[string]string {
	if job.Labels == nil {
		job.Labels = make(map[string]string)
	}
	job.Labels[erdav1beta1.ErdaOperatorLabel] = "true"
	job.Labels[erdav1beta1.ErdaJobNameLabel] = job.Name
	job.Labels[erdav1beta1.ErdaJobTypeLabel] = strings.ToLower(string(job.Type))
	return job.Labels
}

func IsJobFinished(job batchv1.Job) (bool, batchv1.JobCondition) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) &&
			c.Status == corev1.ConditionTrue {
			return true, c
		}
	}

	return false, batchv1.JobCondition{}
}
