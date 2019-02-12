#!/usr/bin/env bash

SPENDINGKEY=""
EXTERNALADDRESS="127.0.0.1"
PORT="9433"
DISCOVERPEERSADDRESS="127.0.0.1:9330"

kill -9 $(pgrep -u root constant)

cd ~/go/src/github.com/ninjadotorg/constant
git pull
/usr/local/go/bin/go build -o constant
./constant --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --spendingkey $SPENDINGKEY --nodemode "auto" --datadir "data/constant" --listen "127.0.0.1:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "constant" --walletpassphrase "12345678" --walletautoinit
cd ~/go/src/github.com/ninjadotorg/constant/bin
