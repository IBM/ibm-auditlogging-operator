//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/IBM/ibm-auditlogging-operator/controllers/k8sutil"

	"github.com/IBM/ibm-auditlogging-operator/controllers/constant"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	certmgr "github.com/ibm/ibm-cert-manager-operator/apis/certmanager/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	cache1 "github.com/IBM/controller-filtered-cache/filteredcache"
	operatorv1 "github.com/IBM/ibm-auditlogging-operator/api/v1"
	operatorv1alpha1 "github.com/IBM/ibm-auditlogging-operator/api/v1alpha1"
	"github.com/IBM/ibm-auditlogging-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))
	utilruntime.Must(certmgr.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	gvkLabelMap := map[schema.GroupVersionKind]cache1.Selector{
		corev1.SchemeGroupVersion.WithKind("Secret"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		corev1.SchemeGroupVersion.WithKind("ConfigMap"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		corev1.SchemeGroupVersion.WithKind("Service"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		corev1.SchemeGroupVersion.WithKind("ServiceAccount"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		appsv1.SchemeGroupVersion.WithKind("Deployment"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		appsv1.SchemeGroupVersion.WithKind("DaemonSet"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		rbacv1.SchemeGroupVersion.WithKind("Role"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		rbacv1.SchemeGroupVersion.WithKind("RoleBinding"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		batchv1.SchemeGroupVersion.WithKind("Job"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		certmgr.SchemeBuilder.GroupVersion.WithKind("Certificate"): {
			LabelSelector: constant.AuditTypeLabel,
		},
		certmgr.SchemeBuilder.GroupVersion.WithKind("Issuer"): {
			LabelSelector: constant.AuditTypeLabel,
		},
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces")
	}

	options := ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "d9301293.ibm.com",
	}
	//type CustomMetricsService struct{}
	//metricsService := &CustomMetricsService{}

	// ServeMetrics serves metrics using a custom HTTP server.

	//  	func (c *CustomMetricsService) ServeMetrics(metricsPath string, addr string, port int) error {
	// 	http.Handle(metricsPath, promhttp.Handler())
	// 	listenAddress := fmt.Sprintf("%s:%d", addr, port)
	// 	return http.ListenAndServe(listenAddress, nil)
	//    }

	//    // Start starts the metrics service.
	//    func (c *CustomMetricsService) Start(stopCh <-chan struct{}) error {
	// 	// Implement any custom start logic if needed
	// 	return nil
	//    }

	//   // Shutdown shuts down the metrics service.
	//   func (c *CustomMetricsService) Shutdown() {
	// 	// Implement any custom shutdown logic if needed
	//    }

	// // Register metrics
	// if err := metrics.Register(); err != nil {
	// 	panic(fmt.Sprintf("unable to register metrics: %v", err))
	// }

	// metricsService := &webhook.Service{}

	// options := ctrl.Options{
	// 	Scheme:             scheme,
	// 	MetricsBindAddress: metricsAddr,
	// 	Port:               9443,
	// 	LeaderElection:     enableLeaderElection,
	// 	LeaderElectionID:   "d9301293.ibm.com",
	// }

	if watchNamespace != "" {
		options.NewCache = k8sutil.NewAuditCache(strings.Split(watchNamespace, ","))
	} else {
		options.NewCache = cache1.NewFilteredCacheBuilder(gvkLabelMap)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	// mgr.Add(metricsAddr)
	// mgr.Add(ports)
	//mgr.Add(metricsService)
	if err = (&controllers.CommonAuditReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("CommonAudit"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("ibm-auditlogging-operator"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CommonAudit")
		os.Exit(1)
	}
	if err = (&controllers.AuditLoggingReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("AuditLogging"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("ibm-auditlogging-operator"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AuditLogging")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
