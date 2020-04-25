// +kubebuilder:object:generate=false
package gql

import (
	context "context"
	"fmt"
	"github.com/bythepowerof/kmake-controller/api/v1"
	"os"
	"strings"
	"sync"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type KmakeListener struct {
	client     client.Client
	manager    manager.Manager
	mutex      sync.Mutex
	changes    map[string]map[int]chan KmakeObject
	index      int
	namespace  string
	ownManager bool
}

func NewKmakeListener(namespace string, mgr manager.Manager) *KmakeListener {

	ownManager := false
	if mgr == nil {
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		_ = bythepowerofv1.AddToScheme(scheme)

		var err error

		// facilitate testing by passing manager in
		mo := manager.Options{Scheme: scheme, MetricsBindAddress: "0"}
		if strings.ToLower(namespace) != "all" {
			mo.Namespace = namespace
		}
		mgr, err = manager.New(ctrl.GetConfigOrDie(), mo)

		if err != nil {
			fmt.Println("failed to create manager")
			os.Exit(1)
		}
		ownManager = true
	}

	return &KmakeListener{
		client:     mgr.GetClient(),
		manager:    mgr,
		mutex:      sync.Mutex{},
		changes:    map[string]map[int]chan KmakeObject{},
		namespace:  namespace,
		ownManager: ownManager,
	}

}

func (r *KmakeListener) AddChangeClient(ctx context.Context, namespace string) (<-chan KmakeObject, error) {
	if r.namespace != "all" && r.namespace != namespace {
		return nil, fmt.Errorf("namespace %s not supported", namespace)
	}

	kmo := make(chan KmakeObject, 1)
	r.mutex.Lock()
	r.index++

	if _, ok := r.changes[namespace]; !ok {
		r.changes[namespace] = make(map[int]chan KmakeObject)
	}

	r.changes[namespace][r.index] = kmo
	r.mutex.Unlock()

	// Delete channel when done
	go func() {
		<-ctx.Done()
		r.mutex.Lock()
		delete(r.changes[r.namespace], r.index)
		r.mutex.Unlock()
	}()
	return kmo, nil
}

func (r *KmakeListener) KmakeChanges(namespace string) error {
	// Create a new Controller that will call the provided Reconciler function in response
	// to events.

	if r.namespace != "all" && r.namespace != namespace {
		return fmt.Errorf("namespace %q not supported", namespace)
	}

	err := r.prepareKmakeWatch()
	if err != nil {
		panic(err)
	}

	err = r.prepareKmakeRunWatch()
	if err != nil {
		panic(err)
	}

	err = r.prepareKmakeScheduleRunWatch()
	if err != nil {
		panic(err)
	}

	err = r.prepareKmakeNowSchedulerWatch()
	if err != nil {
		panic(err)
	}

	// Start the Controllers through the manager.
	if r.ownManager {
		go func() {
			if err := r.manager.Start(signals.SetupSignalHandler()); err != nil {
				panic(err)
			}
		}()
	}

	return nil
}

func (r *KmakeListener) prepareKmakeWatch() error {
	c, err := controller.New("kmake-watch", r.manager, controller.Options{
		Reconciler: reconcile.Func(r.watchKmake),
	})
	if err != nil {
		return err
	}
	// Watch for kmake objects create / update / delete events and call Reconcile
	return c.Watch(&source.Kind{Type: &v1.Kmake{}}, &handler.EnqueueRequestForObject{})
}

func (r *KmakeListener) watchKmake(o reconcile.Request) (reconcile.Result, error) {
	// Your business logic to implement the API by creating, updating, deleting objects goes here.
	ret := &v1.Kmake{}

	err := r.client.Get(context.Background(), o.NamespacedName, ret)
	if err != nil {
		return reconcile.Result{}, err
	}
	if ret.IsBeingDeleted() {
		ret.Status.Status = "Deleting"
	}

	// Notify new message
	r.mutex.Lock()
	for _, ch := range r.changes[o.Namespace] {
		ch <- ret
	}
	r.mutex.Unlock()
	return reconcile.Result{}, nil
}

func (r *KmakeListener) prepareKmakeRunWatch() error {
	c, err := controller.New("kmakerun-watch", r.manager, controller.Options{
		Reconciler: reconcile.Func(r.watchKmakeRun),
	})
	if err != nil {
		return err
	}
	// Watch for kmake objects create / update / delete events and call Reconcile
	return c.Watch(&source.Kind{Type: &v1.KmakeRun{}}, &handler.EnqueueRequestForObject{})
}

func (r *KmakeListener) watchKmakeRun(o reconcile.Request) (reconcile.Result, error) {
	// Your business logic to implement the API by creating, updating, deleting objects goes here.
	ret := &v1.KmakeRun{}

	err := r.client.Get(context.Background(), o.NamespacedName, ret)
	if err != nil {
		return reconcile.Result{}, err
	}
	if ret.IsBeingDeleted() {
		ret.Status.Status = "Deleting"
	}

	// Notify new message
	r.mutex.Lock()
	for _, ch := range r.changes[o.Namespace] {
		ch <- ret
	}
	r.mutex.Unlock()
	return reconcile.Result{}, nil
}

func (r *KmakeListener) prepareKmakeNowSchedulerWatch() error {
	c, err := controller.New("kmakenowscheduler-watch", r.manager, controller.Options{
		Reconciler: reconcile.Func(r.watchKmakeNowScheduler),
	})
	if err != nil {
		return err
	}
	// Watch for kmake objects create / update / delete events and call Reconcile
	return c.Watch(&source.Kind{Type: &v1.KmakeNowScheduler{}}, &handler.EnqueueRequestForObject{})
}

func (r *KmakeListener) watchKmakeNowScheduler(o reconcile.Request) (reconcile.Result, error) {
	// Your business logic to implement the API by creating, updating, deleting objects goes here.
	ret := &v1.KmakeNowScheduler{}

	err := r.client.Get(context.Background(), o.NamespacedName, ret)
	if err != nil {
		return reconcile.Result{}, err
	}
	if ret.IsBeingDeleted() {
		ret.Status.Status = "Deleting"
	}

	// Notify new message
	r.mutex.Lock()
	for _, ch := range r.changes[o.Namespace] {
		ch <- ret
	}
	r.mutex.Unlock()
	return reconcile.Result{}, nil
}

func (r *KmakeListener) prepareKmakeScheduleRunWatch() error {
	c, err := controller.New("kmakeschedulerun-watch", r.manager, controller.Options{
		Reconciler: reconcile.Func(r.watchKmakeScheduleRun),
	})
	if err != nil {
		return err
	}
	// Watch for kmake objects create / update / delete events and call Reconcile
	return c.Watch(&source.Kind{Type: &v1.KmakeScheduleRun{}}, &handler.EnqueueRequestForObject{})
}

func (r *KmakeListener) watchKmakeScheduleRun(o reconcile.Request) (reconcile.Result, error) {
	// Your business logic to implement the API by creating, updating, deleting objects goes here.
	ret := &v1.KmakeScheduleRun{}

	err := r.client.Get(context.Background(), o.NamespacedName, ret)
	if err != nil {
		// if errors.IsNotFound(err) {
		// 	return reconcile.Result{}, nil
		// }
		return reconcile.Result{}, err
	}
	if ret.IsBeingDeleted() {
		ret.Status.Status = "Deleting"
	}

	// Notify new message
	r.mutex.Lock()
	for _, ch := range r.changes[o.Namespace] {
		ch <- ret
	}
	r.mutex.Unlock()
	return reconcile.Result{}, nil
}
