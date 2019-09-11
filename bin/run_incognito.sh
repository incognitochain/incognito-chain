#!/bin/sh
mkdir -p /data
if [ "$1" == "y" ]; then
    find /data -maxdepth 1 -mindepth 1 -type d | xargs rm -rf
fi

if [ -z $NAME ]; then
    NAME="miner";
fi

if [ -z $BOOTNODE_IP ]; then
    BOOTNODE_IP="157.245.128.19:9330";
fi

if [ -z $NODE_PORT ]; then
    NODE_PORT=9433;
fi

if [ -z $PUBLIC_IP ]; then
    PUBLIC_IP=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`;
fi

if [ -z $RPC_PORT ]; then RPC_PORT=9334; fi

if [ -z $WS_PORT ]; then WS_PORT=19334; fi



if [ -n $PRIVATEKEY ]; then
    echo ./incognito -n $NAME --discoverpeers --discoverpeersaddress $BOOTNODE_IP --privatekey $PRIVATEKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" > cmd.sh
    ./incognito -n $NAME --discoverpeers --discoverpeersaddress $BOOTNODE_IP --privatekey $PRIVATEKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" > /data/log.txt 2>/data/error_log.txt
elif [ -n $MININGKEY ]; then
    echo ./incognito -n $NAME --discoverpeers --discoverpeersaddress $BOOTNODE_IP --miningkeys $MININGKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" > cmd.sh
    ./incognito -n $NAME --discoverpeers --discoverpeersaddress $BOOTNODE_IP --miningkeys $MININGKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" > /data/log.txt 2>/data/error_log.txt
fi




