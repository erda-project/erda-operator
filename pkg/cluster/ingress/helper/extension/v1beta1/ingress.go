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

package v1

import (
	"fmt"
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"

	"github.com/erda-project/dice-operator/pkg/cluster/ingress/helper/types"
	"github.com/erda-project/dice-operator/pkg/cluster/ingress/helper/common"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/dice-operator/pkg/spec"
)

type Ingress struct {
	c v1beta1.ExtensionsV1beta1Interface
}

func NewIngress(c v1beta1.ExtensionsV1beta1Interface) *Ingress {
	return &Ingress{c: c}
}

func (i *Ingress) CreateIfNotExists(svcName string, svc *diceyml.Service, cluster *spec.DiceCluster, ownerRefs []metav1.OwnerReference) error {
	generatedIngress, err := buildIngress(svcName, svc, cluster, ownerRefs)
	if err != nil {
		return err
	}

	if _, err = i.c.Ingresses(cluster.Namespace).Get(context.Background(), generatedIngress.Name,
		metav1.GetOptions{}); err != nil && !errors.IsNotFound(err) {
		return nil
	}

	if _, err = i.c.Ingresses(cluster.Namespace).Create(context.Background(), generatedIngress,
		metav1.CreateOptions{}); err != nil {
		return err
	}

	return err
}

func (i *Ingress) CreateOrUpdate(svcName string, svc *diceyml.Service, cluster *spec.DiceCluster, ownerRefs []metav1.OwnerReference) error {
	generatedIngress, err := buildIngress(svcName, svc, cluster, ownerRefs)
	if err != nil {
		return err
	}

	if _, err = i.c.Ingresses(cluster.Namespace).Get(context.Background(), generatedIngress.Name,
		metav1.GetOptions{}); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		if _, err := i.c.Ingresses(cluster.Namespace).Create(context.Background(), generatedIngress,
			metav1.CreateOptions{}); err != nil {
			return err
		}

		return nil
	}

	if _, err := i.c.Ingresses(cluster.Namespace).Update(context.Background(), generatedIngress,
		metav1.UpdateOptions{}); errors.IsForbidden(err) || errors.IsInvalid(err) {
		_ = i.c.Ingresses(cluster.Namespace).Delete(context.Background(), generatedIngress.Name, metav1.DeleteOptions{})
		_, err := i.c.Ingresses(cluster.Namespace).Create(context.Background(), generatedIngress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Ingress) Delete(svcName string, cluster *spec.DiceCluster) error {
	err := i.c.Ingresses(cluster.Namespace).Delete(context.Background(), svcName, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	}

	return err
}

func buildIngress(
	svcName string,
	svc *diceyml.Service,
	cluster *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*extensions.Ingress, error) {
	if len(svc.Expose) == 0 {
		return nil, fmt.Errorf("svc: %v, not exposed any port, check diceyml", svc)
	}
	rules := make([]extensions.IngressRule, 0)
	for _, host := range common.ConvertHost(svcName, cluster) {
		rules = append(rules, extensions.IngressRule{
			Host: host,
			IngressRuleValue: extensions.IngressRuleValue{
				HTTP: &extensions.HTTPIngressRuleValue{
					Paths: GenHTTPIngressPaths(svcName, svc.Expose[0]),
				},
			},
		})
	}

	tls := []extensions.IngressTLS{
		{
			Hosts: common.ConvertHost(svcName, cluster),
		},
	}

	return &extensions.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            svcName,
			Namespace:       cluster.Namespace,
			OwnerReferences: ownerRefs,
			Annotations:     common.Annotations(svcName),
		},
		Spec: extensions.IngressSpec{
			Rules: rules,
			TLS:   tls,
		},
	}, nil
}

func GenHTTPIngressPaths(svcName string, exposePort int) []extensions.HTTPIngressPath {

	var httpIngressPaths []extensions.HTTPIngressPath

	httpIngressPath := extensions.HTTPIngressPath{
		Backend: extensions.IngressBackend{
			ServiceName: svcName,
			ServicePort: intstr.FromInt(exposePort),
		},
	}

	httpIngressPaths = append(httpIngressPaths, httpIngressPath)

	if svcName == types.ErdaServer {
		// TODO: remove register interface
		httpIngressPaths = append(httpIngressPaths, extensions.HTTPIngressPath{
			Backend: extensions.IngressBackend{
				ServiceName: types.ClusterManager,
				ServicePort: intstr.FromInt(80),
			},
			Path: types.AgentRegisterPath,
		})
	}
	return httpIngressPaths
}
