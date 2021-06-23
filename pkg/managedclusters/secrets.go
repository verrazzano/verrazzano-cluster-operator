// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Handles creation/deletion of VerrazzanoManagedClusters

package managedclusters

import (
	"context"

	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/rancher"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util"
	"github.com/verrazzano/pkg/diff"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

// CreateSecret creates/updates a VerrazzanoManagedCluster secret
func CreateSecret(kubeClientSet kubernetes.Interface, secretLister corev1listers.SecretLister, cluster rancher.Cluster) error {
	secretName := util.GetManagedClusterKubeconfigSecretName(cluster.Name)
	zap.S().Debugf("Processing VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
	newSecret := newSecret(secretName, cluster)

	existingSecret, err := secretLister.Secrets(constants.DefaultNamespace).Get(secretName)
	if existingSecret != nil {
		specDiffs := diff.Diff(existingSecret, newSecret)
		if specDiffs != "" {
			zap.S().Infof("Updating VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
			zap.S().Debugf("Spec differences:\n%s", specDiffs)
			_, err = kubeClientSet.CoreV1().Secrets(constants.DefaultNamespace).Update(context.TODO(), newSecret, metav1.UpdateOptions{})
		} else {
			zap.S().Debugf("No need to update existing VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
		}
	} else {
		zap.S().Infof("Creating VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
		_, err = kubeClientSet.CoreV1().Secrets(constants.DefaultNamespace).Create(context.TODO(), newSecret, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	zap.S().Debugf("Successfully processed VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
	return nil
}

// DeleteSecret deletes a VerrazzanoManagedCluster secret
func DeleteSecret(kubeClientSet kubernetes.Interface, secretLister corev1listers.SecretLister, cluster rancher.Cluster) error {
	secretName := util.GetManagedClusterKubeconfigSecretName(cluster.ID)
	zap.S().Debugf("Deleting VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)

	_, err := secretLister.Secrets(constants.DefaultNamespace).Get(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			zap.S().Errorf("VerrazzanoManagedCluster Secret `%s` no longer exists for cluster '%s', for the reason (%v)", secretName, cluster.Name, err)
		}
		return err
	}

	err = kubeClientSet.CoreV1().Secrets(constants.DefaultNamespace).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
	if err != nil {
		zap.S().Errorf("Failed to delete VerrazzanoManagedCluster Secret '%s' for cluster '%s', for the reason (%v)", secretName, cluster.Name, err)
		return err
	}

	zap.S().Debugf("Successfully deleted VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
	return nil
}

// Constructs the secret for the given cluster
func newSecret(secretName string, cluster rancher.Cluster) *corev1.Secret {
	return &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: constants.DefaultNamespace,
			Labels:    util.GetManagedClusterLabels(cluster.Name),
		},
		Data: map[string][]byte{
			constants.KubeconfigSecretKey: []byte(cluster.KubeConfigContents),
		},
	}
}

// GetRancherCACert gets the ca.crt from secret "tls-rancher-ingress" in namespace "cattle-system"
func GetRancherCACert(kubeClientSet kubernetes.Interface) []byte {
	certSecret, err := kubeClientSet.CoreV1().Secrets(rancher.RancherNamespace).Get(context.TODO(), rancher.TLSRancherIngressSecret, metav1.GetOptions{})
	if err != nil {
		zap.S().Warnf("Error getting secret %s/%s in admin cluster: %s", rancher.RancherNamespace, rancher.TLSRancherIngressSecret, err.Error())
		return []byte{}
	}
	if certSecret == nil {
		zap.S().Warnf("Secret %s/%s not found in admin cluster", rancher.RancherNamespace, rancher.TLSRancherIngressSecret)
		return []byte{}
	}
	return certSecret.Data["ca.crt"]
}

// GetRancherCredentials returns username/password from secret "verrazzano" in namespace "verrazzano-system"
func GetRancherCredentials(kubeClientSet kubernetes.Interface) (string, string) {
	const verrazzanoNamespace = "verrazzano-system"
	const credentailsSecret = "verrazzano"
	credsSecret, err := kubeClientSet.CoreV1().Secrets(verrazzanoNamespace).Get(context.TODO(), credentailsSecret, metav1.GetOptions{})
	if err != nil {
		zap.S().Warnf("Error getting secret %s/%s in admin cluster: %s", verrazzanoNamespace, credentailsSecret, err.Error())
		return "", ""
	}
	if credsSecret == nil {
		zap.S().Warnf("Secret %s/%s not found in admin cluster", verrazzanoNamespace, credentailsSecret)
		return "", ""
	}
	zap.S().Infof("The username used to connect to rancher in admin cluster is %s", credsSecret.StringData["username"])
	return credsSecret.StringData["username"], credsSecret.StringData["password"]
}

// GetNginxIngressControllerNodeIPAndPort returns the nginx controller node ip and port
func GetNginxIngressControllerNodeIPAndPort(kubeClientSet kubernetes.Interface) (string, int32) {
	const NginxNamespace = "ingress-nginx"
	nodePort := int32(0)
	service, err := kubeClientSet.CoreV1().Services(NginxNamespace).Get(context.TODO(), "ingress-controller-ingress-nginx-controller", metav1.GetOptions{})
	if err != nil {
		zap.S().Warnf("Error getting servcice for ingress-nginx-controller in admin cluster: %s", err.Error())
		return "", nodePort
	}
	for _, servicePort := range service.Spec.Ports {
		if servicePort.Name == "https" {
			nodePort = servicePort.NodePort
		}
	}
	set := labels.Set(service.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := kubeClientSet.CoreV1().Pods(NginxNamespace).List(context.TODO(), listOptions)
	if err != nil {
		zap.S().Warnf("Error getting pod for ingress-nginx-controller in admin cluster: %s", err.Error())
		return "", nodePort
	}
	for _, pod := range pods.Items {
		zap.S().Infof("The host ip for ingress-nginx-controller pod in admin cluster is %s", pod.Status.HostIP)
		return pod.Status.HostIP, nodePort
	}
	return "", nodePort
}

// GetRancherIngress returns rancher ingress
func GetRancherIngress(kubeClientSet kubernetes.Interface) string {
	const rancherNamespace = "cattle-system"
	const rancherIngressName = "rancher"
	ingress, err := kubeClientSet.ExtensionsV1beta1().Ingresses(rancherNamespace).Get(context.TODO(), rancherIngressName, metav1.GetOptions{})
	if err != nil {
		zap.S().Warnf("Error getting ingress %s/%s in admin cluster: %s", rancherNamespace, rancherIngressName, err.Error())
		return ""
	}
	if ingress == nil {
		zap.S().Warnf("Ingress %s/%s not found in admin cluster", rancherNamespace, rancherIngressName)
		return ""
	}
	for _, rule := range ingress.Spec.Rules {
		url := "https://" + rule.Host
		zap.S().Infof("The URL used to connect to rancher in admin cluster is %s", url)
		return url
	}
	return ""
}
