#!/usr/bin/env bash
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rqJHgJp2TPpNpLNx34aWHB5VH5Pys3hVjjhhf9tctVeCNmX2zQLBqzHau6LpUbSV52kXtG2hRZsuYWkXWF5kw2v24RJq791fWmQxVqy" --nodemode "auto" --datadir "data/shard0-0" --listen "127.0.0.1:8433" --externaladdress "127.0.0.1:8433" --norpcauth --rpclisten "127.0.0.1:8334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "0" --txpoolttl 250
fi
if [ "$1" == "shard0-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rrEEcDQBMnUM5J17qniHZZmckmr8LGCv9nBjP9x5wmGFGUryKTNvEAf1jh2wwW69rxwtANq4m8JmzowfKVPayUHPmAKdwQw5718GKuH" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:8434" --externaladdress "127.0.0.1:8434" --norpcauth --rpclisten "127.0.0.1:8335" --relayshards "0"
fi
if [ "$1" == "shard0-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rnXgYcLCjeHBpnesutJEBfG5xVzhMf8gLVZuL4fya4ov9Ko8pr6Mrj5eEZe5bie4ZVnH7MQNpLjuAxiuSDmPp5WgkuRRGW8NLcJZoF2" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:8435" --externaladdress "127.0.0.1:8435" --norpcauth --rpclisten "127.0.0.1:8336" --relayshards "0"
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rnY5yiM5XJCnAH3hD2UzhKnwoQjPCDDy6cRwsoNevRFJBm7HmYLbpijKvpxG8GD1LsamT9DC6xhs1RqCnDzoDhwFv1sd7ckzSa2Cj16" --nodemode "auto" --datadir "data/shard1-0" --listen "127.0.0.1:8443" --externaladdress "127.0.0.1:8443" --norpcauth --rpclisten "127.0.0.1:8337" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "shard1-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rpDvusPVjYYUECbVd8mSFq9iijuJHtbX2yrEwNT1mebBo8EHAe9neFPNfmd6VmdzAZDHP5ei8QDMxLXE1HRHEqo21gi3twYm8JFiFxJ" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:8444" --externaladdress "127.0.0.1:8444" --norpcauth --rpclisten "127.0.0.1:8338"
fi
if [ "$1" == "shard1-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rq772Lr6tsEri2smLpG5s3v1LrtLNyu9eAavAJk4yMcWz6ziZPoB2sVt8SAjtPY3KT3oGAWXiszXBWXfQxXumJDvgD1rckwiGsS73XU" --nodemode "auto" --datadir "data/shard1-2" --listen "127.0.0.1:8445" --externaladdress "127.0.0.1:8445" --norpcauth --rpclisten "127.0.0.1:8339"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8ruLZgaV3ze37GRikKn8QVnrJDJ5C9Dhtou66vyeBfBDSJ6ZGRSg3k4qTwTjm14kgvwuFX3aAqeU64cGiixDh1ip4nvnmW7xHbSuXpwB" --nodemode "auto" --datadir "data/beacon-0" --listen "127.0.0.1:8423" --externaladdress "127.0.0.1:8423" --norpcauth --rpclisten "127.0.0.1:8340" 
fi
if [ "$1" == "beacon-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rwFSJ8rQ2nuQRH28tgsy5GufZhfQUDQsURdjdDG2jtUyE8ScDGjQNivmFkEJi7HeYX259xYv3tfBir9GPskF9tjEHxp4HcVK2w3rmxJ" --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:8424" --externaladdress "127.0.0.1:8424" --norpcauth --rpclisten "127.0.0.1:8341"
fi
if [ "$1" == "beacon-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rnyr3YjfpLN9unAvAM2TGZhurt7gfcrjCkUv11DWf6gN79raFqrazCHArWsWyosZvGdw5s5cuzzYXgpxDj9sEieyKqKCZ97inpaAXUR" --nodemode "auto" --datadir "data/beacon-2" --listen "127.0.0.1:8425" --externaladdress "127.0.0.1:8425" --norpcauth --rpclisten "127.0.0.1:8342"
fi
######
if [ "$1" == "shard-stake-1" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rtTwTgp4QKJ7rP2p5TyqtFjKYxeFHCUumTwuH4NbCAk7g7H1MvH5eDKyy6N5wvT1FVVLoPrUzrAKKzJeHcCrc2BoSJfTvkDobVSmSZe" --nodemode "auto" --datadir "data/shard-stake" --listen "127.0.0.1:8436" --externaladdress "127.0.0.1:8436" --norpcauth --rpclisten "127.0.0.1:8343" 
fi
if [ "$1" == "shard-stake-2" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rsURTpYQMp3978j2nvYXTbuMa9H7MfLTA4PCJoxyweZNWRR3beMEtsoLBBbc473Bv8NE3uKUXcVA2Jnh6sPhTEnFfmQEpY8opeFytoM" --nodemode "auto" --datadir "data/shard-stake-2" --listen "127.0.0.1:8446" --externaladdress "127.0.0.1:8446" --norpcauth --rpclisten "127.0.0.1:8344"
fi
if [ "$1" == "shard-stake-3" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rzcRLXhra2ouQo4yCZiQt1iEoZdRkD3m6fixXqCLzygjo28L3isePdPjPbXJ7zcxgyxbiNuF4Ex15NFCHVLwHhJD7QL7AHUfUsH78AP" --nodemode "auto" --datadir "data/shard-stake-3" --listen "127.0.0.1:8447" --externaladdress "127.0.0.1:8447" --norpcauth --rpclisten "127.0.0.1:8345"
fi


###### SINGLE_MEMBER
######
# Shard: 0, Role: Proposer
if [ "$1" == "shard0-proposer" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rqJHgJp2TPpNpLNx34aWHB5VH5Pys3hVjjhhf9tctVeCNmX2zQLBqzHau6LpUbSV52kXtG2hRZsuYWkXWF5kw2v24RJq791fWmQxVqy" --nodemode "auto" --datadir "data/shard-0" --listen "127.0.0.1:8433" --externaladdress "127.0.0.1:8433" --norpcauth --rpclisten "127.0.0.1:8334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
# Shard: 1, Role: Proposer
if [ "$1" == "shard1-proposer" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8rrEEcDQBMnUM5J17qniHZZmckmr8LGCv9nBjP9x5wmGFGUryKTNvEAf1jh2wwW69rxwtANq4m8JmzowfKVPayUHPmAKdwQw5718GKuH" --nodemode "auto" --datadir "data/shard-1" --listen "127.0.0.1:8436" --externaladdress "127.0.0.1:8436" --norpcauth --rpclisten "127.0.0.1:8338" --enablewallet --wallet "wallet2" --walletpassphrase "12345678" --walletautoinit
fi
# Beacon, Role: Proposer
if [ "$1" == "beacon-proposer" ]; then
go run *.go --discoverpeersaddress "127.0.0.1:8330" --privatekey "112t8ruLZgaV3ze37GRikKn8QVnrJDJ5C9Dhtou66vyeBfBDSJ6ZGRSg3k4qTwTjm14kgvwuFX3aAqeU64cGiixDh1ip4nvnmW7xHbSuXpwB" --nodemode "auto" --datadir "data/beacon" --listen "127.0.0.1:8423" --externaladdress "127.0.0.1:8423" --norpcauth --rpclisten "127.0.0.1:8340"
fi
# Relay node
if [ "$1" == "relaynode" ]; then
go run *.go --relayshards "all" --datadir "data/relaynode" --listen "127.0.0.1:8435" --externaladdress "127.0.0.1:8435" --norpcauth --rpclisten "127.0.0.1:8336" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit
fi
