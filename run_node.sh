#!/usr/bin/env bash

# Shard 0
if [ "$1" == "shard0-0" ]; then
go run *.go --spendingkey "112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV" --nodemode "auto" --datadir "data/shard0-0" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "0"
fi
if [ "$1" == "shard0-1" ]; then
go run *.go --spendingkey "112t8ruNDweRqN4LvP1FaxjWWoVjDqbhtqQyjx52xtq3rpcRuPhqtwApvuNYpj78TAdZAqDsy4ewwgaEdCQWqK5VAXdBPfWKkMD3QbeRSaSJ" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:9434" --externaladdress "127.0.0.1:9434" --norpcauth --rpclisten "127.0.0.1:9335" --relayshards "0"
fi
if [ "$1" == "shard0-2" ]; then
go run *.go --spendingkey "112t8roGSduAqZKFWWdRMzVazM5qDN6xzii4wWwbn6r4uEbTfEVN2o9sHY9kvHpouUXnhm9HnUcWSUTDvyBDgRLTVqqj5pgdmD6QJoC99bRq" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --rpclisten "127.0.0.1:9336" --relayshards "0"
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
go run *.go --spendingkey "112t8rpxJvKj42esCYsnShYnbe67yERdjrPmWcDWCcTyM6W4Nst3yyBkidWEu2M2M5H5cKwVMtzBLM6XWyqQrB4L4QK2GJFyaBQhRwgC8Vz6" --nodemode "auto" --datadir "data/shard1-0" --listen "127.0.0.1:9443" --externaladdress "127.0.0.1:9443" --norpcauth --rpclisten "127.0.0.1:9337" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit --relayshards "1"
fi
if [ "$1" == "shard1-1" ]; then
go run *.go --spendingkey "112t8rqNdNmv4Z1WeR9Lj8Us9qABCCSFRNJf1fQ3yr2v9WKkkxxHQZzDX6zF1GYq3hC6qBE9GMo9hJgrBq7irZ2qc3HSMaPLMkMTE76YH4Jc" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:9444" --externaladdress "127.0.0.1:9444" --norpcauth --rpclisten "127.0.0.1:9338" --relayshards "1"
fi
if [ "$1" == "shard1-2" ]; then
go run *.go --spendingkey "112t8rpcsM9MQtbinYKXphEmCM1xmd4SmsFupTmNp2kBkboVSsAJj3r7bZ4CwDAdHapc5hdsrNNgoCo7a6MJkc93UuSu875Y6ewNLMDXtgyr" --nodemode "auto" --datadir "data/shard1-2" --listen "127.0.0.1:9445" --externaladdress "127.0.0.1:9445" --norpcauth --rpclisten "127.0.0.1:9339" --relayshards "1"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
go run *.go --spendingkey "112t8rxTdWfGCtgWvAMHnnEw9vN3R1D7YgD1SSHjAnVGL82HCrMq9yyXrHv3kB4gr84cejnMZRQ973RyHhq2G3MksoTWejNKdSWoQYDFf4gQ" --nodemode "auto" --datadir "data/beacon-0" --listen "127.0.0.1:9423" --externaladdress "127.0.0.1:9423" --norpcauth --rpclisten "127.0.0.1:9340" 
fi
if [ "$1" == "beacon-1" ]; then
go run *.go --spendingkey "112t8rnXDNYL1RyTuT85JXeX7mJg1Sc6tCby5akSM7pfEGApgAx83X8C46EDu6dFAK6MVcWfQv2sfTk5nV9HqU3jrkrWdhrmi9z34jEhgHak" --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:9424" --externaladdress "127.0.0.1:9424" --norpcauth --rpclisten "127.0.0.1:9341"
fi
if [ "$1" == "beacon-2" ]; then
go run *.go --spendingkey "112t8rnXmEeG5zsS7rExURJfqaRZhm6r4Pypkeag2gprdhtgDpen3LwV68x1nDPRYz2zhyhJTJCGvq1tUx4P1dvrdxF9W9DH7ME7PeGN2ohZ" --nodemode "auto" --datadir "data/beacon-2" --listen "127.0.0.1:9425" --externaladdress "127.0.0.1:9425" --norpcauth --rpclisten "127.0.0.1:9342"
fi

# Shard: 0, Role: Proposer
if [ "$1" == "shard0-proposer" ]; then
go run *.go --spendingkey "112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "0"
fi
# Shard: 0, Role: Normal
if [ "$1" == "shard0-normal" ]; then
go run *.go --spendingkey "112t8rqnMrtPkJ4YWzXfG82pd9vCe2jvWGxqwniPM5y4hnimki6LcVNfXxN911ViJS8arTozjH4rTpfaGo5i1KKcG1ayjiMsa4E3nABGAqQh" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --rpclisten "127.0.0.1:9336" --enablewallet --wallet "wallet2" --walletpassphrase "12345678" --walletautoinit
fi
# Shard: 1, Role: Proposer
if [ "$1" == "shard1-proposer" ]; then
go run *.go --spendingkey "112t8s2UkZEwS7JtqLHFruRrh4Drj53UzH4A6DrairctKutxVb8Vw2DMzxCReYsAZkXi9ycaSNRHEcB7TJaTwPhyPvqRzu5NnUgTMN9AEKwo" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:9436" --externaladdress "127.0.0.1:9436" --norpcauth --rpclisten "127.0.0.1:9338" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit --relayshards "1"
fi
# Beacon, Role: Proposer
if [ "$1" == "beacon-proposer" ]; then
go run *.go --spendingkey "112t8rxTdWfGCtgWvAMHnnEw9vN3R1D7YgD1SSHjAnVGL82HCrMq9yyXrHv3kB4gr84cejnMZRQ973RyHhq2G3MksoTWejNKdSWoQYDFf4gQ" --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:9430" --externaladdress "127.0.0.1:9430" --norpcauth --rpclisten "127.0.0.1:9337"
fi

