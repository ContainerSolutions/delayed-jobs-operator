package controllers_test

import (
	"context"
	"github.com/containersolutions/delayed-jobs-operator/api/v1alpha1"
	"github.com/containersolutions/delayed-jobs-operator/controllers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

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
	delayedJob := v1alpha1.DelayedJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foobar",
			Labels: map[string]string{
				"app": "foo",
			},
		},
		Spec: v1alpha1.DelayedJobSpec{},
	}
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
