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
	"k8s.io/client-go/kubernetes"
)

func UntilDeploymentReady(ctx context.Context, client kubernetes.Interface, ns, name string) error {
	deploy, err := client.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(2) * time.Second):
		}
		logrus.Infof("deployment %s/%s not exists yet", ns, name)
		return UntilDeploymentReady(ctx, client, ns, name)
	} else if err != nil {
		return err
	}
	if CheckDeploymentAvailable(deploy) {
		return nil
	}
	selector := fmt.Sprintf("metadata.name=%s", name)
	w, err := client.AppsV1().Deployments(ns).Watch(context.Background(), metav1.ListOptions{
		FieldSelector:   selector,
		ResourceVersion: deploy.ResourceVersion,
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
			if e.Type == "DELETED" || e.Type == "ERROR" {
				break
			}
			deploy, ok := e.Object.(*appsv1.Deployment)
			if !ok && e.Object == nil {
				return UntilDaemonsetReady(ctx, client, ns, name)
			}
			if !ok {
				logrus.Errorf("event: %v", e)
				continue
			}
			if CheckDeploymentAvailable(deploy) {
				return nil
			}
		}
	}
	panic("unreachable")
}

func CheckDeploymentAvailable(deploy *appsv1.Deployment) bool {
	if deploy.Spec.Replicas == nil {
		return true
	}
	return deploy.Status.ObservedGeneration != 0 &&
		*deploy.Spec.Replicas == deploy.Status.AvailableReplicas &&
		deploy.Status.UnavailableReplicas == 0
}
