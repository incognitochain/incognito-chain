#!/usr/bin/env bash

# Shard 0
if [ "$1" == "shard0-0" ]; then
go run *.go --spendingkey "112t8rsURTpYQMp3978j2nvYXTbuMa9H7MfLTA4PCJoxyweZNWRR3beMEtsoLBBbc473Bv8NE3uKUXcVA2Jnh6sPhTEnFfmQEpY8opeFytoM" --nodemode "auto" --datadir "data/shard0-0" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "0"
fi
if [ "$1" == "shard0-1" ]; then
go run *.go --spendingkey "112t8rtTwTgp4QKJ7rP2p5TyqtFjKYxeFHCUumTwuH4NbCAk7g7H1MvH5eDKyy6N5wvT1FVVLoPrUzrAKKzJeHcCrc2BoSJfTvkDobVSmSZe" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:9434" --externaladdress "127.0.0.1:9434" --norpcauth --rpclisten "127.0.0.1:9335" --relayshards "0"
fi
if [ "$1" == "shard0-2" ]; then
go run *.go --spendingkey "112t8rnj1Xs9p5D6ue4fuLnVq9mJz5kKtyRMLEVBr8hxxmgd7haNFhruwezHswtEu9YKgAUWrnPrU1uvJE1gvep2pixy55ctDNc593c5pHmB" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --rpclisten "127.0.0.1:9336" --relayshards "0"
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
go run *.go --spendingkey "112t8rpDvusPVjYYUECbVd8mSFq9iijuJHtbX2yrEwNT1mebBo8EHAe9neFPNfmd6VmdzAZDHP5ei8QDMxLXE1HRHEqo21gi3twYm8JFiFxJ" --nodemode "auto" --datadir "data/shard1-0" --listen "127.0.0.1:9443" --externaladdress "127.0.0.1:9443" --norpcauth --rpclisten "127.0.0.1:9337" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit --relayshards "1"
fi
if [ "$1" == "shard1-1" ]; then
go run *.go --spendingkey "112t8rq772Lr6tsEri2smLpG5s3v1LrtLNyu9eAavAJk4yMcWz6ziZPoB2sVt8SAjtPY3KT3oGAWXiszXBWXfQxXumJDvgD1rckwiGsS73XU" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:9444" --externaladdress "127.0.0.1:9444" --norpcauth --rpclisten "127.0.0.1:9338" --relayshards "1"
fi
if [ "$1" == "shard1-2" ]; then
go run *.go --spendingkey "112t8rrEgLjxmpzQTh3i2SFxxV27WntXpAkoe9JbseqFvDBPpaPaudzJWXFctZorJXtivEXv1nPzggnmNfNDyj9d5PKh5S4N3UTs6fHBWgeo" --nodemode "auto" --datadir "data/shard1-2" --listen "127.0.0.1:9445" --externaladdress "127.0.0.1:9445" --norpcauth --rpclisten "127.0.0.1:9339" --relayshards "1"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
go run *.go --spendingkey "112t8ruLZgaV3ze37GRikKn8QVnrJDJ5C9Dhtou66vyeBfBDSJ6ZGRSg3k4qTwTjm14kgvwuFX3aAqeU64cGiixDh1ip4nvnmW7xHbSuXpwB" --nodemode "auto" --datadir "data/beacon-0" --listen "127.0.0.1:9423" --externaladdress "127.0.0.1:9423" --norpcauth --rpclisten "127.0.0.1:9340" 
fi
if [ "$1" == "beacon-1" ]; then
go run *.go --spendingkey "112t8rwFSJ8rQ2nuQRH28tgsy5GufZhfQUDQsURdjdDG2jtUyE8ScDGjQNivmFkEJi7HeYX259xYv3tfBir9GPskF9tjEHxp4HcVK2w3rmxJ" --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:9424" --externaladdress "127.0.0.1:9424" --norpcauth --rpclisten "127.0.0.1:9341"
fi
if [ "$1" == "beacon-2" ]; then
go run *.go --spendingkey "112t8rnyr3YjfpLN9unAvAM2TGZhurt7gfcrjCkUv11DWf6gN79raFqrazCHArWsWyosZvGdw5s5cuzzYXgpxDj9sEieyKqKCZ97inpaAXUR" --nodemode "auto" --datadir "data/beacon-2" --listen "127.0.0.1:9425" --externaladdress "127.0.0.1:9425" --norpcauth --rpclisten "127.0.0.1:9342"
fi

# Shard: 0, Role: Proposer
if [ "$1" == "shard0-proposer" ]; then
go run *.go --spendingkey "112t8rqJHgJp2TPpNpLNx34aWHB5VH5Pys3hVjjhhf9tctVeCNmX2zQLBqzHau6LpUbSV52kXtG2hRZsuYWkXWF5kw2v24RJq791fWmQxVqy" --nodemode "auto" --datadir "data/shard0-1" --listen "127.0.0.1:9433" --externaladdress "127.0.0.1:9433" --norpcauth --rpclisten "127.0.0.1:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --relayshards "0"
fi
# Shard: 0, Role: Normal
if [ "$1" == "shard0-normal" ]; then
go run *.go --spendingkey "112t8rnj1Xs9p5D6ue4fuLnVq9mJz5kKtyRMLEVBr8hxxmgd7haNFhruwezHswtEu9YKgAUWrnPrU1uvJE1gvep2pixy55ctDNc593c5pHmB" --nodemode "auto" --datadir "data/shard0-2" --listen "127.0.0.1:9435" --externaladdress "127.0.0.1:9435" --norpcauth --rpclisten "127.0.0.1:9336" --enablewallet --wallet "wallet2" --walletpassphrase "12345678" --walletautoinit
fi
# Shard: 1, Role: Proposer
if [ "$1" == "shard1-proposer" ]; then
go run *.go --spendingkey "112t8rrEEcDQBMnUM5J17qniHZZmckmr8LGCv9nBjP9x5wmGFGUryKTNvEAf1jh2wwW69rxwtANq4m8JmzowfKVPayUHPmAKdwQw5718GKuH" --nodemode "auto" --datadir "data/shard1-1" --listen "127.0.0.1:9436" --externaladdress "127.0.0.1:9436" --norpcauth --rpclisten "127.0.0.1:9338" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit --relayshards "1"
fi
# Beacon, Role: Proposer
if [ "$1" == "beacon-proposer" ]; then
go run *.go --spendingkey "112t8rnaYDe8T6fdAqRwLj8QcSfoUeT6VC1Wj3q3MmpU4dunS6HnRkYnQibSGSvtHW8y7hJYRBvVJkdhj61o5NW8EjxTgHWLdyCM6LoaTTMs" --nodemode "auto" --datadir "data/beacon-1" --listen "127.0.0.1:9423" --externaladdress "127.0.0.1:9423" --norpcauth --rpclisten "127.0.0.1:9340"
fi

