#!/bin/sh
/constant --discoverpeers --discoverpeersaddress $DISCOVERPEERSADDRESS --spendingkey $SPENDINGKEY --nodemode "auto" --datadir "/data/constant" --listen "0.0.0.0:$PORT" --externaladdress "$EXTERNALADDRESS:$PORT" --norpcauth --enablewallet --wallet "constant" --walletpassphrase "12345678" --walletautoinit
