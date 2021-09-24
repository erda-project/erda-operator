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

type AddonSpecification string

const (
	BasicFormat    AddonSpecification = "basic"
	ProFormat      AddonSpecification = "professional"
	UltimateFormat AddonSpecification = "ultimate"
)

type AddonType string

const (
	AddonMysql         AddonType = "mysql"
	AddonElasticSearch AddonType = "elasticsearch"
	AddonRedis         AddonType = "redis"
	AddonZooKeeper     AddonType = "zookeeper"
	AddonRocketMQ      AddonType = "rocketMQ"
)

type Addon struct {
	Metadata Metadata  `json:"metadata,omitempty"`
	Spec     AddonSpec `yaml:"spec" json:"spec"`
}

//TODO: finish the addon design

type AddonSpec struct {
	Type           AddonType           `yaml:"type" json:"type"`
	Version        string              `yaml:"version" json:"version"`
	Specification  AddonSpecification  `yaml:"specification" json:"specification"`
	Resources      corev1.ResourceList `yaml:"resources" json:"resources"`
	Image          string              `yaml:"image" json:"image"`
	CustomResource bool                `yaml:"custom_resource" json:"custom_resource"`
	Params         map[string]string   `yaml:"params" json:"params"`
}
