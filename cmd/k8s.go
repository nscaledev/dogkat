package cmd

import (
	"github.com/drew-viles/k8s-e2e-tester/resources"
	promclient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	promscheme "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/scheme"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioscheme "istio.io/client-go/pkg/clientset/versioned/scheme"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	ClusterName string
	clientsets  *resources.ClientSets
)

func connectToKubernetes(kubeconfig string) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientsets = &resources.ClientSets{}
	// create the clientsets
	clientsets.K8S, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientsets.Prometheus, err = promclient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientsets.Istio, err = istioclient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	prepareClient()
}

func prepareClient() {
	promscheme.AddToScheme(scheme.Scheme)
	istioscheme.AddToScheme(scheme.Scheme)
}