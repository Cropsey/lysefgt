#!/usr/bin/env bash
set -xeuo pipefail

echo "deploy registry and prometheus"
kubectl create -f /root/assets/registry.yaml
kubectl create -f /root/assets/prometheus.yaml
kubectl create -f /root/assets/grafana.yaml

registry_ip=$(kubectl get svc -nkube-system -o json registry | jq --raw-output '.spec.clusterIP')
echo "Updating internal registry configs - $registry_ip"
echo "$registry_ip registry.kube-system.svc" >> /etc/hosts
cat /etc/docker/daemon.json | jq '. += {"insecure-registries": ["http://registry.kube-system.svc:5000"]}' > /tmp/dd.json
mv /tmp/dd.json /etc/docker/daemon.json
cat << EOF >> /etc/containerd/config.toml
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry.kube-system.svc:5000"]
  endpoint = ["http://registry.kube-system.svc:5000"]
EOF
systemctl restart docker.service containerd.service

echo "install bpftool"

wget https://github.com/libbpf/bpftool/releases/download/v7.2.0/bpftool-v7.2.0-amd64.tar.gz
tar -xzvf bpftool-v7.2.0-amd64.tar.gz
chmod +x bpftool
mv bpftool /usr/sbin/bpftool
rm -rf bpftool-v7.2.0-amd64.tar.gz

echo "install llvm and clang packages"
apt-get install -y llvm clang

echo "waiting for registry and prometheus to be up"
kubectl wait --selector=app=registry --for=condition=ready -n kube-system pod --timeout=180s
kubectl wait --selector=app.kubernetes.io/component=server --for=condition=ready pod --timeout=180s

echo "READY"
