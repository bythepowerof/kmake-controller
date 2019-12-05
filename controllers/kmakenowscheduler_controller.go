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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
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
		bytes, err := json.Marshal(instance.Status.Resources)
		if err != nil {
			return err
		}
		instance.Annotations["bythepowerof.github.io/kmake"] = string(bytes)
		return r.Update(context.Background(), instance)
	}
	return nil
}

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakenowschedulers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakenowschedulers/status,verbs=get;update;patch

func (r *KmakeNowSchedulerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("kmakenowscheduler", req.NamespacedName)

	ctx := context.Background()
	log := r.Log.WithValues("kmakenowscheduler", req.NamespacedName)
	requeue := ctrl.Result{Requeue: true}
	backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(1 * time.Minute))}

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
		err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Main, "")
		if err != nil {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	// env configmap

	currentenvmap := &corev1.ConfigMap{}
	requiredenvmap := &corev1.ConfigMap{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "env", "KmakeNowScheduler"),

		Data: instance.Spec.Variables,
	}

	log.Info(fmt.Sprintf("Checking env map %v", instance.Status.NameConcat(bythepowerofv1.EnvMap)))

	err = r.Get(ctx, instance.NamespacedNameConcat(bythepowerofv1.EnvMap), currentenvmap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found env map %v", instance.Status.NameConcat(bythepowerofv1.EnvMap)))

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
		log.Info(fmt.Sprintf("delete env map %v", instance.Status.NameConcat(bythepowerofv1.EnvMap)))
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

	allRuns := append([]bythepowerofv1.KmakeRunManifest(nil), instance.Status.Runs...)
	for j := 0; j < len(allRuns); j++ {
		element := &allRuns[j]
		element.RunPhase = "Deleted"
	}

	// look at thr kmakerun items
	for _, element := range instance.Spec.Monitor {
		runs := &bythepowerofv1.KmakeRunList{}
		opts := []client.ListOption{
			client.InNamespace(req.NamespacedName.Namespace),
			client.MatchingLabels{"bythepowerof.github.io/scheduler": element},
		}

		err = r.List(ctx, runs, opts...)
		if err != nil {
			return reconcile.Result{}, err
		}

		for _, run := range runs.Items {
			if val, ok := run.GetObjectMeta().GetLabels()["bythepowerof.github.io/kmake"]; ok {
				xx := bythepowerofv1.KmakeRunManifest{
					RunName:   run.GetName(),
					RunPhase:  run.Status.Status,
					KmakeName: val,
				}

				found := false

				for j := 0; j < len(allRuns); j++ {
					i := &allRuns[j]
					if i.RunName == xx.RunName {
						// i.RunPhase = xx.RunPhase
						// i.KmakeName = xx.KmakeName
						found = true
						break
					}
				}
				if !found {
					allRuns = append(allRuns, xx)
				}
			} else {
				log.Info(fmt.Sprintf("run %v not connected to kmake", run.Status.NameConcat(bythepowerofv1.Runs)))
			}
		}
	}

	// look at the scheduleruns just for this instance...
	runs := &bythepowerofv1.KmakeScheduleRunList{}
	opts := []client.ListOption{
		client.MatchingLabels{"bythepowerof.github.io/schedule": instance.GetName()},
	}
	err = r.List(ctx, runs, opts...)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, run := range runs.Items {
		xx := bythepowerofv1.KmakeRunManifest{
			RunName:   run.GetKmakeRunName(),
			RunPhase:  run.Status.Status,
			KmakeName: run.GetKmakeName(),
		}
		found := false

		for j := 0; j < len(allRuns); j++ {
			i := &allRuns[j]

			if i.RunName == xx.RunName {
				i.RunPhase = xx.RunPhase
				i.KmakeName = xx.KmakeName
				found = true
				break
			}
		}
		if !found {
			allRuns = append(allRuns, xx)
		}
	}

	if !(equality.Semantic.DeepEqual(allRuns, instance.Status.Runs)) {
		log.Info(fmt.Sprintf("update runs %v", instance.Status.NameConcat(bythepowerofv1.Runs)))

		instance.Status.Runs = allRuns
		instance.Status.Status = ""
		err = r.Event(instance, bythepowerofv1.Update, bythepowerofv1.Runs, instance.GetName())
	}

	return backoff5, nil
}

func (r *KmakeNowSchedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.KmakeNowScheduler{}).
		Complete(r)
}
