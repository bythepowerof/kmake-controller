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
	const kmakerunname = "foo2"
	const kmakename = "kmake"

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

	Context("New kmake run", func() {
		key := types.NamespacedName{
			Name:      kmakerunname,
			Namespace: namespace,
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
					Labels:    map[string]string{"bythepowerof.github.io/kmake": kmakename},
				},
				Spec: bythepowerofv1.KmakeRunSpec{},
			}

			Expect(k8sClient.Create(context.Background(), kmakerun)).Should(Succeed())

			By("delete kmakerun")

			time.Sleep(time.Second * 5)

			f := &bythepowerofv1.KmakeRun{}
			Expect(k8sClient.Get(context.Background(), key, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			By("kmakerun not exist")
			Eventually(func() error {
				f := &bythepowerofv1.KmakeRun{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

			By("delete kmake")
			key := types.NamespacedName{
				Name:      kmakename,
				Namespace: namespace,
			}
			f2 := &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), key, f2)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f2)).Should(Succeed())
		})
	})
})
