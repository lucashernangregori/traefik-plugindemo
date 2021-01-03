#/bin/bash
apt update
apt install -y docker.io
docker build -t traefik_test:1.0 .