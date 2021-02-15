// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package util

import (
	"fmt"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
)

// GetManagedClusterKubeconfigSecretName returns the secret for a managed cluster
func GetManagedClusterKubeconfigSecretName(clusterID string) string {
	return fmt.Sprintf("%s-%s", constants.ManagedClusterPrefix, clusterID)
}

// GetManagedClusterLabels return labels for a managed cluster
func GetManagedClusterLabels(managedClusterID string) map[string]string {
	return map[string]string{constants.K8SAppLabel: constants.VerrazzanoGroup, constants.VerrazzanoClusterLabel: managedClusterID}
}
