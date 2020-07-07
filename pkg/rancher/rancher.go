// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util"
	"go.uber.org/zap"
)

// interface to expose Rancher APIs
type rancher interface {
	APICall(rancherConfig Config, apiPath string, httpMethod string, parameterMap map[string]string, payload string) (*gabs.Container, error)
}

// The Rancher default implementation
type Rancher struct{}

func getRealPath(path string, clusterID string) string {
	return strings.Replace(path, clusterReplacementString, clusterID, -1)
}

// GetClusters returns Rancher clusters
func GetClusters(r rancher, rancherConfig Config) ([]Cluster, error) {
	var clusters []Cluster

	json, err := r.APICall(rancherConfig, clustersAPIPath, http.MethodGet, defaultParameterMap, defaultPayload)
	if err != nil {
		return nil, err
	}

	clustersMap := json.Path(jsonDataPath).Children()
	for _, clusterInfo := range clustersMap {
		clusterID := clusterInfo.Path(jsonIDPath).Data().(string)

		// generate kubeconfig contents
		kubeconfigContents, err := getGenerateKubeconfig(r, rancherConfig, clusterID)
		if err != nil {
			return nil, err
		}

		// get the k8s api server for this cluster
		server := getValue(clusterInfo, jsonK8sAPIHostPath, "") + ":" + getValue(clusterInfo, jsonK8sAPIPortPath, "")

		clusters = append(
			clusters, Cluster{
				ID:                 clusterID,
				Name:               clusterInfo.Path(jsonNamePath).Data().(string),
				KubeConfigContents: kubeconfigContents,
				ServerAddress:      server,
				Type:               getValue(clusterInfo, jsonTypePath, ""),
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

func getGenerateKubeconfig(r rancher, rancherConfig Config, clusterID string) (string, error) {
	json, err := r.APICall(rancherConfig, getRealPath(generateKubeConfigAPIPath, clusterID), http.MethodPost, generateKubeConfigParameterMap, "")
	if err != nil {
		return "", err
	}
	return json.Path(config).Data().(string), nil
}

// APICall for Generic Rancher API call returning a json object.
func (c Rancher) APICall(rancherConfig Config, apiPath string, httpMethod string, parameterMap map[string]string, payload string) (*gabs.Container, error) {
	defaultHeaders := map[string]string{"Content-Type": "application/json"}
	zap.S().Debugf("[APICall] [%s] url:'%s'", httpMethod, rancherConfig.URL+apiPath)

	response, responseBody, err := util.WaitForSendRequest(httpMethod, rancherConfig.URL+apiPath, rancherConfig.Host,
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
