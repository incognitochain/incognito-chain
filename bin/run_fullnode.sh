#!/bin/sh
mkdir -p /data_v2
if [ "$1" == "y" ]; then
    rm -rf /data_v2/*
fi

echo ./incognito -n $NAME --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --relayshards "all" --datadir "/data_v2" --listen "0.0.0.0:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --metricurl "$METRIC_URL" > cmd.sh

./incognito -n $NAME --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --relayshards "all" --datadir "/data_v2" --listen "0.0.0.0:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --metricurl "$METRIC_URL"> /data_v2/log.txt 2>/data_v2/error_log.txt
