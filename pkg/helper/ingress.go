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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/utils"
)

func ComposeIngressV1(component *erdav1beta1.Component, references []metav1.OwnerReference) *networkingv1.Ingress {
	ingress := &networkingv1.Ingress{
		ObjectMeta: utils.ComposeObjectMetadataFromComponent(component, references),
		Spec: networkingv1.IngressSpec{
			Rules: composeRulesV1(component),
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: composeDomains(component),
				},
			},
		},
	}

	if len(component.Annotations) == 0 {
		return ingress
	}

	// ingress snippet with annotation
	ingAnnotations := component.Annotations[erdav1beta1.AnnotationIngressAnnotation]
	if ingAnnotations == "" {
		return ingress
	}

	fmtIngAnnotations := make(map[string]string)
	if err := yaml.Unmarshal([]byte(ingAnnotations), &fmtIngAnnotations); err != nil {
		// TODO: error tips
		return ingress
	}
	ingress.Annotations = fmtIngAnnotations

	return ingress
}

func ComposeIngressV1SpecFromK8sIngress(ingress *networkingv1.Ingress) networkingv1.IngressSpec {
	ingressSpec := networkingv1.IngressSpec{
		TLS:   ingress.Spec.TLS,
		Rules: ingress.Spec.Rules,
	}
	return ingressSpec
}

func composeDomains(component *erdav1beta1.Component) []string {
	domains := make([]string, 0, len(component.Network.ServiceDiscovery))
	for _, sd := range component.Network.ServiceDiscovery {
		if sd.Domain != "" {
			domains = append(domains, sd.Domain)
		}
	}
	return domains
}

func composeRulesV1(component *erdav1beta1.Component) []networkingv1.IngressRule {
	ingressRules := make([]networkingv1.IngressRule, 0, len(component.Network.ServiceDiscovery))
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
										Number: sd.Port,
									},
								},
							},
							Path:     sd.Path,
							PathType: func(pathType networkingv1.PathType) *networkingv1.PathType { return &pathType }(networkingv1.PathTypeImplementationSpecific),
						},
					},
				},
			},
		}
		ingressRules = append(ingressRules, ingressRule)
	}

	return ingressRules
}
