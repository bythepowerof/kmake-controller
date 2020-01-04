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
	Start   *KmakeScheduleRunStart   `json:"start,omitempty"`
	Restart *KmakeScheduleRunRestart `json:"restart,omitempty"`
	Stop    *KmakeScheduleRunStop    `json:"stop,omitempty"`
	Delete  *KmakeScheduleDelete     `json:"delete,omitempty"`
	Create  *KmakeScheduleCreate     `json:"create,omitempty"`
	Reset   *KmakeScheduleReset      `json:"reset,omitempty"`
	Force   *KmakeScheduleForce      `json:"force,omitempty"`
}

type KmakeScheduleRunStart struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Run      string `json:"run,omitempty"`
}

type KmakeScheduleRunRestart struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run string `json:"run,omitempty"`
}

type KmakeScheduleRunStop struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Run string `json:"run,omitempty"`
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
	Recurse string `json:"recurse,omitempty"`
	Full    string `json:"full,omitempty"`
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
	Resources map[string]string `json:"resources,omitempty"`
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
	if val, ok := kmsr.Labels["bythepowerof.github.io/status"]; ok {
		return strings.Contains(val, "Success") ||
			strings.Contains(val, "Error") ||
			strings.Contains(val, "Abort")
	}
	return false
}

func (kmsr *KmakeScheduleRun) IsActive() bool {
	if val, ok := kmsr.Labels["bythepowerof.github.io/status"]; ok {
		return strings.Contains(val, "Provision") ||
			strings.Contains(val, "Active")
	}
	return true
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

func (kmsr *KmakeScheduleRun) GetKmakeName() string {
	value, ok := kmsr.ObjectMeta.Labels["bythepowerof.github.io/kmake"]
	if ok {
		return value
	} else {
		return ""
	}
}

func (kmsr *KmakeScheduleRun) GetKmakeRunName() string {
	value, ok := kmsr.ObjectMeta.Labels["bythepowerof.github.io/run"]
	if ok {
		return value
	} else {
		return ""
	}
}

func (kmsr *KmakeScheduleRun) GetKmakeScheduleName() string {
	value, ok := kmsr.ObjectMeta.Labels["bythepowerof.github.io/schedule"]
	if ok {
		return value
	} else {
		return ""
	}
}

func (kmsr *KmakeScheduleRun) GetKmakeScheduleEnvName() string {
	value, ok := kmsr.ObjectMeta.Labels["bythepowerof.github.io/schedule-env"]
	if ok {
		return value
	} else {
		return ""
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
