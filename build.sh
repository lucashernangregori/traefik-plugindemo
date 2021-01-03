#/bin/bash
apt update
apt install -y docker.io
docker build -t traefik_test:1.0 .

mkdir log
docker run -d -p 8080:8080 -p 80:80 \
-v $PWD/traefik/traefik.yml:/etc/traefik/traefik.yml \
-v $PWD/traefik/dynamic.yml:/etc/traefik/dynamic.yml \
-v $PWD/log:/var/log \
-v /var/run/docker.sock:/var/run/docker.sock \
--name traefikt \
traefik_test:1.0