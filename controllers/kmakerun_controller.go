/*

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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
)

// KmakeRunReconciler reconciles a KmakeRun object
type KmakeRunReconciler struct {
	client.Client
	Log         logr.Logger
	Recorder    record.EventRecorder
	KReconciler *KmakeReconciler
}

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakeruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakeruns/status,verbs=get;update;patch
func (r *KmakeRunReconciler) Event(instance *bythepowerofv1.KmakeRun, phase bythepowerofv1.Phase, subresource bythepowerofv1.SubResource, name string) {
	m := ""
	if name != "" {
		m = fmt.Sprintf("%v %v (%v)", phase.String(), subresource.String(), name)
	} else {
		m = fmt.Sprintf("%v %v", phase.String(), subresource.String())
	}
	r.Recorder.Event(instance, "Normal", phase.String()+subresource.String(), m)

	log := r.Log.WithValues("kmakerun", instance.GetName())
	log.Info(m)

	if instance.Status.Status != m {
		instance.Status.Status = m

		log.Info(name)

		instance.Status.UpdateSubResource(subresource, name)
		r.Status().Update(context.Background(), instance)
	}
}

func (r *KmakeRunReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kmakerun", req.NamespacedName)

	// requeue := ctrl.Result{Requeue: true}
	backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(1 * time.Minute))}

	// your logic here
	instance := &bythepowerofv1.KmakeRun{}
	err := r.Get(ctx, req.NamespacedName, instance)

	log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer log.Info(fmt.Sprintf("Finish reconcile loop for %v", req.NamespacedName))

	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.IsBeingDeleted() {
		r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Main, "")
		return ctrl.Result{}, nil
	}

	if instance.HasEnded() {
		return ctrl.Result{}, nil
	}

	// get parent (kmake)
	kmake := &bythepowerofv1.Kmake{}
	log.Info(fmt.Sprintf("Checking kmake %v", instance.GetKmakeName()))
	err = r.Get(ctx, types.NamespacedName{
		Namespace: instance.GetNamespace(),
		Name:      instance.GetKmakeName(),
	}, kmake)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found kmake %v", instance.GetKmakeName()))
			r.Event(instance, bythepowerofv1.Error, bythepowerofv1.KMAKE, instance.GetKmakeName())
			// don't requeue
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if instance.IsNew() {
		// defer to its owner kmake to schedule it
		// defer this so we don't get nested reconcile calls
		defer r.KReconciler.AppendRun(kmake, instance)
		r.Event(instance, bythepowerofv1.Wait, bythepowerofv1.Main, instance.GetName())

		// don't requeue
		return ctrl.Result{}, nil
	}

	if instance.IsActive() {
		// check the job
		currentjob := &v1.Job{}
		err = r.Get(ctx, instance.NamespacedNameConcat(bythepowerofv1.Job), currentjob)

		if err != nil {
			if errors.IsNotFound(err) {
				// make sure someone hasn't delete the Job
				if instance.NamespacedNameConcat(bythepowerofv1.Job).Name != "" {
					r.Event(instance, bythepowerofv1.Abort, bythepowerofv1.Job, currentjob.GetName())
					return reconcile.Result{}, nil
				}
			} else {
				return reconcile.Result{}, err
			}
		} else {
			// check the status
			// if currentjob.Status.Failed > 5 &&  currentjob.Status.Active > 0{
			// 	// r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Job, currentjob.GetName())
			// 	// try do scale to zero as its flapping
			// 	currentjob.Sca
			// 	return ctrl.Result{}, nil
			// }
			if currentjob.Status.Active > 0 {
				r.Event(instance, bythepowerofv1.Active, bythepowerofv1.Job, currentjob.GetName())
				return backoff5, nil
			}
			if currentjob.Status.Succeeded > 0 {
				r.Event(instance, bythepowerofv1.Success, bythepowerofv1.Job, currentjob.GetName())
				return ctrl.Result{}, nil
			}
			if currentjob.Status.Failed > 0 {
				r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Job, currentjob.GetName())
				return ctrl.Result{}, nil
			}
			return backoff5, nil
		}
	}

	// build the pod
	requiredjob := &v1.Job{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "job", "KmakeRun"),
	}
	requiredjob.Spec.Template = kmake.Spec.JobTemplate

	// use the image from here if there is one
	if instance.Spec.Image != "" {
		if requiredjob.Spec.Template.Spec.Containers == nil {
			requiredjob.Spec.Template.Spec.Containers = make([]corev1.Container, 1)
		}
		requiredjob.Spec.Template.Spec.Containers[0].Image = instance.Spec.Image
	}

	// use the image from here if there is one
	if instance.Spec.Image != "" {
		if requiredjob.Spec.Template.Spec.Containers == nil {
			requiredjob.Spec.Template.Spec.Containers = make([]corev1.Container, 1)
		}
		requiredjob.Spec.Template.Spec.Containers[0].Image = instance.Spec.Image
	}

	// give it a name
	if requiredjob.Spec.Template.Spec.Containers[0].Name == "" {
		requiredjob.Spec.Template.Spec.Containers[0].Name = "kmake-run"

	}

	// use the command from here if there is one
	if instance.Spec.Command != nil {
		requiredjob.Spec.Template.Spec.Containers[0].Command = instance.Spec.Command
	}

	// override the args
	if instance.Spec.Args != nil {
		requiredjob.Spec.Template.Spec.Containers[0].Args = instance.Spec.Args
	}

	// add in the targets as args
	if requiredjob.Spec.Template.Spec.Containers[0].Args == nil {
		requiredjob.Spec.Template.Spec.Containers[0].Args = instance.Spec.Targets
	} else {
		requiredjob.Spec.Template.Spec.Containers[0].Args = append(requiredjob.Spec.Template.Spec.Containers[0].Args, instance.Spec.Targets...)
	}

	// add in the env mount and env
	if requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts == nil {
		requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = make([]corev1.VolumeMount, 1)
		requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts[0] = corev1.VolumeMount{
			MountPath: "/usr/share/env",
			Name:      kmake.GetSubReference(bythepowerofv1.EnvMap),
		}
	} else {
		requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = append(
			requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				MountPath: "/usr/share/env",
				Name:      kmake.GetSubReference(bythepowerofv1.EnvMap),
			})
	}

	if requiredjob.Spec.Template.Spec.Volumes == nil {
		requiredjob.Spec.Template.Spec.Volumes = make([]corev1.Volume, 1)
		requiredjob.Spec.Template.Spec.Volumes[0] = corev1.Volume{
			Name: kmake.GetSubReference(bythepowerofv1.EnvMap),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: kmake.GetSubReference(bythepowerofv1.EnvMap)},
				},
			},
		}
	} else {
		requiredjob.Spec.Template.Spec.Volumes = append(
			requiredjob.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: kmake.GetSubReference(bythepowerofv1.EnvMap),
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: kmake.GetSubReference(bythepowerofv1.EnvMap)},
					},
				},
			})
	}

	requiredjob.Spec.Template.Spec.Containers[0].EnvFrom = append(
		requiredjob.Spec.Template.Spec.Containers[0].EnvFrom,
		corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: kmake.GetSubReference(bythepowerofv1.EnvMap)},
			},
		})

	// add in the pvc and mount
	requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = append(
		requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts,
		corev1.VolumeMount{
			MountPath: "/usr/share/pvc",
			Name:      kmake.GetSubReference(bythepowerofv1.PVC),
		})

	requiredjob.Spec.Template.Spec.Volumes = append(
		requiredjob.Spec.Template.Spec.Volumes,
		corev1.Volume{
			Name: kmake.GetSubReference(bythepowerofv1.PVC),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: kmake.GetSubReference(bythepowerofv1.PVC),
					ReadOnly:  false,
				},
			},
		})

	// add in the kmake mount
	requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = append(
		requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts,
		corev1.VolumeMount{
			MountPath: "/usr/share/kmake",
			Name:      kmake.GetSubReference(bythepowerofv1.KmakeMap),
		})

	requiredjob.Spec.Template.Spec.Volumes = append(
		requiredjob.Spec.Template.Spec.Volumes,
		corev1.Volume{
			Name: kmake.GetSubReference(bythepowerofv1.KmakeMap),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: kmake.GetSubReference(bythepowerofv1.KmakeMap)},
				},
			},
		})

	// fix the restart policy
	if requiredjob.Spec.Template.Spec.RestartPolicy == "" {
		requiredjob.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	}

	// create it
	err = r.Create(ctx, requiredjob)
	if err != nil {
		r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Job, requiredjob.ObjectMeta.Name)
		return reconcile.Result{}, err
	}
	r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.Job, requiredjob.ObjectMeta.Name)

	// apiVersion: v1
	// kind: Pod
	// metadata:
	//   name: kmake-sample-make
	// spec:
	//   volumes:
	//     - name: kmake-sample-pvc
	//       persistentVolumeClaim:
	//         claimName: kmake-sample-pvc
	//     - name: kmake-sample-env
	//       configMap:
	//         name: kmake-sample-env
	//     - name: kmake-sample-kmake
	//       configMap:
	//         name: kmake-sample-kmake
	//   containers:
	//     - name: make-sample
	//       image: jeremymarshall/make-test:1
	//       command: ["make"]
	//       args: [".KMAKESLEEP"]
	//       envFrom:
	//       - configMapRef:
	//           name: kmake-sample-env
	//       volumeMounts:
	//         - mountPath: "/usr/share/env"
	//           name: kmake-sample-env
	//         - mountPath: "/usr/share/kmake"
	//           name: kmake-sample-kmake
	//         - mountPath: /usr/share/pvc
	//           name: kmake-sample-pvc

	return ctrl.Result{}, nil
}

func (r *KmakeRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.KmakeRun{}).
		Complete(r)
}
