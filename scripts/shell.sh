#!/bin/bash

if [ $# -lt 1 ]; then
  echo "Usage: $0 CONTAINER_NAME"
else
  docker exec -it $1 /bin/bash
fi
