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
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KmakeSpec defines the desired state of Kmake
type KmakeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Variables []KmakeVariable `json:"variables,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
	Rules       []KmakeRule       `json:"rules"`
	MasterImage string            `json:"master_image"`
	StorageSize string            ` json:"storage_size"`
	JobImage    string            `json:"job_image"`
	Folders     []string          `json:"folders,omitempty"`

	JobTemplate batchv1beta1.JobTemplateSpec `json:"jobTemplate,omitempty"`
}

// KmakeStatus defines the observed state of Kmake
type KmakeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Runs   []KmakeRunStatus `json:"runs,omitempty"`
	Status string           `json:"status,omitempty"`
}

// Kmake is the Schema for the kmakes API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="status of the kind"
// +kubebuilder:object:root=true

type Kmake struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KmakeSpec   `json:"spec,omitempty"`
	Status KmakeStatus `json:"status,omitempty"`
}

// func (kmake *Kmake) IsSubmitted() bool {
// 	return kmake.Status.Status != ""
// }

func (kmake *Kmake) IsBeingDeleted() bool {
	return !kmake.ObjectMeta.DeletionTimestamp.IsZero()
}

const KmakeFinalizerName = "kmake.finalizers.bythepowerof.github.com"

func (kmake *Kmake) HasFinalizer(finalizerName string) bool {
	return containsString(kmake.ObjectMeta.Finalizers, finalizerName)
}

func (kmake *Kmake) AddFinalizer(finalizerName string) {
	kmake.ObjectMeta.Finalizers = append(kmake.ObjectMeta.Finalizers, finalizerName)
}

func (kmake *Kmake) RemoveFinalizer(finalizerName string) {
	kmake.ObjectMeta.Finalizers = removeString(kmake.ObjectMeta.Finalizers, finalizerName)
}

// +kubebuilder:object:root=true

// KmakeList contains a list of Kmake
type KmakeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kmake `json:"items"`
}

type KmakeRule struct {
	Targets     []string `json:"targets"`
	DoubleColon bool     `json:"double_colon,omitempty"`
	Commands    []string `json:"commands,omitempty"`
	Prereqs     []string `json:"prereqs,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Kmake{}, &KmakeList{})
}
