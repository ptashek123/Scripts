package main

import "k8s.io/apimachinery/pkg/runtime/schema"

var orderedGVRList = []schema.GroupVersionResource{
	{Group: "", Version: "v1", Resource: "namespaces"},

	{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
	{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	{Group: "", Version: "v1", Resource: "serviceaccounts"},

	{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	{Group: "", Version: "v1", Resource: "persistentvolumes"},

	{Group: "", Version: "v1", Resource: "configmaps"},
	{Group: "", Version: "v1", Resource: "secrets"},

	{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},

	{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	{Group: "networking.k8s.io", Version: "v1", Resource: "ingressclasses"},

	{Group: "apps", Version: "v1", Resource: "deployments"},
	{Group: "apps", Version: "v1", Resource: "statefulsets"},
	{Group: "apps", Version: "v1", Resource: "daemonsets"},
	{Group: "apps", Version: "v1", Resource: "replicasets"},
	{Group: "batch", Version: "v1", Resource: "jobs"},
	{Group: "batch", Version: "v1", Resource: "cronjobs"},

	{Group: "", Version: "v1", Resource: "services"},
	{Group: "", Version: "v1", Resource: "endpoints"},
	{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},

	{Group: "autoscaling", Version: "v1", Resource: "horizontalpodautoscalers"},
	{Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"},
	{Group: "", Version: "v1", Resource: "limitranges"},
	{Group: "", Version: "v1", Resource: "resourcequotas"},

	{Group: "coordination.k8s.io", Version: "v1", Resource: "leases"},
	{Group: "", Version: "v1", Resource: "events"},
	{Group: "", Version: "v1", Resource: "pods"},
	{Group: "storage.k8s.io", Version: "v1", Resource: "volumeattachments"},
	{Group: "storage.k8s.io", Version: "v1", Resource: "csidrivers"},
	{Group: "storage.k8s.io", Version: "v1", Resource: "csinodes"},
	{Group: "node.k8s.io", Version: "v1", Resource: "runtimeclasses"},
	{Group: "scheduling.k8s.io", Version: "v1", Resource: "priorityclasses"},
}

var systemNamespaces = map[string]bool{
	"kube-system":                   true,
	"kube-public":                   true,
	"kube-node-lease":               true,
	"cert-manager":                  true,
	"cert-manager-local":            true,
	"ingress-nginx":                 true,
	"monitoring":                    true,
	"prometheus":                    true,
	"istio-system":                  true,
	"istio-ingress":                 true,
	"linkerd":                       true,
	"calico-system":                 true,
	"tigera-operator":               true,
	"kube-flannel":                  true,
	"cattle-system":                 true,
	"cattle-global-data":            true,
	"local-path-storage":            true,
	"metallb-system":                true,
	"longhorn-system":               true,
	"rook-ceph":                     true,
	"rook-ceph-system":              true,
	"opentelemetry-operator-system": true,
	"argocd":                        true,
	"argocd-operator-system":        true,
	"flux-system":                   true,
	"velero":                        true,
}

func isSystemNamespace(name string) bool {
	return systemNamespaces[name]
}
