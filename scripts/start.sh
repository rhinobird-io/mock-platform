#!/bin/bash

set -eo pipefail
export HOME=$(cd `dirname $0`; cd ..; pwd)
export GW_NAME="tw-gateway"
export PLAT_NAME="tw-platform"
export PLATFORM_REPO=${PLATFORM_REPO:-"http://gitlab.ate-sh.worksdev:8081/team-work/team-work.git"}
export PLATFORM_DIR=${HOME}/docker/platform/repo

echo -e "Starting ${GW_NAME}..."

DEBIAN_IMAGE="debian:wheezy" && sudo docker history $DEBIAN_IMAGE > /dev/null || sudo docker pull $DEBIAN_IMAGE
sudo docker inspect $GW_NAME >/dev/null && sudo docker rm -f $GW_NAME > /dev/null || true
sudo docker run -d --name $GW_NAME -v "$HOME":/gateway -w /gateway -p 8000:8000 $DEBIAN_IMAGE ./gateway
export GW_ADDR=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' ${GW_NAME})

echo -e "Starting ${PLAT_NAME}..."

if [ -d "$PLATFORM_DIR" ]; then
  (
    cd $PLATFORM_DIR
    git pull --rebase origin master
  )
else
  git clone "$PLATFORM_REPO" "$PLATFORM_DIR" 
fi

(
  cd $PLATFORM_DIR/public
  bower install
)

RUBY_IMAGE="ruby:2.1" && sudo docker history $RUBY_IMAGE > /dev/null || sudo docker pull $RUBY_IMAGE
sudo docker inspect $PLAT_NAME > /dev/null && sudo docker rm -f $PLAT_NAME > /dev/null || true
(
  cd $PLATFORM_DIR/../
  sudo docker build -t $PLAT_NAME .
)
sudo docker run -it --rm --name $PLAT_NAME -v "$PLATFORM_DIR":/app -w /app $PLAT_NAME rake db:migrate
sudo docker run -d --name $PLAT_NAME -v "$PLATFORM_DIR":/app -w /app -p 4567:4567 -e "AUTH_URL=http://${GW_ADDR}:8000/auth" $PLAT_NAME ruby app.rb

