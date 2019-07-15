#!/bin/sh
mkdir -p /data
if [ "$1" == "y" ]; then
    find /data -maxdepth 1 -mindepth 1 -type d | xargs rm -rf
fi

if [ -z $NAME ]; then
    NAME="miner";
fi

if [ -z $BOOTNODE_IP ]; then
    BOOTNODE_IP="172.105.115.134";
fi

if [ -z $NODE_PORT ]; then
    NODE_PORT=9433;
fi

if [ -z $PUBLIC_IP ]; then
    PUBLIC_IP=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`;
fi

if [ -z $RPC_PORT ]; then RPC_PORT=9334; fi

echo ./incognito -n $NAME --discoverpeers --discoverpeersaddress $BOOTNODE_IP --privatekey $PRIVATEKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" > cmd.sh

./incognito -n $NAME --discoverpeers --discoverpeersaddress $BOOTNODE_IP --privatekey $PRIVATEKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" > /data/log.txt 2>/data/error_log.txt
