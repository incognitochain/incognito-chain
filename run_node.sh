#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rqEd6a7v4PfWu9CwXFAmRyazCNRNEQi8G8eGBYtrg2Cix6GNYhhs3tz1JRPxJ4FsHq3mH7fJBq3bC8c3DJM3noWAQE8eY4sMeHxXYDe" --nodemode "auto" --datadir "data/shard0-0" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "0" --rpcwslisten "127.0.0.1:19334"
fi
if [ "$1" == "shard0-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rtidnuL3BBzhfVptK6mPmAr2AttXgmcdE8xZwYyQd3t7t5a16dDPPnSSx4KzzWAvM7AgXo8UQBCo5o4Wn1K6Cs27hHCjQpSY8UjV5Gy" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:9434" --externaladdress "127.0.0.1:9434" --norpcauth --rpclisten "127.0.0.1:9335" --relayshards "0"
fi
if [ "$1" == "shard0-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rtxdpGEx5H7MHi5EXGKZhKAxgbEkRv1d8TH1cbqKzc5j4mXEKC2NUCeCSQRA3zDeUQxKUW54sAGzpqFi86JQGaPAh4Ao4Gqo7q51yGw" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --rpclisten "127.0.0.1:9336" --relayshards "0"
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roCAzRKCBy3VpSwujBaM2iPJZ82Sb5dtYJiyS1Cknax3bxd3x9aynNkZmoNsNLJWb8NMtJZXf9S68RHHy5W9iuvVErrd1Q37eW3uerc" --nodemode "auto" --datadir "data/shard1-0" --listen "127.0.0.1:9443" --externaladdress "127.0.0.1:9443" --norpcauth --rpclisten "127.0.0.1:9337" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "shard1-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rpHmXL5Eb9tW4wuKT3sh5ADthEBKJPw8thFreVNJyvY6o9iRTbdomrCC3VchVmLYox6pnxS7pvNzwdPrykokmaNQa1Q1ZYfNVEDMwKS" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:9444" --externaladdress "127.0.0.1:9444" --norpcauth --rpclisten "127.0.0.1:9338"
fi
if [ "$1" == "shard1-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rqXccVRRjL4CGxoJUD3Fya7rzfWXXF5uakQokbNuiiWSabr78SwnXLEqSozeJZzgUWajgZkUbMJmoAUnkVvkDgYTEBzjhKeU9ECNPtY" --nodemode "auto" --datadir "data/shard1-2" --listen "127.0.0.1:9445" --externaladdress "127.0.0.1:9445" --norpcauth --rpclisten "127.0.0.1:9339"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8sDsj4tnDQ7mv5E26hfrHpsjyueJybmyJyKfn2uB7NbVY8pAzQEiLrKURuqihV8EvwrJoMLoz2apTyJWPYTQeEPMtDgnAKXjCKGPmvkb" --nodemode "auto" --datadir "data/beacon-0" --listen "127.0.0.1:9423" --externaladdress "127.0.0.1:9423" --norpcauth --rpclisten "127.0.0.1:9340" 
fi
if [ "$1" == "beacon-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8sFHLcDnYJyndjBYs7UFRQGnzq9MtDcr6xsgQ8wfZZ7kXRFexnGxSA7u1SB7iJXoWXRB2ono5BHzKb1eqzTATEWTA32G82GcGc2krrSm" --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:9424" --externaladdress "127.0.0.1:9424" --norpcauth --rpclisten "127.0.0.1:9341"
fi
if [ "$1" == "beacon-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8sL2pTLHdZuTuZSFAeDq6rzBruAqvyMYwKyav9TKUc1zKTz3ScSbT44MG4GgiN9KxC1234j63fFgBE88pHfE1Wbrz6qbm6vGohhNua2h" --nodemode "auto" --datadir "data/beacon-2" --listen "127.0.0.1:9425" --externaladdress "127.0.0.1:9425" --norpcauth --rpclisten "127.0.0.1:9342"
fi
# FullNode
if [ "$1" == "full_node" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --nodemode "relay" --datadir "data/full_node" --listen "127.0.0.1:9533" --externaladdress "127.0.0.1:9533" --norpcauth --rpclisten "127.0.0.1:9554" --enablewallet --wallet "wallet_fullnode" --walletpassphrase "12345678" --walletautoinit --relayshards "all"  --txpoolmaxtx 100000
fi
######
if [ "$1" == "shard-stake-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rtTwTgp4QKJ7rP2p5TyqtFjKYxeFHCUumTwuH4NbCAk7g7H1MvH5eDKyy6N5wvT1FVVLoPrUzrAKKzJeHcCrc2BoSJfTvkDobVSmSZe" --nodemode "auto" --datadir "data/shard-stake" --listen "127.0.0.1:9436" --externaladdress "127.0.0.1:9436" --norpcauth --rpclisten "127.0.0.1:9343" 
fi
if [ "$1" == "shard-stake-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rsURTpYQMp3978j2nvYXTbuMa9H7MfLTA4PCJoxyweZNWRR3beMEtsoLBBbc473Bv8NE3uKUXcVA2Jnh6sPhTEnFfmQEpY8opeFytoM" --nodemode "auto" --datadir "data/shard-stake-2" --listen "127.0.0.1:9446" --externaladdress "127.0.0.1:9446" --norpcauth --rpclisten "127.0.0.1:9344"
fi
if [ "$1" == "shard-stake-3" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rzcRLXhra2ouQo4yCZiQt1iEoZdRkD3m6fixXqCLzygjo28L3isePdPjPbXJ7zcxgyxbiNuF4Ex15NFCHVLwHhJD7QL7AHUfUsH78AP" --nodemode "auto" --datadir "data/shard-stake-3" --listen "127.0.0.1:9447" --externaladdress "127.0.0.1:9447" --norpcauth --rpclisten "127.0.0.1:9345"
fi
####full node
# go run *.go --discoverpeersaddress "127.0.0.1:9330" --nodemode "relay" --relayshards "all" --datadir "data/fullnode" --listen "127.0.0.1:9436" --externaladdress "127.0.0.1:9436" --norpcauth --rpclisten "127.0.0.1:9343"

###### SINGLE_MEMBER
######
# Shard: 0, Role: Proposer
if [ "$1" == "shard0-proposer" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rqJHgJp2TPpNpLNx34aWHB5VH5Pys3hVjjhhf9tctVeCNmX2zQLBqzHau6LpUbSV52kXtG2hRZsuYWkXWF5kw2v24RJq791fWmQxVqy" --nodemode "auto" --datadir "data/shard-0" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
# Shard: 1, Role: Proposer
if [ "$1" == "shard1-proposer" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rrEEcDQBMnUM5J17qniHZZmckmr8LGCv9nBjP9x5wmGFGUryKTNvEAf1jh2wwW69rxwtANq4m8JmzowfKVPayUHPmAKdwQw5718GKuH" --nodemode "auto" --datadir "data/shard-1" --listen "127.0.0.1:9436" --externaladdress "127.0.0.1:9436" --norpcauth --rpclisten "127.0.0.1:9338" --enablewallet --wallet "wallet2" --walletpassphrase "12345678" --walletautoinit
fi
# Beacon, Role: Proposer
if [ "$1" == "beacon-proposer" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8ruLZgaV3ze37GRikKn8QVnrJDJ5C9Dhtou66vyeBfBDSJ6ZGRSg3k4qTwTjm14kgvwuFX3aAqeU64cGiixDh1ip4nvnmW7xHbSuXpwB" --nodemode "auto" --datadir "data/beacon" --listen "127.0.0.1:9423" --externaladdress "127.0.0.1:9423" --norpcauth --rpclisten "127.0.0.1:9340"
fi
# Relay node
if [ "$1" == "relaynode" ]; then
go run *.go --relayshards "all" --datadir "data/relaynode" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --rpclisten "127.0.0.1:9336" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit
fi
