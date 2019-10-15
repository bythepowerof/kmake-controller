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
	// bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func NamespacedNameConcat(namespacedName types.NamespacedName, suffix string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespacedName.Namespace,
		Name:      namespacedName.Name + "-" + suffix,
	}
}

func ObjectMetaConcat(owner metav1.Object, namespacedName types.NamespacedName, suffix string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespacedName.Namespace,
		Name:      namespacedName.Name + "-" + suffix,
		OwnerReferences: []metav1.OwnerReference{
			metav1.OwnerReference{
				APIVersion: "bythepowerof.github.com/v1",
				Kind:       "Kmake",
				Name:       owner.GetName(),
				UID:        owner.GetUID(),
			},
		},
	}
}

func NameConcat(namespacedName types.NamespacedName, suffix string) string {
	return namespacedName.Name + "-" + suffix
}
