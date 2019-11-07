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
	"fmt"
	"strings"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubResource int

const (
	PVC SubResource = iota
	EnvMap
	KmakeMap
	Main
)

func (d SubResource) String() string {
	return [...]string{"PVC", "EnvMap", "KmakeMap", "Main"}[d]
}

type Phase int

const (
	Provision Phase = iota
	Delete
	BackOff
	Update
)

func (d Phase) String() string {
	return [...]string{"Provision", "Delete", "BackOff", "Update"}[d]
}

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
	JobImage    string            `json:"job_image"`
	Folders     []string          `json:"folders,omitempty"`

	JobTemplate                   batchv1beta1.JobTemplateSpec     `json:"job_template,omitempty"`
	PersistentVolumeClaimTemplate corev1.PersistentVolumeClaimSpec `json:"persistent_volume_claim_template"`
}

func (kmake *KmakeSpec) ToMakefile() (string, error) {
	var b strings.Builder
	hasTargetPattern := false

	for _, rule := range kmake.Rules {
		for _, target := range rule.Targets {
			fmt.Fprintf(&b, "%s ", target)
		}

		if rule.DoubleColon {
			fmt.Fprint(&b, "::")
		} else {
			fmt.Fprint(&b, ":")
		}

		for _, pattern := range rule.TargetPatterns {
			fmt.Fprintf(&b, "%s ", pattern)
			hasTargetPattern = true
		}

		if hasTargetPattern {
			fmt.Fprint(&b, ":")
		}

		for _, prereq := range rule.Prereqs {
			fmt.Fprintf(&b, "%s ", prereq)
		}

		fmt.Fprint(&b, "\n")

		for _, command := range rule.Commands {
			fmt.Fprintf(&b, "\t%s\n", command)
		}

		fmt.Fprint(&b, "\n")
	}
	return b.String(), nil
}

// KmakeStatus defines the observed state of Kmake
type KmakeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Runs      []KmakeRunStatus  `json:"runs,omitempty"`
	Status    string            `json:"status,omitempty"`
	Resources map[string]string `json:"kmake_resources,omitempty"`
}

func (status *KmakeStatus) UpdateSubResource(subresource SubResource, name string) {
	if name == "" {
		return
	}
	if status.Resources == nil {
		status.Resources = map[string]string{}
	}
	status.Resources[subresource.String()] = name
}

func (status *KmakeStatus) NameConcat(subresource SubResource) string {
	return status.Resources[subresource.String()]
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
	Targets        []string `json:"targets"`
	DoubleColon    bool     `json:"doublecolon,omitempty"`
	Commands       []string `json:"commands,omitempty"`
	Prereqs        []string `json:"prereqs,omitempty"`
	TargetPatterns []string `json:"target_patterns,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Kmake{}, &KmakeList{})
}
