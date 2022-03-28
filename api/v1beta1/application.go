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

type WorkspaceStr string

const (
	WS_DEV     WorkspaceStr = "development"
	WS_TEST                 = "test"
	WS_STAGING              = "staging"
	WS_PROD                 = "production"
)

type Application struct {
	Metadata   `yaml:",inline" json:",inline"`
	Envs       []corev1.EnvVar        `yaml:"envs,omitempty" json:"envs,omitempty"`
	EnvFrom    []corev1.EnvFromSource `yaml:"envFrom,omitempty" json:"envFrom,omitempty"`
	Components []Component            `yaml:"components,omitempty" json:"components,omitempty"`
	// TODO: Finish the addons design
	// Addons map[string]Addon `yaml:"addons,omitempty" json:"addons,omitempty"`
}
