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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
	"github.com/erda-project/erda-operator/pkg/utils"
)

// ErdaReconciler reconciles a Erda object
type ErdaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	options
}

type options struct {
}

//+kubebuilder:rbac:groups=erda.terminus.io,resources=erdas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=erda.terminus.io,resources=erdas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=erda.terminus.io,resources=erdas/finalizers,verbs=update
func (r *ErdaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO: log
	log := r.Log.WithValues("erda-operator", req.NamespacedName)
	// TODO: reconcile timeout

	// fetch erda resource
	var erda erdav1beta1.Erda
	if err := r.Get(ctx, req.NamespacedName, &erda); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch Erda")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	references := erda.ComposeOwnerReferences()

	// TODO: There are two questions about pre-job tasks and post-job tasks:
	// TODO: 1. How to deal with the job task when it is failed? delete and recreate it?
	// TODO: 2. How to deal with the job task when it is completed and the user updates the DICE YAML

	//if len(erda.Spec.PreJobs) > 0 {
	//	erda.Status.Condition = erdav1beta1.ConditionPreJobs
	//
	//	erda.Status.PreJobStatus = r.ReconcileJob(erda, erdav1beta1.PreJobType)
	//	if err := r.Status().Update(ctx, &erda); err != nil {
	//		return ctrl.Result{Requeue: true}, err
	//	}
	//}

	dependEnvs := utils.ComposeDependEnvs(erda)

	for _, app := range erda.Spec.Applications {
		for _, component := range app.Components {
			// set component.Namespace value from Erda.Namespace
			component.Namespace = erda.Namespace

			err := r.SyncPersistentVolumeClaim(component)
			if client.IgnoreNotFound(err) != nil {
				log.Error(err, "sync pvc error")
				return ctrl.Result{Requeue: true}, err
			}

			// filter the existed secrets when be used in components
			r.SyncConfigurations(&component)

			if len(component.Network.ServiceDiscovery) > 0 {
				component.Envs = append(component.Envs, utils.ComposeSelfADDREnv(component,
					utils.ParseProtocol(app.Annotations[erdav1beta1.AnnotationSSLEnabled]))...)
			}
			component.Envs = append(component.Envs, utils.ComposeResourceToEnvs(component)...)
			component.Envs = utils.MergeEnvs(app.Envs, component.Envs)
			component.Envs = utils.ReplaceDependsEnv(dependEnvs, component.Envs)
			component.Envs = utils.ReplaceEnvironments(component.Envs)

			if component.EnvFrom != nil && app.EnvFrom != nil {
				component.EnvFrom = append(app.EnvFrom, component.EnvFrom...)
			} else {
				component.EnvFrom = app.EnvFrom
			}

			err, _ = r.ReconcileWorkload(ctx, component, references)
			if err != nil {
				return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
			}
		}
	}

	if err := r.SyncWorkLoadStatus(ctx, &erda); err != nil {
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}

	if erda.Status.Phase != erdav1beta1.PhaseReady {
		return ctrl.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	return ctrl.Result{}, nil

	// TODO: There are two questions about pre-job tasks and post-job tasks:
	// TODO: 1. How to deal with the job task when it is failed? delete and recreate it?
	// TODO: 2. How to deal with the job task when it is completed and the user updates the DICE YAML

	//if len(erda.Spec.PostJobs) > 0 {
	//	erda.Status.Condition = erdav1beta1.ConditionPostJobs
	//	erda.Status.PostJobStatus = r.ReconcileJob(erda, erdav1beta1.PostJobType)
	//	if err := r.Status().Update(ctx, &erda); err != nil {
	//		return ctrl.Result{Requeue: true}, err
	//	}
	//}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ErdaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&erdav1beta1.Erda{}).
		Complete(r)
}
