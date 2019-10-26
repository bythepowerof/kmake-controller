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
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	// "k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// ctrl "sigs.k8s.io/controller-runtime"
	// "sigs.k8s.io/controller-runtime/pkg/client"
	corev1 "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/kubernetes/pkg/api"
	// "k8s.io/apimachinery/pkg/types"
	// "k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
)

// KmakeReconciler reconciles a Kmake object
type KmakeReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *KmakeReconciler) Event(instance *bythepowerofv1.Kmake, reason string, message string, a ...interface{}) {
	m := fmt.Sprintf(message, a...)
	r.Recorder.Event(instance, "Normal", reason, m)

	log := r.Log.WithValues("kmake", instance.GetName())
	log.Info(m)

	if instance.Status.Status != m {
		instance.Status.Status = m
		r.Status().Update(context.Background(), instance)
	}
}

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakes/status,verbs=get;update;patch

func (r *KmakeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kmake", req.NamespacedName)

	requeue := ctrl.Result{Requeue: true}
	backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(5 * time.Minute))}

	// your logic here
	instance := &bythepowerofv1.Kmake{}
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
		r.Event(instance, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	// PVC
	currentpvc := &corev1.PersistentVolumeClaim{}
	requiredpvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "pvc"),

		Spec: instance.Spec.PersistentVolumeClaimTemplate,
	}

	log.Info(fmt.Sprintf("Checking pvc %v", NameConcat(instance, "pvc")))

	err = r.Get(ctx, NamespacedNameConcat(instance, "pvc"), currentpvc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found pvc %v", NameConcat(instance, "pvc")))

			// create it
			err = r.Create(ctx, requiredpvc)
			if err != nil {
				return reconcile.Result{}, err
			}
			log.Info(fmt.Sprintf("Created pvc %v", NameConcat(instance, "pvc")))

			r.Event(instance, "ProvisionPVC", "Created pvc %v", NameConcat(instance, "pvc"))

			return requeue, err

		}
		return reconcile.Result{}, err
	}
	if !(reflect.DeepEqual(currentpvc.Spec.Resources, requiredpvc.Spec.Resources) &&
		reflect.DeepEqual(currentpvc.ObjectMeta.Labels, requiredpvc.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("delete/recreate pvc %v", NameConcat(instance, "pvc")))

		// You don't seem to be able to update pvcs - even the labels so recreate
		// the pvc will not relase if other jobs are in flight
		// Recreate in next reconcile
		// currentpvc.ObjectMeta.Labels = requiredpvc.ObjectMeta.Labels
		// currentpvc.Spec = requiredpvc.Spec
		err = r.Delete(ctx, currentpvc)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Event(instance, "DeletePVC", "Deleted pvc %v", NameConcat(instance, "pvc"))
		return requeue, nil
	}

	if currentpvc.Status.Phase != corev1.ClaimBound {
		r.Event(instance, "BackofPV", "Backoff PV for %v - %v", NameConcat(instance, "pvc"), string(currentpvc.Status.Phase))
		return backoff5, nil
	}

	if strings.Contains(instance.Status.Status, "Backoff PV") {
		// so we need to rerun the master job to copy the files and makefile again
		r.Event(instance, "ProvisionedPVC", "Provisioned pvc for %v -%v", NameConcat(instance, "pvc"), string(currentpvc.Status.Phase))
	}
	// env configmap

	currentenvmap := &corev1.ConfigMap{}
	requiredenvmap := &corev1.ConfigMap{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "env"),

		Data: instance.Spec.Variables,
	}

	log.Info(fmt.Sprintf("Checking env map %v", NameConcat(instance, "env")))

	err = r.Get(ctx, NamespacedNameConcat(instance, "env"), currentenvmap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found env map %v", NameConcat(instance, "env")))

			// create it
			err = r.Create(ctx, requiredenvmap)
			if err != nil {
				return reconcile.Result{}, err
			}
			r.Event(instance, "CreatedEnvMap", "Created env map %v", NameConcat(instance, "env"))
			return requeue, err

		}
		return reconcile.Result{}, err
	}
	if !(reflect.DeepEqual(currentenvmap.Data, requiredenvmap.Data) &&
		reflect.DeepEqual(currentenvmap.ObjectMeta.Labels, requiredenvmap.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("modify env map %v", NameConcat(instance, "env")))
		currentenvmap.ObjectMeta.Labels = requiredenvmap.ObjectMeta.Labels
		currentenvmap.Data = requiredenvmap.Data
		err = r.Update(ctx, currentenvmap)
		if err != nil {
			return reconcile.Result{}, err
		}

		r.Event(instance, "UpdatedEnvMap", "Updated %v", NameConcat(instance, "env"))
		return requeue, nil
	}

	return requeue, nil

	// return ctrl.Result{}, nil
}

func (r *KmakeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.Kmake{}).
		Complete(r)
}
