#!/usr/bin/env bash

# Shard: 0, Role: Proposer
if [ "$1" == "shard0-proposer" ]; then
./constant --spendingkey "112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "[0,1]"
fi
# Shard: 0, Role: Normal
if [ "$1" == "shard0-normal" ]; then
./constant --spendingkey "112t8rqnMrtPkJ4YWzXfG82pd9vCe2jvWGxqwniPM5y4hnimki6LcVNfXxN911ViJS8arTozjH4rTpfaGo5i1KKcG1ayjiMsa4E3nABGAqQh" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --rpclisten "127.0.0.1:9336" --enablewallet --wallet "wallet2" --walletpassphrase "12345678" --walletautoinit --relayshards "[0,1]"
fi
# Shard: 1, Role: Proposer
if [ "$1" == "shard1-proposer" ]; then
./constant --spendingkey "112t8s2UkZEwS7JtqLHFruRrh4Drj53UzH4A6DrairctKutxVb8Vw2DMzxCReYsAZkXi9ycaSNRHEcB7TJaTwPhyPvqRzu5NnUgTMN9AEKwo" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:9436" --externaladdress "127.0.0.1:9436" --norpcauth --rpclisten "127.0.0.1:9338" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit --relayshards "[0,1]"
fi
# Beacon, Role: Proposer
if [ "$1" == "beacon-proposer" ]; then
./constant --spendingkey '112t8rxTdWfGCtgWvAMHnnEw9vN3R1D7YgD1SSHjAnVGL82HCrMq9yyXrHv3kB4gr84cejnMZRQ973RyHhq2G3MksoTWejNKdSWoQYDFf4gQ' --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:9430" --externaladdress "127.0.0.1:9430" --norpcauth --rpclisten "127.0.0.1:9337"
fi