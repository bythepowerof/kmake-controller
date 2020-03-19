/*
Copyright 2019 microsoft.

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
	"k8s.io/apimachinery/pkg/types"
)

const charset = "abcdefghijklmnopqrstuvwxyz"

type SubResource int

const (
	PVC SubResource = iota
	EnvMap
	KmakeMap
	Main
	KMAKE
	Job
	Runs
	Schedule
	SchEnvMap
	Dummy
	FileWait
	Owner
)

func (d SubResource) String() string {
	return [...]string{"PVC", "EnvMap", "KmakeMap", "Main", "Kmake", "Job", "Runs", "Schedule", "SchEnvMap", "Dummy", "FileWait", "Owner"}[d]
}

type Phase int

const (
	Provision Phase = iota
	Delete
	BackOff
	Update
	Error
	Active
	Success
	Abort
	Wait
	Stop
	Restart
	Ready
)

func (d Phase) String() string {
	return [...]string{"Provision", "Delete", "BackOff", "Update", "Error", "Active", "Success", "Abort", "Wait", "Stop", "Restart", "Ready"}[d]
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// KmakeStatus defines the observed state of Kmake things
type KmakeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    string            `json:"status,omitempty"`
	Resources map[string]string `json:"resources,omitempty"`
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

func (status *KmakeStatus) NamespacedNameConcat(subresource SubResource, namespace string) types.NamespacedName {
	if name, ok := status.Resources[subresource.String()]; ok {
		return types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}
	}
	return types.NamespacedName{
		Namespace: namespace,
		Name:      "",
	}
}

func (status *KmakeStatus) GetSubReference(s SubResource) string {
	if name, ok := status.Resources[s.String()]; ok {
		return name
	}
	return ""
}

type KmakeRule struct {
	Targets       []string `json:"targets"`
	DoubleColon   bool     `json:"doublecolon,omitempty"`
	Commands      []string `json:"commands,omitempty"`
	Prereqs       []string `json:"prereqs,omitempty"`
	TargetPattern string   `json:"targetpattern,omitempty"`
}

type KV struct {
	Key   string
	Value string
}
