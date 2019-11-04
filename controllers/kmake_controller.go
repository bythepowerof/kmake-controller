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
	"gopkg.in/yaml.v2"
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

func (r *KmakeReconciler) UpdateSubResource(status *bythepowerofv1.KmakeStatus, subresource SubResource, name string) {
	if name == "" {
		return
	}
	switch subresource {
	case PVC:
		status.Resources.Pvc = name
	case EnvMap:
		status.Resources.Env = name
	case KmakeMap:
		status.Resources.Kmake = name
	}
}

func (r *KmakeReconciler) Event(instance *bythepowerofv1.Kmake, phase Phase, subresource SubResource, name string) {
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
		r.UpdateSubResource(&instance.Status, subresource, name)
		r.Status().Update(context.Background(), instance)
	}
}

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakes/status,verbs=get;update;patch

func (r *KmakeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kmake", req.NamespacedName)

	requeue := ctrl.Result{Requeue: true}
	backoff5 := ctrl.Result{RequeueAfter: time.Until(time.Now().Add(1 * time.Minute))}

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
		r.Event(instance, Delete, Main, "")
		return ctrl.Result{}, nil
	}

	// PVC
	currentpvc := &corev1.PersistentVolumeClaim{}
	requiredpvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "pvc"),

		Spec: instance.Spec.PersistentVolumeClaimTemplate,
	}

	log.Info(fmt.Sprintf("Checking pvc %v", NameConcat(instance.Status, PVC)))

	err = r.Get(ctx, NamespacedNameConcat(instance, PVC), currentpvc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found pvc %v", NameConcat(instance.Status, PVC)))

			// create it
			err = r.Create(ctx, requiredpvc)
			if err != nil {
				return reconcile.Result{}, err
			}
			log.Info(fmt.Sprintf("Created pvc %v", requiredpvc.ObjectMeta.Name))

			r.Event(instance, Provision, PVC, requiredpvc.ObjectMeta.Name)

			return requeue, err

		}
		return reconcile.Result{}, err
	}
	if !(reflect.DeepEqual(currentpvc.Spec.Resources, requiredpvc.Spec.Resources) &&
		reflect.DeepEqual(currentpvc.ObjectMeta.Labels, requiredpvc.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("delete/recreate pvc %v", NameConcat(instance.Status, PVC)))

		// You don't seem to be able to update pvcs - even the labels so recreate
		// the pvc will not relase if other jobs are in flight
		// Recreate in next reconcile
		// currentpvc.ObjectMeta.Labels = requiredpvc.ObjectMeta.Labels
		// currentpvc.Spec = requiredpvc.Spec
		err = r.Delete(ctx, currentpvc)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Event(instance, Delete, PVC, "")
		return requeue, nil
	}

	if currentpvc.Status.Phase != corev1.ClaimBound {
		r.Event(instance, BackOff, PVC, currentpvc.ObjectMeta.Name)
		return backoff5, nil
	}

	if strings.Contains(instance.Status.Status, "BackOff PV") {
		// so we need to rerun the master job to copy the files and makefile again
		r.Event(instance, Provision, PVC, currentpvc.ObjectMeta.Name)
	}
	// env configmap

	currentenvmap := &corev1.ConfigMap{}
	requiredenvmap := &corev1.ConfigMap{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "env"),

		Data: instance.Spec.Variables,
	}

	log.Info(fmt.Sprintf("Checking env map %v", NameConcat(instance.Status, EnvMap)))

	err = r.Get(ctx, NamespacedNameConcat(instance, EnvMap), currentenvmap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found env map %v", NameConcat(instance.Status, EnvMap)))

			// create it
			err = r.Create(ctx, requiredenvmap)
			if err != nil {
				return reconcile.Result{}, err
			}
			r.Event(instance, Provision, EnvMap, requiredenvmap.ObjectMeta.Name)
			return requeue, err

		}
		return reconcile.Result{}, err
	}
	if !(reflect.DeepEqual(currentenvmap.Data, requiredenvmap.Data) &&
		reflect.DeepEqual(currentenvmap.ObjectMeta.Labels, requiredenvmap.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("modify env map %v", NameConcat(instance.Status, EnvMap)))
		currentenvmap.ObjectMeta.Labels = requiredenvmap.ObjectMeta.Labels
		currentenvmap.Data = requiredenvmap.Data
		err = r.Update(ctx, currentenvmap)
		if err != nil {
			return reconcile.Result{}, err
		}

		r.Event(instance, Update, EnvMap, currentenvmap.ObjectMeta.Name)
		return requeue, nil
	}

	// make yaml config map

	y, err := yaml.Marshal(instance.Spec.Rules)
	m, err := ToMakefile(instance.Spec.Rules)

	currentkmakemap := &corev1.ConfigMap{}
	requiredkmakemap := &corev1.ConfigMap{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "kmake"),
		Data: map[string]string{"kmake.yaml": string(y),
			"kmake.mk": m},
	}

	log.Info(fmt.Sprintf("Checking kmake map %v", NameConcat(instance.Status, KmakeMap)))

	err = r.Get(ctx, NamespacedNameConcat(instance, KmakeMap), currentkmakemap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found kmake map %v", NameConcat(instance.Status, KmakeMap)))

			// create it
			err = r.Create(ctx, requiredkmakemap)
			if err != nil {
				return reconcile.Result{}, err
			}
			r.Event(instance, Provision, KmakeMap, requiredkmakemap.ObjectMeta.Name)
			return requeue, err

		}
		return reconcile.Result{}, err
	}
	if !(reflect.DeepEqual(currentkmakemap.Data, requiredkmakemap.Data) &&
		reflect.DeepEqual(currentkmakemap.ObjectMeta.Labels, requiredkmakemap.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("modify kmake map %v", NameConcat(instance.Status, KmakeMap)))
		currentkmakemap.ObjectMeta.Labels = requiredkmakemap.ObjectMeta.Labels
		currentkmakemap.Data = requiredkmakemap.Data
		err = r.Update(ctx, currentkmakemap)
		if err != nil {
			return reconcile.Result{}, err
		}

		r.Event(instance, Update, KmakeMap, currentkmakemap.ObjectMeta.Name)
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
