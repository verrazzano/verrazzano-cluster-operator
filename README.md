# Verrazzano Cluster Operator

The Verrazzano Cluster Operator is an operator that interacts with Kubernetes clusters in the Verrazzano
environment and with the Rancher Server on behalf of the Verrazzano system.

The responsibilities of the Cluster Rancher Operator are the following:
- Watch the Rancher API for available clusters, and creates `VerrazzanoManagedCluster` custopm resources to define
  these for the Super Domain Operator to consume.
- Call the Rancher Catalog API to ensure things prerequisites like Prometheus Operator and Istio are installed in
  the Managed Clusters.

## Project Status

This repo is currently an ALPHA, with some details and tests stubbed out.  What it currently does:
- Sets up informers on k8s secrets and `VerrazzanoManagedClusters` - these are the k8s entities that Verrazzano Rancher Operator interacts with.
- Stubbed out calls to create/update secrets and `VerrazzanoManagedClusters`.
- Stubbed out section to poll Rancher server to discover available clusters.
- Stubbed out section to install prerequisites like Prometheus Operator and Istio.
- Stubbed out integ tests, and no unit tests yet.

## Artifacts

On a successful release (which occurs on a Git tag), this repo publises a Docker image:
- verrazzano-cluster-operator:tag

## Building

To build this operator:

* Go build:
    ```
    make go-install
    ```

* Docker build:
    ```
    make build
    ```

* Docker push:
    ```
    make push
    ```

## Running

### Running locally

The Verrazzano Cluster Operator can be run as an out-of-cluster process during development:

```
export KUBECONFIG=<your_kubeconfig>
make go-run
```

### Running in a Kubernetes Cluster

To deploy the Verrazzano Cluster Operator to a Kubernetes cluster:

```
kubectl apply -f ./k8s/manifests/verrazzano-cluster-operator-serviceaccount.yaml
kubectl apply -f ./k8s/manifests/verrazzano-cluster-operator-deployment.yaml
```

**Note** If you want to use your own Docker image, update the image in the
`verrazzano-cluster-operator-deployment.yaml` file first.

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

## Contributing to Verrazzano

Oracle welcomes contributions to this project from anyone.  Contributions may be reporting an issue with the operator or submitting a pull request.  Before embarking on significant development that may result in a large pull request, it is recommended that you create an issue and discuss the proposed changes with the existing developers first.

If you want to submit a pull request to fix a bug or enhance an existing feature, please first open an issue and link to that issue when you submit your pull request.

If you have any questions about a possible submission, feel free to open an issue too.

## Contributing to the Verrazzano Cluster Operator repository

Pull requests can be made under The Oracle Contributor Agreement (OCA), which is available at [https://www.oracle.com/technetwork/community/oca-486395.html](https://www.oracle.com/technetwork/community/oca-486395.html).

For pull requests to be accepted, the bottom of the commit message must have the following line, using the contributor’s name and e-mail address as it appears in the OCA Signatories list.

```
Signed-off-by: Your Name <you@example.org>
```

This can be automatically added to pull requests by committing with:

```
git commit --signoff
```

Only pull requests from committers that can be verified as having signed the OCA can be accepted.

## Pull request process

*	Fork the repository.
*	Create a branch in your fork to implement the changes. We recommend using the issue number as part of your branch name, for example, `1234-fixes`.
*	Ensure that any documentation is updated with the changes that are required by your fix.
*	Ensure that any samples are updated if the base image has been changed.
*	Submit the pull request. Do not leave the pull request blank. Explain exactly what your changes are meant to do and provide simple steps on how to validate your changes. Ensure that you reference the issue you created as well. We will assign the pull request to 2-3 people for review before it is merged.

## Introducing a new dependency

Please be aware that pull requests that seek to introduce a new dependency will be subject to additional review.  In general, contributors should avoid dependencies with incompatible licenses, and should try to use recent versions of dependencies.  Standard security vulnerability checklists will be consulted before accepting a new dependency.  Dependencies on closed-source code, including WebLogic Server, will most likely be rejected.
