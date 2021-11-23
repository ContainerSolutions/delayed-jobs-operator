package controllers_test

import (
	"context"
	"github.com/containersolutions/delayed-jobs-operator/api/v1alpha1"
	"github.com/containersolutions/delayed-jobs-operator/controllers"
	types2 "github.com/containersolutions/delayed-jobs-operator/pkg/types"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	testing2 "k8s.io/utils/clock/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
)

func getSimpleJobSpec() batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "pi",
						Image: "perl",
						Command: []string{
							"perl",
							"-Mbignum=bpi",
							"-wle",
							"print bpi(2000)",
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}
}

func getSimpleDelayedJobSpec() v1alpha1.DelayedJob {
	return v1alpha1.DelayedJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foobar",
			Labels: map[string]string{
				"app": "foo",
			},
			UID: "some-value",
		},
		Spec: v1alpha1.DelayedJobSpec{
			JobSpec:    getSimpleJobSpec(),
			DelayUntil: types2.Epoch(0),
		},
	}
}

func TestDelayedJobReconciler_ReconcilesWithoutErrorIfNoObjects(t *testing.T) {
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
	}
	_, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "foobar",
		},
	})
	if err != nil {
		t.Errorf("Failed to reconcile without error (%v)", err)
	}
}

func TestDelayedJobReconciler_ReconcilesWithoutErrorIfObjects(t *testing.T) {
	delayedJob := getSimpleDelayedJobSpec()
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
	}
	_, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "foobar",
		},
	})
	if err != nil {
		t.Errorf("Failed to reconcile without error (%v)", err)
	}
}

func TestDelayedJobReconciler_ReconcileCreatesJob(t *testing.T) {
	delayedJob := getSimpleDelayedJobSpec()
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  testing2.NewFakeClock(time.Now()),
	}

	// Before we reconcile we want to make sure No Job exists inside the client
	job := &batchv1.Job{}
	err := controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, job)
	if err == nil {
		// We are expecting an error for NotFound.
		// If we don't receive an error, the test should fail
		t.Fatalf("Expected NotFound error when looking for Job")
	}
	if !errors.IsNotFound(err) {
		// If the error isn't NotFound, something else is wrong
		t.Fatalf("Unexpected error returned when looking for Job: (%v)", err)
	}

	_, err = controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, job)
	if err != nil {
		if errors.IsNotFound(err) {
			t.Errorf("The reconciler never created the job")
		} else {
			t.Fatalf("Failed to fetch created job (%v)", err)
		}
	}
}

func TestDelayedJobReconciler_ReconcileCreatesConditionCompleteTrueForCreatedJob(t *testing.T) {
	delayedJob := getSimpleDelayedJobSpec()
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  testing2.NewFakeClock(time.Now()),
	}

	// Before we reconcile we want to make sure No Job exists inside the client
	_, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	job := &batchv1.Job{}
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, job)
	if err != nil {
		if errors.IsNotFound(err) {
			t.Errorf("The reconciler never created the job")
		} else {
			t.Fatalf("Failed to fetch created job (%v)", err)
		}
	}

	// Get a fresh DelayedJob
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, &delayedJob)
	if !meta.IsStatusConditionTrue(delayedJob.Status.Conditions, v1alpha1.ConditionCompleted) {
		t.Error("Expected DelayedJob Condition Completed to be True")
	}
}

func TestDelayedJobReconciler_ReconcileCreatesConditionCompleteFalseForCreatedJob(t *testing.T) {
	delayedJob := getSimpleDelayedJobSpec()
	delayedJob.Spec.DelayUntil = types2.Epoch(time.Now().Unix() + 60)
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  testing2.NewFakeClock(time.Now()),
	}

	// Before we reconcile we want to make sure No Job exists inside the client
	_, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	job := &batchv1.Job{}
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, job)
	if err != nil {
		if !errors.IsNotFound(err) {
			t.Fatalf("Failed to fetch created job (%v)", err)
		}
	}

	// Get a fresh DelayedJob
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, &delayedJob)
	if !meta.IsStatusConditionFalse(delayedJob.Status.Conditions, v1alpha1.ConditionCompleted) {
		t.Error("Expected DelayedJob Condition Completed to be False")
	}
}

func TestDelayedJobReconciler_ReconcileCreatesConditionAwaitingDelayFalseForCreatedJob(t *testing.T) {
	delayedJob := getSimpleDelayedJobSpec()
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  testing2.NewFakeClock(time.Now()),
	}

	// Before we reconcile we want to make sure No Job exists inside the client
	_, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	job := &batchv1.Job{}
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, job)
	if err != nil {
		if errors.IsNotFound(err) {
			t.Errorf("The reconciler never created the job")
		} else {
			t.Fatalf("Failed to fetch created job (%v)", err)
		}
	}

	// Get a fresh DelayedJob
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, &delayedJob)
	if !meta.IsStatusConditionFalse(delayedJob.Status.Conditions, v1alpha1.ConditionAwaitingDelay) {
		t.Error("Expected DelayedJob Condition AwaitingDelay to be False")
	}
}

func TestDelayedJobReconciler_ReconcileCreatesConditionAwaitingDelayTrueForUncreatedJob(t *testing.T) {
	delayedJob := getSimpleDelayedJobSpec()
	delayedJob.Spec.DelayUntil = types2.Epoch(time.Now().Unix() + 60)
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  testing2.NewFakeClock(time.Now()),
	}

	// Before we reconcile we want to make sure No Job exists inside the client
	_, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	job := &batchv1.Job{}
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, job)
	if err != nil {
		if !errors.IsNotFound(err) {
			t.Fatalf("Failed to fetch created job (%v)", err)
		}
	}

	// Get a fresh DelayedJob
	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, &delayedJob)
	if !meta.IsStatusConditionTrue(delayedJob.Status.Conditions, v1alpha1.ConditionAwaitingDelay) {
		t.Error("Expected DelayedJob Condition AwaitingDelay to be False")
	}
}

