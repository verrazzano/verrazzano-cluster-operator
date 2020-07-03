// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package integ_test

import (
	"bytes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"path/filepath"
	"strings"
)

const Namespace string = "default"

var _ = Describe("Verrazzano Cluster Operator", func() {
	It("is deployed", func() {
		deployment, err := getClientSet().AppsV1().Deployments(Namespace).Get("verrazzano-cluster-operator", metav1.GetOptions{})
		Expect(err).To(BeNil(), "Received an error while trying to get the verrazzano-cluster-operator deployment")
		Expect(deployment.Spec.Template.Spec.Containers[0].Name).To(Equal("verrazzano-cluster-operator"))
	})

	It("is running (within 5m)", func() {
		isPodRunningYet := func() bool {
			return isPodRunning("verrazzano-cluster-operator", Namespace)
		}
		Eventually(isPodRunningYet, "5m", "5s").Should(BeTrue(),
			"The verrazzano-cluster-operator pod should be in the Running state")
	})
})

func getKubeconfig() string {
	var kubeconfig string
	if home := homeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		Fail("Could not get kubeconfig")
	}
	return kubeconfig
}

func getClientSet() *kubernetes.Clientset {
	kubeconfig := getKubeconfig()

	// use the current context in the kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		Fail("Could not get current context from kubeconfig " + kubeconfig)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Fail("Could not get clientset from config")
	}

	return clientset
}

func isPodRunning(name string, namespace string) bool {
	GinkgoWriter.Write([]byte("[DEBUG] checking if there is a running pod named " + name + "* in namespace " + namespace + "\n"))
	clientset := getClientSet()
	podInterface := clientset.CoreV1().Pods(namespace)
	pods, err := podInterface.List(metav1.ListOptions{})
	if err != nil {
		Fail("Could not get list of pods")
	}
	GinkgoWriter.Write([]byte(fmt.Sprintf("PODS: %v\n", pods)))
	var podName string
	for i := range pods.Items {
		GinkgoWriter.Write([]byte("\nPOD details: " + pods.Items[i].Name + "\n"))
		if strings.HasPrefix(pods.Items[i].Name, name) {
			podName = pods.Items[i].Name
			conditions := pods.Items[i].Status.Conditions
			for j := range conditions {
				GinkgoWriter.Write([]byte(fmt.Sprintf("Condition: %v, Status: %v\n", conditions[j].Type, conditions[j].Status)))
				if conditions[j].Type == "Ready" {
					if conditions[j].Status == "True" {
						return true
					}
				}
			}
		}
	}
	podLogs := getPodLogs(podName, podInterface)
	GinkgoWriter.Write([]byte(fmt.Sprintf("***** POD LOG START ***** %v ***** POD LOG END *****\n\n", podLogs)))

	return false
}

func getPodLogs(podName string, podInterface clientV1.PodInterface) string {
	req := podInterface.GetLogs(podName, &v1.PodLogOptions{})
	podLogs, err := req.Stream()
	if err != nil {
		return fmt.Sprintf("ERROR Opening Stream: %v", err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return fmt.Sprintf("Failed to copy log information to buf: %v", err)
	}
	return buf.String()
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return ""
}
