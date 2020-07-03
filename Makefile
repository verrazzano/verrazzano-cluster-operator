# Copyright (C) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

NAME:=verrazzano-cluster-operator
CLUSTER_NAME = v8o-cluster-operator

DOCKER_IMAGE_NAME ?= ${NAME}-dev
TAG=$(shell git rev-parse HEAD)
DOCKER_IMAGE_TAG = ${TAG}

CREATE_LATEST_TAG=0

ifeq ($(MAKECMDGOALS),$(filter $(MAKECMDGOALS),push push-tag))
ifndef DOCKER_REPO
    $(error DOCKER_REPO must be defined as the name of the docker repository where image will be pushed)
endif
ifndef DOCKER_NAMESPACE
    $(error DOCKER_NAMESPACE must be defined as the name of the docker namespace where image will be pushed)
endif
    DOCKER_IMAGE_FULLNAME = ${DOCKER_REPO}/${DOCKER_NAMESPACE}/${DOCKER_IMAGE_NAME}
endif


DIST_DIR:=dist
K8S_NAMESPACE:=default
WATCH_NAMESPACE:=
EXTRA_PARAMS=
INTEG_RUN_ID=
ENV_NAME=verrazzano-cluster-operator
GO ?= GO111MODULE=on GOPRIVATE=github.com/verrazzano go

.PHONY: all
all: build

#
# Go build related tasks
#
.PHONY: go-install
go-install: go-mod
	git config core.hooksPath hooks
	$(GO) install ./cmd/...

.PHONY: go-run
go-run: go-install
	$(GO) run cmd/main.go --kubeconfig=${KUBECONFIG} --v=4 --watchNamespace=${WATCH_NAMESPACE} ${EXTRA_PARAMS}

.PHONY: go-fmt
go-fmt:
	gofmt -s -e -d $(shell find . -name "*.go" | grep -v /vendor/)

.PHONY: go-vet
go-vet:
	echo go vet $(shell go list ./... | grep -v /vendor/)

.PHONY: go-mod
go-mod:
	$(GO) mod vendor

#
# Docker-related tasks
#
.PHONY: docker-clean
docker-clean:
	rm -rf ${DIST_DIR}

.PHONY: build
build: go-mod
	docker build --pull --no-cache \
		-t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .

.PHONY: push
push: build
	docker tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}
	docker push ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}

	if [ "${CREATE_LATEST_TAG}" == "1" ]; then \
		docker tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:latest; \
		docker push ${DOCKER_IMAGE_FULLNAME}:latest; \
	fi

.PHONY: push-tag
push-tag:
	docker pull ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}
	docker tag ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:${TAG_NAME}
	docker push ${DOCKER_IMAGE_FULLNAME}:${TAG_NAME}

#
# Tests-related tasks
#
.PHONY: unit-test
unit-test: go-install
	$(GO) test -v ./pkg/... ./cmd/...


PHONY: thirdparty-check
thirdparty-check:
	./build/scripts/thirdparty_check.sh

.PHONY: coverage
coverage:
	./build/scripts/coverage.sh html

.PHONY: integ-test
integ-test: go-install build
	echo 'Install KinD...'
	GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0
ifdef JENKINS_URL
	./build/scripts/cleanup.sh ${CLUSTER_NAME}
endif
	echo 'Create cluster...'
	time kind create cluster \
	    --name ${CLUSTER_NAME} \
	    --wait 5m \
		--config=test/kind-config.yaml

	kubectl config set-context kind-${CLUSTER_NAME}
ifdef JENKINS_URL
	cat ${HOME}/.kube/config | grep server
	# this ugly looking line of code will get the ip address of the container running the kube apiserver
	# and update the kubeconfig file to point to that address, instead of localhost
	sed -i -e "s|127.0.0.1.*|`docker inspect ${CLUSTER_NAME}-control-plane | jq '.[].NetworkSettings.IPAddress' | sed 's/"//g'`:6443|g" ${HOME}/.kube/config
	cat ${HOME}/.kube/config | grep server
endif
	kubectl cluster-info

	kubectl wait --for=condition=ready nodes --all
	kubectl get nodes
	echo 'Copy operator Docker image into KinD...'
	kind load --name ${CLUSTER_NAME} docker-image ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

	kubectl apply -f deploy/service_account.yaml
	kubectl apply -f deploy/role.yaml
	kubectl apply -f deploy/role_binding.yaml

	echo 'Deploy operator...'
	cat deploy/operator.yaml | sed -e 's|REPLACE_IMAGE|${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}|g' | kubectl apply -f -
	echo 'Run tests...'
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u github.com/onsi/gomega/...

	ginkgo -v --keepGoing -cover test/integ/... || IGNORE=FAILURE

#
# Cleanup Kind cluster and docker containers
#
.PHONY: clean-cluster
clean-cluster:
	./build/scripts/cleanup.sh ${CLUSTER_NAME}

#
# Kubernetes-related tasks
#
.PHONY: k8s-deploy
k8s-deploy:
	mkdir -p ${DIST_DIR}/manifests
	cp -r k8s/manifests/* $(DIST_DIR)/manifests

	# Fill in Docker image and tag that's being tested
	sed -i.bak "s|${DOCKER_REPO}/${DOCKER_NAMESPACE}/${NAME}:latest|${DOCKER_REPO}/${DOCKER_NAMESPACE}/${DOCKER_IMAGE_NAME}:$(DOCKER_IMAGE_TAG)|g" $(DIST_DIR)/manifests/verrazzano-cluster-operator-deployment.yaml
	sed -i.bak "s|namespace: default|namespace: ${K8S_NAMESPACE}|g" $(DIST_DIR)/manifests/verrazzano-cluster-operator-deployment.yaml
	sed -i.bak "s|namespace: default|namespace: ${K8S_NAMESPACE}|g" $(DIST_DIR)/manifests/verrazzano-cluster-operator-serviceaccount.yaml
	kubectl apply -f ${DIST_DIR}/manifests
