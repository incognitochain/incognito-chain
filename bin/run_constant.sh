#!/usr/bin/env bash

cd ~/go/src/github.com/ninjadotorg/constant
./constant --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --spendingkey $SPENDINGKEY --nodemode "auto" --datadir "data/constant" --listen "127.0.0.1:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "constant" --walletpassphrase "12345678" --walletautoinit