func TestDelayedJobReconciler_ReconcileCreatesJobOnlyAfterDelayUntilHasPassed(t *testing.T) {
	// Setup fake clock first
	fakeClock := testing2.NewFakeClock(time.Now())
	delayedJob := getSimpleDelayedJobSpec()
	delayedJob.Spec.DelayUntil = types2.Epoch(fakeClock.Now().Unix() + 60)
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	}
	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  fakeClock,
	}

	// The time has not yet passed, so if we reconcile, the job should not be created
	_, err := controller.Reconcile(context.TODO(), request)
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	// Check that the job does not yet exist
	job := &batchv1.Job{}
	err = controller.Client.Get(context.TODO(), client.ObjectKeyFromObject(&delayedJob), job)
	if err == nil {
		// We are expecting an error for NotFound.
		// If we don't receive an error, the test should fail
		t.Fatalf("Expected NotFound error when looking for Job. Job was created too early.")
	}
	if !errors.IsNotFound(err) {
		// If the error isn't NotFound, something else is wrong
		t.Fatalf("Unexpected error returned when looking for Job: (%v)", err)
	}

	fakeClock.SetTime(time.Now().Add(time.Duration(61) * time.Second))
	controller.Clock = fakeClock
	_, err = controller.Reconcile(context.TODO(), request)
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	// Now we should see the Job in the client
	err = controller.Client.Get(context.TODO(), client.ObjectKeyFromObject(&delayedJob), job)
	if err != nil {
		if errors.IsNotFound(err) {
			t.Errorf("The reconciler never created the job")
		} else {
			t.Fatalf("Failed to fetch created job (%v)", err)
		}
	}
}

func TestDelayedJobReconciler_ReconcileReturnsRequeueWithDelayEqualToDelayDifference(t *testing.T) {
	// Setup fake clock first
	fakeClock := testing2.NewFakeClock(time.Now())
	delayedJob := getSimpleDelayedJobSpec()
	delayedJob.Spec.DelayUntil = types2.Epoch(fakeClock.Now().Unix() + 60)
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}
	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  fakeClock,
	}

	// The time has not yet passed, so if we reconcile, the job should not be created
	result, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	if result.RequeueAfter != 60*time.Second {
		t.Errorf("Expected RequeuAfter to equal Duration between DelayUntil and Now(). Expected %d, got %d", 60*time.Second, result.RequeueAfter)
	}

}

func TestDelayedJobReconciler_ReconcileCreatedJobHasItsOwnerRefSetToTheDelayedJob(t *testing.T) {
	delayedJob := getSimpleDelayedJobSpec()
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  testing2.NewFakeClock(time.Now()),
	}

	_, err := controller.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: delayedJob.Namespace,
			Name:      delayedJob.Name,
		},
	})
	if err != nil {
		t.Fatalf("Failed to reconcile without error (%v)", err)
	}

	job := &batchv1.Job{}
	// Now we should see the Job in the client
	controller.Client.Get(context.TODO(), client.ObjectKey{
		Namespace: delayedJob.Namespace,
		Name:      delayedJob.Name,
	}, job)

	if !metav1.IsControlledBy(job, &delayedJob) {
		t.Error("Expected Job to be controlled by DelayedJob")
	}
}

func TestDelayedJobReconciler_JobExists_ReturnsTrueIfJobAlreadyExists(t *testing.T) {
	// Setup fake clock first
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}
	fakeClock := testing2.NewFakeClock(time.Now())

	delayedJob := getSimpleDelayedJobSpec()
	delayedJob.Spec.DelayUntil = types2.Epoch(fakeClock.Now().Unix() + 60)
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       delayedJob.Name,
			GenerateName:               delayedJob.GenerateName,
			Namespace:                  delayedJob.Namespace,
			DeletionGracePeriodSeconds: delayedJob.DeletionGracePeriodSeconds,
			Labels:                     delayedJob.Labels,
			Annotations:                delayedJob.Annotations,
			Finalizers:                 delayedJob.Finalizers,
		},
		Spec:   getSimpleJobSpec(),
		Status: batchv1.JobStatus{},
	}

	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob, &job)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  fakeClock,
	}

	// The time has not yet passed, so if we reconcile, the job should not be created
	exists, err := controller.JobExists(context.TODO(), &delayedJob)
	if err != nil {
		t.Fatalf("Did not expect error from JobExists when Job exists, (%v)", err)
	}
	if !exists {
		t.Error("Expected JobExists to return True when job exists")
	}
}

func TestDelayedJobReconciler_JobExists_ReturnsFalseIfJobDoesNotExists(t *testing.T) {
	// Setup fake clock first
	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add DelayedJob scheme: (%v)", err)
	}
	fakeClock := testing2.NewFakeClock(time.Now())

	delayedJob := getSimpleDelayedJobSpec()
	delayedJob.Spec.DelayUntil = types2.Epoch(fakeClock.Now().Unix() + 60)
	clientBuilder := fake.NewClientBuilder()
	clientBuilder.WithObjects(&delayedJob)
	controller := controllers.DelayedJobReconciler{
		Client: clientBuilder.Build(),
		Scheme: s,
		Clock:  fakeClock,
	}

	// The time has not yet passed, so if we reconcile, the job should not be created
	exists, err := controller.JobExists(context.TODO(), &delayedJob)
	if err != nil {
		t.Fatalf("Did not expect error from JobExists when Job does not exist, (%v)", err)
	}
	if exists {
		t.Error("Expected JobExists to return False when job does not exist")
	}
}
