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
	"fmt"
	"strings"

	"github.com/gogo/protobuf/sortkeys"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/utils"
)

const (
	HTTPProtocolType  = "HTTP"
	HTTPSProtocolType = "HTTPS"
	GRPCProtocolType  = "GRPC"
	TCPProtocolType   = "TCP"
	UDPProtocolType   = "UDP"
)

func ComposeKubernetesService(
	component *erdav1beta1.Component,
	references []metav1.OwnerReference) *corev1.Service {

	k8sService := &corev1.Service{
		Spec: corev1.ServiceSpec{
			SessionAffinity: corev1.ServiceAffinityNone,
			Selector: utils.AppendLabels(component.Labels, map[string]string{
				erdav1beta1.ErdaComponentLabel: component.Name,
			}),
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	k8sService.ObjectMeta = utils.ComposeObjectMetadataFromComponent(component, references)

	servicePorts := []corev1.ServicePort{}

	sdMap := map[int32]erdav1beta1.ServiceDiscovery{}
	sdKeys := []int32{}
	for _, sd := range component.Network.ServiceDiscovery {
		if _, ok := sdMap[sd.Port]; !ok {
			sdMap[sd.Port] = sd
			sdKeys = append(sdKeys, sd.Port)
		}
	}
	sortkeys.Int32s(sdKeys)

	for _, key := range sdKeys {
		servicePort := corev1.ServicePort{}
		servicePort.Port = sdMap[key].Port
		servicePort.Protocol = GetKubernetesProtocol(sdMap[key].Protocol)
		servicePort.Name = fmt.Sprintf("%s-%d",
			strings.ToLower(string(servicePort.Protocol)), servicePort.Port)
		servicePort.TargetPort = intstr.FromInt(int(sdMap[key].Port))
		servicePorts = append(servicePorts, servicePort)
	}
	k8sService.Spec.Ports = servicePorts
	return k8sService
}

func GetKubernetesProtocol(protocol string) corev1.Protocol {
	switch strings.ToUpper(protocol) {
	case UDPProtocolType:
		return corev1.ProtocolUDP
	case HTTPProtocolType, HTTPSProtocolType, TCPProtocolType, GRPCProtocolType:
		return corev1.ProtocolTCP
	default:
		return corev1.ProtocolTCP
	}
}

func ComposeKubernetesServiceSpecFromK8sService(service *corev1.Service) corev1.ServiceSpec {
	k8sServiceSpec := corev1.ServiceSpec{
		SessionAffinity: corev1.ServiceAffinityNone,
		Selector:        service.Spec.Selector,
		Type:            service.Spec.Type,
	}
	for _, port := range service.Spec.Ports {
		k8sServiceSpec.Ports = append(k8sServiceSpec.Ports, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.Port,
			TargetPort: port.TargetPort,
		})
	}
	return k8sServiceSpec
}
