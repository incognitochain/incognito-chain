#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard0_keyset_2_privatekey" --nodemode "auto" --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334" 
fi
if [ "$1" == "shard0-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard0_keyset_3_privatekey" --nodemode "auto" --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335" 
fi
if [ "$1" == "shard0-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard0_keyset_4_privatekey" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336" --rpcwslisten "0.0.0.0:19336" 
fi
if [ "$1" == "shard0-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard0_keyset_5_privatekey" --nodemode "auto" --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337" --rpcwslisten "0.0.0.0:19337" 
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard1_keyset_2_privatekey" --nodemode "auto" --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --rpcwslisten "0.0.0.0:19338" 
fi
if [ "$1" == "shard1-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard1_keyset_3_privatekey" --nodemode "auto" --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "0.0.0.0:19339" 
fi
if [ "$1" == "shard1-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard1_keyset_4_privatekey" --nodemode "auto" --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340" --rpcwslisten "0.0.0.0:19340" 
fi
if [ "$1" == "shard1-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard1_keyset_5_privatekey" --nodemode "auto" --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341" --rpcwslisten "0.0.0.0:19341" 
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard0_keyset_0_privatekey" --nodemode "auto" --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350" 
fi
if [ "$1" == "beacon-1" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard0_keyset_1_privatekey" --nodemode "auto" --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351" 
fi
if [ "$1" == "beacon-2" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard1_keyset_0_privatekey" --nodemode "auto" --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352" 
fi
if [ "$1" == "beacon-3" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "shard1_keyset_1_privatekey" --nodemode "auto" --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353" 
fi
# FullNode
if [ "$1" == "full_node" ]; then
./incognito --discoverpeersaddress "0.0.0.0:9330" --nodemode "relay" --datadir "data/full_node" --listen "0.0.0.0:9454" --externaladdress "0.0.0.0:9454" --norpcauth --rpclisten "0.0.0.0:9354" --relayshards "all"  --txpoolmaxtx 100000
fi
######
