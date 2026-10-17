// Command manager runs the ShortLink operator (lesson 15.4).
package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortlink-operator/api/v1alpha1"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortlink-operator/internal/controller"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortlink-operator/internal/shortener"
)

// [start main]

func main() {
	var (
		apiURL         = flag.String("shortener-url", "http://shortener", "base URL of the shortener API")
		metricsAddr    = flag.String("metrics-bind-address", ":8080", "address of the metrics endpoint")
		probeAddr      = flag.String("health-probe-bind-address", ":8081", "address of the health endpoints")
		leaderElection = flag.Bool("leader-elect", false, "run as one of several replicas, with only the leader active")
	)
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	log := ctrl.Log.WithName("setup")

	// The scheme maps Go types to API kinds: the built-in ones, and ours.
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		log.Error(err, "scheme")
		os.Exit(1)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		log.Error(err, "scheme")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: *metricsAddr},
		HealthProbeBindAddress: *probeAddr,
		LeaderElection:         *leaderElection,
		LeaderElectionID:       "shortlink-operator.shortener.example.com",
		// Give up the lease when stopping, so a standby replica takes over
		// at once instead of waiting for the lease to expire (15s). Safe
		// because the process exits right after the manager stops.
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		log.Error(err, "create manager")
		os.Exit(1)
	}

	r := &controller.ShortLinkReconciler{Client: mgr.GetClient(), API: shortener.New(*apiURL)}
	if err := r.SetupWithManager(mgr); err != nil {
		log.Error(err, "set up controller")
		os.Exit(1)
	}
	mgr.AddHealthzCheck("healthz", healthz.Ping)
	mgr.AddReadyzCheck("readyz", healthz.Ping)

	log.Info("starting manager", "shortener", *apiURL)
	// SetupSignalHandler cancels on SIGTERM; Start returns after the
	// controllers have stopped and in-flight reconciles have finished.
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "manager stopped")
		os.Exit(1)
	}
}

// [end main]
