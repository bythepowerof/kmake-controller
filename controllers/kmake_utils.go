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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type SubResource int

const (
	PVC SubResource = iota
	EnvMap
	KmakeMap
	Main
	KMAKE
	Job
)

func (d SubResource) String() string {
	return [...]string{"PVC", "EnvMap", "KmakeMap", "Main", "Kmake", "Job"}[d]
}
func ObjectMetaConcat(owner metav1.Object, namespacedName types.NamespacedName, suffix string, kind string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace:    namespacedName.Namespace,
		GenerateName: namespacedName.Name + "-" + suffix + "-",
		Labels:       owner.GetLabels(),
		OwnerReferences: []metav1.OwnerReference{
			metav1.OwnerReference{
				APIVersion: "bythepowerof.github.com/v1",
				Kind:       kind,
				Name:       owner.GetName(),
				UID:        owner.GetUID(),
			},
		},
	}
}

type KmakeAnnotation struct {
	Namespace string
	Resources map[string]string `json:"kmake_resources,omitempty"`
}

func (kma *KmakeAnnotation) UpdateSubResource(subresource SubResource, name string) {
	if name == "" {
		return
	}
	if kma.Resources == nil {
		kma.Resources = map[string]string{}
	}
	kma.Resources[subresource.String()] = name
}

func (kma *KmakeAnnotation) NameConcat(subresource SubResource) string {
	return kma.Resources[subresource.String()]
}

func (kma *KmakeAnnotation) GetSubReference(s SubResource) string {
	return kma.Resources[s.String()]
}

func (kma *KmakeAnnotation) NamespacedNameConcat(subresource SubResource) types.NamespacedName {
	if _, ok := kma.Resources[subresource.String()]; ok {
		return types.NamespacedName{
			Namespace: kma.GetNamespace(),
			Name:      kma.Resources[subresource.String()],
		}
	}
	return types.NamespacedName{
		Namespace: kmake.GetNamespace(),
		Name:      "",
	}
}
