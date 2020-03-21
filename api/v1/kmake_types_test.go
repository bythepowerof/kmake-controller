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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("Kmake", func() {
	var (
		key              types.NamespacedName
		created, fetched *Kmake
	)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Create API", func() {
		It("should create an object successfully", func() {
			key = types.NamespacedName{
				Name:      "foo",
				Namespace: "default",
			}
			created = &Kmake{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: KmakeSpec{
					Variables: map[string]string{
						"VAR1": "Value1",
						"VAR2": "Value2",
					},
					Rules: []KmakeRule{
						KmakeRule{
							Targets:       []string{"Rule1"},
							TargetPattern: "Rule%",
							DoubleColon:   true,
							Commands:      []string{"@echo $@"},
						},
						KmakeRule{
							Targets:  []string{"Rule2"},
							Commands: []string{"@echo $@"},
						},
					},
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &Kmake{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("generating a Makefile")
			Expect(created.Spec.ToMakefile()).To(Equal("Rule1:: Rule%:  \n\t@echo $@\nRule2:  \n\t@echo $@\n"))

			By("checking status field")
			Expect(fetched.GetStatus()).To(Equal(""))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})

		It("should correctly handle finalizers", func() {
			kmake := &Kmake{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
				},
			}
			Expect(kmake.IsBeingDeleted()).To(BeTrue())

			kmake.AddFinalizer(KmakeFinalizerName)
			Expect(len(kmake.GetFinalizers())).To(Equal(1))
			Expect(kmake.HasFinalizer(KmakeFinalizerName)).To(BeTrue())

			kmake.RemoveFinalizer(KmakeFinalizerName)
			Expect(len(kmake.GetFinalizers())).To(Equal(0))
			Expect(kmake.HasFinalizer(KmakeFinalizerName)).To(BeFalse())
		})
	})

})
