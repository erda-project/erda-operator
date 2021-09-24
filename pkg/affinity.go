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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
)

func ComposeAffinityByService(component *erdav1beta1.Component) *corev1.Affinity {

	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/master",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
				{
					Weight: 80,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/lb",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
			},
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      "erda/component",
								Operator: "In",
								Values:   []string{component.Name},
							}},
						},
					},
				},
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: "kubernetes.io/zone",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{{
								Key:      "erda/component",
								Operator: "In",
								Values:   []string{component.Name},
							}},
						},
					},
				},
			},
		},
	}

	affinity.NodeAffinity = composeNodeAffinity(component.Affinity, affinity.NodeAffinity)
	affinity.PodAntiAffinity = composePodAntiAffinity(component.Labels, affinity.PodAntiAffinity)
	return affinity
}

func ComposeAffinityByJob(job *erdav1beta1.Job) *corev1.Affinity {
	affinity := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/master",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
				{
					Weight: 80,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "node-role.kubernetes.io/lb",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
			},
		},
	}

	affinity.NodeAffinity = composeNodeAffinity(job.Spec.Affinity, affinity.NodeAffinity)
	return affinity
}

func composePodAntiAffinity(labels map[string]string, podAntiAffinity *corev1.PodAntiAffinity) *corev1.PodAntiAffinity {
	if podAntiAffinity == nil {
		podAntiAffinity = &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  []corev1.PodAffinityTerm{},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{},
		}
	}
	preferredTerm := podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	_, cpuOK := labels[erdav1beta1.CPUBound]
	_, ioOK := labels[erdav1beta1.IOBound]
	if cpuOK && ioOK {
		preferredTerm = append(preferredTerm,
			corev1.WeightedPodAffinityTerm{
				Weight: 100,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      fmt.Sprintf("%s/%s", erdav1beta1.ErdaPrefix, erdav1beta1.CPUBound),
								Operator: metav1.LabelSelectorOpExists,
							},
							{
								Key:      fmt.Sprintf("%s/%s", erdav1beta1.ErdaPrefix, erdav1beta1.IOBound),
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		)
	} else {
		preferredTerm = append(preferredTerm,
			[]corev1.WeightedPodAffinityTerm{
				{
					Weight: 50,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      fmt.Sprintf("%s/%s", erdav1beta1.ErdaPrefix, erdav1beta1.CPUBound),
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				{
					Weight: 50,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      fmt.Sprintf("%s/%s", erdav1beta1.ErdaPrefix, erdav1beta1.IOBound),
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}...,
		)
	}
	podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferredTerm
	return podAntiAffinity
}

func composeNodeAffinity(affinities []erdav1beta1.Affinity, nodeAffinity *corev1.NodeAffinity) *corev1.NodeAffinity {

	for _, affinity := range affinities {

		nodeSelectorTerm := corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				composeNodeSelectorRequirement(affinity),
			},
		}

		if affinity.Type == erdav1beta1.NodePreferredAffinityType {
			if nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
				nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.PreferredSchedulingTerm{}
			}
			preferred := corev1.PreferredSchedulingTerm{
				Weight:     100,
				Preference: nodeSelectorTerm,
			}
			nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, preferred)
		}

		if affinity.Type == erdav1beta1.NodeRequestedAffinityType {
			if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
				nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{}}
			}
			nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, nodeSelectorTerm)
		}

	}
	return nodeAffinity
}

func composeNodeSelectorRequirement(affinity erdav1beta1.Affinity) corev1.NodeSelectorRequirement {
	nodeSelectorRequirement := corev1.NodeSelectorRequirement{
		Key: affinity.Key,
	}
	if affinity.Exist == false {
		nodeSelectorRequirement.Operator = corev1.NodeSelectorOpDoesNotExist
	} else {
		nodeSelectorRequirement.Operator = corev1.NodeSelectorOpExists
	}
	if affinity.Value != "" {
		nodeSelectorRequirement.Operator = corev1.NodeSelectorOpIn
		nodeSelectorRequirement.Values = []string{affinity.Value}
	}
	return nodeSelectorRequirement
}
