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

package utils

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"strconv"
)

// ConvertInt32ToPointInt32 convert int32 to int32 pointer
func ConvertInt32ToPointInt32(input int32) *int32 {
	return &input
}

// ConvertDeletePropagationToPoint convert DeletionPropagation in k8s.io/apimachinery/pkg/apis/meta/v1
// to DeletionPropagation pointer
func ConvertDeletePropagationToPoint(propagation metav1.DeletionPropagation) *metav1.DeletionPropagation {
	return &propagation
}

func ComposeObjectMetadataFromComponent(component *erdav1beta1.Component, references []metav1.OwnerReference) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      component.Name,
		Namespace: component.Namespace,
		Labels: AppendLabels(component.Labels, map[string]string{
			erdav1beta1.ErdaOperatorLabel:  "true",
			erdav1beta1.ErdaComponentLabel: component.Name,
		}),
		OwnerReferences: references,
	}
}
func ComposeObjectMetadataFromJob(job *erdav1beta1.Job, references []metav1.OwnerReference) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:            job.Name,
		Namespace:       job.Namespace,
		Labels:          job.Labels,
		Annotations:     job.Annotations,
		OwnerReferences: references,
	}
}

func ReplaceEnvironments(envs []corev1.EnvVar) []corev1.EnvVar {
	replaceEnvs := map[string]string{}
	for _, env := range envs {
		if strings.HasPrefix(env.Name, "_") &&
			strings.HasSuffix(env.Name, "_") {
			env.Name = strings.TrimSuffix(strings.TrimPrefix(env.Name, "_"), "_")
			replaceEnvs[env.Name] = env.Value
		}
	}
	for index, env := range envs {
		if strings.HasSuffix(env.Name, "_") &&
			strings.HasPrefix(env.Name, "_") {
			continue
		}
		if v, ok := replaceEnvs[env.Name]; ok {
			envs[index].Value = v
		}
	}
	return envs
}

func AppendLabels(originLabels, appendLabels map[string]string) map[string]string {
	if originLabels == nil {
		originLabels = map[string]string{}
	}
	for k, v := range appendLabels {
		originLabels[k] = v
	}
	return originLabels
}

func MergeEnvs(originEnvs []corev1.EnvVar, destEnvs []corev1.EnvVar) []corev1.EnvVar {
	mergeEnvs := make([]corev1.EnvVar, 0)
	destVisited := make([]int, len(destEnvs), cap(destEnvs))

	for _, originEnv := range originEnvs {
		var isExist bool
		for i := 0; i < len(destEnvs); i++ {
			if destEnvs[i].Name == originEnv.Name {
				destVisited[i] = 1
				mergeEnvs = append(mergeEnvs, destEnvs[i])
				isExist = true
				break
			}
		}
		if !isExist {
			mergeEnvs = append(mergeEnvs, originEnv)
		}
	}

	for index, destEnv := range destEnvs {
		if destVisited[index] == 0 {
			mergeEnvs = append(mergeEnvs, destEnv)
		}
	}
	return mergeEnvs
}

func ComposeResourceToEnvs(component erdav1beta1.Component) []corev1.EnvVar {
	maxCpu := MaxFloat64(component.Resources.Requests.Cpu().AsApproximateFloat64(),
		component.Resources.Limits.Cpu().AsApproximateFloat64())
	maxMem := MaxInt64(component.Resources.Requests.Memory().Value()/1024/1024,
		component.Resources.Limits.Memory().Value()/1024/1024)
	return []corev1.EnvVar{
		{
			Name:  "DICE_CPU_REQUEST",
			Value: fmt.Sprintf("%f", component.Resources.Requests.Cpu().AsApproximateFloat64()),
		},
		{
			Name:  "DICE_MEM_REQUEST",
			Value: fmt.Sprintf("%d", component.Resources.Requests.Memory().Value()/1024/1024),
		},
		{
			Name:  "DICE_CPU_ORIGIN",
			Value: fmt.Sprintf("%f", maxCpu),
		},
		{
			Name:  "DICE_MEM_ORIGIN",
			Value: fmt.Sprintf("%d", maxMem),
		},
		{
			Name:  "DICE_CPU_LIMIT",
			Value: fmt.Sprintf("%f", maxCpu),
		},
		{
			Name:  "DICE_MEM_LIMIT",
			Value: fmt.Sprintf("%d", maxMem),
		},
	}
}

func ComposeDependEnvs(erda erdav1beta1.Erda) []corev1.EnvVar {
	envs := make([]corev1.EnvVar, 0)
	for _, app := range erda.Spec.Applications {
		for _, component := range app.Components {
			if component.WorkLoad == erdav1beta1.PerNode ||
				component.Network.Type == erdav1beta1.NetworkKindHost {
				continue
			}
			if len(component.Network.ServiceDiscovery) > 0 {
				sd := component.Network.ServiceDiscovery[0]
				convertedCompName := strings.ToUpper(strings.ReplaceAll(component.Name, "-", "_"))
				envs = append(envs, corev1.EnvVar{
					Name: fmt.Sprintf("%s_ADDR", convertedCompName),
					// TODO: Kubernetes service name, svc.cluster.local is default
					Value: fmt.Sprintf("%s.%s.svc.cluster.local:%d", component.Name, erda.Namespace, sd.Port),
				})
				if sd.Domain == "" {
					continue
				}
				envs = append(envs, []corev1.EnvVar{
					{
						Name:  fmt.Sprintf("%s_PUBLIC_URL", convertedCompName),
						Value: fmt.Sprintf("%s://%s", ParseProtocol(app.Annotations[erdav1beta1.AnnotationSSLEnabled]), sd.Domain),
					},
					{
						Name:  fmt.Sprintf("%s_PUBLIC_ADDR", convertedCompName),
						Value: sd.Domain,
					},
				}...)
			}
		}
	}
	return envs
}

func ComposeSelfADDREnv(component erdav1beta1.Component, protocol string) []corev1.EnvVar {
	envs := make([]corev1.EnvVar, 0)
	envs = append(envs, corev1.EnvVar{
		Name: "SELF_ADDR",
		// TODO: kubernetes services
		Value: fmt.Sprintf("%s.%s.svc.cluster.local:%d", component.Name, component.Namespace, component.Network.ServiceDiscovery[0].Port),
	})
	if component.Network.ServiceDiscovery[0].Domain != "" {
		envs = append(envs, []corev1.EnvVar{
			{
				Name:  "SELF_PUBLIC_URL",
				Value: fmt.Sprintf("%s://%s", protocol, component.Network.ServiceDiscovery[0].Domain),
			},
			{
				Name:  "SELF_PUBLIC_ADDR",
				Value: component.Network.ServiceDiscovery[0].Domain,
			},
		}...)
	}
	return envs
}

func ReplaceDependsEnv(dependEnvs []corev1.EnvVar, envs []corev1.EnvVar) []corev1.EnvVar {
	for _, denv := range dependEnvs {
		isExisted := false
		for index, env := range envs {
			if env.Name == denv.Name {
				env.Value = denv.Value
				isExisted = true
				envs[index] = env
				break
			}
		}
		if !isExisted {
			envs = append(envs, denv)
		}
	}
	return envs
}

func ParseProtocol(source string) string {
	ssl, _ := strconv.ParseBool(source)
	if ssl {
		return "https"
	}
	return "http"
}
