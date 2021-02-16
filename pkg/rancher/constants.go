// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

// Structure maintaining the details of a Cluster obtained from the Rancher URL

// Config contains Rancher Server endpoint URL and credentials structure
type Config struct {
	URL                      string
	Username                 string
	Password                 string
	NodeIP                   string
	NodePort                 string
	CertificateAuthorityData []byte
}

// Cluster contains Rancher Managed cluster structure
type Cluster struct {
	ID                 string
	Name               string
	KubeConfigContents string
	PrometheusURL      string
	ServerAddress      string
	Type               string
}

// Rancher API URLs
const (
	clusterReplacementString  = "##CLUSTER_ID##"
	clustersAPIPath           = "/v3/clusters"
	generateKubeConfigAPIPath = "/v3/clusters/" + clusterReplacementString
)

// Rancher API configurations
var (
	defaultPayload                 = ""
	defaultParameterMap            = map[string]string{}
	generateKubeConfigParameterMap = map[string]string{"action": "generateKubeconfig"}
)

// Rancher Response json paths
const (
	jsonDataPath       = "data"
	jsonIDPath         = "id"
	jsonNamePath       = "name"
	jsonK8sAPIHostPath = "labels.k8sApiHost"
	jsonK8sAPIPortPath = "labels.k8sApiPort"
	jsonTypePath       = "labels.type"
	config             = "config"
)

// RancherNamespace contains constant for Rancher namespace
const RancherNamespace = "cattle-system"

// TLSRancherIngressSecret contains constant for tls-rancher-ingress
const TLSRancherIngressSecret = "tls-rancher-ingress"
