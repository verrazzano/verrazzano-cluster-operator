// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package controller

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog"
	clientset "github.com/verrazzano/verrazzano-crd-generator/pkg/client/clientset/versioned"
	clientsetscheme "github.com/verrazzano/verrazzano-crd-generator/pkg/client/clientset/versioned/scheme"
	informers "github.com/verrazzano/verrazzano-crd-generator/pkg/client/informers/externalversions"
	listers "github.com/verrazzano/verrazzano-crd-generator/pkg/client/listers/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/constants"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/managedclusters"
	"github.com/verrazzano/verrazzano-cluster-operator/pkg/rancher"
	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	extclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

const controllerAgentName = "verrazzano-rancher-controller"

// Primary controller structure
type Controller struct {
	kubeClientSet        kubernetes.Interface
	kubeExtClientSet     apiextensionsclient.Interface
	superDomainClientSet clientset.Interface

	// Local cluster listers and informers
	secretLister                     corev1listers.SecretLister
	secretInformer                   cache.SharedIndexInformer
	verrazzanoManagedClusterLister   listers.VerrazzanoManagedClusterLister
	verrazzanoManagedClusterInformer cache.SharedIndexInformer

	// Rancher cluster
	rancherConfig rancher.Config

	// Misc
	watchNamespace string
	stopCh         <-chan struct{}

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new Super Domain Operator controller
func NewController(kubeconfig string, masterURL string, watchNamespace string, rancherURL string, rancherHost string, rancherUsername string, rancherPassword string) (*Controller, error) {
	//
	// Instantiate connection and clients to local k8s cluster
	//

	// create logger for new controller
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Rancher").Str("name", rancherHost).Logger()

	logger.Debug().Msg("Building config")
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logger.Fatal().Msgf("Error building kubeconfig: %v", err)
	}

	logger.Debug().Msg("Building kubernetes clientset")
	kubeClientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Fatal().Msgf("Error building kubernetes clientset: %v", err)
	}

	logger.Debug().Msg("Building kubernetes apiextensions apiserver clientset")
	kubeExtClientSet, err := extclientset.NewForConfig(cfg)
	if err != nil {
		logger.Fatal().Msgf("Error building kubernetes apiextensions apiserverclientset: %v", err)
	}

	logger.Debug().Msg("Building superdomain clientset")
	superDomainClientSet, err := clientset.NewForConfig(cfg)
	if err != nil {
		logger.Fatal().Msgf("Error building superdomain clientset: %v", err)
	}

	//
	// Set up informers and listers for the local k8s cluster
	//
	var kubeInformerFactory kubeinformers.SharedInformerFactory
	var superDomainInformerFactory informers.SharedInformerFactory
	if watchNamespace == "" {
		// Consider all namespaces if our namespace is left wide open our set to default
		kubeInformerFactory = kubeinformers.NewSharedInformerFactory(kubeClientSet, constants.ResyncPeriod)
		superDomainInformerFactory = informers.NewSharedInformerFactory(superDomainClientSet, constants.ResyncPeriod)

	} else {
		// Otherwise, restrict to a specific namespace
		kubeInformerFactory = kubeinformers.NewFilteredSharedInformerFactory(kubeClientSet, constants.ResyncPeriod, watchNamespace, nil)
		superDomainInformerFactory = informers.NewFilteredSharedInformerFactory(superDomainClientSet, constants.ResyncPeriod, watchNamespace, nil)
	}
	secretsInformer := kubeInformerFactory.Core().V1().Secrets()
	verrazzanoManagedClusterInformer := superDomainInformerFactory.Verrazzano().V1beta1().VerrazzanoManagedClusters()

	clientsetscheme.AddToScheme(scheme.Scheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Info().Msgf)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	// If the Rancher host name is explicitly passed in, we'll use that, otherwise we'll just use the host name from the
	// URL.  Having the rancherHost parameter allows rancher-operator to work in a Verrazzano environment without external DNS.
	if rancherHost == "" {
		rancherUrlObj, err := url.Parse(rancherURL)
		if err != nil {
			logger.Fatal().Msgf("Invalid Rancher URL '%s': %v", rancherURL, err)
		}
		rancherHost = rancherUrlObj.Host
	}
	rancherConfig := rancher.Config{rancherURL,
		rancherUsername,
		rancherPassword,
		rancherHost,
		managedclusters.GetRancherCACert(kubeClientSet)}

	controller := &Controller{
		rancherConfig:                    rancherConfig,
		watchNamespace:                   watchNamespace,
		kubeClientSet:                    kubeClientSet,
		kubeExtClientSet:                 kubeExtClientSet,
		superDomainClientSet:             superDomainClientSet,
		secretLister:                     secretsInformer.Lister(),
		secretInformer:                   secretsInformer.Informer(),
		verrazzanoManagedClusterLister:   verrazzanoManagedClusterInformer.Lister(),
		verrazzanoManagedClusterInformer: verrazzanoManagedClusterInformer.Informer(),
		recorder: recorder,
	}

	// Set up signals so we handle the first shutdown signal gracefully
	logger.Debug().Msg("Setting up signals")
	stopCh := make(chan struct{})

	go kubeInformerFactory.Start(stopCh)
	go superDomainInformerFactory.Start(stopCh)

	return controller, nil
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
//
func (c *Controller) Run(threadiness int) error {
	defer runtime.HandleCrash()

	// create logger for running
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Rancher").Str("name", "handler").Logger()

	// Start the informer factories to begin populating the informer caches
	logger.Info().Msg("Starting Verrazzano Rancher controller")

	// Wait for the caches to be synced before starting watchers
	logger.Info().Msg("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(c.stopCh, c.secretInformer.HasSynced, c.verrazzanoManagedClusterInformer.HasSynced); !ok {
		return errors.New("failed to wait for caches to sync")
	}

	logger.Info().Msg("Starting watchers")

	c.secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(new interface{}) { c.processRancherSecret(new.(*corev1.Secret)) },
		UpdateFunc: func(old, new interface{}) { c.processRancherSecret(new.(*corev1.Secret)) },
	})

	go c.startRancherWatcher(c.stopCh)

	<-c.stopCh
	return nil
}

