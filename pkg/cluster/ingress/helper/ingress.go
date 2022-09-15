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

package helper

import (
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/dice-operator/pkg/spec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"github.com/erda-project/dice-operator/pkg/utils"
	"github.com/erda-project/dice-operator/pkg/cluster/ingress/helper/networking/v1"
	v1beta1 "github.com/erda-project/dice-operator/pkg/cluster/ingress/helper/extension/v1beta1"
)

type IngressHelper interface {
	CreateIfNotExists(
		svcName string,
		svc *diceyml.Service,
		cluster *spec.DiceCluster,
		ownerRefs []metav1.OwnerReference,
	) error

	CreateOrUpdate(
		svcName string,
		svc *diceyml.Service,
		cluster *spec.DiceCluster,
		ownerRefs []metav1.OwnerReference,
	) error

	Delete(
		svcName string,
		cluster *spec.DiceCluster,
	) error
}

func New(c kubernetes.Interface) IngressHelper {
	if utils.VersionHas(extensionsv1beta1.SchemeGroupVersion.String()) {
		logrus.Debugf("ingress helper use version: %s", extensionsv1beta1.SchemeGroupVersion.String())
		return v1beta1.NewIngress(c.ExtensionsV1beta1())
	}

	logrus.Debugf("ingress helper use version: %s", networkingv1.SchemeGroupVersion.String())
	return v1.NewIngress(c.NetworkingV1())
}
