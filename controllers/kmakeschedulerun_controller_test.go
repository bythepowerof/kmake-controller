package controllers

import (
	"fmt"
	"golang.org/x/net/context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controllers/KmakeRunController", func() {
	const timeout = time.Second * 30
	const timeout2 = time.Second * 120
	const interval = time.Second * 1
	const namespace = "default"
	const kmsrname = "foo4"
	const kmakerunname = "foo10"
	const kmakename = "kmake6"
	const schedenv = "schedenv"

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
			Name:      kmsrname,
			Namespace: namespace,
		}

		kmsrMeta := metav1.ObjectMeta{
			Name:      kmsrname,
			Namespace: namespace,
			Labels: map[string]string{
				"bythepowerof.github.io/schedule-instance": "test",
				"bythepowerof.github.io/run":               kmakerunname,
				"bythepowerof.github.io/kmake":             kmakename,
				"bythepowerof.github.io/schedule-env":      schedenv,
				"bythepowerof.github.io/workload":          "yes",
				"bythepowerof.github.io/status":            "Provision",
			},
		}

		pvcName := ""

		kmakekey := types.NamespacedName{
			Name:      kmakename,
			Namespace: namespace,
		}

		pvcExists := func() {
			By("Creating pvc")
			Eventually(func() string {
				f := &bythepowerofv1.Kmake{}
				k8sClient.Get(context.Background(), kmakekey, f)
				pvcName = f.Status.GetSubReference(bythepowerofv1.PVC)
				return pvcName
			}, timeout, interval).Should(MatchRegexp(fmt.Sprintf("(?i)%s-%s", kmakename, bythepowerofv1.PVC.String())))

			By("pending pvc")
			Eventually(func() corev1.PersistentVolumeClaimPhase {
				f := &corev1.PersistentVolumeClaim{}
				k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      pvcName,
					Namespace: namespace,
				}, f)
				return f.Status.Phase
			}, timeout, interval).Should(Equal(corev1.ClaimPending))
		}

		batchJobExists := func() {
			name := ""
			By("Creating batch.job")
			Eventually(func() string {
				f := &bythepowerofv1.KmakeScheduleRun{}
				k8sClient.Get(context.Background(), key, f)
				name = f.Status.GetSubReference(bythepowerofv1.Job)
				return name
			}, timeout, interval).Should(MatchRegexp(fmt.Sprintf("(?i)%s-%s", name, bythepowerofv1.Job.String())))

			By("created batch.job")
			Eventually(func() error {
				f := &v1.Job{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}, f)
			}, timeout, interval).Should(Succeed())
		}

		It("Should create start successfully", func() {
			By("Create kmake for run")

			cap := &corev1.ResourceList{
				"storage": resource.MustParse("3Ki"),
			}

			storageClass := ""

			kmake := &bythepowerofv1.Kmake{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kmakename,
					Namespace: namespace,
				},
				Spec: bythepowerofv1.KmakeSpec{
					PersistentVolumeClaimTemplate: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
						Resources: corev1.ResourceRequirements{
							Requests: *cap,
						},
						StorageClassName: &storageClass,
					},
					Variables: map[string]string{
						"VAR1": "Value1",
						"VAR2": "Value2",
					},
					Rules: []bythepowerofv1.KmakeRule{
						bythepowerofv1.KmakeRule{
							Targets:       []string{"Rule1"},
							TargetPattern: "Rule%",
							DoubleColon:   true,
							Commands:      []string{"@echo $@"},
						},
						bythepowerofv1.KmakeRule{
							Targets:  []string{"Rule2"},
							Commands: []string{"@echo $@"},
						},
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), kmake)).Should(Succeed())
			pvcExists()

			By("Create kmake run")

			kmakerun := &bythepowerofv1.KmakeRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kmakerunname,
					Namespace: namespace,
					Labels: map[string]string{
						"bythepowerof.github.io/kmake":     kmakename,
						"bythepowerof.github.io/scheduler": "test",
					},
				},
				Spec: bythepowerofv1.KmakeRunSpec{
					KmakeRunOperation: bythepowerofv1.KmakeRunOperation{
						Job: &bythepowerofv1.KmakeRunJob{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										corev1.Container{
											Name:    "test",
											Command: []string{"make"},
											Image:   "jeremymarshall/make-test:1",
											Args: []string{
												"-f",
												"/usr/share/kmake/kmake.mk",
											},
										},
									},
								},
							},
							Targets: []string{"Rule1"},
						},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmakerun)).Should(Succeed())

			By("Create sched env")

			schenv := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      schedenv,
					Namespace: namespace,
				},
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			}

			Expect(k8sClient.Create(context.Background(), schenv)).Should(Succeed())

			By("Create kmake chedule run - start")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: kmsrMeta,
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Start: &bythepowerofv1.KmakeScheduleRunStart{},
					},
				}}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())
			batchJobExists()

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
							Run: kmakerunname,
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

		// It("Should create stop successfully", func() {
		// 	By("Create kmake schedule run")

		// 	kmsr := &bythepowerofv1.KmakeScheduleRun{
		// 		ObjectMeta: kmsrMeta,
		// 		Spec: bythepowerofv1.KmakeScheduleRunSpec{
		// 			KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
		// 				Stop: &bythepowerofv1.KmakeScheduleRunStop{
		// 					Run: kmakerunname,
		// 				},
		// 			},
		// 		},
		// 	}

		// 	Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

		// 	By("delete kmsr")
		// 	time.Sleep(time.Second * 5)

		// 	f := &bythepowerofv1.KmakeScheduleRun{}
		// 	Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
		// 	Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

		// 	By("kmsr not exist")
		// 	Eventually(func() error {
		// 		f := &bythepowerofv1.KmakeScheduleRun{}
		// 		return k8sClient.Get(context.Background(), key, f)
		// 	}, timeout, interval).ShouldNot(Succeed())

		// 	By("delete kmakerun")
		// 	key2 := types.NamespacedName{
		// 		Name:      kmakerunname,
		// 		Namespace: namespace,
		// 	}

		// 	f2 := &bythepowerofv1.KmakeRun{}
		// 	Expect(k8sClient.Get(context.Background(), key2, f2)).Should(Succeed())
		// 	Expect(k8sClient.Delete(context.Background(), f2)).Should(Succeed())

		// 	By("delete kmake")
		// 	key3 := types.NamespacedName{
		// 		Name:      kmakename,
		// 		Namespace: namespace,
		// 	}
		// 	f3 := &bythepowerofv1.Kmake{}
		// 	Expect(k8sClient.Get(context.Background(), key3, f3)).Should(Succeed())
		// 	Expect(k8sClient.Delete(context.Background(), f3)).Should(Succeed())
		// })

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

		// It("Should create reset successfully", func() {
		// 	By("Create kmake schedule run")

		// 	kmsr := &bythepowerofv1.KmakeScheduleRun{
		// 		ObjectMeta: kmsrMeta,
		// 		Spec: bythepowerofv1.KmakeScheduleRunSpec{
		// 			KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
		// 				Reset: &bythepowerofv1.KmakeScheduleReset{},
		// 			},
		// 		},
		// 	}

		// 	Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

		// 	By("delete kmsr")
		// 	time.Sleep(time.Second * 5)

		// 	f := &bythepowerofv1.KmakeScheduleRun{}
		// 	Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
		// 	Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

		// 	By("kmsr not exist")
		// 	Eventually(func() error {
		// 		f := &bythepowerofv1.KmakeScheduleRun{}
		// 		return k8sClient.Get(context.Background(), key, f)
		// 	}, timeout, interval).ShouldNot(Succeed())
		// })

		// It("Should create force successfully", func() {
		// 	By("Create kmake schedule run")

		// 	kmsr := &bythepowerofv1.KmakeScheduleRun{
		// 		ObjectMeta: kmsrMeta,
		// 		Spec: bythepowerofv1.KmakeScheduleRunSpec{
		// 			KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
		// 				Force: &bythepowerofv1.KmakeScheduleForce{},
		// 			},
		// 		},
		// 	}

		// 	Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

		// 	By("delete kmsr")
		// 	time.Sleep(time.Second * 5)

		// 	f := &bythepowerofv1.KmakeScheduleRun{}
		// 	Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
		// 	Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

		// 	By("kmsr not exist")
		// 	Eventually(func() error {
		// 		f := &bythepowerofv1.KmakeScheduleRun{}
		// 		return k8sClient.Get(context.Background(), key, f)
		// 	}, timeout, interval).ShouldNot(Succeed())

		// })
	})
})
