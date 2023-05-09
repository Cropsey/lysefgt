#!/usr/bin/env bash
set -xeuo pipefail

echo waiting for environment setup
while [ ! -f /tmp/ready ]; do sleep 1; done
kubectl wait --selector=app.kubernetes.io/component=server --for=condition=ready pod
