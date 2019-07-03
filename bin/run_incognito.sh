#!/bin/sh
mkdir -p /data
if [ "$1" == "y" ]; then
    rm -rf /data/*
fi

if [ -z $DISCOVERPEERSADDRESS ]; then DISCOVERPEERSADDRESS="172.105.115.134" ; fi
if [ -z $PORT ]; then PORT=9433; fi
if [ -z $EXTERNALADDRESS ]; then EXTERNALADDRESS=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`; fi
if [ -z $RPC_PORT ]; then RPC_PORT=9334; fi

echo ./incognito -n $NAME --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --privatekey $PRIVATEKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" > cmd.sh

./incognito -n $NAME --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --privatekey $PRIVATEKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" > /data/log.txt 2>/data/error_log.txt
