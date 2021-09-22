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

package status

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/erda-project/dice-operator/pkg/crd"
	"github.com/erda-project/dice-operator/pkg/spec"
)

func UpdatePhase(client rest.Interface, clus *spec.DiceCluster, namespace, name string, phase spec.ClusterPhase) error {
	return auxUpdateStatus(client, namespace, name, func(old spec.ClusterStatus) spec.ClusterStatus {
		old.Phase = phase
		clus.Status.Phase = phase
		return old
	})
}

func UpdateConditionAndPhase(client rest.Interface, clus *spec.DiceCluster, namespace, name string, cond spec.Condition, phase spec.ClusterPhase) error {
	return auxUpdateStatus(client, namespace, name, func(old spec.ClusterStatus) spec.ClusterStatus {
		if len(old.Conditions) > 20 {
			old.Conditions = old.Conditions[:20]
		}
		old.Conditions = append([]spec.Condition{{
			TransitionTime: time.Now().Format(time.RFC3339),
			Reason:         cond.Reason,
		}}, old.Conditions...)
		old.Phase = phase
		clus.Status.Phase = phase
		return old
	})
}

func UpdateComponentStatus(client rest.Interface, namespace, name string, status map[string]spec.ComponentStatus) error {
	return auxUpdateStatus(client, namespace, name, func(old spec.ClusterStatus) spec.ClusterStatus {
		if old.Components == nil {
			old.Components = map[string]spec.ComponentStatus{}
		}
		for component, s := range status {
			old.Components[component] = s
		}
		for k := range old.Components {
			if _, ok := status[k]; !ok {
				delete(old.Components, k)
			}
		}
		return old
	})
}

func auxUpdateStatus(client rest.Interface, namespace, name string, f func(old spec.ClusterStatus) spec.ClusterStatus) error {
	raw, err := client.Get().
		Prefix("apis", crd.GetCRDGroupVersion()).
		Namespace(namespace).
		Resource(crd.GetCRDPlural()).
		Name(name).
		DoRaw(context.Background())
	if err != nil {
		logrus.Errorf("Failed to fetch spec: %s/%s, err: %v, body: %v", namespace, name, err, string(raw))
		return err
	}
	before := &spec.DiceCluster{}
	if err := json.Unmarshal(raw, before); err != nil {
		logrus.Errorf("Failed to unmarshal spec.DiceCluster, err: %v", err)
		return err
	}
	if len(before.Status.Conditions) > 20 {
		before.Status.Conditions = before.Status.Conditions[:20]
	}
	afterStatus := f(before.Status)
	before.Status = afterStatus
	after, err := json.Marshal(before)
	if err != nil {
		return err
	}
	r, err := client.Put().
		Prefix("apis", crd.GetCRDGroupVersion()).
		Namespace(namespace).
		Resource(crd.GetCRDPlural()).
		Name(name).
		SubResource("status").
		Body(after).
		DoRaw(context.Background())
	if err != nil {
		logrus.Errorf("Failed to updateSpec: r: %v, err: %v", string(r), err)
		return err
	}
	return nil
}

func RevertResetStatus(client rest.Interface, namespace, name string) error {
	raw, err := client.Get().
		Prefix("apis", crd.GetCRDGroupVersion()).
		Namespace(namespace).
		Resource(crd.GetCRDPlural()).
		Name(name).
		DoRaw(context.Background())
	if err != nil {
		logrus.Errorf("Failed to fetch spec: %s/%s, err: %v, body: %v", namespace, name, err, string(raw))
		return err
	}
	before := &spec.DiceCluster{}
	if err := json.Unmarshal(raw, before); err != nil {
		logrus.Errorf("Failed to unmarshal spec.DiceCluster, err: %v", err)
		return err
	}
	before.Spec.ResetStatus = false
	after, err := json.Marshal(before)
	if err != nil {
		return err
	}
	r, err := client.Put().
		Prefix("apis", crd.GetCRDGroupVersion()).
		Namespace(namespace).
		Resource(crd.GetCRDPlural()).
		Name(name).
		Body(after).
		DoRaw(context.Background())
	if err != nil {
		logrus.Errorf("Failed to updateSpec: r: %v, err: %v", string(r), err)
		return err
	}
	return nil
}
