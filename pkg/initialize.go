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

package pkg

import (
	"os"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/erda-project/dice-operator/pkg/conf"
	"github.com/erda-project/dice-operator/pkg/controller"
	"github.com/erda-project/dice-operator/pkg/crd"
)

const (
	EnableConfigMapNamespace = "ENABLE_CONFIGMAP_NAMESPACE"
)

func Initialize() {
	conf.Load()
	if conf.Debug() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("DEBUG MODE")
	}
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBE_CONFIG_PATH"))
	if err != nil {
		logrus.Fatalf("Failed to create config: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Failed to create client: %v", err)
	}
	if err := crd.CreateCRD(config); err != nil {
		logrus.Fatalf("Failed to create crd: %v", err)
	}

	var namespace = metav1.NamespaceDefault
	if os.Getenv(EnableConfigMapNamespace) != "" {
		namespace = os.Getenv(EnableConfigMapNamespace)
	}

	ctlr := controller.New(client.RESTClient(), client, &controller.Config{Namespace: namespace}, config)
	ctlr.Run()
}
