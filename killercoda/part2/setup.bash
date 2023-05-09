#!/usr/bin/env bash
set -xeuo pipefail

kubectl create -f /root/assets/prometheus.yaml
echo done > /tmp/ready
