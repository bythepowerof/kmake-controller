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

package main

import (
	"github.com/namsral/flag"
	"os"
	"strings"

	bythepowerofv1 "github.com/bythepowerof/kmake-controller/api/v1"
	"github.com/bythepowerof/kmake-controller/controllers"
	"github.com/bythepowerof/kmake-controller/logrusr"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = bythepowerofv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var enablePrettyPrint bool
	var namespace string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enablePrettyPrint, "enable-pretty-print", false,
		"Enable pretty print JSON logging")
	flag.StringVar(&namespace, "namespace", "all",
		"Namespace to watch - use 'all' for all namespaces")

	flag.Parse()

	// Create new "foo" logger that's enabled and has a verbosity level of 1.
	l := logrus.New()

	// PrettyPrint should be false/not set
	l.SetFormatter(&logrus.JSONFormatter{PrettyPrint: enablePrettyPrint})

	logger := logrusr.New("kmake", *l)

	ctrl.SetLogger(logger)

	options := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	}

	setupLog.Info("watching", "namespace", namespace)

	if strings.ToLower(namespace) != "all" {
		options.Namespace = namespace
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.KmakeReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("kmake").WithName(namespace),
		Recorder: mgr.GetEventRecorderFor("kmake-controller"),
		Scheme:   scheme,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Kmake")
		os.Exit(1)
	}

	if err = (&controllers.KmakeNowSchedulerReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("KmakeNowScheduler").WithName(namespace),
		Recorder: mgr.GetEventRecorderFor("kmake-now-scheduler-controller"),
		Scheme:   scheme,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KmakeNowScheduler")
		os.Exit(1)
	}
	if err = (&controllers.KmakeScheduleRunReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("KmakeScheduleRun").WithName(namespace),
		Recorder: mgr.GetEventRecorderFor("kmake-schedule-run-controller"),
		Scheme:   scheme,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KmakeScheduleRun")
		os.Exit(1)
	}
	if err = (&controllers.KmakeRunReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("KmakeRun").WithName(namespace),
		Recorder: mgr.GetEventRecorderFor("kmake-run-controller"),
		Scheme:   scheme,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KmakeRun")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
