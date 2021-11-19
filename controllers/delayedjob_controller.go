/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/containersolutions/delayed-jobs-operator/pkg/types"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1alpha1 "github.com/containersolutions/delayed-jobs-operator/api/v1alpha1"
)

// DelayedJobReconciler reconciles a DelayedJob object
type DelayedJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock  clock.Clock
}

//+kubebuilder:rbac:groups=batch.container-solutions.com,resources=delayedjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.container-solutions.com,resources=delayedjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.container-solutions.com,resources=delayedjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DelayedJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *DelayedJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling")

	delayedJob := &batchv1alpha1.DelayedJob{}
	err := r.Get(ctx, req.NamespacedName, delayedJob)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("DelayedJob not found. Most likely deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if types.Epoch(r.Clock.Now().Unix()) > delayedJob.Spec.DelayUntil {
		logger.Info("Creating job for DelayedJob")
		// We need to create a job from
		job := r.GetNewJob(delayedJob)
		err = ctrl.SetControllerReference(delayedJob, job, r.Scheme)
		if err != nil {
			logger.Error(err, "Could not set DelayedJob as owner of Job")
		}
		err = r.Client.Create(context.TODO(), job)
		if err != nil {
			logger.Error(err, "Could not create job for DelayedJob")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *DelayedJobReconciler) GetNewJob(delayedJob *batchv1alpha1.DelayedJob) *v1.Job {
	return &v1.Job{
		TypeMeta: delayedJob.TypeMeta,
		ObjectMeta: ctrl.ObjectMeta{
			Name:                       delayedJob.Name,
			GenerateName:               delayedJob.GenerateName,
			Namespace:                  delayedJob.Namespace,
			DeletionGracePeriodSeconds: delayedJob.DeletionGracePeriodSeconds,
			Labels:                     delayedJob.Labels,
			Annotations:                delayedJob.Annotations,
			Finalizers:                 delayedJob.Finalizers,
		},
		Spec: delayedJob.Spec.JobSpec,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DelayedJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.DelayedJob{}).
		Complete(r)
}
