/*
Copyright 2019 microsoft.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
)

func (r *KmakeNowSchedulerReconciler) addFinalizer(instance *bythepowerofv1.KmakeNowScheduler) error {
	instance.AddFinalizer(bythepowerofv1.KmakeNowSchedulerFinalizerName)
	err := r.Update(context.Background(), instance)
	if err != nil {
		return fmt.Errorf("failed to update kmake now scheduler finalizer: %v", err)
	}
	return nil
}

func (r *KmakeNowSchedulerReconciler) handleFinalizer(instance *bythepowerofv1.KmakeNowScheduler) error {
	if instance.HasFinalizer(bythepowerofv1.KmakeNowSchedulerFinalizerName) {
		// // remove all kmake runs owned by us
		del := &bythepowerofv1.KmakeScheduleRun{}

		do := &client.DeleteAllOfOptions{}
		do.ApplyOptions([]client.DeleteAllOfOption{
			client.InNamespace(instance.Namespace)})
		labels := client.MatchingLabels{}
		labels = bythepowerofv1.SetDomainLabel(labels, bythepowerofv1.ScheduleInstLabel, instance.Name)

		policy := metav1.DeletePropagationBackground
		o := &client.DeleteAllOfOptions{DeleteOptions: client.DeleteOptions{PropagationPolicy: &policy}}

		do.ApplyToDeleteAllOf(o)

		do.ApplyOptions([]client.DeleteAllOfOption{labels})
		if err := r.DeleteAllOf(context.Background(), del, do); err != nil {
			return err
		}
		instance.RemoveFinalizer(bythepowerofv1.KmakeNowSchedulerFinalizerName)
		if err := r.Update(context.Background(), instance); err != nil {
			return err
		}

	}
	// Our finalizer has finished, so the reconciler can do nothing.
	return nil
}
