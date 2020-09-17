// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Jeffail/gabs/v2"
)

// mock rancher implementation
type TestRancher struct{}

func (c TestRancher) APICall(rancherConfig Config, apiPath string, httpMethod string, parameterMap map[string]string, payload string) (*gabs.Container, error) {
	if rancherConfig.URL == "bad-url" {
		return nil, fmt.Errorf("got %s", rancherConfig.URL)
	}
	var responseBody string
	if apiPath == "/v3/clusters" {
		responseBody = "{ \"data\": [{\"id\": \"c-ndvgb\", \"name\": \"foo-managed-1\", \"labels\": {\"type\": \"oke\", \"k8sApiHost\": \"130.35.130.66\", \"k8sApiPort\": \"6443\"}},{\"id\": \"c-r998z\", \"name\": \"foo-managed-2\", \"labels\": {\"type\": \"oke\", \"k8sApiHost\": \"147.154.97.197\", \"k8sApiPort\": \"6443\"}},{\"id\": \"local\", \"name\": \"local\", \"labels\": {\"type\": \"oke\", \"k8sApiHost\": \"147.154.96.26\", \"k8sApiPort\": \"6443\"}}]}"
	} else if httpMethod == http.MethodPost && parameterMap["action"] == "generateKubeconfig" {
		ss := strings.Split(apiPath, "/")
		clusterID := ss[len(ss)-1]
		responseBody = fmt.Sprintf("{\"config\": \"generatedKubeConfigOutput:%s\"}", clusterID)
	} else {
		return nil, fmt.Errorf("unrecognized request: %s", apiPath)
	}
	json, err := gabs.ParseJSON([]byte(responseBody))
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body to json: %s", responseBody)
	}
	return json, nil
}

func TestGetClusters(t *testing.T) {
	password := generateRandomString()
	type args struct {
		r             rancher
		rancherConfig Config
	}
	tests := []struct {
		name    string
		args    args
		want    []Cluster
		wantErr bool
	}{
		// test 1 - get 3 clusters
		{
			name: "get 3 clusters",
			args: struct {
				r             rancher
				rancherConfig Config
			}{
				r: TestRancher{},
				rancherConfig: Config{
					URL:                      "https://rancher.foo.verrazzano.example.com/",
					Username:                 "user1",
					Password:                 password,
					Host:                     "123.123.123.0",
					CertificateAuthorityData: []byte{},
				},
			},
			want: []Cluster{
				{
					ID:                 "c-ndvgb",
					Name:               "foo-managed-1",
					KubeConfigContents: "generatedKubeConfigOutput:c-ndvgb",
					PrometheusURL:      "",
					ServerAddress:      "130.35.130.66:6443",
					Type:               "oke",
				}, {
					ID:                 "c-r998z",
					Name:               "foo-managed-2",
					KubeConfigContents: "generatedKubeConfigOutput:c-r998z",
					PrometheusURL:      "",
					ServerAddress:      "147.154.97.197:6443",
					Type:               "oke",
				}, {
					ID:                 "local",
					Name:               "local",
					KubeConfigContents: "generatedKubeConfigOutput:local",
					PrometheusURL:      "",
					ServerAddress:      "147.154.96.26:6443",
					Type:               "oke",
				},
			},
			wantErr: false,
		},
		// test 2 - bad-url
		{
			name: "bad-url",
			args: struct {
				r             rancher
				rancherConfig Config
			}{
				r: TestRancher{},
				rancherConfig: Config{
					URL:                      "bad-url",
					Username:                 "user1",
					Password:                 password,
					Host:                     "123.123.123.0",
					CertificateAuthorityData: []byte{},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetClusters(tt.args.r, tt.args.rancherConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClusters() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// generateRandomString returns a base64 encoded generated random string.
func generateRandomString() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
