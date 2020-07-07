// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package managedclusters

import (
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/rancher"
	"testing"
)

func TestNewVerrazzanoManagedCluster(t *testing.T) {
	cluster := rancher.Cluster{
		Id: "id",
		Name: "name",
		KubeConfigContents: "some stuff",
		ServerAddress: "123.123.123.0:1234",
		Type: "oke",
	}
	c := newVerrazzanoManagedCluster(cluster)

	if c.ObjectMeta.Name != "name" {
		t.Fatalf("expected ObjectMeta.Name to be %s, but got %s", "name", c.ObjectMeta.Name)
	}
	if c.ObjectMeta.Namespace != constants.DefaultNamespace {
		t.Fatalf("expected ObjectMeta.Namespace to be %s, but got %s", constants.DefaultNamespace, c.ObjectMeta.Namespace)
	}
	if c.Spec.ServerAddress != "123.123.123.0:1234" {
		t.Fatalf("expected Spec.ServerAddress to be %s, but got %s", "123.123.123.0:1234", c.Spec.ServerAddress)
	}
	if c.Spec.Type != "oke" {
		t.Fatalf("expected Spec.Type to be %s, but got %s", "oke", c.Spec.Type)
	}
}
