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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KmakeNowSchedulerSpec defines the desired state of KmakeNowScheduler
type KmakeNowSchedulerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Variables map[string]string `json:"variables,omitempty"`
	Monitor   []string          `json:"monitor"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// KmakeNowScheduler is the Schema for the kmakenowschedulers API
type KmakeNowScheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KmakeNowSchedulerSpec `json:"spec,omitempty"`
	Status KmakeStatus           `json:"status,omitempty"`
}

func (kmns *KmakeNowScheduler) IsBeingDeleted() bool {
	return !kmns.ObjectMeta.DeletionTimestamp.IsZero()
}

func (kmns *KmakeNowScheduler) GetSubReference(s SubResource) string {
	return kmns.Status.Resources[s.String()]
}

func (kmns *KmakeNowScheduler) NamespacedNameConcat(subresource SubResource) types.NamespacedName {
	if _, ok := kmns.Status.Resources[subresource.String()]; ok {
		return types.NamespacedName{
			Namespace: kmns.GetNamespace(),
			Name:      kmns.Status.Resources[subresource.String()],
		}
	}
	return types.NamespacedName{
		Namespace: kmns.GetNamespace(),
		Name:      "",
	}
}

// +kubebuilder:object:root=true

// KmakeNowSchedulerList contains a list of KmakeNowScheduler
type KmakeNowSchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KmakeNowScheduler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KmakeNowScheduler{}, &KmakeNowSchedulerList{})
}
