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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	// "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// "k8s.io/client-go/tools/record"
	// ctrl "sigs.k8s.io/controller-runtime"
	// "sigs.k8s.io/controller-runtime/pkg/client"
	corev1 "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/kubernetes/pkg/api"
	// "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
)

// KmakeReconciler reconciles a Kmake object
type KmakeReconciler struct {
	client.Client
	Log logr.Logger

	// Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bythepowerof.github.com,resources=kmakes/status,verbs=get;update;patch

func (r *KmakeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("kmake", req.NamespacedName)

	// your logic here
	instance := &bythepowerofv1.Kmake{}
	err := r.Get(context.Background(), req.NamespacedName, instance)

	r.Log.Info(fmt.Sprintf("Starting reconcile loop for %v", req.NamespacedName))
	defer r.Log.Info(fmt.Sprintf("Finish reconcile loop for %v", req.NamespacedName))

	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// if instance.IsBeingDeleted() {
	// 	err := r.handleFinalizer(instance)
	// 	if err != nil {
	// 		return reconcile.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
	// 	}
	// 	// r.Recorder.Event(instance, "Normal", "Deleted", "Object finalizer is deleted")
	// 	return ctrl.Result{}, nil
	// }

	// if !instance.HasFinalizer(bythepowerofv1.KmakeFinalizerName) {
	// 	err = r.addFinalizer(instance)
	// 	if err != nil {
	// 		return reconcile.Result{}, fmt.Errorf("error when handling kmakefinalizer: %v", err)
	// 	}
	// 	// r.Recorder.Event(instance, "Normal", "Added", "Object finalizer is added")
	// 	return ctrl.Result{}, nil
	// }

	currentpvc := &corev1.PersistentVolumeClaim{}
	requiredpvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: ObjectMetaConcat(instance, req.NamespacedName, "pvc"),
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:  NameConcat(req.NamespacedName, "pvc"),
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(instance.Spec.StorageSize),
				},
			},
		},
	}

	r.Log.Info(fmt.Sprintf("Checking pvc %v", NameConcat(req.NamespacedName, "pvc")))

	err = r.Get(context.Background(), NamespacedNameConcat(req.NamespacedName, "pvc"), currentpvc)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("Not found pvc %v", NameConcat(req.NamespacedName, "pvc")))

			// create it
			err = r.Create(context.Background(), requiredpvc)
			if err != nil {
				return reconcile.Result{}, err
			}
			r.Log.Info(fmt.Sprintf("Created pvc %v", NameConcat(req.NamespacedName, "pvc")))

		}
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *KmakeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bythepowerofv1.Kmake{}).
		Complete(r)
}
