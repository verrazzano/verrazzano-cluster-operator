// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util"
)

// interface to expose Rancher APIs
type rancher interface {
    APICall(rancherConfig Config, apiPath string, httpMethod string, parameterMap map[string]string, payload string) (*gabs.Container, error)
}

// default rancher implementation
type Rancher struct{}

func getRealPath(path string, clusterId string) string {
	return strings.Replace(path, CLUSTER_REPLACMENT_STRING, clusterId, -1)
}

func GetClusters(r rancher, rancherConfig Config) ([]Cluster, error) {
	var clusters []Cluster

	json, err := r.APICall(rancherConfig, CLUSTERS_API_PATH, http.MethodGet, DEFAULT_PARAMETER_MAP, DEFAULT_PAYLOAD)
	if err != nil {
		return nil, err
	}

	clustersMap := json.Path(JSON_DATA_PATH).Children()
	for _, clusterInfo := range clustersMap {
		clusterId := clusterInfo.Path(JSON_ID_PATH).Data().(string)

		// generate kubeconfig contents
		kubeconfigContents, err := getGenerateKubeconfig(r, rancherConfig, clusterId)
		if err != nil {
			return nil, err
		}

		// get the k8s api server for this cluster
		server := getValue(clusterInfo, JSON_K8S_API_HOST_PATH, "") + ":" + getValue(clusterInfo, JSON_K8S_API_PORT_PATH, "")

		clusters = append(
			clusters, Cluster{
				Id:                 clusterId,
				Name:               clusterInfo.Path(JSON_NAME_PATH).Data().(string),
				KubeConfigContents: kubeconfigContents,
				ServerAddress:      server,
				Type:               getValue(clusterInfo, JSON_TYPE_PATH, ""),
			})
	}

	return clusters, nil
}

// get a value at the given path from the given container
func getValue(info *gabs.Container, path string, def string) string {
	value := info.Path(path).Data()
	if value != nil {
		return value.(string)
	}
	return def
}

func getGenerateKubeconfig(r rancher, rancherConfig Config, clusterId string) (string, error) {
	json, err := r.APICall(rancherConfig, getRealPath(GENERATE_KUBECONFIG_API_PATH, clusterId), http.MethodPost, GENERATE_KUBECONFIG_PARAMETER_MAP, "")
	if err != nil {
		return "", err
	}
	return json.Path(CONFIG).Data().(string), nil
}

// Generic Rancher API call which return a json object.
func (c Rancher) APICall(rancherConfig Config, apiPath string, httpMethod string, parameterMap map[string]string, payload string) (*gabs.Container, error) {
	defaultHeaders := map[string]string{"Content-Type": "application/json"}

	// create logger for API call
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Rancher").Str("name", rancherConfig.Host).Logger()

	logger.Debug().Msgf("[APICall] [%s] url:'%s'", httpMethod, rancherConfig.Url+apiPath)

	response, responseBody, err := util.WaitForSendRequest(httpMethod, rancherConfig.Url+apiPath, rancherConfig.Host,
		defaultHeaders, parameterMap, payload, rancherConfig.Username, rancherConfig.Password, rancherConfig.CertificateAuthorityData, util.DefaultRetry)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected response code %d from POST but got %d: %v", http.StatusOK, response.StatusCode, response)
	}

	json, err := gabs.ParseJSON([]byte(responseBody))
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body to json: %s", responseBody)
	}

	return json, nil
}
