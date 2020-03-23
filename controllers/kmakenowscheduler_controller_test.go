package controllers

import (
	"golang.org/x/net/context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
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
	const kmnsname = "foo5"
	const kmakerunname = "foo6"
	const kmakename = "kmake3"

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

	Context("New kmake now scheduler", func() {

		schedMapName := ""
		kmsrName := ""

		key := types.NamespacedName{
			Name:      kmnsname,
			Namespace: namespace,
		}

		mapExists := func(mapKey bythepowerofv1.SubResource, mapName *string) {
			By("config map exists")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeNowScheduler{}
				k8sClient.Get(context.Background(), key, f)
				*mapName = f.Status.GetSubReference(mapKey)

				g := &corev1.ConfigMap{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      *mapName,
					Namespace: namespace,
				}, g)
			}, timeout, interval).Should(BeNil())
		}

		kmsrExists := func() {
			By("kmsr exists")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeNowScheduler{}
				k8sClient.Get(context.Background(), key, f)
				kmsrName = f.Status.GetSubReference(bythepowerofv1.Runs)

				g := &bythepowerofv1.KmakeScheduleRun{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      kmsrName,
					Namespace: namespace,
				}, g)
			}, timeout, interval).Should(BeNil())
		}

		mapNotExists := func(mapName string) {
			By("config map not exists")
			Eventually(func() error {
				f := &corev1.ConfigMap{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      mapName,
					Namespace: namespace,
				}, f)
			}, timeout, interval).ShouldNot(Succeed())
		}

		It("Should create successfully", func() {

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
				Spec: bythepowerofv1.KmakeRunSpec{},
			}

			Expect(k8sClient.Create(context.Background(), kmakerun)).Should(Succeed())

			By("Create kmake now scheduler")

			kmns := &bythepowerofv1.KmakeNowScheduler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kmnsname,
					Namespace: namespace,
				},
				Spec: bythepowerofv1.KmakeNowSchedulerSpec{
					Monitor: []string{"test"},
					Variables: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmns)).Should(Succeed())
		})
		time.Sleep(time.Second * 5)

		It("Should create config map", func() {
			mapExists(bythepowerofv1.EnvMap, &schedMapName)
		})

		It("Should recreate config map", func() {
			f := &bythepowerofv1.KmakeNowScheduler{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())

			f.Spec.Variables["key3"] = "value3"

			Expect(k8sClient.Update(context.Background(), f)).Should(Succeed())
			time.Sleep(time.Second * 5)

			mapNotExists(schedMapName)
			time.Sleep(time.Second * 5)
			schedMapName = "unset"
			mapExists(bythepowerofv1.EnvMap, &schedMapName)
		})

		It("Should create a job", func() {
			f := &bythepowerofv1.KmakeNowScheduler{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())

			kmsrExists()
		})

		It("Should delete", func() {

			By("delete now scheduler")
			f := &bythepowerofv1.KmakeNowScheduler{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmake now scheduler not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeNowScheduler{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

			By("delete kmakerun")
			f4 := &bythepowerofv1.KmakeRun{}
			key4 := types.NamespacedName{
				Name:      kmakerunname,
				Namespace: namespace,
			}
			Expect(k8sClient.Get(context.Background(), key4, f4)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f4)).Should(Succeed())

			By("delete kmake")
			key2 := types.NamespacedName{
				Name:      kmakename,
				Namespace: namespace,
			}
			f2 := &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), key2, f2)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f2)).Should(Succeed())

			By("delete kmsr")
			key3 := types.NamespacedName{
				Name:      kmsrName,
				Namespace: namespace,
			}
			f3 := &bythepowerofv1.KmakeScheduleRun{}
			// deleting the sheduler should delete this
			Expect(k8sClient.Get(context.Background(), key3, f3)).ShouldNot(Succeed())
			// Expect(k8sClient.Delete(context.Background(), f3)).Should(Succeed())
		})
	})
})
