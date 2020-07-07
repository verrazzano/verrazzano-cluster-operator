// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Handles creation/deletion of VerrazzanoManagedClusters

package managedclusters

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/rancher"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/util/diff"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateVerrazzanoManagedCluster creates/updates a VerrazzanoManagedCluster resource
func CreateVerrazzanoManagedCluster(sdoClientSet sdoClientSet.Interface, tmcLister listers.VerrazzanoManagedClusterLister, cluster rancher.Cluster) error {
	// create logger for creating managed cluster
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Cluster").Str("name", cluster.Name).Logger()

	logger.Debug().Msgf("Processing VerrazzanoManagedCluster CR '%s' for cluster '%s'", cluster.ID, cluster.Name)

	// Construct the expected VerrazzanoManagedCluster
	newTmc := newVerrazzanoManagedCluster(cluster)

	existingTmc, err := tmcLister.VerrazzanoManagedClusters(constants.DefaultNamespace).Get(newTmc.Name)
	if existingTmc != nil {
		specDiffs := diff.CompareIgnoreTargetEmpties(existingTmc, newTmc)
		if specDiffs != "" {
			logger.Info().Msgf("Updating VerrazzanoManagedCluster CR '%s'", newTmc.Name)
			logger.Debug().Msgf("Spec differences:\n%s", specDiffs)
			newTmc.ResourceVersion = existingTmc.ResourceVersion
			_, err = sdoClientSet.VerrazzanoV1beta1().VerrazzanoManagedClusters(constants.DefaultNamespace).Update(context.TODO(), newTmc, metav1.UpdateOptions{})
		} else {
			logger.Debug().Msgf("No need to update existing VerrazzanoManagedCluster CR '%s'", newTmc.Name)
		}
	} else {
		logger.Info().Msgf("Creating VerrazzanoManagedCluster CR '%s'", newTmc.Name)
		_, err = sdoClientSet.VerrazzanoV1beta1().VerrazzanoManagedClusters(constants.DefaultNamespace).Create(context.TODO(), newTmc, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	logger.Debug().Msgf("Successfully processed VerrazzanoManagedCluster CR '%s' for cluster '%s'", cluster.ID, cluster.Name)
	return nil
}

// DeleteVerrazzanoManagedCluster deletes a VerrazzanoManagedCluster resource
func DeleteVerrazzanoManagedCluster(sdoClientSet sdoClientSet.Interface, tmcLister listers.VerrazzanoManagedClusterLister, cluster rancher.Cluster) error {
	// create logger for deleting managed cluster
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Cluster").Str("name", cluster.Name).Logger()

	logger.Debug().Msgf("Deleting VerrazzanoManagedCluster CR '%s' for cluster '%s'", cluster.ID, cluster.Name)

	_, err := tmcLister.VerrazzanoManagedClusters(constants.DefaultNamespace).Get(cluster.ID)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error().Msgf("VerrazzanoManagedCluster CR `%s` no longer exists for cluster '%s', for the reason (%v)", cluster.ID, cluster.Name, err)
		}
		return err
	}

	err = sdoClientSet.VerrazzanoV1beta1().VerrazzanoManagedClusters(constants.DefaultNamespace).Delete(context.TODO(), cluster.ID, metav1.DeleteOptions{})
	if err != nil {
		logger.Error().Msgf("Failed to delete VerrazzanoManagedCluster CR '%s' for cluster '%s', for the reason (%v)", cluster.ID, cluster.Name, err)
		return err
	}

	logger.Debug().Msgf("Successfully deleted VerrazzanoManagedCluster CR '%s' for cluster '%s'", cluster.ID, cluster.Name)
	return nil
}

// Constructs a VerrazzanoManagedCluster from the given Cluster
func newVerrazzanoManagedCluster(cluster rancher.Cluster) *v1beta1.VerrazzanoManagedCluster {
	return &v1beta1.VerrazzanoManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: constants.DefaultNamespace,
			Labels:    util.GetManagedClusterLabels(cluster.Name),
		},
		Spec: v1beta1.VerrazzanoManagedClusterSpec{
			KubeconfigSecret: util.GetManagedClusterKubeconfigSecretName(cluster.Name),
			ServerAddress:    cluster.ServerAddress,
			Type:             cluster.Type,
		},
	}
}
