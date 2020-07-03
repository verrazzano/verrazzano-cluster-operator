// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/controller"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util/logs"
)

var (
	masterURL       string
	kubeconfig      string
	watchNamespace  string
	rancherURL      string
	rancherHost     string
	rancherUserName string
	rancherPassword string
)

func main() {
	flag.Parse()

	if rancherURL == "" || rancherUserName == "" || rancherPassword == "" {
		glog.Fatalf("Rancher URL and/or credentials not specified!")
	}

	logs.InitLogs()

	glog.V(6).Infof("Creating new controller watching namespace %s.", watchNamespace)
	newController, err := controller.NewController(kubeconfig, masterURL, watchNamespace, rancherURL, rancherHost, rancherUserName, rancherPassword)
	if err != nil {
		glog.Fatalf("Error creating the controller: %s", err.Error())
	}

	if err = newController.Run(2); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&watchNamespace, "watchNamespace", "", "Optionally, a namespace to watch exclusively.  If not set, all namespaces will be watched.")
	flag.StringVar(&rancherURL, "rancherURL", "", "Rancher URL.")
	flag.StringVar(&rancherHost, "rancherHost", "", "Optional host name to use in host headers when accessing Rancher.")
	flag.StringVar(&rancherUserName, "rancherUserName", "", "Rancher username.")
	flag.StringVar(&rancherPassword, "rancherPassword", "", "Rancher password.")
}
