// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package util

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	gabs "github.com/Jeffail/gabs/v2"
	"github.com/rs/zerolog"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

func GetManagedClusterKubeconfigSecretName(clusterId string) string {
	return fmt.Sprintf("%s-%s", constants.ManagedClusterPrefix, clusterId)
}

func GetManagedClusterLabels(managedClusterId string) map[string]string {
	return map[string]string{constants.K8SAppLabel: constants.VerrazzanoGroup, constants.VerrazzanoClusterLabel: managedClusterId}
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
	if proxyUrl := os.Getenv("https_proxy"); proxyUrl != "" {
		return proxyUrl
	}
	if proxyUrl := os.Getenv("HTTPS_PROXY"); proxyUrl != "" {
		return proxyUrl
	}
	if proxyUrl := os.Getenv("http_proxy"); proxyUrl != "" {
		return proxyUrl
	}
	if proxyUrl := os.Getenv("HTTP_PROXY"); proxyUrl != "" {
		return proxyUrl
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

// Send http request
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

// WaitForSendRequest:  Waits for the given request to return results
func WaitForSendRequest(action, reqURL, host string, headers, parameterMap map[string]string, payload, reqUserName, reqPassword string, caData []byte, backoff wait.Backoff) (latestResponse *http.Response, latestResponseBody string, err error) {
	// create logger for request results
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Request").Str("name", host).Logger()

	expectedStatusCode := http.StatusOK
	logger.Debug().Msgf("Waiting for %s to reach status code %d...\n", reqURL, expectedStatusCode)
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
	logger.Debug().Msgf("Wait time: %s \n", time.Since(startTime))
	return latestResponse, latestResponseBody, err
}

//
// JSon helper
//

// Retrieves the current resource from k8s as a JSON entity
func GetJson(restClient rest.RESTClient, resource string, namespace string, name string) (*gabs.Container, error) {
	result, err := restClient.Get().Resource(resource).Namespace(namespace).Name(name).Do(context.TODO()).Raw()
	if err != nil {
		return nil, err
	}
	json, err := gabs.ParseJSON(result)
	if err != nil {
		return nil, err
	}
	return json, nil
}

// Sets the given attribute from the given resource JSON entity
func SetJsonAttr(json *gabs.Container, path string, value interface{}) {
	// create logger for Json attribute setting
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "JsonAttr").Str("name", path).Logger()

	if _, err := json.SetP(value, path); err != nil {
		logger.Error().Msgf("Error setting '%s=%s' on %v: %v", path, value, json, err)
	}
}

// Deletes the given attribute from the given resource JSON entity
func DeleteJsonAttr(json *gabs.Container, path string) {
	// create logger for Json attribute setting
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "JsonAttr").Str("name", path).Logger()

	if err := json.DeleteP(path); err != nil {
		logger.Error().Msgf("Error deleteing '%s' from %v: %v", path, json, err)
	}
}

// Converts contained object back to a pretty JSON formatted string
func GetPrettyJson(json *gabs.Container) string {
	return string(json.EncodeJSON(gabs.EncodeOptIndent("", "\t")))
}
