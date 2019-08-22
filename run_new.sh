#!/usr/bin/env bash
if [ "$1" == "1" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:14DrMrsYZ58EqrjbuyBuTy3NBmcXB" --nodemode "auto" --datadir "data/1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "2" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1eDXPfd9jXhMVbhLAALewJymUPob" --nodemode "auto" --datadir "data/2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336"
fi
if [ "$1" == "3" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12v9jokoYDWFQ7731mR6QjjkvUG1Q" --nodemode "auto" --datadir "data/3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi