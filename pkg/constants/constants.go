// Copyright (C) 2020, Oracle Corporation and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package constants

import (
	"time"
)

// Retry configs
const ResyncPeriod = 30 * time.Second
const RancherPollInterval = 30 * time.Second
const DefaultNamespace = "default"

// Naming
const VerrazzanoPrefix = "verrazzano"
const ManagedClusterPrefix = VerrazzanoPrefix + "-" + "managed-cluster"
const KubeconfigSecretKey = "kubeconfig"

// Labels and values
const VerrazzanoGroup = "verrazzano.oracle.com"
const K8SAppLabel = "k8s-app"
const VerrazzanoClusterLabel = "verrazzano.cluster"
