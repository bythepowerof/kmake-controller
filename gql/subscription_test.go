package gql

import (
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
	const kmakerunname = "foo2"
	const kmnsname = "foo54"
	const kmsrname = "foo4"

	storageClass := ""

	keyk := types.NamespacedName{
		Name:      kmakename,
		Namespace: namespace,
	}

	keykr := types.NamespacedName{
		Name:      kmakerunname,
		Namespace: namespace,
	}

	keykmns := types.NamespacedName{
		Name:      kmnsname,
		Namespace: namespace,
	}

	keykmsr := types.NamespacedName{
		Name:      kmsrname,
		Namespace: namespace,
	}

	cap := &corev1.ResourceList{
		"storage": resource.MustParse("3Ki"),
	}

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

	Context("New Listener", func() {

		It("Should create successfully", func() {
			By("starting listener controllers")
			listener := NewKmakeListener(namespace, k8sManager)
			listener.KmakeChanges(namespace)

			Expect(listener).ShouldNot(BeNil())

			By("starting listener controllers")
			ch, err := listener.AddChangeClient(context.Background(), "default")

			Expect(err).Should(BeNil())
			Expect(ch).ShouldNot(BeClosed())

			By("Create kmake")

			kmake := &bythepowerofv1.Kmake{
				ObjectMeta: metav1.ObjectMeta{
					Name:       kmakename,
					Namespace:  namespace,
					Finalizers: []string{"123-abc"},
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

			Eventually(func() string {
				select {
				case result := <-ch:
					return result.GetName()
				default:
					return ""
				}
			}, timeout, interval).Should(Equal(kmakename))

			By("kmake modify")
			// time.Sleep(time.Second * 5)

			f := &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), keyk, f)).Should(Succeed())
			f.Spec.Variables["VAR3"] = "Value3"
			Expect(k8sClient.Update(context.Background(), f)).Should(Succeed())

			Eventually(func() string {
				select {
				case result := <-ch:
					return result.GetName()
				default:
					return ""
				}
			}, timeout, interval).Should(Equal(kmakename))

			By("kmakerun create")

			kmakerun := &bythepowerofv1.KmakeRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:       kmakerunname,
					Namespace:  namespace,
					Labels:     map[string]string{"bythepowerof.github.io/kmake": kmakename},
					Finalizers: []string{"123-abc"},
				},
				Spec: bythepowerofv1.KmakeRunSpec{},
			}

			Expect(k8sClient.Create(context.Background(), kmakerun)).Should(Succeed())

			// time.Sleep(time.Second * 5)

			Eventually(func() string {
				select {
				case result := <-ch:
					return result.GetName()
				default:
					return ""
				}
			}, timeout, interval).Should(Equal(kmakerunname))

			By("Create kmake now scheduler")

			kmns := &bythepowerofv1.KmakeNowScheduler{
				ObjectMeta: metav1.ObjectMeta{
					Name:       kmnsname,
					Namespace:  namespace,
					Finalizers: []string{"123-abc"},
				},
				Spec: bythepowerofv1.KmakeNowSchedulerSpec{
					Monitor: []string{"test2"},
					Variables: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), kmns)).Should(Succeed())

			Eventually(func() string {
				select {
				case result := <-ch:
					return result.GetName()
				default:
					return ""
				}
			}, timeout, interval).Should(Equal(kmnsname))

			By("Create kmake schedule run")

			kmsr := &bythepowerofv1.KmakeScheduleRun{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kmsrname,
					Namespace: namespace,
					Labels: map[string]string{
						"bythepowerof.github.io/schedule-instance": "test",
						"bythepowerof.github.io/run":               kmakerunname,
						"bythepowerof.github.io/kmake":             kmakename,
						"bythepowerof.github.io/schedule-env":      "test",
						"bythepowerof.github.io/workload":          "yes",
						"bythepowerof.github.io/status":            "Provision",
					},
					Finalizers: []string{"123-abc"},
				},
				Spec: bythepowerofv1.KmakeScheduleRunSpec{
					KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
						Start: &bythepowerofv1.KmakeScheduleRunStart{},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), kmsr)).Should(Succeed())

			By("delete kmsr")
			f4 := &bythepowerofv1.KmakeScheduleRun{}

			Eventually(func() error {
				return k8sClient.Get(context.Background(), keykmsr, f4)
			}, timeout, interval).Should(Succeed())

			// Expect(k8sClient.Get(context.Background(), keykmsr, f4)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f4)).Should(Succeed())

			// Eventually(func() string {
			// 	result := <-ch
			// 	return result.GetStatus()
			// }, timeout, interval).Should(Equal("Deleting"))

			By("delete now scheduler")
			f3 := &bythepowerofv1.KmakeNowScheduler{}
			Expect(k8sClient.Get(context.Background(), keykmns, f3)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f3)).Should(Succeed())

			// Eventually(func() string {
			// 	select {
			// 	case result := <-ch:
			// 		return result.GetStatus()
			// 	default:
			// 		return ""
			// 	}
			// }, timeout, interval).Should(Equal("Deleting"))

			By("kmakerun delete")

			f2 := &bythepowerofv1.KmakeRun{}
			Expect(k8sClient.Get(context.Background(), keykr, f2)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f2)).Should(Succeed())

			// time.Sleep(time.Second * 5)

			// Eventually(func() string {
			// 	select {
			// 	case result := <-ch:
			// 		return result.GetStatus()
			// 	default:
			// 		return ""
			// 	}
			// }, timeout, interval).Should(Equal("Deleting"))

			By("kmake delete")

			f = &bythepowerofv1.Kmake{}
			Expect(k8sClient.Get(context.Background(), keyk, f)).Should(Succeed())
			Expect(k8sClient.Delete(context.Background(), f)).Should(Succeed())

			// these don't come in the deleted order...
			// var kmakeStatus string

			// Eventually(func() string {
			// 	select {
			// 	case result := <-ch:
			// 		kmakeStatus = result.GetStatus()
			// 		return result.GetName()
			// 	default:
			// 		kmakeStatus = ""
			// 		return ""
			// 	}
			// }, timeout, interval).Should(Equal(kmakename))
			// Expect(kmakeStatus).Should(Equal("Deleting"))
		})
	})
})