// if the secret cattle-system/tls-rancher-ingressis updated, update CertificateAuthorityData in rancherConfig
func (c *Controller) processRancherSecret(newSecret *corev1.Secret) {
	// create logger for secret processor
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Rancher").Str("name", "secretProcessor").Logger()

	if newSecret.Name == rancher.TlsRancherIngressSecret &&
		newSecret.Namespace == rancher.RancherNamespace &&
		bytes.Compare(newSecret.Data["ca.crt"], c.rancherConfig.CertificateAuthorityData) != 0 {
		logger.Info().Msgf("Reloading secret %s/%s...", newSecret.Namespace, newSecret.Name)
		c.rancherConfig.CertificateAuthorityData = newSecret.Data["ca.crt"]
	}
}

// Start polling the Rancher Server for updates
func (c *Controller) startRancherWatcher(<-chan struct{}) {
	// create logger for rancher watcher
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Rancher").Str("name", "watcher").Logger()

	for {
		clusters, err := rancher.GetClusters(rancher.Rancher{}, c.rancherConfig)
		if err != nil {
			logger.Error().Msgf("Failed to get Rancher managed clusters: %v", err)
		} else {
			for _, cluster := range clusters {
				logger.Info().Msgf("Syncing Verrazzano Managed Cluster: Id='%s', Name='%s'", cluster.Id, cluster.Name)

				// Generate the resources to inform the Super Domain Operator about this cluster
				c.generateSuperDomainOperatorResources(cluster)

				logger.Info().Msgf("Successfully synced Verrazzano Managed Cluster: Id='%s', Name='%s'", cluster.Id, cluster.Name)
			}
			logger.Info().Msgf("Successfully synced Rancher.")
		}

		// Check available clusters every perdefined interval in seconds
		<-time.After(constants.RancherPollInterval)
	}
}

// Generates the resources used by the Super Domain Operator for the given cluster
func (c *Controller) generateSuperDomainOperatorResources(cluster rancher.Cluster) {
	/*********************
	 * Create or Update VerrazzanoManagedClusters Secret if needed
	 **********************/

	// create logger for rancher watcher
	logger := zerolog.New(os.Stderr).With().Timestamp().Str("kind", "Cluster").Str("name", cluster.Name).Logger()

	err := managedclusters.CreateSecret(c.kubeClientSet, c.secretLister, cluster)
	if err != nil {
		logger.Error().Msgf("Failed to create/update VerrazzanoManagedCluster Secret for cluster %s, for the reason (%v)", cluster.Name, err)
	}

	/*********************
	 * Create or Update VerrazzanoManagedClusters if needed
	 **********************/
	err = managedclusters.CreateVerrazzanoManagedCluster(c.superDomainClientSet, c.verrazzanoManagedClusterLister, cluster)
	if err != nil {
		logger.Error().Msgf("Failed to create/update VerrazzanoManagedCluster CR for cluster %s, for the reason (%v)", cluster.Name, err)
	}
}

// Configures cluster prereqs via the Rancher API
func (c *Controller) configureClusterPrereqs(cluster rancher.Cluster) {
}
