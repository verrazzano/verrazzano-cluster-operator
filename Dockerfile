# Copyright (C) 2020, Oracle Corporation and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

FROM oraclelinux:7-slim AS build_base

RUN yum update -y \
    && yum-config-manager --save --setopt=ol7_ociyum_config.skip_if_unavailable=true \
    && yum install -y oracle-golang-release-el7 \
    && yum-config-manager --add-repo http://yum.oracle.com/repo/OracleLinux/OL7/developer/golang113/x86_64 \
    && yum install -y git gcc make golang-1.13.1-1.el7 \
    && yum clean all \
    && go version

# Compile to /usr/bin
ENV GOBIN=/usr/bin

# Set go path
ENV GOPATH=/go

ARG BUILDVERSION
ARG BUILDDATE


# Need to use specific WORKDIR to match verrazzano-cluster-operator's source packages
WORKDIR /root/go/src/github.com/verrazzano/verrazzano-cluster-operator
COPY . .

ENV CGO_ENABLED 0
RUN go version
RUN go env

RUN GO111MODULE=on go build \
    -mod=vendor \
    -ldflags '-extldflags "-static"' \
    -ldflags "-X main.buildVersion=${BUILDVERSION} -X main.buildDate=${BUILDDATE}" \
    -o /usr/bin/verrazzano-cluster-operator ./cmd

FROM oraclelinux:7-slim

RUN yum install -y ca-certificates curl openssl && yum clean all && rm -rf /var/cache/yum

COPY --from=build_base /usr/bin/verrazzano-cluster-operator /usr/local/bin/verrazzano-cluster-operator

# Copy source tree to image
RUN mkdir -p go/src/github.com/verrazzano/verrazzano-cluster-operator
COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano-cluster-operator go/src/github.com/verrazzano/verrazzano-cluster-operator

WORKDIR /usr/local/bin/
RUN groupadd -r verrazzano-cluster-operator && useradd --no-log-init -r -g verrazzano-cluster-operator -u 1000 verrazzano-cluster-operator
RUN chown 1000:verrazzano-cluster-operator /usr/local/bin/verrazzano-cluster-operator && chmod 500 /usr/local/bin/verrazzano-cluster-operator
USER 1000

ENTRYPOINT ["/usr/local/bin/verrazzano-cluster-operator"]
