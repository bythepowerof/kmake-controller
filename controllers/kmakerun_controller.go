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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

func (r *KmakeRunReconciler) Event(instance *bythepowerofv1.KmakeRun, phase bythepowerofv1.Phase, subresource bythepowerofv1.SubResource, name string) error {
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

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakeruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakeruns/status,verbs=get;update;patch

func (r *KmakeRunReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kmakerun", req.NamespacedName)

	// requeue := ctrl.Result{Requeue: true}
	backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(1 * time.Minute))}

	log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer log.Info(fmt.Sprintf("Finish reconcile loop for %v", req.NamespacedName))

	instance := &bythepowerofv1.KmakeRun{}
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

	if !instance.HasFinalizer(bythepowerofv1.KmakeRunFinalizerName) {
		err = r.addFinalizer(instance)
		if err != nil {
			r.Event(instance, bythepowerofv1.Error, bythepowerofv1.Main, "finalizer")

			return reconcile.Result{}, fmt.Errorf("error when handling kmakerun finalizer: %v", err)
		}
		r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.Main, "finalizer")

		return ctrl.Result{}, nil
	}

	kmakename := instance.GetKmakeName()
	kmake := &bythepowerofv1.Kmake{}

	log.Info(fmt.Sprintf("Checking kmake %v", kmakename))
	err = r.Get(ctx, types.NamespacedName{
		Namespace: instance.GetNamespace(),
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

	// just add in the kmake as an owner - leave any other owners alone
	if instance.OwnerReferences == nil {
		ctrl.SetControllerReference(kmake, instance, r.Scheme)

		r.Event(instance, bythepowerofv1.Update, bythepowerofv1.KMAKE, kmakename)

		err = r.Update(ctx, instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	for _, owner := range instance.OwnerReferences {
		if owner.Kind == "Kmake" {
			return ctrl.Result{}, nil
		}
	}
	ctrl.SetControllerReference(kmake, instance, r.Scheme)
	r.Event(instance, bythepowerofv1.Update, bythepowerofv1.KMAKE, kmakename)

	err = r.Update(ctx, instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *KmakeRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.KmakeRun{}).
		Complete(r)
}
