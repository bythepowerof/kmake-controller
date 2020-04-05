package controllers

import (
	"golang.org/x/net/context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controllers/KmakeRunController", func() {
	const timeout = time.Second * 30
	const timeout2 = time.Second * 120
	const interval = time.Second * 1
	const namespace = "default"
	const kmnsname = "foo4"

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
	Context("Run directly without existing job", func() {
		It("Should create successfully", func() {
			Expect(1).To(Equal(1))
		})
	})

	Context("New kmake schedule run", func() {
		key := types.NamespacedName{
			Name:      kmnsname,
			Namespace: namespace,
		}

		kmsrMeta := metav1.ObjectMeta{
			Name:      kmnsname,
			Namespace: namespace,
			Labels: map[string]string{
				"bythepowerof.github.io/schedule-instance": "test",
				"bythepowerof.github.io/run":               "test-run",
			},
		}

		It("Should create start successfully", func() {
			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Start: &bythepowerofv1.KmakeScheduleRunStart{},
					},
				}}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeScheduleRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmsr not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("Should create restart successfully", func() {
			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Restart: &bythepowerofv1.KmakeScheduleRunRestart{
							Run: "test-run",
						},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeScheduleRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmsr not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("Should create stop successfully", func() {
			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Stop: &bythepowerofv1.KmakeScheduleRunStop{},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeScheduleRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmsr not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("Should create delete successfully", func() {
			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Delete: &bythepowerofv1.KmakeScheduleDelete{},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeScheduleRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmsr not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("Should create create successfully", func() {
			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Create: &bythepowerofv1.KmakeScheduleCreate{},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeScheduleRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmsr not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("Should create reset successfully", func() {
			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Reset: &bythepowerofv1.KmakeScheduleReset{},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeScheduleRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmsr not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("Should create force successfully", func() {
			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Force: &bythepowerofv1.KmakeScheduleForce{},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeScheduleRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmsr not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
