package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

type resourceInfo struct {
	GVR        schema.GroupVersionResource
	Namespaced bool
}

type Exporter struct {
	discoveryClient discovery.DiscoveryClient
	dynamicClient   dynamic.Interface
	outputDir       string
	namespace       string
	allNamespaces   bool
	excludeSystem   bool
	dryRun          bool
	verbose         bool
	fileCount       int
}

func NewExporter(
	dc discovery.DiscoveryClient,
	dyn dynamic.Interface,
	outputDir string,
	namespace string,
	allNamespaces bool,
	excludeSystem bool,
	dryRun bool,
	verbose bool,
) *Exporter {
	return &Exporter{
		discoveryClient: dc,
		dynamicClient:   dyn,
		outputDir:       outputDir,
		namespace:       namespace,
		allNamespaces:   allNamespaces,
		excludeSystem:   excludeSystem,
		dryRun:          dryRun,
		verbose:         verbose,
	}
}

func (e *Exporter) Run() error {
	namespaces, err := e.getNamespaces()
	if err != nil {
		return err
	}

	infos, err := e.discoverResources()
	if err != nil {
		return err
	}

	var clusterInfos, namespaceInfos []resourceInfo
	for _, info := range infos {
		if info.Namespaced {
			namespaceInfos = append(namespaceInfos, info)
		} else {
			clusterInfos = append(clusterInfos, info)
		}
	}

	clusterInfos = orderResources(clusterInfos)
	namespaceInfos = orderResources(namespaceInfos)

	fmt.Println("[STEP] Exporting cluster-scoped resources...")
	for _, info := range clusterInfos {
		if err := e.exportResource(info.GVR, "", true); err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] %v\n", err)
		}
	}

	fmt.Println("[STEP] Exporting namespace-scoped resources...")
	for _, ns := range namespaces {
		for _, info := range namespaceInfos {
			if err := e.exportResource(info.GVR, ns, false); err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] %v\n", err)
			}
		}
	}

	fmt.Println("========================================")
	if e.dryRun {
		fmt.Println("Dry-run complete!")
	} else {
		fmt.Println("Export complete!")
		fmt.Printf("Result: %s\n", e.outputDir)
		fmt.Printf("Total files: %d\n", e.fileCount)
	}
	fmt.Println("========================================")

	return nil
}

func (e *Exporter) getNamespaces() ([]string, error) {
	if e.namespace != "" {
		fmt.Printf("[INFO] Namespace: %s\n", e.namespace)
		return []string{e.namespace}, nil
	}

	nsGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	list, err := e.dynamicClient.Resource(nsGVR).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var namespaces []string
	var skipped []string
	for _, item := range list.Items {
		name := item.GetName()
		if e.excludeSystem && isSystemNamespace(name) {
			skipped = append(skipped, name)
			continue
		}
		namespaces = append(namespaces, name)
	}

	if len(skipped) > 0 {
		fmt.Printf("[INFO] Namespaces: %s (skipped: %s)\n",
			strings.Join(namespaces, ", "), strings.Join(skipped, ", "))
	} else {
		fmt.Printf("[INFO] Namespaces: %s\n", strings.Join(namespaces, ", "))
	}
	return namespaces, nil
}

func (e *Exporter) discoverResources() ([]resourceInfo, error) {
	lists, err := e.discoveryClient.ServerPreferredResources()
	if err != nil {
		if discovery.IsGroupDiscoveryFailedError(err) {
			if groups, ok := discovery.GroupDiscoveryFailedErrorGroups(err); ok {
				for gv, gvErr := range groups {
					fmt.Fprintf(os.Stderr, "[WARN] Skipping unavailable API group %s: %v\n", gv, gvErr)
				}
			}
		} else {
			return nil, fmt.Errorf("discovery failed: %w", err)
		}
	}

	var infos []resourceInfo
	for _, list := range lists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, r := range list.APIResources {
			if strings.Contains(r.Name, "/") {
				continue // sub-resource
			}
			if !containsVerb(r.Verbs, "list") {
				continue
			}
			infos = append(infos, resourceInfo{
				GVR:        gv.WithResource(r.Name),
				Namespaced: r.Namespaced,
			})
		}
	}
	return infos, nil
}

func (e *Exporter) exportResource(gvr schema.GroupVersionResource, namespace string, isCluster bool) error {
	var ri dynamic.ResourceInterface
	if isCluster {
		ri = e.dynamicClient.Resource(gvr)
	} else {
		ri = e.dynamicClient.Resource(gvr).Namespace(namespace)
	}

	list, err := ri.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if apierrors.IsForbidden(err) {
			fmt.Fprintf(os.Stderr, "[WARN] Forbidden to list %s: %v\n", gvr.String(), err)
			return nil
		}
		if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
			fmt.Fprintf(os.Stderr, "[WARN] Skipping %s: %v\n", gvr.String(), err)
			return nil
		}
		return fmt.Errorf("failed to list %s: %w", gvr.String(), err)
	}

	if len(list.Items) == 0 {
		return nil
	}

	dir := filepath.Join(e.outputDir, gvr.Resource)
	if !e.dryRun {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	for i := range list.Items {
		u := list.Items[i].DeepCopy()
		Clean(u)

		data, err := yaml.Marshal(u.Object)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[WARN] Failed to marshal %s/%s: %v\n", gvr.Resource, u.GetName(), err)
			continue
		}

		fileName := u.GetName() + ".yaml"
		path := filepath.Join(dir, fileName)

		if e.dryRun {
			fmt.Printf("[DRY-RUN] Would write: %s\n", path)
		} else {
			if err := os.WriteFile(path, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "[WARN] Failed to write %s: %v\n", path, err)
				continue
			}
			e.fileCount++
			fmt.Printf("[INFO] %s\n", path)
		}
	}
	return nil
}

func containsVerb(verbs metav1.Verbs, want string) bool {
	for _, v := range verbs {
		if v == want {
			return true
		}
	}
	return false
}

func orderResources(infos []resourceInfo) []resourceInfo {
	priority := make(map[schema.GroupVersionResource]int, len(orderedGVRList))
	for i, gvr := range orderedGVRList {
		priority[gvr] = i
	}

	sort.SliceStable(infos, func(i, j int) bool {
		pi, iok := priority[infos[i].GVR]
		pj, jok := priority[infos[j].GVR]

		if iok && jok {
			return pi < pj
		}
		if iok {
			return true
		}
		if jok {
			return false
		}
		return infos[i].GVR.String() < infos[j].GVR.String()
	})
	return infos
}