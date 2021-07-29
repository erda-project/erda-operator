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

package check

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func UntilDaemonsetReady(ctx context.Context, client kubernetes.Interface, ns, name string) error {
	ds, err := client.AppsV1().DaemonSets(ns).Get(context.Background(), name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		time.Sleep(2 * time.Second)
		return UntilDaemonsetReady(ctx, client, ns, name)
	} else if err != nil {
		return err
	}
	if CheckDaemonsetAvailable(ds) {
		return nil
	}
	selector := fmt.Sprintf("metadata.name=%s", name)
	w, err := client.AppsV1().DaemonSets(ns).Watch(context.Background(), metav1.ListOptions{
		FieldSelector:   selector,
		ResourceVersion: ds.ResourceVersion,
	})
	if err != nil {
		return err
	}
	resultch := w.ResultChan()
	for {
		select {
		case <-ctx.Done():
			w.Stop()
			return ctx.Err()
		case e := <-resultch:
			if e.Type == watch.Deleted || e.Type == watch.Error {
				break
			}
			ds, ok := e.Object.(*appsv1.DaemonSet)
			if !ok && e.Object == nil {
				return UntilDaemonsetReady(ctx, client, ns, name)
			}
			if !ok {
				logrus.Errorf("event: %v", e)
				continue
			}
			if CheckDaemonsetAvailable(ds) {
				return nil
			}
		}
	}
	panic("unreachable")
}

func CheckDaemonsetAvailable(ds *appsv1.DaemonSet) bool {
	return ds.Status.ObservedGeneration != 0 &&
		ds.Status.DesiredNumberScheduled == ds.Status.NumberAvailable &&
		ds.Status.NumberUnavailable == 0
}
