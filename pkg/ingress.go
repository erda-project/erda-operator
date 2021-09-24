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

	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/utils"
)

func ComposeIngressV1(
	component *erdav1beta1.Component, references []metav1.OwnerReference) *networkingv1.Ingress {
	ingress := &networkingv1.Ingress{
		Spec: networkingv1.IngressSpec{
			Rules: composeRulesV1(component),
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: composeDomains(component),
				},
			},
		},
	}

	metadata := utils.ComposeObjectMetadataFromComponent(component, references)
	ingress.ObjectMeta = metadata

	return ingress
}

func ComposeIngressV1SpecFromK8sIngress(ingress *networkingv1.Ingress) networkingv1.IngressSpec {
	ingressSpec := networkingv1.IngressSpec{
		TLS:   ingress.Spec.TLS,
		Rules: ingress.Spec.Rules,
	}
	return ingressSpec
}

func ComposeIngressV1Beta1(component *erdav1beta1.Component, references []metav1.OwnerReference) *networkingv1beta1.Ingress {
	ingress := &networkingv1beta1.Ingress{
		Spec: networkingv1beta1.IngressSpec{
			TLS: []networkingv1beta1.IngressTLS{
				{
					Hosts: composeDomains(component),
				},
			},
			Rules: composeRulesV1Beta1(component),
		},
	}
	ingress.ObjectMeta = utils.ComposeObjectMetadataFromComponent(component, references)
	return ingress
}

func ComposeIngressV1Beta1SpecFromK8sIngress(ingress *networkingv1beta1.Ingress) networkingv1beta1.IngressSpec {
	ingressSpec := networkingv1beta1.IngressSpec{
		TLS:   ingress.Spec.TLS,
		Rules: ingress.Spec.Rules,
	}
	return ingressSpec
}

func composeDomains(component *erdav1beta1.Component) []string {
	domains := []string{}
	for _, sd := range component.Network.ServiceDiscovery {
		if sd.Domain != "" {
			domains = append(domains, sd.Domain)
		}
	}
	return domains
}

func composeRulesV1(component *erdav1beta1.Component) []networkingv1.IngressRule {
	ingressRules := []networkingv1.IngressRule{}

	for _, sd := range component.Network.ServiceDiscovery {
		ingressRule := networkingv1.IngressRule{
			Host: sd.Domain,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: component.Name,
									Port: networkingv1.ServiceBackendPort{
										Name:   fmt.Sprintf("%s", component.Name),
										Number: sd.Port,
									},
								},
							},
							Path: sd.Path,
						},
					},
				},
			},
		}
		ingressRules = append(ingressRules, ingressRule)
	}

	return ingressRules
}

func composeRulesV1Beta1(component *erdav1beta1.Component) []networkingv1beta1.IngressRule {
	ingressRules := []networkingv1beta1.IngressRule{}
	for _, sd := range component.Network.ServiceDiscovery {
		if sd.Domain != "" {
			ingressRule := networkingv1beta1.IngressRule{
				Host: sd.Domain,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{
							{
								Backend: networkingv1beta1.IngressBackend{
									ServiceName: component.Name,
									ServicePort: intstr.FromInt(int(sd.Port)),
								},
								Path:     sd.Path,
								PathType: func(pathType networkingv1beta1.PathType) *networkingv1beta1.PathType { return &pathType }(networkingv1beta1.PathTypeImplementationSpecific),
							},
						},
					},
				},
			}
			ingressRules = append(ingressRules, ingressRule)
		}
	}

	return ingressRules
}
