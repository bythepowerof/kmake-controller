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
	"k8s.io/apimachinery/pkg/types"
	"strings"
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
	Start  *KmakeScheduleRunStart `json:"start,omitempty" protobuf:"bytes,1,opt,name=start"`
	Stop   *KmakeScheduleRunStop  `json:"stop,omitempty" protobuf:"bytes,2,opt,name=stop"`
	Delete *KmakeScheduleDelete   `json:"delete,omitempty" protobuf:"bytes,3,opt,name=delete"`
	Create *KmakeScheduleCreate   `json:"create,omitempty" protobuf:"bytes,4,opt,name=create"`
	Reset  *KmakeScheduleReset    `json:"reset,omitempty" protobuf:"bytes,5,opt,name=reset"`
	Force  *KmakeScheduleForce    `json:"force,omitempty" protobuf:"bytes,6,opt,name=force"`
}

type KmakeScheduleRunStart struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// 	Run      string `json:"run,omitempty"`
	// 	Schedule string `json:"schedule,omitempty"`
}

type KmakeScheduleRunStop struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Run string `json:"run,omitempty"`
}

type KmakeScheduleDelete struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Schedule string `json:"schedule,omitempty"`
}

type KmakeScheduleCreate struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Run      string `json:"run,omitempty"`
	// Schedule string `json:"schedule,omitempty"`
}

type KmakeScheduleReset struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Run      string `json:"run,omitempty"`
	// Schedule string `json:"schedule,omitempty"`
	Recurse string `json:"recurse,omitempty"`
}

type KmakeScheduleForce struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Run       string `json:"run,omitempty"`
	// Schedule  string `json:"schedule,omitempty"`
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

func (status *KmakeScheduleRunStatus) UpdateSubResource(subresource SubResource, name string) {
	if name == "" {
		return
	}
	if status.Resources == nil {
		status.Resources = map[string]string{}
	}
	status.Resources[subresource.String()] = name
}

func (status *KmakeScheduleRunStatus) NameConcat(subresource SubResource) string {
	return status.Resources[subresource.String()]
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

func (kmsr *KmakeScheduleRun) IsBeingDeleted() bool {
	return !kmsr.ObjectMeta.DeletionTimestamp.IsZero()
}

func (kmsr *KmakeScheduleRun) HasEnded() bool {
	return strings.Contains(kmsr.Status.Status, "Success") ||
		strings.Contains(kmsr.Status.Status, "Error") ||
		strings.Contains(kmsr.Status.Status, "Abort")
}

func (kmsr *KmakeScheduleRun) IsActive() bool {
	return strings.Contains(kmsr.Status.Status, "Provision") ||
		strings.Contains(kmsr.Status.Status, "Active")
}

func (kmsr *KmakeScheduleRun) IsNew() bool {
	return kmsr.Status.Status == ""
}

func (kmsr *KmakeScheduleRun) IsScheduled() bool {
	return false
}

func (kmsr *KmakeScheduleRun) NamespacedNameConcat(subresource SubResource) types.NamespacedName {
	if _, ok := kmsr.Status.Resources[subresource.String()]; ok {
		return types.NamespacedName{
			Namespace: kmsr.GetNamespace(),
			Name:      kmsr.Status.Resources[subresource.String()],
		}
	}
	return types.NamespacedName{
		Namespace: kmsr.GetNamespace(),
		Name:      "",
	}
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
