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

package v1alpha1

import (
	"github.com/containersolutions/delayed-jobs-operator/pkg/types"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DelayedJobSpec defines the desired state of DelayedJob
type DelayedJobSpec struct {
	// DelayUntil sets the minimum UNIX Timestamp before Job is created
	DelayUntil types.Epoch `json:"delayUntil,omitempty"`
	v1.JobSpec `json:",inline"`
}

const (
	ConditionAwaitingDelay = "AwaitingDelay"
	ConditionCompleted     = "Completed"

	PhaseAwaitingDelay = "AwaitingDelay"
	PhaseCompleted     = "Completed"
)

// DelayedJobStatus defines the observed state of DelayedJob
type DelayedJobStatus struct {
	Phase      string             `json:"phase,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DelayedJob is the Schema for the delayedjobs API
type DelayedJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DelayedJobSpec   `json:"spec,omitempty"`
	Status DelayedJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DelayedJobList contains a list of DelayedJob
type DelayedJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DelayedJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DelayedJob{}, &DelayedJobList{})
}
