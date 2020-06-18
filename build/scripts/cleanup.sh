#!/bin/bash
#
# Cleanup Kind cluster and docker containers
# delete kind cluster if it exists
echo "Doing cleanup of KIND clusters and Docker containers"
kind delete cluster --name "$1"

containers=$(docker ps -q --filter label=io.x-k8s.kind.cluster | wc -l)
if [ "$containers" -gt 1 ]
then
  # stop all running Kind containers
  echo "Stopping Kind Containers ..."
  docker stop $(docker ps -q --filter label=io.x-k8s.kind.cluster)

  echo "Deleting Kind Containers ..."
  # delete all Kind containers
  docker rm $(docker ps -aq --filter label=io.x-k8s.kind.cluster)
fi

