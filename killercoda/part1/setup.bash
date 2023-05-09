#!/usr/bin/env bash
set -xeuo pipefail

git config --global user.email "lysefgt@lysefgt.com"
git config --global user.name  "lysefgt"

wget https://github.com/libbpf/bpftool/releases/download/v7.2.0/bpftool-v7.2.0-amd64.tar.gz
tar -xzvf bpftool-v7.2.0-amd64.tar.gz
chmod +x bpftool
mv bpftool /usr/sbin/bpftool
rm -rf bpftool-v7.2.0-amd64.tar.gz

apt-get install -y llvm clang #linux-oem-5.6-tools-common libelf-dev
