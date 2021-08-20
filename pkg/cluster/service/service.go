// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func CreateIfNotExists(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*corev1.Service, error) {
	if len(dicesvc.Ports) == 0 {
		return nil, nil
	}
	generatedSvc := buildService(dicesvcname, dicesvc, clus, ownerRefs)
	svc, err := client.CoreV1().Services(clus.Namespace).Get(context.Background(), generatedSvc.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		svc, err = client.CoreV1().Services(clus.Namespace).Create(context.Background(), generatedSvc, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
	}
	return svc, nil
}

func CreateOrUpdate(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*corev1.Service, error) {
	if len(dicesvc.Ports) == 0 {
		return nil, nil
	}
	generatedSvc := buildService(dicesvcname, dicesvc, clus, ownerRefs)
	s, err := client.CoreV1().Services(clus.Namespace).Get(context.Background(), generatedSvc.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return client.CoreV1().Services(clus.Namespace).Create(context.Background(), generatedSvc, metav1.CreateOptions{})
	}
	generatedSvc.Spec.ClusterIP = s.Spec.ClusterIP
	generatedSvc.ObjectMeta.ResourceVersion = s.ObjectMeta.ResourceVersion
	svc, err := client.CoreV1().Services(clus.Namespace).Update(context.Background(), generatedSvc, metav1.UpdateOptions{})
	if errors.IsForbidden(err) || errors.IsInvalid(err) {
		client.CoreV1().Services(clus.Namespace).Delete(context.Background(), generatedSvc.Name, metav1.DeleteOptions{})
		return client.CoreV1().Services(clus.Namespace).Create(context.Background(), generatedSvc, metav1.CreateOptions{})
	}
	return svc, nil
}

func Delete(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster) error {
	if dicesvc != nil && len(dicesvc.Ports) == 0 {
		return nil
	}
	err := client.CoreV1().Services(clus.Namespace).Delete(context.Background(), convertServiceName(dicesvcname), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

func buildService(
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) *corev1.Service {

	ports := []corev1.ServicePort{}
	for _, port := range dicesvc.Ports {
		// TODO: better impl, 目前diceyml中的port字段没有区分 udp, tcp 的配置
		if port.Protocol == "" {
			port.Protocol = "TCP"
			port.L4Protocol = corev1.ProtocolTCP
			if port.Port == 8125 || port.Port == 8094 {
				port.Protocol = "UDP"
				port.L4Protocol = corev1.ProtocolUDP
			}
		}
		ports = append(ports, corev1.ServicePort{
			Name:       fmt.Sprintf("%s-%d", strings.ToLower(port.Protocol), port.Port),
			Port:       int32(port.Port),
			Protocol:   port.L4Protocol,
			TargetPort: intstr.FromInt(port.Port),
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            convertServiceName(dicesvcname),
			Namespace:       clus.Namespace,
			OwnerReferences: ownerRefs,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"dice/component":    dicesvcname,
				"dice/koperator":    "true",
				"dice/cluster-name": clus.Name,
			},
			Ports: ports,
		},
	}
}

func convertServiceName(dicesvcname string) string {
	switch dicesvcname {
	case "officer":
		return "officer"
	default:
		return dicesvcname
	}
}
