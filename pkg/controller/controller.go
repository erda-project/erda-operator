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

package controller

import (
	"context"
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/erda-project/dice-operator/pkg/cluster"
	"github.com/erda-project/dice-operator/pkg/crd"
	"github.com/erda-project/dice-operator/pkg/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	ErdaClusterInfoConfigMapEnv = "ERDA_CLUSTER_INFO"
	ErdaAddonInfoConfigMapEnv   = "ERDA_ADDON_INFO"
)

type Config struct {
	// TODO: make it configable
	Namespace string
}

type Controller struct {
	cfg          *Config
	client       rest.Interface
	k8sclient    kubernetes.Interface
	clientconfig *rest.Config
	diceClusters map[string]*cluster.Cluster
}

func New(client rest.Interface, k8sclient kubernetes.Interface, cfg *Config, clientcfg *rest.Config) *Controller {
	return &Controller{
		cfg:          cfg,
		client:       client,
		k8sclient:    k8sclient,
		clientconfig: clientcfg,
		diceClusters: make(map[string]*cluster.Cluster),
	}
}

var lastMaxGeneration int64 = 0

func compareAndUpdateGeneration(obj *spec.DiceCluster) (ignore bool) {
	if obj.Generation > lastMaxGeneration {
		lastMaxGeneration = obj.Generation
		return false
	}
	return true
}

type rawEvent struct {
	Type   string
	Object json.RawMessage
}

func (c *Controller) Run() {
	for {
		resourceversion, err := c.initDiceClusters()
		if err != nil {
			logrus.Fatalf("Failed to init diceclusters: %v", err)
		}
		c.monitor(resourceversion, c.onAdd, c.onUpdate, c.onDelete)
	}
}

func (c *Controller) monitor(resourceversion string, AddFunc, UpdateFunc, DeleteFunc func(interface{})) {
	monitorFailCount := 0
	for { // apiserver will close stream periodically? Yes
		body, err := c.client.Get().
			Prefix("apis", crd.GetCRDGroupVersion()).
			Namespace(c.cfg.Namespace).
			Resource(crd.GetCRDPlural()).
			VersionedParams(&metav1.ListOptions{ResourceVersion: resourceversion, Watch: true},
				metav1.ParameterCodec).Stream(context.Background())
		if err != nil {
			logrus.Errorf("monitor failed: resourceversion: %s, err: %v", resourceversion, err)
			monitorFailCount++
			if monitorFailCount > 3 {
				os.Exit(0)
			}
			continue
		}
		decoder := json.NewDecoder(body)
		for {
			rawevent := rawEvent{}
			if err := decoder.Decode(&rawevent); err != nil {
				break
			}
			switch strutil.ToUpper(rawevent.Type) {
			case "ADDED":
				obj := spec.DiceCluster{}
				if err := json.Unmarshal(rawevent.Object, &obj); err != nil {
					logrus.Errorf("Failed to unmarshal event obj, err: %v, raw: %s", err, string(rawevent.Object))
					continue
				}
				resourceversion = obj.ObjectMeta.GetResourceVersion()
				logrus.Debugf("monitor ADDED, namespace: %s, name: %s", obj.Namespace, obj.Name)
				AddFunc(&obj)
			case "MODIFIED":
				obj := spec.DiceCluster{}
				if err := json.Unmarshal(rawevent.Object, &obj); err != nil {
					logrus.Errorf("Failed to unmarshal event obj, err: %v, raw: %s", err, string(rawevent.Object))
					continue
				}
				resourceversion = obj.ObjectMeta.GetResourceVersion()
				logrus.Debugf("monitor MODIFIED, namespace: %s, name: %s", obj.Namespace, obj.Name)
				UpdateFunc(&obj)
			case "DELETED":
				obj := spec.DiceCluster{}
				if err := json.Unmarshal(rawevent.Object, &obj); err != nil {
					logrus.Errorf("Failed to unmarshal event obj, err: %v, raw: %s", err, string(rawevent.Object))
					continue
				}
				if compareAndUpdateGeneration(&obj) {
					continue
				}
				resourceversion = obj.ObjectMeta.GetResourceVersion()
				logrus.Debugf("monitor DELETED, namespace: %s, name: %s", obj.Namespace, obj.Name)
				DeleteFunc(&obj)
			case "ERROR":
				logrus.Errorf("Watching got Error event: %v", rawevent)
				// restart monitor with new resourceversion
				return
			case "BOOKMARK":
				panic("unreachable")
			}
		}
		body.Close()
	}
}

func (c *Controller) listDiceClusters() (*spec.DiceClusterList, error) {
	result := &spec.DiceClusterList{}
	req := c.client.Get().
		Prefix("apis", crd.GetCRDGroupVersion()).
		Namespace(c.cfg.Namespace).
		Resource(crd.GetCRDPlural()).
		VersionedParams(&metav1.ListOptions{}, metav1.ParameterCodec)
	r, err := req.DoRaw(context.Background())
	if err != nil {
		logrus.Errorf("listDiceClusters: %s, err: %v, req: %v", string(r), err, req) // debug print
		return nil, err
	}
	if err := json.Unmarshal(r, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Controller) initDiceClusters() (string, error) {
	clusterList, err := c.listDiceClusters()
	if err != nil {
		return "", err
	}
	for _, clus := range clusterList.Items {
		if os.Getenv(ErdaAddonInfoConfigMapEnv) != "" {
			clus.Spec.AddonConfigMap = os.Getenv(ErdaAddonInfoConfigMapEnv)
		}
		if os.Getenv(ErdaClusterInfoConfigMapEnv) != "" {
			clus.Spec.ClusterinfoConfigMap = os.Getenv(ErdaClusterInfoConfigMapEnv)
		}
		dc, err := cluster.New(&clus, c.client, c.k8sclient, c.clientconfig)
		if err != nil {
			return "", err
		}
		c.diceClusters[clus.Name] = dc
	}
	logrus.Infof("resourceversion=%s", clusterList.ObjectMeta.ResourceVersion)
	return clusterList.ObjectMeta.ResourceVersion, nil
}
