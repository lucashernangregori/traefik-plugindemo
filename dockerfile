FROM traefik:v2.3.6
COPY ./traefik /home/ubuntu/traefik
COPY ./plugin /home/ubuntu/traefik/plugins/src/plugindemo