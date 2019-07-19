#!/usr/bin/env bash

docker -v || bash -c "wget -qO- https://get.docker.com/ | sh"
dataDir="data"
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
docker pull incognitochain/incognito

docker rm -f inc_miner
docker rm -f inc_geth

# docker run -e NAME=miner -p $miner_port:$miner_port -p $miner_rpc:$miner_rpc -e DISCOVERPEERSADDRESS='172.104.39.6:9330' -v /${dataDir}:/data -e PRIVATEKEY="${privatekey}" -e EXTERNALADDRESS="${ip}:9330"  -e PORT=$miner_port -e RPC_PORT=$miner_rpc -d --name constant_miner hungngoautonomous/incognito /run_constant.sh $clear


docker network create --driver bridge inc_net || true

docker run -it --net inc_net -p 8545:8545 -p 30303:30303 -d -v $PWD/${ethDataDir}:/root/.ethereum --name inc_geth ethereum/client-go --testnet --syncmode "light" --maxpeers 0 --rpc --rpcaddr "0.0.0.0" --rpcapi db,eth,net,web3,personal,admin --rpccorsdomain "*" --rpcvhosts="*"

docker run --net inc_net -d -e GETH_NAME=inc_geth -e NAME=miner -p $miner_port:$miner_port -p $miner_rpc:$miner_rpc -e DISCOVERPEERSADDRESS='172.104.39.6:9330' -v $PWD/${dataDir}:/data -e PRIVATEKEY="${privatekey}" -e EXTERNALADDRESS="${ip}:9330" -e PORT=$miner_port -e RPC_PORT=$miner_rpc --name inc_miner incognitochain/incognito /run_incognito.sh $clear
