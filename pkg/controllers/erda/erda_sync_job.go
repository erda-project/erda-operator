// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package erda

import (
	"context"
	"strings"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/helper"
)

func (r *ErdaReconciler) ReconcileJob(ctx context.Context, erda *erdav1beta1.Erda, references []metav1.OwnerReference) error {
	erdaJobMap := make(map[string]*erdav1beta1.Job)
	// init status
	if erda.Status == nil {
		erda.Status = &erdav1beta1.ErdaStatus{}
	}

	if len(erda.Spec.Jobs) == 0 {
		return nil
	}

	for _, job := range erda.Spec.Jobs {
		if _, ok := erdaJobMap[job.Name]; ok {
			erda.Status.Phase = erdav1beta1.PhaseFailed
			if err := r.Status().Update(ctx, erda); err != nil {
				return err
			}
			return fmt.Errorf("job name is duplicated, job: %s", job.Name)
		}
		erdaJobMap[job.Name] = &job
	}

	erdaJobStatusMap := make(map[string]erdav1beta1.StatusType)

	// init job status
	if erda.Status.Jobs == nil {
		// reset all status, wait deploying
		for _, job := range erda.Spec.Jobs {
			erdaJobStatusMap[job.Name] = erdav1beta1.StatusUnKnown
		}

		erda.Status.Jobs = erdaJobStatusMap
		erda.Status.Phase = erdav1beta1.PhaseInitialization
		if err := r.Status().Update(ctx, erda); err != nil {
			return err
		}
	} else {
		// load current jobs status
		for name, eJobStatus := range erda.Status.Jobs {
			if _, ok := erdaJobMap[name]; !ok {
				continue
			}
			erdaJobStatusMap[name] = eJobStatus
		}
	}

	// check all jobs completed or not
	isCompleted := true
	for name, _ := range erdaJobMap {
		if erdaJobStatusMap[name] != erdav1beta1.StatusCompleted {
			isCompleted = false
		}
	}

	if isCompleted {
		// if all pre jobs completed, start to deploy applications
		if erda.Status.Phase == erdav1beta1.PhaseInitialization {
			erda.Status.Phase = erdav1beta1.PhaseDeploying
			err := r.Status().Update(ctx, erda)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// list all jobs via labels
	k8sJobs := batchv1.JobList{}
	if err := r.List(context.Background(), &k8sJobs, client.InNamespace(erda.Namespace),
		client.MatchingLabels{
			erdav1beta1.ErdaOperatorLabel: "true",
			erdav1beta1.ErdaJobTypeLabel:  strings.ToLower(erdav1beta1.PreJobType),
		}); err != nil {
		r.Log.Error(err, "list job err", "resource", erda.Name, "namespace", erda.Namespace)
		return err
	}

	// job items
	for _, kJob := range k8sJobs.Items {
		erdaJobName := kJob.Labels[erdav1beta1.ErdaJobNameLabel]
		_, ok := erdaJobMap[erdaJobName]
		if !ok {
			continue
		}
		// remove deployed job from map
		delete(erdaJobMap, erdaJobName)
		exeRes, jobCondition := helper.IsJobFinished(kJob)
		if !exeRes {
			erdaJobStatusMap[erdaJobName] = erdav1beta1.StatusRunning
			continue
		}
		switch jobCondition.Type {
		case batchv1.JobComplete:
			erdaJobStatusMap[erdaJobName] = erdav1beta1.StatusCompleted
		case batchv1.JobFailed:
			erdaJobStatusMap[erdaJobName] = erdav1beta1.StatusFailed
			erda.Status.Jobs = erdaJobStatusMap
			erda.Status.Phase = erdav1beta1.PhaseFailed
			if err := r.Status().Update(ctx, erda); err != nil {
				return err
			}
			return nil
		}
	}

	// job need to deploying
	for _, eJob := range erdaJobMap {
		eJob.Namespace = erda.Namespace
		kJob := helper.ComposeKubernetesJob(erda.Name, eJob, references)
		if err := r.Client.Create(context.Background(), &kJob); err != nil {
			return err
		}
		erdaJobStatusMap[eJob.Name] = erdav1beta1.StatusRunning
	}
	erda.Status.Jobs = erdaJobStatusMap
	erda.Status.Phase = erdav1beta1.PhaseInitialization
	if err := r.Status().Update(ctx, erda); err != nil {
		return err
	}
	return nil
}
