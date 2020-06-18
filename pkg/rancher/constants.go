// Copyright (C) 2020, Oracle Corporation and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

// Structure maintaining the details of a Cluster obtained from the Rancher URL

// Rancher Server endpoint URL and credentials structure
type Config struct {
	Url                      string
	Username                 string
	Password                 string
	Host                     string
	CertificateAuthorityData []byte
}

// Rancher Managed cluster structure
type Cluster struct {
	Id                 string
	Name               string
	KubeConfigContents string
	PrometheusURL      string
	ServerAddress      string
	Type			   string
}

// Rancher API URLs
const (
	CLUSTER_REPLACMENT_STRING    = "##CLUSTER_ID##"
	CLUSTERS_API_PATH            = "/v3/clusters"
	GENERATE_KUBECONFIG_API_PATH = "/v3/clusters/" + CLUSTER_REPLACMENT_STRING
)

// Rancher API configurations
var (
	DEFAULT_PAYLOAD                   = ""
	DEFAULT_PARAMETER_MAP             = map[string]string{}
	GENERATE_KUBECONFIG_PARAMETER_MAP = map[string]string{"action": "generateKubeconfig"}
)

// Rancher Response json paths
const (
	JSON_DATA_PATH = "data"
	JSON_ID_PATH   = "id"
	JSON_NAME_PATH = "name"
	JSON_K8S_API_HOST_PATH = "labels.k8sApiHost"
	JSON_K8S_API_PORT_PATH = "labels.k8sApiPort"
	JSON_TYPE_PATH = "labels.type"
	CONFIG         = "config"
)

// Rancher
const RancherNamespace = "cattle-system"
const TlsRancherIngressSecret = "tls-rancher-ingress"
