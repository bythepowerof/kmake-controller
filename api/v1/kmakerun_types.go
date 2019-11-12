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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KmakeRunSpec defines the desired state of KmakeRun
type KmakeRunSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Image   string   `json:"image,omitempty"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Kmake   string   `json:"kmake"`
	Targets []string `json:"targets,omitempty"`
}

// KmakeRunStatus defines the observed state of KmakeRun
type KmakeRunStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Targets       []string          `json:"targets,omitempty"`
	StartTime     int64             `json:"start_time,omitempty"`
	TerminateTime int64             `json:"terminate_time,omitempty"`
	Status        string            `json:"status,omitempty"`
	ExicCode      int64             `json:"exit_code,omitempty"`
	Resources     map[string]string `json:"kmake_resources,omitempty"`
}

func (status *KmakeRunStatus) UpdateSubResource(subresource SubResource, name string) {
	if name == "" {
		return
	}
	if status.Resources == nil {
		status.Resources = map[string]string{}
	}
	status.Resources[subresource.String()] = name
}

func (status *KmakeRunStatus) NameConcat(subresource SubResource) string {
	return status.Resources[subresource.String()]
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="status of the run"
// +k8s:openapi-gen=true
// KmakeRun is the Schema for the kmakeruns API
type KmakeRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KmakeRunSpec   `json:"spec,omitempty"`
	Status KmakeRunStatus `json:"status,omitempty"`
}

func (kmakeRun *KmakeRun) IsBeingDeleted() bool {
	return !kmakeRun.ObjectMeta.DeletionTimestamp.IsZero()
}

func (kmakeRun *KmakeRun) HasEnded() bool {
	return strings.Contains(kmakeRun.Status.Status, "Success") ||
		strings.Contains(kmakeRun.Status.Status, "Error") ||
		strings.Contains(kmakeRun.Status.Status, "Abort")
}

func (kmakeRun *KmakeRun) IsActive() bool {
	return strings.Contains(kmakeRun.Status.Status, "Provision") ||
		strings.Contains(kmakeRun.Status.Status, "Active")
}

func (kmakeRun *KmakeRun) NamespacedNameConcat(subresource SubResource) types.NamespacedName {
	if _, ok := kmakeRun.Status.Resources[subresource.String()]; ok {
		return types.NamespacedName{
			Namespace: kmakeRun.GetNamespace(),
			Name:      kmakeRun.Status.Resources[subresource.String()],
		}
	}
	return types.NamespacedName{
		Namespace: kmakeRun.GetNamespace(),
		Name:      "",
	}
}

const KmakeRunFinalizerName = "kmakerun.finalizers.bythepowerof.github.com"

func (kmakeRun *KmakeRun) HasFinalizer(finalizerName string) bool {
	return containsString(kmakeRun.ObjectMeta.Finalizers, finalizerName)
}

func (kmake *KmakeRun) AddFinalizer(finalizerName string) {
	kmake.ObjectMeta.Finalizers = append(kmake.ObjectMeta.Finalizers, finalizerName)
}

func (kmake *KmakeRun) RemoveFinalizer(finalizerName string) {
	kmake.ObjectMeta.Finalizers = removeString(kmake.ObjectMeta.Finalizers, finalizerName)
}

func (kmake *KmakeRun) GetKmakeName() string {
	return kmake.Spec.Kmake
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
