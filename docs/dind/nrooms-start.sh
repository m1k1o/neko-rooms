#!/bin/sh

#
# wait for docker to start
#

until docker ps
do
  echo "waiting for docker..."
  sleep 1
done

#
# create traefik network
#

docker network create --attachable traefik

#
# (re)start traefik
#

docker stop traefik
docker rm traefik
docker run -d \
    --restart unless-stopped \
    --name traefik \
    --network=traefik \
    -p "80:80" \
    -v "/var/run/docker.sock:/var/run/docker.sock:ro" \
    -e "TZ=${TZ}" \
    traefik:2.4 \
      --providers.docker=true \
      --providers.docker.watch=true \
      --providers.docker.exposedbydefault=false \
      --providers.docker.network=traefik \
      --entrypoints.web.address=:80;

# pull some neko images...
docker pull ghcr.io/m1k1o/neko/firefox
docker pull ghcr.io/m1k1o/neko/chromium

#
# (re)start neko-rooms
#

docker stop nrooms
docker rm nrooms
docker run -t \
    --restart unless-stopped \
    --name nrooms \
    --network=traefik \
    -v "/var/run/docker.sock:/var/run/docker.sock" \
    -v "/data:/data" \
    -e "TZ=${TZ}" \
    -e "NEKO_ROOMS_EPR=${NEKO_ROOMS_EPR}" \
    -e "NEKO_ROOMS_NAT1TO1=${NEKO_ROOMS_NAT1TO1}" \
    -e "NEKO_ROOMS_INSTANCE_URL=${NEKO_ROOMS_INSTANCE_URL}" \
    -e "NEKO_ROOMS_TRAEFIK_DOMAIN=*" \
    -e "NEKO_ROOMS_TRAEFIK_ENTRYPOINT=web" \
    -e "NEKO_ROOMS_TRAEFIK_NETWORK=traefik" \
    -e "NEKO_ROOMS_STORAGE_ENABLED=true" \
    -e "NEKO_ROOMS_STORAGE_INTERNAL=/data" \
    -e "NEKO_ROOMS_STORAGE_EXTERNAL=/data" \
    -l "traefik.enable=true" \
    -l "traefik.http.services.neko-rooms-frontend.loadbalancer.server.port=8080" \
    -l "traefik.http.routers.neko-rooms.entrypoints=web" \
    -l 'traefik.http.routers.neko-rooms.rule=HostRegexp(`{host:.+}`)' \
    -l 'traefik.http.routers.neko-rooms.priority=1' \
    m1k1o/neko-rooms:latest;
