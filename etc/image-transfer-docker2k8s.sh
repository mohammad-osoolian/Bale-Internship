#!/bin/bash
docker save -o ./transfer-image.tar $1
cp ./transfer-image.tar /var/lib/docker/volumes/minikube/_data/data/
rm -f ./transfer-image.tar
docker exec -it minikube docker load -i /data/transfer-image.tar
rm -f /var/lib/docker/volumes/minikube/_data/data/transfer-image.tar
