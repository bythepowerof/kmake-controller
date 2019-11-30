/*

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KmakeScheduleRunSpec defines the desired state of KmakeScheduleRun
type KmakeScheduleRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	KmakeScheduleRunOperation `json:"operation"`
}

type KmakeScheduleRunOperation struct {
	KmakeScheduleRunStart `json:"start,omitempty"`
	KmakeScheduleRunStop  `json:"stop,omitempty"`
	KmakeScheduleDelete   `json:"delete,omitempty"`
	KmakeScheduleCreate   `json:"create,omitempty"`
	KmakeScheduleReset    `json:"reset,omitempty"`
	KmakeScheduleForce    `json:"force,omitempty"`
}

type KmakeScheduleRunStart struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run      string `json:"run,omitempty"`
	Schedule string `json:"schedule,omitempty"`
}

type KmakeScheduleRunStop struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run string `json:"run,omitempty"`
}

type KmakeScheduleDelete struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Schedule string `json:"schedule,omitempty"`
}

type KmakeScheduleCreate struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run      string `json:"run,omitempty"`
	Schedule string `json:"schedule,omitempty"`
}

type KmakeScheduleReset struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run      string `json:"run,omitempty"`
	Schedule string `json:"schedule,omitempty"`
	Recurse  string `json:"recurse,omitempty"`
}

type KmakeScheduleForce struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run       string `json:"run,omitempty"`
	Schedule  string `json:"schedule,omitempty"`
	Operation string `json:"operation,omitempty"`
	Recurse   string `json:"recurse,omitempty"`
}

// KmakeScheduleRunStatus defines the observed state of KmakeScheduleRun
type KmakeScheduleRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    string            `json:"status,omitempty"`
	Resources map[string]string `json:"kmake_resources,omitempty"`
}

// +kubebuilder:object:root=true

// KmakeScheduleRun is the Schema for the kmakescheduleruns API
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="status of the schedule run"
type KmakeScheduleRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KmakeScheduleRunSpec   `json:"spec,omitempty"`
	Status KmakeScheduleRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KmakeScheduleRunList contains a list of KmakeScheduleRun
type KmakeScheduleRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KmakeScheduleRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KmakeScheduleRun{}, &KmakeScheduleRunList{})
}
