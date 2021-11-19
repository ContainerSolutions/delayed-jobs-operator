package controllers_test

import (
	"context"
	"github.com/containersolutions/delayed-jobs-operator/api/v1alpha1"
	"github.com/containersolutions/delayed-jobs-operator/controllers"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
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
			JobSpec: getSimpleJobSpec(),
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
