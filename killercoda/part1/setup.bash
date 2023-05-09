#!/usr/bin/env bash
set -xeuo pipefail

git config --global user.email "lysefgt@lysefgt.com"
git config --global user.name  "lysefgt"
apt-get install -y llvm clang
