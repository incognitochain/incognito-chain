#!/usr/bin/env bash

# Shard 0
if [ "$1" == "shard0-0" ]; then
go run *.go --spendingkey "112t8rqGc71CqjrDCuReGkphJ4uWHJmiaV7rVczqNhc33pzChmJRvikZNc3Dt5V7quhdzjWW9Z4BrB2BxdK5VtHzsG9JZdZ5M7yYYGidKKZV" --nodemode "auto" --datadir "data/shard0-0" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "0"
fi
if [ "$1" == "shard0-1" ]; then
go run *.go --spendingkey "112t8s2UkZEwS7JtqLHFruRrh4Drj53UzH4A6DrairctKutxVb8Vw2DMzxCReYsAZkXi9ycaSNRHEcB7TJaTwPhyPvqRzu5NnUgTMN9AEKwo" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:9434" --externaladdress "127.0.0.1:9434" --norpcauth --norpc --relayshards "0"
fi
if [ "$1" == "shard0-2" ]; then
go run *.go --spendingkey "112t8rnYY8UbXGVJ3PsrWxssjr1JXaTPNCPDrneXcQgVQs2MFYwgCzPmTsgqPPbeq8c4QxkrjpHYRaG39ZjtwCmHMJBNh2MxaQvKWw5eUGTM" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --norpc --relayshards "0"
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
go run *.go --spendingkey "112t8rnYoj4LesSwRsseGCCYi4J2Py5QxytKKF2WixwEYP4opKUNL2Av9bR2zjfLewf3PQeKcNnuRTTPKgZSJaZH8dfoqY2rmHNekmGMBNDX" --nodemode "auto" --datadir "data/shard1-0" --listen "127.0.0.1:9443" --externaladdress "127.0.0.1:9443" --norpcauth --rpclisten "127.0.0.1:9338" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit --relayshards "1"
fi
if [ "$1" == "shard1-1" ]; then
go run *.go --spendingkey "112t8rnZ5aGQqJw9bg6fR8AiGe9NFRtSmn73Scd4oNJcE5BNY4Rbju2amkTRW5PUaFpETkKAdSJUMqptjFYb3B8PVAcQhrqooieNFXe5jzTj" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:9444" --externaladdress "127.0.0.1:9444" --norpc --relayshards "1"
fi
if [ "$1" == "shard1-2" ]; then
go run *.go --spendingkey "112t8rnZUKcW5CBDojVmMD6PmDJzR3VtfqFGWG6HRT9PocB6aewekjebWMm9aQnSncgwDV2GMqAWzspzFYL2vs3C3KnZB9H5YSE4s1SdotHb" --nodemode "auto" --datadir "data/shard1-2" --listen "127.0.0.1:9445" --externaladdress "127.0.0.1:9445" --norpc --relayshards "1"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
go run *.go --spendingkey "112t8rxTdWfGCtgWvAMHnnEw9vN3R1D7YgD1SSHjAnVGL82HCrMq9yyXrHv3kB4gr84cejnMZRQ973RyHhq2G3MksoTWejNKdSWoQYDFf4gQ" --nodemode "auto" --datadir "data/beacon-0" --listen "127.0.0.1:9430" --externaladdress "127.0.0.1:9430" --norpcauth --rpclisten "127.0.0.1:9337"
fi
if [ "$1" == "beacon-1" ]; then
go run *.go --spendingkey "112t8rnXDNYL1RyTuT85JXeX7mJg1Sc6tCby5akSM7pfEGApgAx83X8C46EDu6dFAK6MVcWfQv2sfTk5nV9HqU3jrkrWdhrmi9z34jEhgHak" --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:9431" --externaladdress "127.0.0.1:9430" --norpcauth --norpc
fi
if [ "$1" == "beacon-2" ]; then
go run *.go --spendingkey "112t8rnXmEeG5zsS7rExURJfqaRZhm6r4Pypkeag2gprdhtgDpen3LwV68x1nDPRYz2zhyhJTJCGvq1tUx4P1dvrdxF9W9DH7ME7PeGN2ohZ" --nodemode "auto" --datadir "data/beacon-2" --listen "127.0.0.1:9432" --externaladdress "127.0.0.1:9430" --norpcauth --norpc
fi
