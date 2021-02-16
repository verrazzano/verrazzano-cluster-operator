// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/Jeffail/gabs/v2"
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

	response, responseBody, err := WaitForSendRequest(httpMethod, rancherConfig, apiPath, defaultHeaders, parameterMap, payload, DefaultRetry)
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

// DefaultRetry is the default backoff for e2e tests.
var DefaultRetry = wait.Backoff{
	Steps:    12,
	Duration: 5 * time.Second,
	Factor:   1.0,
	Jitter:   0.1,
}

// Retry executes the provided function repeatedly, retrying until the function
// returns done = true, errors, or exceeds the given timeout.
func Retry(backoff wait.Backoff, fn wait.ConditionFunc) error {
	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		done, err := fn()
		if err != nil {
			lastErr = err
		}
		return done, err
	})
	if err == wait.ErrWaitTimeout {
		if lastErr != nil {
			err = lastErr
		}
	}
	return err
}

// Retrieve proxy url from environment
func getProxyURL() string {
	if proxyURL := os.Getenv("https_proxy"); proxyURL != "" {
		return proxyURL
	}
	if proxyURL := os.Getenv("HTTPS_PROXY"); proxyURL != "" {
		return proxyURL
	}
	if proxyURL := os.Getenv("http_proxy"); proxyURL != "" {
		return proxyURL
	}
	if proxyURL := os.Getenv("HTTP_PROXY"); proxyURL != "" {
		return proxyURL
	}
	return ""
}

func rootCertPool(caData []byte) *x509.CertPool {
	if len(caData) == 0 {
		return nil
	}

	// if we have caData, use it
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caData)
	return certPool
}

// SendRequest sends http request
func SendRequest(action string, rancherConfig Config, apiPath string, headers, parameterMap map[string]string, payload string) (*http.Response, string, error) {
	proxyURL := getProxyURL()

	// Create Transport object
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{RootCAs: rootCertPool(rancherConfig.CertificateAuthorityData)},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Add the proxy URL to the Transport object
	if proxyURL != "" {
		tURL := url.URL{}
		tURLProxy, _ := tURL.Parse(proxyURL)
		tr.Proxy = http.ProxyURL(tURLProxy)
	}

	// Add the Transport object to the http Client
	client := &http.Client{Transport: tr, Timeout: 300 * time.Second}

	// Generate the HTTP GET request
	reqURL := rancherConfig.URL + apiPath
	req, err := http.NewRequest(action, reqURL, strings.NewReader(payload))
	if err != nil {
		return nil, "", err
	}

	req.Header.Add("Accept", "*/*")

	// Add any headers to the request
	for k := range headers {
		req.Header.Add(k, headers[k])
	}

	// Set resolve
	urlObj, err := url.Parse(reqURL)
	if err != nil {
		zap.S().Fatalf("Invalid URL '%s': %v", reqURL, err)
		return nil, "", err
	}
	parsedHost := urlObj.Host
	nodeIP := rancherConfig.NodeIP
	nodePort := rancherConfig.NodePort
	zap.S().Debugf("resolve address: %s:%s \n", nodeIP, nodePort)
	if nodeIP != "" && nodePort != "" && nodeIP != parsedHost {
		// When the in-cluster accessible host is different from the outside accessible URL's host (parsedHost),
		// do a 'curl --resolve' equivalent
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			zap.S().Debugf("address original: %s \n", addr)
			if addr == parsedHost+":443" {
				addr = nodeIP + ":" + nodePort
				zap.S().Debugf("address modified: %s \n", addr)
			}
			return dialer.DialContext(ctx, network, addr)
		}
	}

	// Set basic auth
	if rancherConfig.Username != "" && rancherConfig.Password != "" {
		req.SetBasicAuth(rancherConfig.Username, rancherConfig.Password)
	}

	// Add parameters
	query := req.URL.Query()
	for key, value := range parameterMap {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// Extract the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return resp, string(body), err
}

// WaitForSendRequest waits for the given request to return results
func WaitForSendRequest(action string, rancherConfig Config, apiPath string, headers, parameterMap map[string]string, payload string, backoff wait.Backoff) (latestResponse *http.Response, latestResponseBody string, err error) {
	expectedStatusCode := http.StatusOK
	zap.S().Debugf("Waiting for %s to reach status code %d...\n", rancherConfig.URL, expectedStatusCode)
	startTime := time.Now()

	err = Retry(backoff, func() (bool, error) {
		response, responseBody, reqErr := SendRequest(action, rancherConfig, apiPath, headers, parameterMap, payload)
		latestResponse = response
		latestResponseBody = responseBody
		if reqErr != nil {
			return false, reqErr
		}
		if response.StatusCode == expectedStatusCode {
			return true, nil
		}
		return false, nil
	})
	zap.S().Debugf("Wait time: %s \n", time.Since(startTime))
	return latestResponse, latestResponseBody, err
}
