#!/bin/sh
mkdir -p /data
if [ "$1" == "y" ]; then
    rm -rf /data/*
fi

./constant -n $NAME --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --spendingkey $SPENDINGKEY --nodemode "auto" --datadir "/data" --listen "0.0.0.0:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "constant" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" > /data/log.txt
