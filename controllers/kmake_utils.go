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
	"strings"

	// bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	"k8s.io/apimachinery/pkg/types"
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

func NamespacedNameConcat(owner *bythepowerofv1.Kmake, subresource SubResource) types.NamespacedName {
	switch subresource {
	case PVC:
		return types.NamespacedName{
			Namespace: owner.GetNamespace(),
			Name:      owner.Status.Resources.Pvc,
		}
	case EnvMap:
		return types.NamespacedName{
			Namespace: owner.GetNamespace(),
			Name:      owner.Status.Resources.Env,
		}
	case KmakeMap:
		return types.NamespacedName{
			Namespace: owner.GetNamespace(),
			Name:      owner.Status.Resources.Kmake,
		}
	}
	return types.NamespacedName{}
}

func ObjectMetaConcat(owner metav1.Object, namespacedName types.NamespacedName, suffix string) metav1.ObjectMeta {

	return metav1.ObjectMeta{
		Namespace:    namespacedName.Namespace,
		GenerateName: namespacedName.Name + "-" + suffix + "-",
		Labels:       owner.GetLabels(),
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

func NameConcat(status bythepowerofv1.KmakeStatus, subresource SubResource) string {
	switch subresource {
	case PVC:
		return status.Resources.Pvc
	case EnvMap:
		return status.Resources.Env
	case KmakeMap:
		return status.Resources.Kmake
	}
	return ""

}

func ToMakefile(rules []bythepowerofv1.KmakeRule) (string, error) {
	var b strings.Builder
	hasTargetPattern := false

	for _, rule := range rules {
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
