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

package controllers

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func ObjectMetaConcat(owner metav1.Object, namespacedName types.NamespacedName, suffix string, kind string) metav1.ObjectMeta {

	// isController := true
	return metav1.ObjectMeta{
		Namespace:    namespacedName.Namespace,
		GenerateName: namespacedName.Name + "-" + suffix + "-",
		Labels:       owner.GetLabels(),
	}
}

// SetOwnerReference sets owner as a OwnerReference on owned.
// This is used for garbage collection of the owned object and for
// reconciling the owner object on changes to owned (with a Watch + EnqueueRequestForOwner).
func SetOwnerReference(owner, object metav1.Object, scheme *runtime.Scheme) error {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("%T is not a runtime.Object, cannot call SetControllerReference", owner)
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}

	// Create a new ref
	ref := *NewOwnerRef(owner, schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})

	existingRefs := object.GetOwnerReferences()
	fi := -1
	for i, r := range existingRefs {
		if referSameObject(ref, r) {
			fi = i
		}
	}
	if fi == -1 {
		existingRefs = append(existingRefs, ref)
	} else {
		existingRefs[fi] = ref
	}

	// Update owner references
	object.SetOwnerReferences(existingRefs)
	return nil
}

// NewOwnerRef creates an OwnerReference pointing to the given owner.
func NewOwnerRef(owner metav1.Object, gvk schema.GroupVersionKind) *metav1.OwnerReference {
	blockOwnerDeletion := false
	isController := false
	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}

// Returns true if a and b point to the same object
func referSameObject(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV == bGV && a.Kind == b.Kind && a.Name == b.Name
}

type OwnerReferencePatch struct {
	ApiVersion              string `yaml:"apiVersion"`
	Kind                    string `yaml:"kind"`
	*OwnerReferenceMetadata `yaml:"metadata"`
}

type OwnerReferenceMetadata struct {
	Name            string            `yaml:"name"`
	OwnerReferences []*OwnerReference `yaml:"ownerReferences"`
}

type OwnerReference struct {
	APIVersion         string    `yaml:"apiVersion"`
	Kind               string    `yaml:"kind"`
	Name               string    `yaml:"name"`
	UID                types.UID `yaml:"uid"`
	Controller         *bool     `yaml:"controller,omitempty"`
	BlockOwnerDeletion *bool     `yaml:"blockOwnerDeletion,omitempty"`
}

func NewOwnerReferencePatch(owner metav1.Object, scheme *runtime.Scheme) *OwnerReferencePatch {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return nil
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return nil
	}
	t := true

	return &OwnerReferencePatch{
		ApiVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		OwnerReferenceMetadata: &OwnerReferenceMetadata{
			Name: "not-important",
			OwnerReferences: []*OwnerReference{{
				APIVersion:         gvk.GroupVersion().String(),
				Kind:               gvk.Kind,
				Name:               owner.GetName(),
				UID:                owner.GetUID(),
				BlockOwnerDeletion: &t,
				Controller:         &t,
			}},
		},
	}
}
