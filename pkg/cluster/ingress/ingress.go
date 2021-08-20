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

package ingress

import (
	"context"
	"fmt"

	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func HasIngress(dicesvc *diceyml.Service) bool {
	if dicesvc == nil {
		return false
	}
	return len(dicesvc.Expose) > 0
}

func CreateIfNotExists(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*extensions.Ingress, error) {
	generatedIngress, err := buildIngress(dicesvcname, dicesvc, clus, ownerRefs)
	if err != nil {
		return nil, err
	}
	ingress, err := client.ExtensionsV1beta1().Ingresses(clus.Namespace).Get(context.Background(), generatedIngress.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return client.ExtensionsV1beta1().Ingresses(clus.Namespace).Create(context.Background(), generatedIngress, metav1.CreateOptions{})
	}
	return ingress, err
}

func CreateOrUpdate(
	client kubernetes.Interface,
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*extensions.Ingress, error) {
	generatedIngress, err := buildIngress(dicesvcname, dicesvc, clus, ownerRefs)
	if err != nil {
		return nil, err
	}
	_, err = client.ExtensionsV1beta1().Ingresses(clus.Namespace).Get(context.Background(), generatedIngress.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return client.ExtensionsV1beta1().Ingresses(clus.Namespace).Create(context.Background(), generatedIngress, metav1.CreateOptions{})
	}
	ing, err := client.ExtensionsV1beta1().Ingresses(clus.Namespace).Update(context.Background(), generatedIngress, metav1.UpdateOptions{})
	if errors.IsForbidden(err) || errors.IsInvalid(err) {
		client.ExtensionsV1beta1().Ingresses(clus.Namespace).Delete(context.Background(), generatedIngress.Name, metav1.DeleteOptions{})
		return client.ExtensionsV1beta1().Ingresses(clus.Namespace).Create(context.Background(), generatedIngress, metav1.CreateOptions{})
	}
	return ing, nil
}

func GenHTTPIngressPaths(diceSvcName string, exposePort int) []extensions.HTTPIngressPath {

	var httpIngressPaths []extensions.HTTPIngressPath

	httpIngressPath := extensions.HTTPIngressPath{
		Backend: extensions.IngressBackend{
			ServiceName: diceSvcName,
			ServicePort: intstr.FromInt(exposePort),
		},
	}

	if diceSvcName == "cluster-dialer" {
		httpIngressPath.Path = "/clusteragent/connect"
	}

	httpIngressPaths = append(httpIngressPaths, httpIngressPath)
	return httpIngressPaths
}

func Delete(
	client kubernetes.Interface,
	dicesvcname string,
	clus *spec.DiceCluster) error {
	err := client.ExtensionsV1beta1().Ingresses(clus.Namespace).Delete(context.Background(), dicesvcname, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

func buildIngress(
	dicesvcname string,
	dicesvc *diceyml.Service,
	clus *spec.DiceCluster,
	ownerRefs []metav1.OwnerReference) (*extensions.Ingress, error) {
	if len(dicesvc.Expose) == 0 {
		return nil, fmt.Errorf("dicesvc: %v, not exposed any port, check diceyml", dicesvc)
	}
	rules := []extensions.IngressRule{}
	for _, host := range convertHost(dicesvcname, clus) {
		rules = append(rules, extensions.IngressRule{
			Host: host,
			IngressRuleValue: extensions.IngressRuleValue{
				HTTP: &extensions.HTTPIngressRuleValue{
					Paths: GenHTTPIngressPaths(dicesvcname, dicesvc.Expose[0]),
				},
			},
		})
	}

	tls := []extensions.IngressTLS{
		{
			Hosts: convertHost(dicesvcname, clus),
		},
	}

	return &extensions.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dicesvcname,
			Namespace:       clus.Namespace,
			OwnerReferences: ownerRefs,
			Annotations:     annotations(dicesvcname),
		},
		Spec: extensions.IngressSpec{
			Rules: rules,
			TLS:   tls,
		},
	}, nil
}

func convertHost(dicesvcname string, clus *spec.DiceCluster) []string {
	customdomainMap := clus.Spec.CustomDomain

	if domains, ok := customdomainMap[dicesvcname]; ok {
		return strutil.Map(strutil.Split(domains, ",", true), func(s string) string { return strutil.Trim(s) })
	}
	r, ok := map[string][]string{
		"ui": {
			fmt.Sprintf("dice.%s", clus.Spec.PlatformDomain),
			fmt.Sprintf("*.%s", clus.Spec.PlatformDomain),
		},
	}[dicesvcname]
	if !ok {
		return []string{fmt.Sprintf("%s.%s", dicesvcname, clus.Spec.PlatformDomain)}
	}
	return r
}

func annotations(dicesvcname string) map[string]string {

	annotation := map[string]string{
		"nginx.ingress.kubernetes.io/enable-access-log": "false",
	}

	switch dicesvcname {
	case "gittar", "openapi", "ui":
		annotation["nginx.ingress.kubernetes.io/proxy-body-size"] = "0"
	default:
	}

	return annotation
}
