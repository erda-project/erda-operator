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

package v1beta1

import corev1 "k8s.io/api/core/v1"

type Job struct {
	Metadata `yaml:",inline" json:",inline"`
	JobSpec  `yaml:",inline" json:",inline"`
}

type JobSpec struct {
	//+kubebuilder:validation:Enum={PreJob}
	Type      JobType                     `yaml:"type" json:"type"`
	Retries   *int32                      `yaml:"retries,omitempty" json:"retries,omitempty"`
	ImageInfo ImageInfo                   `yaml:"imageInfo" json:"imageInfo"`
	Command   []string                    `yaml:"command,omitempty" json:"command,omitempty"`
	Envs      []corev1.EnvVar             `yaml:"envs,omitempty" json:"envs,omitempty"`
	Resources corev1.ResourceRequirements `yaml:"resources,omitempty" json:"resources,omitempty"`
	Affinity  []Affinity                  `yaml:"affinity,omitempty" json:"affinity,omitempty"`
	Storage   Storage                     `yaml:"storage,omitempty" json:"storage,omitempty"`
	Hosts     []string                    `yaml:"hosts,omitempty" json:"hosts,omitempty"`
}
