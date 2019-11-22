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
	"k8s.io/apimachinery/pkg/api/equality"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
)

// KmakeReconciler reconciles a Kmake object
type KmakeReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *KmakeReconciler) Event(instance *bythepowerofv1.Kmake, phase bythepowerofv1.Phase, subresource bythepowerofv1.SubResource, name string) error {
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

func (r *KmakeReconciler) AppendRun(kmake *bythepowerofv1.Kmake, run *bythepowerofv1.KmakeRun) {
	kmake.Status.Runs = append(kmake.Status.Runs, &bythepowerofv1.KmakeRuns{
		RunName:  run.GetName(),
		RunPhase: bythepowerofv1.Wait.String(),
	})
	r.Status().Update(context.Background(), kmake)
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
		err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.Main, "")
		if err != nil {
			return reconcile.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// PVC
	currentpvc := &corev1.PersistentVolumeClaim{}
	requiredpvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "pvc", "Kmake"),

		Spec: instance.Spec.PersistentVolumeClaimTemplate,
	}

	log.Info(fmt.Sprintf("Checking pvc %v", instance.Status.NameConcat(bythepowerofv1.PVC)))

	err = r.Get(ctx, instance.NamespacedNameConcat(bythepowerofv1.PVC), currentpvc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found pvc %v", instance.Status.NameConcat(bythepowerofv1.PVC)))

			// create it
			err = r.Create(ctx, requiredpvc)
			if err != nil {
				return reconcile.Result{}, err
			}
			log.Info(fmt.Sprintf("Created pvc %v", requiredpvc.ObjectMeta.Name))

			err = r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.PVC, requiredpvc.ObjectMeta.Name)
			if err != nil {
				return reconcile.Result{}, err
			}
			return requeue, err

		}
		return reconcile.Result{}, err
	}
	if !(equality.Semantic.DeepEqual(currentpvc.Spec.Resources, requiredpvc.Spec.Resources) &&
		equality.Semantic.DeepEqual(currentpvc.ObjectMeta.Labels, requiredpvc.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("delete/recreate pvc %v", instance.Status.NameConcat(bythepowerofv1.PVC)))

		err = r.Delete(ctx, currentpvc)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.PVC, "")
		if err != nil {
			return reconcile.Result{}, err
		}
		return requeue, nil
	}

	if currentpvc.Status.Phase != corev1.ClaimBound {
		err = r.Event(instance, bythepowerofv1.BackOff, bythepowerofv1.PVC, currentpvc.ObjectMeta.Name)
		if err != nil {
			return reconcile.Result{}, err
		}
		return backoff5, nil
	}

	if strings.Contains(instance.Status.Status, "BackOff PV") {
		// so we need to rerun the master job to copy the files and makefile again
		err = r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.PVC, currentpvc.ObjectMeta.Name)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// env configmap

	currentenvmap := &corev1.ConfigMap{}
	requiredenvmap := &corev1.ConfigMap{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "env", "Kmake"),

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

	// make yaml config map

	y, err := yaml.Marshal(map[string][]bythepowerofv1.KmakeRule{"rules": instance.Spec.Rules})
	m, err := instance.Spec.ToMakefile()

	currentkmakemap := &corev1.ConfigMap{}
	requiredkmakemap := &corev1.ConfigMap{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "kmake", "Kmake"),
		Data: map[string]string{"kmake.yaml": string(y),
			"kmake.mk": m},
	}

	log.Info(fmt.Sprintf("Checking kmake map %v", instance.Status.NameConcat(bythepowerofv1.KmakeMap)))

	err = r.Get(ctx, instance.NamespacedNameConcat(bythepowerofv1.KmakeMap), currentkmakemap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Not found kmake map %v", instance.Status.NameConcat(bythepowerofv1.KmakeMap)))

			// create it
			err = r.Create(ctx, requiredkmakemap)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.Event(instance, bythepowerofv1.Provision, bythepowerofv1.KmakeMap, requiredkmakemap.ObjectMeta.Name)
			if err != nil {
				return reconcile.Result{}, err
			}
			return requeue, err
		}
		return reconcile.Result{}, err
	}
	if !(equality.Semantic.DeepEqual(currentkmakemap.Data, requiredkmakemap.Data) &&
		equality.Semantic.DeepEqual(currentkmakemap.ObjectMeta.Labels, requiredkmakemap.ObjectMeta.Labels)) {
		log.Info(fmt.Sprintf("delete kmake map %v", instance.Status.NameConcat(bythepowerofv1.KmakeMap)))
		err = r.Delete(ctx, currentkmakemap)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.Event(instance, bythepowerofv1.Delete, bythepowerofv1.KmakeMap, "")
		if err != nil {
			return reconcile.Result{}, err
		}
		return requeue, nil
	}
	// return requeue, nil
	return ctrl.Result{}, nil
}

func (r *KmakeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.Kmake{}).
		Complete(r)
}
