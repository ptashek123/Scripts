package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var (
		outputDir     string
		namespace     string
		allNamespaces bool
		excludeSystem bool
		kubeconfig    string
		dryRun        bool
		verbose       bool
	)

	defaultOutput := fmt.Sprintf("./k8s-export-%s", time.Now().Format("20060102-150405"))

	flag.StringVar(&outputDir, "output", defaultOutput, "Output directory")
	flag.StringVar(&outputDir, "o", defaultOutput, "Output directory (shorthand)")
	flag.StringVar(&namespace, "namespace", "", "Export only specified namespace")
	flag.StringVar(&namespace, "n", "", "Export only specified namespace (shorthand)")
	flag.BoolVar(&allNamespaces, "all-namespaces", false, "Export all namespaces including system ones")
	flag.BoolVar(&allNamespaces, "A", false, "Export all namespaces including system ones (shorthand)")
	flag.BoolVar(&excludeSystem, "exclude-system", true, "Exclude system namespaces")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig")
	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be done without writing files")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&verbose, "v", false, "Verbose output (shorthand)")

	flag.Parse()

	if allNamespaces {
		excludeSystem = false
	}

	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			home := homedir.HomeDir()
			kubeconfigPath := filepath.Join(home, ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	config.QPS = 50
	config.Burst = 100

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating discovery client: %v\n", err)
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[INFO] Connected to cluster: %s\n", config.Host)
	fmt.Printf("[INFO] Output directory: %s\n", outputDir)

	exp := NewExporter(
		discoveryClient,
		dynamicClient,
		outputDir,
		namespace,
		allNamespaces,
		excludeSystem,
		dryRun,
		verbose,
	)
	if err := exp.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
		os.Exit(1)
	}
}