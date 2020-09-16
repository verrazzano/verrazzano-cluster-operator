// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package constants

import (
	"time"
)

// ResyncPeriod is interval when informer is resynced
const ResyncPeriod = 30 * time.Second

// RancherPollInterval is interval to poll Rancher Server for updates
const RancherPollInterval = 30 * time.Second

// DefaultNamespace is constant for the default namespace
const DefaultNamespace = "default"

// ManagedClusterPrefix is the constant for the managed cluster prefix
const ManagedClusterPrefix = "verrazzano-managed-cluster"

// KubeconfigSecretKey is the constant for kubeconfig
const KubeconfigSecretKey = "kubeconfig"

// VerrazzanoGroup is the constant for the Verrazzano group
const VerrazzanoGroup = "verrazzano.oracle.com"

// K8SAppLabel is the constant for k8s-app label
const K8SAppLabel = "k8s-app"

// VerrazzanoClusterLabel is the constant the verrazzano.cluster label
const VerrazzanoClusterLabel = "verrazzano.cluster"
