package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	defaultNamespace string          = "default"
	patchType        types.PatchType = types.MergePatchType
)

type Metadata struct {
	Labels `json:"labels"`
}

type patch struct {
	Metadata `json:"metadata"`
}

type Labels map[string]string

func main() {
	// define cmd flags
	nodeName := flag.String("node-name", "", "this node status will be used")
	podNamespacedName := flag.String("pod-namespaced-name", "", "this pod labels will be updated (use namespace/pod format)")
	addAddresses := flag.Bool("add-addresses", true, "node addresses will be add into pod labels")

	flag.Parse()

	podNamespace, podName, err := getPodNamespaceAndName(*podNamespacedName)
	if err != nil {
		log.Fatalf("failed to parse pod namespace and name: %w", err)
	}

	// creates the in-cluster config
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatalf("failed to get kubeconfig: %w", err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		log.Fatalf("failed to init clientset: %w", err)
	}

	// get node meta
	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), *nodeName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("failed to get node info: %w", err)
	}

	labels := Labels{}
	if *addAddresses {
		labels.addNodeAddresses(node.Status.Addresses)
	}

	data, err := json.Marshal(patch{Metadata: Metadata{Labels: labels}})
	if err != nil {
		log.Fatalf("failed to marshal patch request body: %w", err)
	}

	log.Printf("patch request body: %s\n", data)

	// patch pod
	_, err = clientset.CoreV1().Pods(podNamespace).Patch(
		context.TODO(),
		podName,
		patchType,
		data,
		metav1.PatchOptions{},
	)
	if err != nil {
		log.Fatalf("failed to make patch request: %w", err)
	}

	log.Printf("succefully patched\n")
}

func getPodNamespaceAndName(podNamespacedName string) (string, string, error) {
	v := strings.Split(podNamespacedName, "/")
	if len(v) != 2 {
		return "", "", errors.New("incorrect flag value")
	}

	return v[0], v[1], nil
}

func (l *Labels) addNodeAddresses(addrs []v1.NodeAddress) {
	for _, addr := range addrs {
		switch addr.Type {
		case v1.NodeHostName:
			(*l)["node.status.addresses/hostname"] = addr.Address
		case v1.NodeInternalIP:
			(*l)["node.status.addresses/internal-ip"] = addr.Address
		case v1.NodeExternalIP:
			(*l)["node.status.addresses/external-ip"] = addr.Address
		case v1.NodeInternalDNS:
			(*l)["node.status.addresses/internal-dns"] = addr.Address
		case v1.NodeExternalDNS:
			(*l)["node.status.addresses/external-dns"] = addr.Address
		}
	}
}
