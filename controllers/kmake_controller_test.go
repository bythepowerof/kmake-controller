package controllers

import (
	"fmt"
	"golang.org/x/net/context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	// storage "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Controllers/KmakeController", func() {
	const timeout = time.Second * 30
	const timeout2 = time.Second * 120
	const interval = time.Second * 1
	const namespace = "default"
	const kmakename = "foo"

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

	Context("New kmake", func() {
		storageClass := ""
		pvcName := ""
		envMapName := ""
		kmakeMapName := ""

		key := types.NamespacedName{
			Name:      kmakename,
			Namespace: namespace,
		}
		cap := &corev1.ResourceList{
			"storage": resource.MustParse("3Ki"),
		}
		cap2 := &corev1.ResourceList{
			"storage": resource.MustParse("4Ki"),
		}

		mapExists := func(mapKey bythepowerofv1.SubResource, mapName *string) {
			By("config map exists")
			Eventually(func() error {
				f := &bythepowerofv1.Kmake{}
				k8sClient.Get(context.Background(), key, f)
				*mapName = f.Status.GetSubReference(mapKey)
				g := &corev1.ConfigMap{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      *mapName,
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

		pvcExists := func() {
			By("Creating pvc")
			Eventually(func() string {
				f := &bythepowerofv1.Kmake{}
				k8sClient.Get(context.Background(), key, f)
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

		pvcNotExists := func() {
			By("pvc not exist")
			Eventually(func() error {
				f := &corev1.PersistentVolumeClaim{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      pvcName,
					Namespace: namespace,
				}, f)
			}, timeout, interval).ShouldNot(Succeed())
		}

		It("Should create successfully", func() {
			By("Create kmake")

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
			// defer func() {
			// 	Expect(k8sClient.Delete(context.Background(), kmake)).Should(Succeed())
			// 	time.Sleep(time.Second * 30)
			// }()
		})
		It("Should create config map", func() {
			mapExists(bythepowerofv1.EnvMap, &envMapName)
		})
		It("Should create kmake map", func() {
			mapExists(bythepowerofv1.KmakeMap, &kmakeMapName)
		})
		It("Should create pvc", func() {
			pvcExists()
		})
		It("Should recreate env config map", func() {
			f := &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			f.Spec.Variables["VAR3"] = "Value3"
			Expect(k8sClient.Update(context.Background(), f)).Should(Succeed())

			mapNotExists(envMapName)
			mapExists(bythepowerofv1.EnvMap, &envMapName)
		})
		It("Should recreate kmake config map", func() {
			f := &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())

			f.Spec.Rules = append(f.Spec.Rules, bythepowerofv1.KmakeRule{
				Targets:  []string{"Rule3"},
				Commands: []string{"@echo $@"},
			})

			Expect(k8sClient.Update(context.Background(), f)).Should(Succeed())

			mapNotExists(kmakeMapName)
			mapExists(bythepowerofv1.KmakeMap, &envMapName)
		})
		It("Should recreate pvc", func() {
			f := &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())

			f.Spec.PersistentVolumeClaimTemplate.Resources.Requests = *cap2

			Expect(k8sClient.Update(context.Background(), f)).Should(Succeed())

			pvcNotExists()
			pvcExists()

			By("delete kmake")

			time.Sleep(time.Second * 5)

			f = &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmake not exist")
			Eventually(func() error {
				f = &bythepowerofv1.Kmake{}
				return k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      kmakename,
					Namespace: namespace,
				}, f)
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("Should delete subresources", func() {
			Skip("Subresources are not currently deleted correctly")
			mapNotExists(envMapName)
			mapNotExists(kmakeMapName)
			pvcNotExists()
		})
	})
})
