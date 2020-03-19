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
	v11 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("Kmake", func() {
	var (
		key              types.NamespacedName
		created, fetched *KmakeRun
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
			created = &KmakeRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
					Labels:    map[string]string{"bythepowerof.github.io/kmake": "kmake-name"},
				},
				Spec: KmakeRunSpec{
					KmakeRunOperation: KmakeRunOperation{
						Job: &KmakeRunJob{
							Template: v11.PodTemplateSpec{
								Spec: v11.PodSpec{
									Containers: []v11.Container{
										v11.Container{
											Command: []string{"command text"},
											Image:   "image:latest",
											Args:    []string{"arg1", "arg2"},
										},
									},
								},
							},
						},
						Dummy:    &KmakeRunDummy{},
						FileWait: &KmakeRunFileWait{},
					},
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &KmakeRun{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("checking status field")
			Expect(fetched.GetStatus()).To(Equal(""))

			By("checking kmake name label")
			Expect(fetched.GetKmakeName()).To(Equal("kmake-name"))

			By("checking dummy function")
			Expect(fetched.Spec.KmakeRunOperation.Job.Dummy()).To(Equal("KmakeRunJob"))
			Expect(fetched.Spec.KmakeRunOperation.Dummy.Dummy()).To(Equal("KmakeRunDummy"))
			Expect(fetched.Spec.KmakeRunOperation.FileWait.Dummy()).To(Equal("KmakeRunFileWait"))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})

		It("should correctly handle finalizers", func() {
			kmakerun := &KmakeRun{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
				},
			}
			Expect(kmakerun.IsBeingDeleted()).To(BeTrue())

			kmakerun.AddFinalizer(KmakeFinalizerName)
			Expect(len(kmakerun.GetFinalizers())).To(Equal(1))
			Expect(kmakerun.HasFinalizer(KmakeFinalizerName)).To(BeTrue())

			kmakerun.RemoveFinalizer(KmakeFinalizerName)
			Expect(len(kmakerun.GetFinalizers())).To(Equal(0))
			Expect(kmakerun.HasFinalizer(KmakeFinalizerName)).To(BeFalse())
		})
	})

})
