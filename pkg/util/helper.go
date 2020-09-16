// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/util/wait"
)

// GetManagedClusterKubeconfigSecretName returns the secret for a managed cluster
func GetManagedClusterKubeconfigSecretName(clusterID string) string {
	return fmt.Sprintf("%s-%s", constants.ManagedClusterPrefix, clusterID)
}

// GetManagedClusterLabels return labels for a managed cluster
func GetManagedClusterLabels(managedClusterID string) map[string]string {
	return map[string]string{constants.K8SAppLabel: constants.VerrazzanoGroup, constants.VerrazzanoClusterLabel: managedClusterID}
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

// Retrieve poxy url from environment
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
func SendRequest(action, reqURL, host string, headers map[string]string, parameterMap map[string]string, payload string, reqUserName string, reqPassword string, caData []byte) (*http.Response, string, error) {
	proxyURL := getProxyURL()

	// Create Transport object
	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{RootCAs: rootCertPool(caData)},
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
	req, err := http.NewRequest(action, reqURL, strings.NewReader(payload))
	if err != nil {
		return nil, "", err
	}

	req.Header.Add("Accept", "*/*")

	// Add any headers to the request
	for k := range headers {
		req.Header.Add(k, headers[k])
	}

	// Set host
	if host != "" {
		req.Host = host
	}

	// Set basic auth
	if reqUserName != "" && reqPassword != "" {
		req.SetBasicAuth(reqUserName, reqPassword)
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
func WaitForSendRequest(action, reqURL, host string, headers, parameterMap map[string]string, payload, reqUserName, reqPassword string, caData []byte, backoff wait.Backoff) (latestResponse *http.Response, latestResponseBody string, err error) {
	expectedStatusCode := http.StatusOK
	glog.V(7).Infof("Waiting for %s to reach status code %d...\n", reqURL, expectedStatusCode)
	startTime := time.Now()

	err = Retry(backoff, func() (bool, error) {
		response, responseBody, reqErr := SendRequest(action, reqURL, host, headers, parameterMap, payload, reqUserName, reqPassword, caData)
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
	glog.V(7).Infof("Wait time: %s \n", time.Since(startTime))
	return latestResponse, latestResponseBody, err
}
