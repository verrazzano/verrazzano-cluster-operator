// Copyright (C) 2020, Oracle Corporation and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Handles creation/deletion of VerrazzanoManagedClusters

package managedclusters

import (
	"github.com/golang/glog"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/rancher"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util/diff"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

func CreateSecret(kubeClientSet kubernetes.Interface, secretLister corev1listers.SecretLister, cluster rancher.Cluster) error {
	secretName := util.GetManagedClusterKubeconfigSecretName(cluster.Name)
	glog.V(6).Infof("Processing VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
	newSecret := newSecret(secretName, cluster)

	existingSecret, err := secretLister.Secrets(constants.DefaultNamespace).Get(secretName)
	if existingSecret != nil {
		specDiffs := diff.CompareIgnoreTargetEmpties(existingSecret, newSecret)
		if specDiffs != "" {
			glog.V(4).Infof("Updating VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
			glog.V(6).Infof("Spec differences:\n%s", specDiffs)
			_, err = kubeClientSet.CoreV1().Secrets(constants.DefaultNamespace).Update(newSecret)
		} else {
			glog.V(6).Infof("No need to update existing VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
		}
	} else {
		glog.V(4).Infof("Creating VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
		_, err = kubeClientSet.CoreV1().Secrets(constants.DefaultNamespace).Create(newSecret)
	}
	if err != nil {
		return err
	}

	glog.V(6).Infof("Successfully processed VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
	return nil
}

func DeleteSecret(kubeClientSet kubernetes.Interface, secretLister corev1listers.SecretLister, cluster rancher.Cluster) error {
	secretName := util.GetManagedClusterKubeconfigSecretName(cluster.Id)
	glog.V(6).Infof("Deleting VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)

	_, err := secretLister.Secrets(constants.DefaultNamespace).Get(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			glog.Errorf("VerrazzanoManagedCluster Secret `%s` no longer exists for cluster '%s', for the reason (%v)", secretName, cluster.Name, err)
		}
		return err
	}

	err = kubeClientSet.CoreV1().Secrets(constants.DefaultNamespace).Delete(secretName, &metav1.DeleteOptions{})
	if err != nil {
		glog.Errorf("Failed to delete VerrazzanoManagedCluster Secret '%s' for cluster '%s', for the reason (%v)", secretName, cluster.Name, err)
		return err
	}

	glog.V(6).Infof("Successfully deleted VerrazzanoManagedCluster Secret '%s' for cluster '%s'", secretName, cluster.Name)
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

// get the ca.crt from secret "tls-rancher-ingress" in namespace "cattle-system"
func GetRancherCACert(kubeClientSet kubernetes.Interface) []byte {
	certSecret, err := kubeClientSet.CoreV1().Secrets(rancher.RancherNamespace).Get(rancher.TlsRancherIngressSecret, metav1.GetOptions{})
	if err != nil {
		glog.Warningf("Error getting secret %s/%s in management cluster: %s", rancher.RancherNamespace, rancher.TlsRancherIngressSecret, err.Error())
		return []byte{}
	}
	if certSecret == nil {
		glog.Warningf("Secret %s/%s not found in management cluster", rancher.RancherNamespace, rancher.TlsRancherIngressSecret)
		return []byte{}
	}
	return certSecret.Data["ca.crt"]
}
