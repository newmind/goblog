#!/bin/bash

# docker service rm viz
# docker service create \
#   --name=viz \
#   --publish=8080:8000/tcp \
#   --constraint=node.role==manager \
#   --mount=type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock \
#   manomarks/visualizer

docker service rm dvizz  
docker service create \
  --constraint node.role==manager \
  --replicas 1 --name dvizz -p 6969:6969 \
  --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
  --network my_network \
  eriklupander/dvizz