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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
)

// KmakeNowSchedulerReconciler reconciles a KmakeNowScheduler object
type KmakeNowSchedulerReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

func (r *KmakeNowSchedulerReconciler) Event(instance *bythepowerofv1.KmakeNowScheduler, phase bythepowerofv1.Phase, subresource bythepowerofv1.SubResource, name string) error {
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

		var err error
		instance.Annotations, err = bythepowerofv1.SetDomainAnnotation(instance.Annotations, instance.Status.Resources)
		if err != nil {
			return err
		}
		return r.Update(context.Background(), instance)
	}
	return nil
}

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakenowschedulers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakenowschedulers/status,verbs=get;update;patch

func (r *KmakeNowSchedulerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	log := r.Log.WithValues("kmakenowscheduler", req.NamespacedName)
	requeue := ctrl.Result{Requeue: true}
	backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(1 * time.Minute))}
	// backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(10 * time.Second))}

	// your logic here
	instance := &bythepowerofv1.KmakeNowScheduler{}
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

	if !instance.HasFinalizer(bythepowerofv1.KmakeNowSchedulerFinalizerName) {
		err = r.addFinalizer(instance)
		if err != nil {
			r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Main, "finalizer")
			return reconcile.Result{}, fmt.Errorf("error when handling kmakenowscheduler finalizer: %v", err)
		}
		r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.Main, "finalizer")
		return ctrl.Result{}, nil
	}

	// env configmap

	currentenvmap := &corev1.ConfigMap{}
	requiredenvmap := &corev1.ConfigMap{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, bythepowerofv1.EnvMap),

		Data: instance.Spec.Variables,
	}
	ctrl.SetControllerReference(instance, requiredenvmap, r.Scheme)
	log.Info(fmt.Sprintf("Checking env map %v", instance.Status.GetSubReference(bythepowerofv1.EnvMap)))

	err = r.Get(ctx, instance.Status.NamespacedNameConcat(bythepowerofv1.EnvMap, instance.GetNamespace()), currentenvmap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found env map %v", instance.Status.GetSubReference(bythepowerofv1.EnvMap)))

			// create it
			err = r.Create(ctx, requiredenvmap)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.EnvMap, requiredenvmap.ObjectMeta.Name)
			if err != nil {
				return reconcile.Result{}, err
			}
			return requeue, err

		}
		return reconcile.Result{}, err
	}
	if !(equality.Semantic.DeepEqual(currentenvmap.Data, requiredenvmap.Data) &&
		equality.Semantic.DeepEqual(currentenvmap.ObjectMeta.Labels, requiredenvmap.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("delete env map %v", instance.Status.GetSubReference(bythepowerofv1.EnvMap)))
		err = r.Delete(ctx, currentenvmap)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.EnvMap, "")
		if err != nil {
			return reconcile.Result{}, err
		}
		return requeue, nil
	}

	// search for things label bythepowerof.github.io/scheduler

	// look at the scheduleruns just for this instance...
	runs := &bythepowerofv1.KmakeScheduleRunList{}
	opts := []client.ListOption{
		client.InNamespace(req.NamespacedName.Namespace),
		client.MatchingLabels{bythepowerofv1.ScheduleLabel.String(): instance.GetName()},
	}
	err = r.List(ctx, runs, opts...)
	if err != nil {
		return reconcile.Result{}, err
	}

	allRuns := make([]string, 0)

	for _, run := range runs.Items {
		allRuns = append(allRuns, run.GetKmakeRunName())
	}

	// look at the kmakerun items
	for _, element := range instance.Spec.Monitor {
		runs := &bythepowerofv1.KmakeRunList{}
		opts := []client.ListOption{
			client.InNamespace(req.NamespacedName.Namespace),
			client.MatchingLabels{bythepowerofv1.ScheduleLabel.String(): element},
		}

		err = r.List(ctx, runs, opts...)
		if err != nil {
			return reconcile.Result{}, err
		}

		for _, run := range runs.Items {
			kmakeName := bythepowerofv1.GetDomainLabel(run.Labels, bythepowerofv1.KmakeLabel)
			if kmakeName != "" {
				found := false

				for _, i := range allRuns {
					if i == run.GetName() {
						found = true
						break
					}
				}

				if !found {
					kmsr := &bythepowerofv1.KmakeScheduleRun{
						ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, bythepowerofv1.ScheduleRun),
						Spec: bythepowerofv1.KmakeScheduleRunSpec{
							KmakeScheduleRunOperation: bythepowerofv1.KmakeScheduleRunOperation{
								Start: &bythepowerofv1.KmakeScheduleRunStart{},
							},
						},
					}
					ctrl.SetControllerReference(instance, kmsr, r.Scheme)
					SetOwnerReference(&run, kmsr, r.Scheme)

					kmsr.SetLabels(map[string]string{
						bythepowerofv1.KmakeLabel.String():       kmakeName,
						bythepowerofv1.ScheduleLabel.String():    instance.Name,
						bythepowerofv1.ScheduleEnvLabel.String(): currentenvmap.GetName(),
						bythepowerofv1.RunLabel.String():         run.GetName(),
						bythepowerofv1.WorkloadLabel.String():    "yes",
						bythepowerofv1.StatusLabel.String():      "Provision",
					})

					err = r.Create(ctx, kmsr)
					if err != nil {
						return reconcile.Result{}, err
					}
					err = r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.Runs, kmsr.GetName())
					if err != nil {
						return reconcile.Result{}, err
					}
					// return requeue, nil
					allRuns = append(allRuns, run.GetName())
				}
			} else {
				log.Info(fmt.Sprintf("run %v not connected to kmake", instance.GetName()))
			}
		}
	}

	_ = r.Event(instance, bythepowerofv1.Ready, bythepowerofv1.Main, "")
	return backoff5, nil
}

func (r *KmakeNowSchedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	kmsrOwnerKey := ".metadata.controller"
	apiGVStr := bythepowerofv1.GroupVersion.String()

	if err := mgr.GetFieldIndexer().IndexField(&bythepowerofv1.KmakeScheduleRun{}, kmsrOwnerKey, func(rawObj runtime.Object) []string {
		// grab the run object, extract the owner...
		kmsr := rawObj.(*bythepowerofv1.KmakeScheduleRun)
		owner := metav1.GetControllerOf(kmsr)
		if owner == nil {
			return nil
		}
		// ...make sure it's a now scheduler...
		if owner.APIVersion != apiGVStr || owner.Kind != "KmakeNowScheduler" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.KmakeNowScheduler{}).
		Owns(&bythepowerofv1.KmakeScheduleRun{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
