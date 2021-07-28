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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func UntilJobFinished(ctx context.Context, client kubernetes.Interface, ns, name string) error {
	job, err := client.BatchV1().Jobs(ns).Get(context.Background(), name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		time.Sleep(2 * time.Second)
		logrus.Infof("job %s/%s not exists yet", ns, name)
		return UntilJobFinished(ctx, client, ns, name)
	} else if err != nil {
		return err
	}

	if finished, err := CheckJobFinished(job); finished {
		return err
	}
	selector := fmt.Sprintf("metadata.name=%s", name)
	w, err := client.BatchV1().Jobs(ns).Watch(context.Background(), metav1.ListOptions{
		FieldSelector:   selector,
		ResourceVersion: job.ResourceVersion,
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
			job := e.Object.(*batchv1.Job)
			if finished, err := CheckJobFinished(job); finished {
				return err
			}
		}
	}
	panic("unreachable")
}

func CheckJobFinished(job *batchv1.Job) (bool, error) {
	if len(job.Status.Conditions) == 0 {
		return false, nil
	}

	for _, cond := range job.Status.Conditions {
		switch cond.Type {
		case batchv1.JobComplete:
			if cond.Status == corev1.ConditionTrue {
				return true, nil
			} else {
				return false, nil
			}
		case batchv1.JobFailed:
			if cond.Status == corev1.ConditionTrue {
				return true, fmt.Errorf("%s", cond.Message)
			} else {
				return false, nil
			}
		}
	}
	return false, nil
}
