#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXB47RhSdyVRU41TEf78nxbtWGtmjutwSp9YqsNaCpFxQGXcnwcXTtBkCGDk1KLBRBeWMvb2aXG5SeDUJRHtFV8jTB3weHEkbMJ1AL" --nodemode "auto" --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit --rpcwslisten "0.0.0.0:19334" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard0-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXVdfBqBMigSs5fm9NSS8rgsVVURUxArpv6DxYmPZujKqomqUa2H9wh1zkkmDGtDn2woK4NuRDYnYRtVkUhK34TMfbUF4MShSkrCw5" --nodemode "auto" --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard0-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXi8eKJ5RYJjyQYcFMThfbXHgaL6pq5AF5bWsDXwfsw8pqQUreDv6qgWyiABoDdphvqE7NFr9K92aomX7Gi5Nm1e4tEoV3qRLVdfSR" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard0-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY42xRqJghQX3zvhgEa2ZJBwSzJ46SXyVQEam1yNpN4bfAqJwh1SsobjHAz8wwRvwnqJBfxrbwUuTxqgEbuEE8yMu6F14QmwtwyM43" --nodemode "auto" --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXBPJQWJTyPdzWsfsUCFTDhcas3y2MYsauKo66euh1udG8dSh2ZszSbfqHwCpYHPRSpFTxYkUcVa619XUM6DjdV7FfUWvYoziWE2Bm" --nodemode "auto" --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --enablewallet --wallet "wallet2" --walletpassphrase "12345678" --walletautoinit --rpcwslisten "127.0.0.1:19338" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard1-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXN2SLxQncPYvFdzEivznKjBxK5byYmPbAhnEEv8TderLG7NUD7nwAEDu7DJ7pnCKw9N5PuTuELCHz8qKc7z9S9jF8QG41u7Vomc6L" --nodemode "auto" --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "127.0.0.1:19339" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard1-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXs5os49h71E7utfHatnWGQnirbVF2b5Ua8h1ttidk1S5AFcUqHCDmpMziiFC15BG8W1LQKK5tYcvr2CM7DyYgsfVmAWYh4kQ6f33T" --nodemode "auto" --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "shard1-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXvcE6sxwt7nQ6mH6KdPMWyQRv6xAd3WWorzS7k26YPjm4mvFtC51bRaU18yubQm1N3gBeDJJyXqWmxi5QdCkqYExCEkSqNpD1Wzpz" --nodemode "auto" --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
# Shard 2
if [ "$1" == "shard2-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnX2Fngsy1KuJv5GLLNUeB3gZwoHMss2cqRr1ECa2ibR7FQNUyE7kMFvq7rGtqVJULo8XfAxThuLxUwd8vv76MojbL3wPhxmTvbcd2S" --nodemode "auto" --datadir "data/shard2-0" --listen "0.0.0.0:9442" --externaladdress "0.0.0.0:9442" --norpcauth --rpclisten "0.0.0.0:9342"
fi
if [ "$1" == "shard2-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXRrZ1gC7MYNFVUA1paZrE3iSiAb9AR7Z5quNBgR2ovrcfj8p4kTb3ynx6ddjnoPey3qA2vRiP17tCvpCHU9xBDwMq8D1Mg2GBM9eC" --nodemode "auto" --datadir "data/shard2-1" --listen "0.0.0.0:9443" --externaladdress "0.0.0.0:9443" --norpcauth --rpclisten "0.0.0.0:9343"
fi
if [ "$1" == "shard2-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXbsJF4f5xtzM2jPW6dCYHChNksth7X64iUkLR4bGi2JBVjgJQRLeKRsdbiFaYMsxzrfbfKAp4TELGre45QkxHWCnwVXPGnnZjJKVL" --nodemode "auto" --datadir "data/shard2-2" --listen "0.0.0.0:9444" --externaladdress "0.0.0.0:9444" --norpcauth --rpclisten "0.0.0.0:9344"
fi
if [ "$1" == "shard2-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY4DeSGZYb8r8sSN5WJr6ZL3NCafYAQ7f7Am9KXQDGSc3Qddpn7BfHW1i6CoVVk8vKEzJ25vA9uc9EdhoLU98eoUw7fMrPPrBdNB7Q" --nodemode "auto" --datadir "data/shard2-3" --listen "0.0.0.0:9445" --externaladdress "0.0.0.0:9445" --norpcauth --rpclisten "0.0.0.0:9345"
fi
# Shard 3
if [ "$1" == "shard3-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXBg35mhRzZDw7nNQCiRcM9QLg2DTHewY7kRZ3XCmgkdQa4iVcMUwMd5DTvvjEcCvv3SnCo5zSQpS93zskAxG6tdvR1QPBxtCmaBCK" --nodemode "auto" --datadir "data/shard3-0" --listen "0.0.0.0:9446" --externaladdress "0.0.0.0:9446" --norpcauth --rpclisten "0.0.0.0:9346"
fi
if [ "$1" == "shard3-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXa3USo2wMdyuZTfurZrYe1ZJhqibnp9GHx8Jkf9dh5cU39sjqBTKoWPtNHvVZn2eqGj6V26PmELez85bUUBMBKG6tqFQrer2GkuJ4" --nodemode "auto" --datadir "data/shard3-1" --listen "0.0.0.0:9447" --externaladdress "0.0.0.0:9447" --norpcauth --rpclisten "0.0.0.0:9347"
fi
if [ "$1" == "shard3-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXc38wPUsDdRasThfk58ME3KoSdeyXJS3tsyRuxPrSkvuc8MBT1gxiQjSCMBbifqibxEVAHimGcfnLbjVEEett8FtwFuBQ8zAUHTet" --nodemode "auto" --datadir "data/shard3-2" --listen "0.0.0.0:9448" --externaladdress "0.0.0.0:9448" --norpcauth --rpclisten "0.0.0.0:9348"
fi
if [ "$1" == "shard3-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY6BgeU6bk7Nui1TwB2w2MsG3UeSXrgJ1n2tGVs66Qk9YT6E15PbXR4ai7eW1qyPrW7a2AUtN2otuXtBAZKtm2DDiUxh3ngZ6JPYHv" --nodemode "auto" --datadir "data/shard3-3" --listen "0.0.0.0:9449" --externaladdress "0.0.0.0:9449" --norpcauth --rpclisten "0.0.0.0:9349"
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXCerQX2RRd8KhPfsFCj2rrBYUx42FZJKgRFcdBfg36Mid3ygKyMn5LSc5LBHsxqapRaN6xMav7bGhA6VtGUzNNYuA9Y78CB5oGkti" --nodemode "auto" --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "beacon-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXYgxipKvTJJfHg7tQhcdmA2R1jPpCPmXg37Xi1VfgrFzWFuNy4U6828q1yfbD7VEdutD63HfVYAqL6U32joXVjqdkfUP52LnNGXda" --nodemode "auto" --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "beacon-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnXe3Jxg5d1Rejg2fB1NwnqNsr94RCT3PX14h5NNDjrdgLeEWFkqcMNamKCHask1Gx46g5WYZDKHKx7kzLVD7h1cgvU6NxNijkyGmA9" --nodemode "auto" --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
fi
if [ "$1" == "beacon-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY2gqonwhnhGD6rKeEXkbJDB7DHUtZQKC8SfLci6ABb5eCEj4o7ezWBZWaGbu7CJ1R1mrADGqmRjugg42GeA6jhaXbNDeP2HUr8udw" --nodemode "auto" --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353" --btcclient 1 --btcclientip "159.65.142.153" --btcclientport "8332" --btcclientusername "admin" --btcclientpassword "autonomous"
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
