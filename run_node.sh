#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXCerQX2RRd8KhPfsFCj2rrBYUx42FZJKgRFcdBfg36Mid3ygKyMn5LSc5LBHsxqapRaN6xMav7bGhA6VtGUzNNYuA9Y78CB5oGkti" --nodemode "auto" --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --rpcwslisten "0.0.0.0:19334" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard0-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rtidnuL3BBzhfVptK6mPmAr2AttXgmcdE8xZwYyQd3t7t5a16dDPPnSSx4KzzWAvM7AgXo8UQBCo5o4Wn1K6Cs27hHCjQpSY8UjV5Gy" --nodemode "auto" --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard0-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rtxdpGEx5H7MHi5EXGKZhKAxgbEkRv1d8TH1cbqKzc5j4mXEKC2NUCeCSQRA3zDeUQxKUW54sAGzpqFi86JQGaPAh4Ao4Gqo7q51yGw" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard0-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8ruUPhxinaPG89z372Gza7B7G7ZZHkHaXa3PvARVpqdAjB6Nwe32Qtt8fV3yCEA4jgnZGuYHsS2Un89juYwv9Uqpw9v5b8cunXmtoULJ" --nodemode "auto" --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8roCAzRKCBy3VpSwujBaM2iPJZ82Sb5dtYJiyS1Cknax3bxd3x9aynNkZmoNsNLJWb8NMtJZXf9S68RHHy5W9iuvVErrd1Q37eW3uerc" --nodemode "auto" --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --enablewallet --wallet "wallet2" --walletpassphrase "12345678" --walletautoinit --rpcwslisten "127.0.0.1:19338" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard1-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rpHmXL5Eb9tW4wuKT3sh5ADthEBKJPw8thFreVNJyvY6o9iRTbdomrCC3VchVmLYox6pnxS7pvNzwdPrykokmaNQa1Q1ZYfNVEDMwKS" --nodemode "auto" --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "127.0.0.1:19339" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard1-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rqXccVRRjL4CGxoJUD3Fya7rzfWXXF5uakQokbNuiiWSabr78SwnXLEqSozeJZzgUWajgZkUbMJmoAUnkVvkDgYTEBzjhKeU9ECNPtY" --nodemode "auto" --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard1-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rrYgmYrxDRyBoWHw7RMksy2RoETHC9DRXTMLTDrN6gid6M3wsvYfUZ9Gn9vsFHC25tujQBkvYKeJNAtUAActG69V6Rp6rgy69g8zcrA" --nodemode "auto" --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
# Shard 2
if [ "$1" == "shard2-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rno8tY7zampmGdpNPfMGLe1gNguZda5ei7UcGSi8skGE8qPvtqgUAte4YL8yNKMKJwTzUSiycooeph1C1kRs2yNggLy1ysG9DLPCe9h" --nodemode "auto" --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "shard2-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8ro3MXHpHRDvXtcrktzyQq75G6Hxn4rXU5j5qvfsABu3puRQqLt9xeTyHAWPmMP35Ud4xTq7BjhV3jFMHVQHzih7GjGRvBGzKGCyF5s1" --nodemode "auto" --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "shard2-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rpkKoAoJTH6CMwsa2daiDSQCSfvAuev9xP5Yd1aHa22CafxpJnVMwsivLqXPm2cQZDR7SgztVrSaYhVKJp7SNSiTiu4Fnco9ydqJMm5" --nodemode "auto" --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "shard2-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rqfirrZBpBi939Rua8BFFs5UEZRsi3ePHExQdoMDsyLoJyQKXMUAT38mGbFzXvupyoAPE4jHSkooCBtz39SCVq9Fq2aYwFFZSg9Lbrn" --nodemode "auto" --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
# Shard 3
if [ "$1" == "shard3-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8roCAzRKCBy3VpSwujBaM2iPJZ82Sb5dtYJiyS1Cknax3bxd3x9aynNkZmoNsNLJWb8NMtJZXf9S68RHHy5W9iuvVErrd1Q37eW3uerc" --nodemode "auto" --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi
if [ "$1" == "shard3-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rpHmXL5Eb9tW4wuKT3sh5ADthEBKJPw8thFreVNJyvY6o9iRTbdomrCC3VchVmLYox6pnxS7pvNzwdPrykokmaNQa1Q1ZYfNVEDMwKS" --nodemode "auto" --datadir "data/shard3-1" --listen "0.0.0.0:9447" --externaladdress "0.0.0.0:9447" --norpcauth --rpclisten "0.0.0.0:9347"
fi
if [ "$1" == "shard3-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rqXccVRRjL4CGxoJUD3Fya7rzfWXXF5uakQokbNuiiWSabr78SwnXLEqSozeJZzgUWajgZkUbMJmoAUnkVvkDgYTEBzjhKeU9ECNPtY" --nodemode "auto" --datadir "data/shard3-2" --listen "0.0.0.0:9448" --externaladdress "0.0.0.0:9448" --norpcauth --rpclisten "0.0.0.0:9348"
fi
if [ "$1" == "shard3-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rrYgmYrxDRyBoWHw7RMksy2RoETHC9DRXTMLTDrN6gid6M3wsvYfUZ9Gn9vsFHC25tujQBkvYKeJNAtUAActG69V6Rp6rgy69g8zcrA" --nodemode "auto" --datadir "data/shard3-3" --listen "0.0.0.0:9449" --externaladdress "0.0.0.0:9449" --norpcauth --rpclisten "0.0.0.0:9349"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXCerQX2RRd8KhPfsFCj2rrBYUx42FZJKgRFcdBfg36Mid3ygKyMn5LSc5LBHsxqapRaN6xMav7bGhA6VtGUzNNYuA9Y78CB5oGkti" --nodemode "auto" --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "beacon-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sFHLcDnYJyndjBYs7UFRQGnzq9MtDcr6xsgQ8wfZZ7kXRFexnGxSA7u1SB7iJXoWXRB2ono5BHzKb1eqzTATEWTA32G82GcGc2krrSm" --nodemode "auto" --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "beacon-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sL2pTLHdZuTuZSFAeDq6rzBruAqvyMYwKyav9TKUc1zKTz3ScSbT44MG4GgiN9KxC1234j63fFgBE88pHfE1Wbrz6qbm6vGohhNua2h" --nodemode "auto" --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "beacon-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8sP8ZgYgFTjMTVmzZ9YpPpESfoe5b78a7TqvWn9PAkVPUrQsccx5QB62TAupWm4zCPVh9DKFeorMRxYpgQtRGEBCDv4MCEkePXZ4gbek" --nodemode "auto" --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
# FullNode
#if [ "$1" == "full_node" ]; then
#./incognito --discoverpeersaddress "0.0.0.0:9330" --nodemode "relay" --datadir "data/full_node" --listen "0.0.0.0:9454" --externaladdress "0.0.0.0:9454" --norpcauth --rpclisten "0.0.0.0:9354" --enablewallet --wallet "wallet_fullnode" --walletpassphrase "12345678" --walletautoinit --relayshards "all"  --txpoolmaxtx 100000
#fi
######
if [ "$1" == "shard-candidate0-1" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rtTwTgp4QKJ7rP2p5TyqtFjKYxeFHCUumTwuH4NbCAk7g7H1MvH5eDKyy6N5wvT1FVVLoPrUzrAKKzJeHcCrc2BoSJfTvkDobVSmSZe" --nodemode "auto" --datadir "data/shard-stake" --listen "127.0.0.1:9455" --externaladdress "127.0.0.1:9455" --norpcauth --rpclisten "127.0.0.1:9355"
fi
if [ "$1" == "shard-candidate0-2" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rsURTpYQMp3978j2nvYXTbuMa9H7MfLTA4PCJoxyweZNWRR3beMEtsoLBBbc473Bv8NE3uKUXcVA2Jnh6sPhTEnFfmQEpY8opeFytoM" --nodemode "auto" --datadir "data/shard-stake-2" --listen "127.0.0.1:9456" --externaladdress "127.0.0.1:9456" --norpcauth --rpclisten "127.0.0.1:9356"
fi
if [ "$1" == "shard-candidate0-3" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rsq5Xx45T1ZKH4N45aBztqBJiDAR9Nw5wMb8Fe5PnFCqDiUAgVzoMr3xBznNJTfu2CSW3HC6M9rGHxTyUzUBbZHjv6wCMnucDDKbHT4" --nodemode "auto" --datadir "data/shard-stake-6" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
fi
if [ "$1" == "shard-candidate1-1" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rotpPVSeHrknwVUTLQgy2avatUWKh2oV9EjVMw6eEtwyJT1FsrHGzvBaLpHL4gPVfJjuSUWvTtiTKuWGNNwGuLo8SHCgfA36ttJ5J7u" --nodemode "auto" --datadir "data/shard-stake-3" --listen "0.0.0.0:9457" --externaladdress "0.0.0.0:9457" --norpcauth --rpclisten "0.0.0.0:9357"
fi
if [ "$1" == "shard-candidate1-2" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8roEu1K8gUXF3cyRBv1GSGHDXNaSFh45sxgCCSi4Q6KWbW91DtYFqJdAP5MFBJEziejirb4cdDhE2Mxi4PSg8wf277vnCryQyL3VtBK2" --nodemode "auto" --datadir "data/shard-stake-4" --listen "0.0.0.0:9458" --externaladdress "0.0.0.0:9458" --norpcauth --rpclisten "0.0.0.0:9358"
fi
if [ "$1" == "shard-candidate1-3" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rrEgLjxmpzQTh3i2SFxxV27WntXpAkoe9JbseqFvDBPpaPaudzJWXFctZorJXtivEXv1nPzggnmNfNDyj9d5PKh5S4N3UTs6fHBWgeo" --nodemode "auto" --datadir "data/shard-stake-5" --listen "0.0.0.0:9459" --externaladdress "0.0.0.0:9459" --norpcauth --rpclisten "0.0.0.0:9359"
fi
if [ "$1" == "shard-candidate1-4" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnZk7K75qq4GWuVUJwxfRy5EAcrvCyJnMs4ayznzsiV4MWqHdW1DqBPohCutvJzaiwWff1VZvVaH9s6CYHnbvhAhEMtA5dhvMzRDLQv" --nodemode "auto" --datadir "data/shard-stake-7" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
fi
if [ "$1" == "shard-candidate1-5" ]; then
./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rna7M8BYBfNjNHmw3Tie6Yir9mQgp5rSRgUngTqn6A6iSRvAPex4sXsmGxVzXcpUUDfnRfRys3QrPnTHauiipdUNtj7Ef6t3mHUwiC3" --nodemode "auto" --datadir "data/shard-stake-8" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
fi
#if [ "$1" == "shard-stake-9" ]; then
#./incognito --discoverpeersaddress "127.0.0.1:9330" --privatekey "112t8rnaLC8yRN5im7BgETP2y6nDbWrxfn2sfQaJvDqV7siRoLLaYnaehad7dY4L7n3dTd4XbYFfbr867vFq2uqCm36PmTq9usop6oH3MKQf" --nodemode "auto" --datadir "data/shard-stake-9" --listen "0.0.0.0:9463" --externaladdress "0.0.0.0:9463" --norpcauth --rpclisten "0.0.0.0:9363"
#fi
####full node
# ./incognito --discoverpeersaddress "0.0.0.0:9330" --nodemode "relay" --relayshards "all" --datadir "data/fullnode" --listen "0.0.0.0:9459" --externaladdress "0.0.0.0:9459" --norpcauth --rpclisten "0.0.0.0:9359"

###### SINGLE_MEMBER
######
## Shard: 0, Role: Proposer
#if [ "$1" == "shard0-proposer" ]; then
#./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rqJHgJp2TPpNpLNx34aWHB5VH5Pys3hVjjhhf9tctVeCNmX2zQLBqzHau6LpUbSV52kXtG2hRZsuYWkXWF5kw2v24RJq791fWmQxVqy" --nodemode "auto" --datadir "data/shard-0" --listen "0.0.0.0:9460" --externaladdress "0.0.0.0:9460" --norpcauth --rpclisten "0.0.0.0:9360"
#fi
## Shard: 1, Role: Proposer
#if [ "$1" == "shard1-proposer" ]; then
#./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rrEEcDQBMnUM5J17qniHZZmckmr8LGCv9nBjP9x5wmGFGUryKTNvEAf1jh2wwW69rxwtANq4m8JmzowfKVPayUHPmAKdwQw5718GKuH" --nodemode "auto" --datadir "data/shard-1" --listen "0.0.0.0:9461" --externaladdress "0.0.0.0:9461" --norpcauth --rpclisten "0.0.0.0:9361"
#fi
## Beacon, Role: Proposer
#if [ "$1" == "shard2-proposer" ]; then
#./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rotpPVSeHrknwVUTLQgy2avatUWKh2oV9EjVMw6eEtwyJT1FsrHGzvBaLpHL4gPVfJjuSUWvTtiTKuWGNNwGuLo8SHCgfA36ttJ5J7u" --nodemode "auto" --datadir "data/beacon" --listen "0.0.0.0:9462" --externaladdress "0.0.0.0:9462" --norpcauth --rpclisten "0.0.0.0:9362"
#fi
## Relay node
#if [ "$1" == "relaynode" ]; then
#./incognito --relayshards "all" --datadir "data/relaynode" --listen "0.0.0.0:9463" --externaladdress "0.0.0.0:9463" --norpcauth --rpclisten "0.0.0.0:9363" --enablewallet --wallet "wallet3" --walletpassphrase "12345678" --walletautoinit
#fi
