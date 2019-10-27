package cli

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/golang/glog"
	cli "github.com/jawher/mow.cli"
	freenasProvisioner "github.com/travisghansen/freenas-iscsi-provisioner/provisioner"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

const (
	exponentialBackOffOnError = false
	failedRetryThreshold      = 5
	leasePeriod               = controller.DefaultLeaseDuration
	retryPeriod               = controller.DefaultRetryPeriod
	renewDeadline             = controller.DefaultRenewDeadline
)

var (
	// cli parameters
	kubeconfig      *string
	identifier      *string
	provisionerName *string

	// controller tweaks
	controllerThreadiness                 *int
	controllerCreateProvisionedPVInterval *int
	controllerLeaseDuration               *int
	controllerRenewDeadline               *int
	controllerRetryPeriod                 *int
	controllerMetricsPort                 *int
)

// Process all command line parameters
func Process(appName, appDesc, appVersion string) {
	syscall.Umask(0)
	flag.Set("logtostderr", "true")

	app := cli.App(appName, appDesc)
	app.Version("v version", fmt.Sprintf("%s version %s", appName, appVersion))

	kubeconfig = app.String(cli.StringOpt{
		Name:   "kubeconfig",
		Desc:   "Path to kubernetes configuration file (for out of cluster execution)",
		EnvVar: "KUBECONFIG",
	})
	identifier = app.String(cli.StringOpt{
		Name:   "i identifier",
		Value:  "freenas-iscsi-provisioner",
		Desc:   "Provisioner identifier (e.g. if unsure set it to current node name)",
		EnvVar: "IDENTIFIER",
	})
	provisionerName = app.String(cli.StringOpt{
		Name:   "provisioner-name",
		Value:  "freenas.org/iscsi",
		Desc:   "Provisioner Name (e.g. 'provisioner' attribute of storage-class)",
		EnvVar: "PROVISIONER_NAME",
	})

	controllerThreadiness = app.Int(cli.IntOpt{
		Name:   "controller-threadiness",
		Value:  4,
		Desc:   "Number of controller threads to handle provisioner tasks",
		EnvVar: "CONTROLLER_THREADINESS",
	})

	controllerCreateProvisionedPVInterval = app.Int(cli.IntOpt{
		Name:   "controller-create-provisioned-pv-interval",
		Value:  10,
		Desc:   "controller create provisioned pv interval",
		EnvVar: "CONTROLLER_CREATE_PROVISIONED_PV_INTERVAL",
	})

	controllerLeaseDuration = app.Int(cli.IntOpt{
		Name:   "controller-lease-duration",
		Value:  15,
		Desc:   "controller lease duration",
		EnvVar: "CONTROLLER_LEASE_DURATION",
	})

	controllerRenewDeadline = app.Int(cli.IntOpt{
		Name:   "controller-renew-deadline",
		Value:  10,
		Desc:   "controller renew deadline",
		EnvVar: "CONTROLLER_RENEW_DEADLINE",
	})

	controllerRetryPeriod = app.Int(cli.IntOpt{
		Name:   "controller-retry-period",
		Value:  2,
		Desc:   "controller retry period",
		EnvVar: "CONTROLLER_RETRY_PERIOD",
	})

	controllerMetricsPort = app.Int(cli.IntOpt{
		Name:   "controller-metrics-port",
		Value:  0,
		Desc:   "port to use for prometheus metrics",
		EnvVar: "CONTROLLER_METRICS_PORT",
	})

	app.Action = execute
	app.Run(os.Args)
}

func execute() {
	var err error
	var config *rest.Config

	/* Params checking */
	var msgs []string
	if *identifier == "" {
		msgs = append(msgs, "Identifier parameter must be specified")
	}

	// Print all parameters' error and exist if need be
	if len(msgs) > 0 {
		fmt.Fprintf(os.Stderr, "The following error(s) occured:\n")
		for _, m := range msgs {
			fmt.Fprintf(os.Stderr, "  - %s\n", m)
		}
		os.Exit(1)
	}
	/* End params checking */

	if *kubeconfig != "" {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		// Create an InClusterConfig and use it to create a client for the controller
		// to use to communicate with Kubernetes
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting server version: %v", err)
	}

	clientFreenasProvisioner := freenasProvisioner.New(
		clientset,
		*identifier,
	)

	pc := controller.NewProvisionController(
		clientset,
		*provisionerName,
		clientFreenasProvisioner,
		serverVersion.GitVersion,
		controller.Threadiness(*controllerThreadiness),
		controller.CreateProvisionedPVInterval(time.Duration(*controllerCreateProvisionedPVInterval)*time.Second),
		controller.LeaseDuration(time.Duration(*controllerLeaseDuration)*time.Second),
		controller.RenewDeadline(time.Duration(*controllerRenewDeadline)*time.Second),
		controller.RetryPeriod(time.Duration(*controllerRetryPeriod)*time.Second),
		controller.MetricsPort(int32(*controllerMetricsPort)),
	)

	pc.Run(wait.NeverStop)
}
