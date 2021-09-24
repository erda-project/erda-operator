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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusType string

const (
	StatusReady     StatusType = "Ready"
	StatusDeploying StatusType = "Deploying"
	StatusUnReady   StatusType = "Unready"
	StatusRunning   StatusType = "Running"
	StatusFailed    StatusType = "Failed"
	StatusCompleted StatusType = "Completed"
	StatusUnKnown   StatusType = "UnKnown"
)

type PhaseType string

const (
	PhaseReady     PhaseType = "Ready"
	PhaseDeploying PhaseType = "Deploying"
)

const (
	ErdaPrefix  = "erda"
	PreJobType  = "PreJob"
	PostJobType = "PostJob"
)

const (
	ErdaJobTypeLabel  = "erda.io/job-type"
	ErdaOperatorLabel = "erda.io/erda-operator"
	ErdaOperatorApp   = "erda.io/erda-operator-app"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ErdaList contains a list of Erda
type ErdaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Erda `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="the erda status phase"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Erda is the Schema for the erdas API
type Erda struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *ErdaSpec   `json:"spec,omitempty"`
	Status            *ErdaStatus `json:"status,omitempty"`
}

// ErdaSpec defines the desired state of Erda
type ErdaSpec struct {
	// The Env List of Erda will be injected into all subOjbject
	// and the env will be overwritten by subObject env when the key of env is the same
	Applications []Application `yaml:"applications" json:"applications"`

	// TODO: There are two questions about pre-job tasks and post-job tasks:
	// TODO: 1. How to deal with the job task when it is failed? delete and recreate it?
	// TODO: 2. How to deal with the job task when it is completed and the user updates the DICE YAML
	//PreJobs      []Job                  `yaml:"preJobs,omitempty" json:"preJobs,omitempty"`
	//PostJobs     []Job                  `yaml:"postJobs,omitempty" json:"postJobs,omitempty"`

	// TODO: Finish the addons design
	// Addons       map[string]Addon       `yaml:"envs,omitempty" json:"addons"`
}

// ErdaStatus defines the observed state of Erda
type ErdaStatus struct {
	Phase        PhaseType           `yaml:"phase,omitempty" json:"phase,omitempty"`
	Applications []ApplicationStatus `yaml:"applications,omitempty"json:"applications,omitempty"`
	//PreJobStatus      map[string]Status        `json:"preJobStatus,omitempty"`
	//PostJobStatus     map[string]Status        `json:"postJobStatus,omitempty"`
}

type ApplicationStatus struct {
	Name       string            `json:"name"`
	Status     StatusType        `json:"status"`
	Components []ComponentStatus `json:"components"`
}

type ComponentStatus struct {
	Name   string     `json:"name"`
	Status StatusType `json:"status"`
}

func (e *Erda) ComposeOwnerReferences() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion: e.APIVersion,
			Kind:       e.Kind,
			Name:       e.Name,
			UID:        e.UID,
		},
	}
}

func init() {
	SchemeBuilder.Register(&Erda{}, &ErdaList{})
}
