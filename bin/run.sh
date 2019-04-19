#!/usr/bin/env bash

docker -v || bash -c "wget -qO- https://get.docker.com/ | sh"

if [  -z "$privatekey" ]; then
    echo "Private key not set!"
    echo "Type: privatekey=your_private_key"
    echo "Then run again"
    exit
fi

clear="n"
read -p "Clear database?[y/n]"  clear

ip=`curl http://checkip.amazonaws.com/`
docker rm -f constant_miner

docker run -e NAME=miner -p 9330:9330 -p 9334:9334 -e DISCOVERPEERSADDRESS='172.104.39.6:9330' -v /data:/data -e PRIVATEKEY="${privatekey}" -e EXTERNALADDRESS="${ip}:9330"  -e PORT=9330 -e RPC_PORT=9334 -d --name constant_miner dungvanautonomous/constant /run_constant.sh $clear
