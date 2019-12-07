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
	"encoding/json"
	"fmt"
	"time"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// KmakeScheduleRunReconciler reconciles a KmakeScheduleRun object
type KmakeScheduleRunReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *KmakeScheduleRunReconciler) Event(instance *bythepowerofv1.KmakeScheduleRun, phase bythepowerofv1.Phase, subresource bythepowerofv1.SubResource, name string) error {
	m := ""
	if name != "" {
		m = fmt.Sprintf("%v %v (%v)", phase.String(), subresource.String(), name)
	} else {
		m = fmt.Sprintf("%v %v", phase.String(), subresource.String())
	}
	r.Recorder.Event(instance, "Normal", phase.String()+subresource.String(), m)

	log := r.Log.WithValues("kmake", instance.GetName())
	log.Info(m)

	if instance.Status.Status != m {
		instance.Status.Status = m

		log.Info(name)

		instance.Status.UpdateSubResource(subresource, name)
		r.Status().Update(context.Background(), instance)
		bytes, err := json.Marshal(instance.Status.Resources)
		if err != nil {
			return err
		}
		if instance.Annotations == nil {
			instance.Annotations = make(map[string]string)
		}
		instance.Annotations["bythepowerof.github.io/kmake"] = string(bytes)
		return r.Update(context.Background(), instance)
	}
	return nil
}

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakescheduleruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakescheduleruns/status,verbs=get;update;patch

func (r *KmakeScheduleRunReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	// your logic here
	ctx := context.Background()
	log := r.Log.WithValues("kmakeschedulerun", req.NamespacedName)

	// requeue := ctrl.Result{Requeue: true}
	backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(1 * time.Minute))}

	// your logic here

	// instance is kmsr
	log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer log.Info(fmt.Sprintf("Finish reconcile loop for %v", req.NamespacedName))

	instance := &bythepowerofv1.KmakeScheduleRun{}
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

	if instance.Spec.Start != nil {

		kmakename := instance.GetKmakeName()
		kmakerun := instance.GetKmakeRunName()
		kmakescheduleEnv := instance.GetKmakeScheduleEnvName()

		// get kmakerun
		run := &bythepowerofv1.KmakeRun{}
		err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: kmakerun}, run)

		if err != nil {
			if errors.IsNotFound(err) {
				return reconcile.Result{}, nil
			}
			return reconcile.Result{}, err
		}

		// if run.IsBeingDeleted() {
		// 	r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Main, "")
		// 	return ctrl.Result{}, nil
		// }

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

		kmake := &bythepowerofv1.Kmake{}
		log.Info(fmt.Sprintf("Checking kmake %v", kmakename))
		err = r.Get(ctx, types.NamespacedName{
			Namespace: run.GetNamespace(),
			Name:      kmakename,
		}, kmake)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info(fmt.Sprintf("Not found kmake %v", kmakename))
				r.Event(instance, bythepowerofv1.Error, bythepowerofv1.KMAKE, kmakename)
				// don't requeue
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, err
		}
		// build the pod
		requiredjob := &v1.Job{
			ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "job", "KMSR"),
		}
		requiredjob.Spec.Template = run.Spec.JobTemplate

		// add in the targets as args
		if requiredjob.Spec.Template.Spec.Containers[0].Args == nil {
			requiredjob.Spec.Template.Spec.Containers[0].Args = run.Spec.Targets
		} else {
			requiredjob.Spec.Template.Spec.Containers[0].Args = append(requiredjob.Spec.Template.Spec.Containers[0].Args, run.Spec.Targets...)
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

		// add in the sched env mount and env
		requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = append(
			requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				MountPath: "/usr/share/schedule",
				Name:      kmakescheduleEnv,
			})

		requiredjob.Spec.Template.Spec.Volumes = append(
			requiredjob.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: kmakescheduleEnv,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: kmakescheduleEnv},
					},
				},
			})

		requiredjob.Spec.Template.Spec.Containers[0].EnvFrom = append(
			requiredjob.Spec.Template.Spec.Containers[0].EnvFrom,
			corev1.EnvFromSource{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: kmakescheduleEnv},
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
		return ctrl.Result{}, nil
	}
	// }
	return ctrl.Result{}, nil
}

func (r *KmakeScheduleRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.KmakeScheduleRun{}).
		Complete(r)
}
