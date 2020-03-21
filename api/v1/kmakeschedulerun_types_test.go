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
		created, fetched *KmakeScheduleRun
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
			created = &KmakeScheduleRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: KmakeScheduleRunOperation{
						Start:   &KmakeScheduleRunStart{},
						Restart: &KmakeScheduleRunRestart{},
						Stop:    &KmakeScheduleRunStop{},
						Delete:  &KmakeScheduleDelete{},
						Create:  &KmakeScheduleCreate{},
						Reset:   &KmakeScheduleReset{},
						Force:   &KmakeScheduleForce{},
					},
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &KmakeScheduleRun{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("checking status field")
			Expect(fetched.GetStatus()).To(Equal(""))

			By("checking dummy function")
			Expect(fetched.Spec.KmakeScheduleRunOperation.Start.Dummy()).To(Equal("KmakeScheduleRunStart"))
			Expect(fetched.Spec.KmakeScheduleRunOperation.Restart.Dummy()).To(Equal("KmakeScheduleRunRestart"))
			Expect(fetched.Spec.KmakeScheduleRunOperation.Stop.Dummy()).To(Equal("KmakeScheduleRunStop"))
			Expect(fetched.Spec.KmakeScheduleRunOperation.Delete.Dummy()).To(Equal("KmakeScheduleDelete"))
			Expect(fetched.Spec.KmakeScheduleRunOperation.Create.Dummy()).To(Equal("KmakeScheduleCreate"))
			Expect(fetched.Spec.KmakeScheduleRunOperation.Reset.Dummy()).To(Equal("KmakeScheduleReset"))
			Expect(fetched.Spec.KmakeScheduleRunOperation.Force.Dummy()).To(Equal("KmakeScheduleForce"))

			By("checking status logic")
			Expect(fetched.HasEnded()).To(Equal(false))
			Expect(fetched.IsActive()).To(Equal(false))
			Expect(fetched.IsNew()).To(Equal(true))
			Expect(fetched.IsScheduled()).To(Equal(false))

			By("checking references")
			Expect(fetched.GetKmakeName()).To(Equal(""))
			Expect(fetched.GetKmakeRunName()).To(Equal(""))
			Expect(fetched.GetKmakeScheduleName()).To(Equal(""))
			Expect(fetched.GetKmakeScheduleEnvName()).To(Equal(""))
			Expect(fetched.GetJobName()).To(Equal(""))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})

		It("should correctly handle finalizers", func() {
			kmakeschedulerun := &KmakeScheduleRun{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
				},
			}
			Expect(kmakeschedulerun.IsBeingDeleted()).To(BeTrue())

			kmakeschedulerun.AddFinalizer(KmakeFinalizerName)
			Expect(len(kmakeschedulerun.GetFinalizers())).To(Equal(1))
			Expect(kmakeschedulerun.HasFinalizer(KmakeFinalizerName)).To(BeTrue())

			kmakeschedulerun.RemoveFinalizer(KmakeFinalizerName)
			Expect(len(kmakeschedulerun.GetFinalizers())).To(Equal(0))
			Expect(kmakeschedulerun.HasFinalizer(KmakeFinalizerName)).To(BeFalse())
		})
	})

})
