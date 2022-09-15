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

package crd

import (
	"context"
	"fmt"
	"os"
	"strings"

	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/api/errors"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/pointer"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/dice-operator/pkg/utils"
)

const (
	CRDKind          string = "Dice"
	CRDSingular      string = "dice"
	CRDPlural        string = "dices"
	CRDGroup         string = "dice.terminus.io"
	CRDVersion       string = "v1beta1"
	CRDKindSpecified string = "CRD_KIND_SPECIFIED"
)

// CreateCRD create 'dice' crd if not exists yet
func CreateCRD(rc *rest.Config) error {
	client, err := apiextension.NewForConfig(rc)
	if err != nil {
		return err
	}

	if utils.VersionHas(GetCRDGroupVersion()) {
		logrus.Infof("crd %s already existed, skip create.", GetCRDGroupVersion())
		return nil
	}

	if utils.VersionHas(apiextensionv1beta1.SchemeGroupVersion.String()) {
		_, err = client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
			context.Background(), &apiextensionv1beta1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: GetCRDFullName()},
				Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
					Group: GetCRDGroup(),
					Versions: []apiextensionv1beta1.CustomResourceDefinitionVersion{{
						Name:    CRDVersion,
						Served:  true,
						Storage: true,
					}},
					Scope: apiextensionv1beta1.NamespaceScoped,
					Names: apiextensionv1beta1.CustomResourceDefinitionNames{
						Plural:   GetCRDPlural(),
						Singular: GetCRDSingular(),
						Kind:     GetCRDKind(),
					},
					AdditionalPrinterColumns: []apiextensionv1beta1.CustomResourceColumnDefinition{
						{
							Name:        "Status",
							Type:        "string",
							Description: "Dice cluster current status",
							JSONPath:    ".status.phase",
						},
						{
							Name:        "LastMessage",
							Type:        "string",
							Description: "last message",
							JSONPath:    ".status.conditions[0].reason",
						},
					},
					Subresources: &apiextensionv1beta1.CustomResourceSubresources{
						Status: &apiextensionv1beta1.CustomResourceSubresourceStatus{},
					},
				},
			}, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}

		logrus.Infof("discover apiextension v1beta1, created crd %s.", GetCRDGroupVersion())

		return nil
	}

	if utils.VersionHas(apiextensionv1.SchemeGroupVersion.String()) {
		_, err = client.ApiextensionsV1().CustomResourceDefinitions().Create(
			context.Background(), &apiextensionv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: GetCRDFullName()},
				Spec: apiextensionv1.CustomResourceDefinitionSpec{
					Group: GetCRDGroup(),
					Names: apiextensionv1.CustomResourceDefinitionNames{
						Plural:   GetCRDPlural(),
						Singular: GetCRDSingular(),
						Kind:     GetCRDKind(),
					},
					Scope: apiextensionv1.NamespaceScoped,
					Versions: []apiextensionv1.CustomResourceDefinitionVersion{{
						Name:    CRDVersion,
						Served:  true,
						Storage: true,
						Schema: &apiextensionv1.CustomResourceValidation{
							OpenAPIV3Schema: &apiextensionv1.JSONSchemaProps{
								Type: "object",
								Properties: map[string]apiextensionv1.JSONSchemaProps{
									"spec": {
										Type:                   "object",
										XPreserveUnknownFields: pointer.Bool(true),
									},
									"status": {
										Type:                   "object",
										XPreserveUnknownFields: pointer.Bool(true),
									},
								},
							},
						},
						Subresources: &apiextensionv1.CustomResourceSubresources{
							Status: &apiextensionv1.CustomResourceSubresourceStatus{},
						},
						AdditionalPrinterColumns: []apiextensionv1.CustomResourceColumnDefinition{
							{
								Name:        "Status",
								Type:        "string",
								Description: "Dice cluster current status",
								JSONPath:    ".status.phase",
							},
							{
								Name:        "LastMessage",
								Type:        "string",
								Description: "last message",
								JSONPath:    ".status.conditions[0].reason",
							},
						},
					}},
				},
			}, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}

		logrus.Infof("discover apiextension v1, created crd %s.", GetCRDGroupVersion())

		return nil
	}

	return fmt.Errorf("unkonwn schema version finded")
}

func GetCRDPlural() string {
	if os.Getenv(CRDKindSpecified) != "" {
		return fmt.Sprintf("%ss", strings.ToLower(os.Getenv(CRDKindSpecified)))
	}
	return CRDPlural
}

func GetCRDSingular() string {
	if os.Getenv(CRDKindSpecified) != "" {
		return fmt.Sprintf("%s", strings.ToLower(os.Getenv(CRDKindSpecified)))
	}
	return CRDSingular
}

func GetCRDKind() string {
	if os.Getenv(CRDKindSpecified) != "" {
		return os.Getenv(CRDKindSpecified)
	}
	return CRDKind
}

func GetCRDGroup() string {
	if os.Getenv(CRDKindSpecified) != "" {
		return fmt.Sprintf("%s.terminus.io", strings.ToLower(os.Getenv(CRDKindSpecified)))
	}
	return CRDGroup
}

func GetCRDFullName() string {
	plural := GetCRDPlural()
	group := GetCRDGroup()
	return fmt.Sprintf("%s.%s", plural, group)
}

func GetCRDGroupVersion() string {
	group := GetCRDGroup()
	return fmt.Sprintf("%s/%s", group, CRDVersion)
}
