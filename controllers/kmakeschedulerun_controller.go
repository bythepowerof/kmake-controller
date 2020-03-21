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
	"gopkg.in/yaml.v2"
	"time"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	// "k8s.io/apimachinery/pkg/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	Scheme   *runtime.Scheme
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
		instance.Labels["bythepowerof.github.io/status"] = phase.String()
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

	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.IsBeingDeleted() {
		err = r.handleFinalizer(instance)
		if err != nil {
			r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Main, "finalizer")
			return reconcile.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Main, "")
		if err != nil {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if !instance.HasFinalizer(bythepowerofv1.KmakeScheduleRunFinalizerName) {
		err = r.addFinalizer(instance)
		if err != nil {
			r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Main, "finalizer")

			return reconcile.Result{}, fmt.Errorf("error when handling kmakeschedulerun finalizer: %v", err)
		}
		r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.Main, "finalizer")

		return ctrl.Result{}, nil
	}

	var runType map[string]*json.RawMessage
	data, err := json.Marshal(instance.Spec.KmakeScheduleRunOperation)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = json.Unmarshal(data, &runType)
	if err != nil {
		return reconcile.Result{}, err
	}

	for k := range runType {
		switch k {
		case "start":
			kmakename := instance.GetKmakeName()
			kmakerun := instance.GetKmakeRunName()
			kmakescheduleEnv := instance.GetKmakeScheduleEnvName()

			if instance.HasEnded() {
				return ctrl.Result{}, nil
			}

			if !instance.IsActive() {
				// make sure the job isn't pending...
				return ctrl.Result{}, nil
			}

			// get kmakerun
			run := &bythepowerofv1.KmakeRun{}
			err = r.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: kmakerun}, run)

			if err != nil {
				if errors.IsNotFound(err) {
					return reconcile.Result{}, nil
				}
				return reconcile.Result{}, err
			}

			if instance.IsActive() {
				// check the job
				currentjob := &v1.Job{}
				err = r.Get(ctx, instance.Status.NamespacedNameConcat(bythepowerofv1.Job, instance.GetNamespace()), currentjob)

				if err != nil {
					if errors.IsNotFound(err) {
						// make sure someone hasn't delete the Job
						if instance.Status.NamespacedNameConcat(bythepowerofv1.Job, instance.GetNamespace()).Name != "" {
							r.Event(instance, bythepowerofv1.Abort, bythepowerofv1.Job, currentjob.GetName())
							return reconcile.Result{}, nil
						}
					} else {
						return reconcile.Result{}, err
					}
				} else {
					if currentjob.Status.Active > 0 {
						r.Event(instance, bythepowerofv1.Active, bythepowerofv1.Job, currentjob.GetName())
						return reconcile.Result{}, nil
					}
					if currentjob.Status.Succeeded > 0 {
						r.Event(instance, bythepowerofv1.Success, bythepowerofv1.Job, currentjob.GetName())
						return ctrl.Result{}, nil
					}
					if currentjob.Status.Failed > 0 {
						r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Job, currentjob.GetName())
						return ctrl.Result{}, nil
					}
					return reconcile.Result{}, nil
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
					r.Event(instance, bythepowerofv1.BackOff, bythepowerofv1.KMAKE, kmakename)
					// wait for kmake
					return backoff5, nil
				}
				return ctrl.Result{}, err
			}
			pvcName := kmake.Status.GetSubReference(bythepowerofv1.PVC)
			if pvcName == "" {
				log.Info(fmt.Sprintf("Not found kmake PVC %v", kmakename))
				r.Event(instance, bythepowerofv1.BackOff, bythepowerofv1.PVC, kmakename)
				// wait for kmake
				return backoff5, nil
			}

			// Job
			if run.Spec.KmakeRunOperation.Job != nil {
				// build the pod
				requiredjob := &v1.Job{
					ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "job", "KMSR"),
				}

				requiredjob.Labels["bythepowerof.github.io/schedulerun"] = instance.GetName()
				ctrl.SetControllerReference(instance, requiredjob, r.Scheme)

				if err := SetOwnerReference(kmake, requiredjob, r.Scheme); err != nil {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.KMAKE, requiredjob.ObjectMeta.Name)
					return reconcile.Result{}, err
				}
				if err = SetOwnerReference(run, requiredjob, r.Scheme); err != nil {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, requiredjob.ObjectMeta.Name)
					return reconcile.Result{}, err
				}
				if err = ctrl.SetControllerReference(instance, requiredjob, r.Scheme); err != nil {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Schedule, requiredjob.ObjectMeta.Name)
					return reconcile.Result{}, err
				}

				requiredjob.Spec.Template = run.Spec.KmakeRunOperation.Job.Template

				// add in the targets as args
				if requiredjob.Spec.Template.Spec.Containers[0].Args == nil {
					requiredjob.Spec.Template.Spec.Containers[0].Args = run.Spec.KmakeRunOperation.Job.Targets
				} else {
					requiredjob.Spec.Template.Spec.Containers[0].Args = append(requiredjob.Spec.Template.Spec.Containers[0].Args, run.Spec.KmakeRunOperation.Job.Targets...)
				}

				// add in the env mount and env
				if requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts == nil {
					requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = make([]corev1.VolumeMount, 1)
					requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts[0] = corev1.VolumeMount{
						MountPath: "/usr/share/env",
						Name:      kmake.Status.GetSubReference(bythepowerofv1.EnvMap),
					}
				} else {
					requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = append(
						requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts,
						corev1.VolumeMount{
							MountPath: "/usr/share/env",
							Name:      kmake.Status.GetSubReference(bythepowerofv1.EnvMap),
						})
				}

				if requiredjob.Spec.Template.Spec.Volumes == nil {
					requiredjob.Spec.Template.Spec.Volumes = make([]corev1.Volume, 1)
					requiredjob.Spec.Template.Spec.Volumes[0] = corev1.Volume{
						Name: kmake.Status.GetSubReference(bythepowerofv1.EnvMap),
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: kmake.Status.GetSubReference(bythepowerofv1.EnvMap)},
							},
						},
					}
				} else {
					requiredjob.Spec.Template.Spec.Volumes = append(
						requiredjob.Spec.Template.Spec.Volumes,
						corev1.Volume{
							Name: kmake.Status.GetSubReference(bythepowerofv1.EnvMap),
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: kmake.Status.GetSubReference(bythepowerofv1.EnvMap)},
								},
							},
						})
				}

				requiredjob.Spec.Template.Spec.Containers[0].EnvFrom = append(
					requiredjob.Spec.Template.Spec.Containers[0].EnvFrom,
					corev1.EnvFromSource{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: kmake.Status.GetSubReference(bythepowerofv1.EnvMap)},
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
						Name:      pvcName,
					})

				requiredjob.Spec.Template.Spec.Volumes = append(
					requiredjob.Spec.Template.Spec.Volumes,
					corev1.Volume{
						Name: kmake.Status.GetSubReference(bythepowerofv1.PVC),
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: kmake.Status.GetSubReference(bythepowerofv1.PVC),
								ReadOnly:  false,
							},
						},
					})

				// add in the kmake mount
				requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = append(
					requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts,
					corev1.VolumeMount{
						MountPath: "/usr/share/kmake",
						Name:      kmake.Status.GetSubReference(bythepowerofv1.KmakeMap),
					})

				requiredjob.Spec.Template.Spec.Volumes = append(
					requiredjob.Spec.Template.Spec.Volumes,
					corev1.Volume{
						Name: kmake.Status.GetSubReference(bythepowerofv1.KmakeMap),
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: kmake.Status.GetSubReference(bythepowerofv1.KmakeMap)},
							},
						},
					})

				// add in the owner config map
				j, err := json.Marshal(requiredjob.OwnerReferences)
				y, err := yaml.Marshal(requiredjob.OwnerReferences)
				km, err := yaml.Marshal(NewOwnerReferencePatch(kmake, r.Scheme))
				kmr, err := yaml.Marshal(NewOwnerReferencePatch(run, r.Scheme))
				kms, err := yaml.Marshal(NewOwnerReferencePatch(instance, r.Scheme))

				ownerconfigmap := &corev1.ConfigMap{
					ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "owner", "owner"),
					Data: map[string]string{
						"owner.yaml":                         string(y),
						"owner.json":                         string(j),
						"kmake-owner-patch.yaml":             string(km),
						"kmakerun-owner-patch.yaml":          string(kmr),
						"kmake-schedulerun-owner-patch.yaml": string(kms),
					},
				}
				ctrl.SetControllerReference(instance, ownerconfigmap, r.Scheme)
				err = r.Create(ctx, ownerconfigmap)
				if err != nil {
					return reconcile.Result{}, err
				}

				requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts = append(
					requiredjob.Spec.Template.Spec.Containers[0].VolumeMounts,
					corev1.VolumeMount{
						MountPath: "/usr/share/owner",
						Name:      ownerconfigmap.GetName(),
					})

				requiredjob.Spec.Template.Spec.Volumes = append(
					requiredjob.Spec.Template.Spec.Volumes,
					corev1.Volume{
						Name: ownerconfigmap.GetName(),
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: ownerconfigmap.GetName()},
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
				return reconcile.Result{}, nil
			}
			if run.Spec.KmakeRunOperation.Dummy != nil {
				err := r.Event(instance, bythepowerofv1.Success, bythepowerofv1.Dummy, instance.GetName())
				return reconcile.Result{}, err
			}
			if run.Spec.KmakeRunOperation.FileWait != nil {
				err := r.Event(instance, bythepowerofv1.Success, bythepowerofv1.FileWait, instance.GetName())
				return reconcile.Result{}, err
			}

		case "reset":
			if instance.IsNew() {

				del := &bythepowerofv1.KmakeScheduleRun{}

				do := &client.DeleteAllOfOptions{}
				labels := client.MatchingLabels{}

				if scheduler, ok := instance.GetLabels()["bythepowerof.github.io/schedule-instance"]; ok {
					do.ApplyOptions([]client.DeleteAllOfOption{
						client.InNamespace(req.NamespacedName.Namespace)})
					labels["bythepowerof.github.io/schedule-instance"] = scheduler
				} else {
					err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Runs, "No scheduler set")
					return reconcile.Result{}, fmt.Errorf("No scheduler set")
				}

				if instance.Spec.Reset.Full == "" || instance.Spec.Reset.Full == "no" {
					labels["bythepowerof.github.io/workload"] = "yes"
				}
				do.ApplyOptions([]client.DeleteAllOfOption{labels})

				err := r.DeleteAllOf(ctx, del, do)
				if err != nil {
					if !errors.IsNotFound(err) {
						return reconcile.Result{}, err
					}
					err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Runs, "No resources found")
					return reconcile.Result{}, err

				}
				err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Runs, "")
				return reconcile.Result{}, err
			}
		case "stop":
			if instance.IsNew() {
				var si, kmr string
				var ok bool
				if si, ok = instance.Labels["bythepowerof.github.io/schedule-instance"]; !ok {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, "No scheduler set")
					return reconcile.Result{}, err
				}
				if kmr, ok = instance.Labels["bythepowerof.github.io/run"]; !ok {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, "No kmakerun set")
					return reconcile.Result{}, err
				}

				do := &client.DeleteAllOfOptions{}
				del := &bythepowerofv1.KmakeScheduleRun{}
				do.ApplyOptions([]client.DeleteAllOfOption{
					client.InNamespace(req.NamespacedName.Namespace),
					client.MatchingLabels{"bythepowerof.github.io/schedule-instance": si,
						// "bythepowerof.github.io/status": "Active",
						"bythepowerof.github.io/run":      kmr,
						"bythepowerof.github.io/workload": "yes"},
				})

				err = r.DeleteAllOf(ctx, del, do)
				if err != nil {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, instance.GetName())
					return reconcile.Result{}, err
				}
				r.Event(instance, bythepowerofv1.Stop, bythepowerofv1.Runs, instance.GetName())
				return reconcile.Result{}, nil
			}
		case "restart":
			if instance.IsNew() {
				var si string
				var ok bool
				if si, ok = instance.Labels["bythepowerof.github.io/schedule-instance"]; !ok {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, "No scheduler set")
					return reconcile.Result{}, err
				}
				if instance.Spec.Restart.Run == "" {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, "No kmakerun set")
					return reconcile.Result{}, err
				}

				do := &client.DeleteAllOfOptions{}
				del := &bythepowerofv1.KmakeScheduleRun{}

				// x, err := labels.Parse("x in (foo,,baz),y,z notin ()")

				do.ApplyOptions([]client.DeleteAllOfOption{
					client.InNamespace(req.NamespacedName.Namespace),
					client.MatchingLabels{"bythepowerof.github.io/schedule-instance": si,
						// "bythepowerof.github.io/status": "Stop",
						"bythepowerof.github.io/run":      instance.Spec.Restart.Run,
						"bythepowerof.github.io/workload": "no"},
					// client.MatchingLabelsSelector{Selector: x},
				})

				err = r.DeleteAllOf(ctx, del, do)
				if err != nil {
					r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, instance.GetName())
					return reconcile.Result{}, err
				}
				r.Event(instance, bythepowerofv1.Restart, bythepowerofv1.Runs, instance.GetName())
				return reconcile.Result{}, nil
			}
		default:
			r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Runs, "Unknown operation")
			return reconcile.Result{}, nil
		}
		break // because we only expect one key
	}

	return reconcile.Result{}, nil
}

func (r *KmakeScheduleRunReconciler) SetupWithManager(mgr ctrl.Manager) error {

	jobOwnerKey := ".metadata.controller"
	apiGVStr := bythepowerofv1.GroupVersion.String()

	if err := mgr.GetFieldIndexer().IndexField(&v1.Job{}, jobOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*v1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Run...
		if owner.APIVersion != apiGVStr || owner.Kind != "KmakeScheduleRun" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.KmakeScheduleRun{}).
		Owns(&v1.Job{}).
		Complete(r)
}
