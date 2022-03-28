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

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type NetworkType string

const (
	NetworkKindHost NetworkType = "host"
)

type WorkLoadType string

const (
	Stateful  WorkLoadType = "Stateful"
	Stateless WorkLoadType = "Stateless"
	PerNode   WorkLoadType = "PerNode"
)

type AffinityType string

const (
	NodePreferredAffinityType AffinityType = "NodePreferred"
	NodeRequestedAffinityType AffinityType = "NodeRequested"
)

const (
	CPUBound string = "cpu_bound"
	IOBound  string = "io_bound"
)

type ConfigurationType string

const (
	ConfigurationSecret    = "Secret"
	ConfigurationConfigMap = "ConfigMap"
)

const (
	AnnotationSSLEnabled          = "erda.erda.cloud/ssl-enabled"
	AnnotationIngressAnnotation   = "erda.erda.cloud/ingress-annotation"
	AnnotationComponentSA         = "erda.erda.cloud/component-service-account"
	AnnotationComponentPrivileged = "erda.erda.cloud/component-security-context-privileged"
)

type Component struct {
	Metadata      `yaml:",inline" json:",inline"`
	ComponentSpec `yaml:",inline" json:",inline"`
}

type ComponentSpec struct {
	WorkLoad       WorkLoadType                `yaml:"workload" json:"workload"`
	ImageInfo      ImageInfo                   `yaml:"imageInfo" json:"imageInfo"`
	Replicas       *int32                      `yaml:"replicas" json:"replicas"`
	Resources      corev1.ResourceRequirements `yaml:"resources" json:"resources"`
	Affinity       []Affinity                  `yaml:"affinity,omitempty" json:"affinity,omitempty"`
	Envs           []corev1.EnvVar             `yaml:"envs,omitempty" json:"envs,omitempty"`
	EnvFrom        []corev1.EnvFromSource      `yaml:"envFrom,omitempty" json:"envFrom,omitempty"`
	Command        []string                    `yaml:"command,omitempty" json:"command,omitempty"`
	Storage        Storage                     `yaml:"storage,omitempty" json:"storage,omitempty"`
	Hosts          []string                    `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	Network        *Network                    `yaml:"network,omitempty" json:"network,omitempty"`
	HealthCheck    *HealthCheck                `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
	Configurations []Configuration             `yaml:"configurations,omitempty" json:"configurations,omitempty"`

	// TODO: impl
	DependsOn []string `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
}

type HealthCheck struct {
	Duration  int32      `yaml:"duration,omitempty" json:"duration,omitempty"`
	HTTPCheck *HTTPCheck `yaml:"httpCheck,omitempty" json:"httpCheck,omitempty"`
	ExecCheck *ExecCheck `yaml:"execCheck,omitempty" json:"execCheck,omitempty"`
}

type HTTPCheck struct {
	Port int    `yaml:"port,omitempty" json:"port,omitempty"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

type ExecCheck struct {
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

type Configuration struct {
	Name       string            `yaml:"name" json:"name"`
	Type       ConfigurationType `yaml:"type" json:"type"`
	TargetPath string            `yaml:"targetPath" json:"targetPath"`
	Data       map[string][]byte `yaml:"data,omitempty" json:"data,omitempty"`
	StringData map[string]string `yaml:"stringData,omitempty" json:"stringData,omitempty"`
}

type Affinity struct {
	Key   string       `yaml:"key" json:"key"`
	Value string       `yaml:"value,omitempty" json:"value,omitempty"`
	Exist bool         `yaml:"exist" json:"exist"`
	Type  AffinityType `yaml:"type" json:"type"`
}

type ImageInfo struct {
	Image      string `yaml:"image" json:"image"`
	UserName   string `yaml:"userName,omitempty" json:"userName,omitempty"`
	Password   string `yaml:"password,omitempty" json:"password,omitempty"`
	PullPolicy string `yaml:"pullPolicy,omitempty" json:"pullPolicy,omitempty"`
	PullSecret string `yaml:"pullSecret,omitempty" json:"pullSecret,omitempty"`
}

type Storage struct {
	Volumes []Volume `yaml:"volumes,omitempty" json:"volumes,omitempty"`
}

// TODO: finish the design
type Volume struct {
	Size         *resource.Quantity `yaml:"size,omitempty" json:"size,omitempty"`
	StorageClass string             `yaml:"storageClass,omitempty" json:"storageClass,omitempty"`
	SourcePath   string             `yaml:"sourcePath,omitempty" json:"sourcePath,omitempty"`
	TargetPath   string             `yaml:"targetPath,omitempty" json:"targetPath,omitempty"`
	ReadOnly     bool               `yaml:"readOnly,omitempty" json:"readOnly,omitempty"`
	Snapshot     *VolumeSnapshot    `yaml:"snapshot,omitempty" json:"snapshot,omitempty"`
}

type VolumeSnapshot struct {
	SnapShotClass string `yaml:"snapshotClass,omitempty" json:"snapshotClass,omitempty"`
	MaxHistory    int32  `yaml:"maxHistory,omitempty" json:"maxHistory,omitempty"`
}

type Network struct {
	Type NetworkType `yaml:"type,omitempty" json:"type,omitempty"`
	//use the first port as domain port
	ServiceDiscovery []ServiceDiscovery `yaml:"serviceDiscovery,omitempty" json:"serviceDiscovery,omitempty"`
	Microservices    *Microservices     `yaml:"microservice,omitempty" json:"microservices,omitempty"`
}

type ServiceDiscovery struct {
	Port     int32  `yaml:"port" json:"port"`
	Protocol string `yaml:"protocol" json:"protocol"`
	Domain   string `yaml:"domain,omitempty" json:"domain,omitempty"`
	Path     string `yaml:"path,omitempty" json:"path,omitempty"`
}

type TrafficSecurity struct {
	Mode string `yaml:"mode,omitempty" json:"mode,omitempty"`
}

type Microservices struct {
	MeshEnable      *bool           `yaml:"meshEnable,omitempty" json:"meshEnable,omitempty"`
	TrafficSecurity TrafficSecurity `yaml:"trafficSecurity,omitempty" json:"trafficSecurity,omitempty"`
	Endpoints       []Endpoint      `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`
}

type Endpoint struct {
	Domain      string           `yaml:"domain,omitempty" json:"domain,omitempty"`
	Path        string           `yaml:"path,omitempty" json:"path,omitempty"`
	BackendPath string           `yaml:"backend_path,omitempty" json:"backend_path,omitempty"`
	Policies    EndpointPolicies `yaml:"policies,omitempty" json:"policies,omitempty"`
}

type EndpointPolicies struct {
	Cors      *map[string]apiextensionsv1.JSON `yaml:"cors,omitempty" json:"cors,omitempty"`
	RateLimit *map[string]apiextensionsv1.JSON `yaml:"rateLimit,omitempty" json:"rateLimit,omitempty"`
}
