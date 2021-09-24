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

package controllers

//
//import (
//	"context"
//	"fmt"
//	"strings"
//
//	batchv1 "k8s.io/api/batch/v1"
//	"k8s.io/apimachinery/pkg/types"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//
//	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
//	"github.com/erda-project/erda-operator/pkg"
//	"github.com/erda-project/erda-operator/pkg/utils"
//)
//
//func (r *ErdaReconciler) ReconcileJob(erda erdav1beta1.Erda, jobType string) map[string]erdav1beta1.Status {
//
//	erdaJobs := erda.Spec.PreJobs
//	result := erda.Status.PreJobStatus
//
//	if jobType == erdav1beta1.PostJobType {
//		erdaJobs = erda.Spec.PostJobs
//		result = erda.Status.PostJobStatus
//	}
//
//	if result == nil {
//		result = map[string]erdav1beta1.Status{}
//	}
//
//	// list all kubernetes jobs via labels
//	k8sJobs := batchv1.JobList{}
//	if err := r.List(context.Background(), &k8sJobs, client.InNamespace(erda.Namespace),
//		client.MatchingLabels{
//			erdav1beta1.ErdaOperatorLabel: erda.Name,
//			erdav1beta1.ErdaJobTypeLabel:  strings.ToLower(erdav1beta1.PreJobType),
//		}); err != nil {
//		r.Log.Error(err, fmt.Sprintf("list %v job err", strings.ToLower(jobType)))
//		return result
//	}
//
//	// create and check job status in order
//	for index, job := range erdaJobs {
//		// set the name and namespace of ErdaJob, those only can be set by the controller
//		job.Name = fmt.Sprintf("%s-%s-%d", erda.Name, strings.ToLower(jobType), index)
//		job.Namespace = erda.Namespace
//		namespacedName := types.NamespacedName{Name: job.Name, Namespace: job.Namespace}
//
//		// check kubernetes job status if it is existed
//		isJobExist := false
//		for _, k8sJob := range k8sJobs.Items {
//			if k8sJob.Name == job.Name {
//				condition := pkg.IsJobFinished(k8sJob)
//				switch condition.Type {
//				case batchv1.JobComplete:
//					result[job.Name] = erdav1beta1.Status{Status: erdav1beta1.ComponentStatusCompleted, Message: ""}
//				case batchv1.JobFailed:
//					result[job.Name] = erdav1beta1.Status{Status: erdav1beta1.ComponentStatusFailed, Message: condition.Message}
//					return result
//				case "":
//					result[job.Name] = erdav1beta1.Status{Status: erdav1beta1.ComponentStatusRunning}
//					return result
//				}
//				isJobExist = true
//				break
//			}
//		}
//
//		// create the kubernetes job if it isn't exist
//		if !isJobExist {
//			job.Labels = utils.AppendLabels(job.Labels, map[string]string{
//				erdav1beta1.ErdaOperatorLabel: erda.Name,
//				erdav1beta1.ErdaJobTypeLabel:  strings.ToLower(jobType),
//			})
//			newK8sJob := pkg.ComposeKubernetesJob(&job, erda.ComposeOwnerReferences())
//			if err := r.Client.Create(context.Background(), &newK8sJob); err != nil {
//				errMsg := fmt.Sprintf("create job %v failed", namespacedName)
//				r.Log.Error(err, errMsg)
//				result[job.Name] = erdav1beta1.Status{Status: erdav1beta1.ComponentStatusFailed, Message: errMsg}
//				return result
//			}
//			result[job.Name] = erdav1beta1.Status{Status: erdav1beta1.ComponentStatusDeploying, Message: ""}
//			return result
//		}
//
//	}
//	return result
//}
