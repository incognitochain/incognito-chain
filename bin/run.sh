#!/usr/bin/env bash

docker -v || bash -c "wget -qO- https://get.docker.com/ | sh"
dataDir="/data"
ethDataDir="eth-data"

source miner.conf

if [  -z "$privatekey" ]; then
    echo "Private key not set!"
    exit
fi
if [  -z "$miner_port" ]; then
    echo "port not set!"
    exit
fi
if [  -z "$miner_rpc" ]; then
    echo "rpc not set!"
    exit
fi

clear="n"
read -p "Clear database?[y/n]"  clear

ip=`curl http://checkip.amazonaws.com/`
docker pull hungngoautonomous/incognito

docker rm -f constant_miner

docker run -e NAME=miner -p $miner_port:$miner_port -p $miner_rpc:$miner_rpc -e DISCOVERPEERSADDRESS='172.104.39.6:9330' -v /${dataDir}:/data -e PRIVATEKEY="${privatekey}" -e EXTERNALADDRESS="${ip}:9330"  -e PORT=$miner_port -e RPC_PORT=$miner_rpc -d --name constant_miner hungngoautonomous/incognito /run_constant.sh $clear


runETHLightNode="n"
read -p "Run ethereum light node? [y/n] "  runETHLightNode
if [ $runETHLightNode == "y" ]; then
    docker pull ethereum/client-go
    docker run -it -p 8545:8545 -d -v /${ethDataDir}:/root/.ethereum ethereum/client-go --testnet --syncmode "light" --maxpeers 0 --rpc --rpcaddr "0.0.0.0" --rpcapi db,eth,net,web3,personal,admin --rpccorsdomain "*"
fi
