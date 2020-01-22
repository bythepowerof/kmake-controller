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
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz"

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

func randomStringWithCharset(length int, charset string) string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	return randomStringWithCharset(length, charset)
}

// KmakeStatus defines the observed state of Kmake things
type KmakeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Runs      []*KmakeRuns      `json:"runs,omitempty"`
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

func (status *KmakeStatus) NameConcat(subresource SubResource) string {
	return status.Resources[subresource.String()]
}

type KmakeRule struct {
	Targets       []string `json:"targets"`
	DoubleColon   bool     `json:"doublecolon,omitempty"`
	Commands      []string `json:"commands,omitempty"`
	Prereqs       []string `json:"prereqs,omitempty"`
	TargetPattern string   `json:"targetpattern,omitempty"`
}

// +kubebuilder:object:generate=false
type KmakeScheduler interface {
	GetName() string
	GetNamespace() string
	Variables() []KV
	Monitor() []string
}

type KV struct {
	Key   string
	Value string
}
