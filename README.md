
# Verrazzano Rancher Operator

The Verrazzano Rancher Operator is an operator that interacts with Rancher Server on behalf of the Verrazzano System.

The responsibilities of the Verrazzano Rancher Operator are the following:
- Watch the Rancher API for available clusters, and creates VerrazzanoManagedCluster CRs to define these for the Super Domain Operator to consume.
- Call the Rancher Catalog API to ensure things prerequisites like Prometheus Operator and Istio are installed in the Managed Clusters.

Splitting the functionality that interacts with Rancher Server into its own repo reduces the complexity of the Super Domain Operator,
making it more nimble and easier to test.  Verrazzano Rancher Operator will require a Rancher Server instance to test again, but can
mock it out for unit testing purposes.

## Project Status

This repo is currently a skeleton, with details and tests stubbed out.  What it currently does:
- Sets up informers on k8s secrets and VerrazzanoManagedClusters - these are the k8s entities that Verrazzano Rancher Operator interacts with.
- Stubbed out calls to create/update secrets and VerrazzanoManagedClusters.
- Stubbed out section to poll Rancher server to discover available clusters.
- Stubbed out section to install prerequisites like Prometheus Operator and Istio.
- Stubbed out integ tests, and no unit tests yet.

## Artifacts

On a successful release (which occurs on a Git tag), this repo publises a Docker image:
- phx.ocir.io/stevengreenberginc/verrazzano/verrazzano-cluster-operator:tag

## Building

Go build:
```
make go-install
```

Docker build:
```
make build
```

Docker push:
```
make push
```

## Running

### Running locally

While developing, it's usually most efficient to run the Super Domain Operator as an out-of-cluster process,
pointing it to your Kubernetes cluster:

```
export KUBECONFIG=<your_kubeconfig>
make go-run
```

### Running in a Kubernetes Cluster

```
kubectl apply -f ./k8s/manifests/verrazzano-cluster-operator-serviceaccount.yaml
kubectl apply -f ./k8s/manifests/verrazzano-cluster-operator-deployment.yaml
```

**Note:** - if you don't intend to use the latest official Docker image, fill in your own Docker image in
`verrazzano-cluster-operator-deployment.yaml` above.

## Demo

TBD

## Development

### Running Tests

To run unit tests:

```
make unit-test
```

To run integration tests:

```
make integ-test
```
