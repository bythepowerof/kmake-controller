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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KmakeRunSpec defines the desired state of KmakeRun
type KmakeRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Targets     []string               `json:"targets,omitempty"`
	JobTemplate corev1.PodTemplateSpec `json:"job_template"`
}

// KmakeRunStatus defines the observed state of KmakeRun
type KmakeRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// KmakeRun is the Schema for the kmakeruns API
type KmakeRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KmakeRunSpec `json:"spec,omitempty"`
	// Status KmakeRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// KmakeRunList contains a list of KmakeRun
type KmakeRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KmakeRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KmakeRun{}, &KmakeRunList{})
}
