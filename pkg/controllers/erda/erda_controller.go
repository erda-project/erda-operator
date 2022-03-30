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
	"k8s.io/apimachinery/pkg/api/errors"

	erdav1beta1 "github.com/erda-project/erda-operator/api/v1beta1"
)

var (
	requeueTime = 5 * time.Second
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

//+kubebuilder:rbac:groups=core.erda.cloud,resources=erdas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.erda.cloud,resources=erdas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.erda.cloud,resources=erdas/finalizers,verbs=update
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

	if len(erda.Spec.Jobs) > 0 {
		if err := r.ReconcileJob(ctx, &erda, references); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
		}
		switch erda.Status.Phase {
		case erdav1beta1.PhaseInitialization:
			return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, nil
		case erdav1beta1.PhaseFailed:
			return ctrl.Result{}, nil
		}
	}

	if err := r.ReconcileApplication(ctx, &erda, references); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
	}

	if erda.Status.Phase != erdav1beta1.PhaseReady {
		return ctrl.Result{Requeue: true, RequeueAfter: requeueTime}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ErdaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&erdav1beta1.Erda{}).
		Complete(r)
}
