#!/bin/sh
mkdir -p /data
cron

if [ "$1" == "y" ]; then
    find /data -maxdepth 1 -mindepth 1 -type d | xargs rm -rf
fi

if [ -z "$NAME" ]; then
    NAME="miner";
fi

if [ -z "$TESTNET" ]; then
    TESTNET=true;
fi

if [ -z $BOOTNODE_IP ]; then
    BOOTNODE_IP="testnet-bootnode.incognito.org:9330";
fi

if [ -z $MONITOR ]; then
   export MONITOR="http://51.91.72.45:33333";
fi

if [ -z "$NODE_PORT" ]; then
    NODE_PORT=9433;
fi

if [ -z "$LIMIT_FEE" ]; then
    LIMIT_FEE=1000000;
fi

if [ -z "$LOG_LEVEL" ]; then
    LOG_LEVEL="info";
fi

if [ -z "$PUBLIC_IP" ]; then
    PUBLIC_IP=`dig -4 @resolver1.opendns.com A myip.opendns.com. +short`;
fi
CONTRACT_IP=`echo $PUBLIC_IP | cut -d '.' -f 1,4`

if [ -z "$RPC_PORT" ]; then RPC_PORT=9334; fi

if [ -z "$WS_PORT" ]; then WS_PORT=19334; fi

IS_BACKUP=""
if [ -n "$BACKUP" ]; then IS_BACKUP="--backup"; fi

BOOTSTRAP_ADD=""
if [ -n "$BOOTSTRAP" ]; then BOOTSTRAP_ADD="--bootstrap ${BOOTSTRAP}" ; fi

IS_FULL_VALIDATION=""
if [ "$FULL_VALIDATION" == "1" ]; then
  IS_FULL_VALIDATION="--is_full_validation";
fi

if [ -z "$NUM_INDEXER_WORKERS" ]; then
    NUM_INDEXER_WORKERS=0;
fi

if [ -n "$FULLNODE" ] &&  [ "$FULLNODE" = "1" ]; then
    echo ./incognito --relayshards "all" $IS_BACKUP $BOOTSTRAP_ADD --usecoindata --coindatapre="__coins__" --numindexerworkers=$NUM_INDEXER_WORKERS --indexeraccesstoken=$INDEXER_ACCESS_TOKEN $IS_FULL_VALIDATION --discoverpeers --discoverpeersaddress $BOOTNODE_IP --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --enablewallet --wallet "wallet" --walletpassphrase "12345678" --walletautoinit --limitfee $LIMIT_FEE --norpcauth --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" --loglevel "$LOG_LEVEL" > cmd.sh
    ./incognito --relayshards "all" $IS_BACKUP $BOOTSTRAP_ADD --usecoindata --coindatapre="__coins__" --numindexerworkers=$NUM_INDEXER_WORKERS --indexeraccesstoken=$INDEXER_ACCESS_TOKEN $IS_FULL_VALIDATION --discoverpeers --discoverpeersaddress $BOOTNODE_IP --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --enablewallet --wallet "wallet" --walletpassphrase "12345678" --walletautoinit --limitfee $LIMIT_FEE --norpcauth --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" --loglevel "$LOG_LEVEL" --rpcmaxclients 1500 2>/data/error.log | cronolog /data/$CONTRACT_IP-%Y-%m-%d.log
elif [ -n "$PRIVATEKEY" ]; then
    echo ./incognito --relayshards "$RELAY_SHARD" $BOOTSTRAP_ADD --usecoindata --coindatapre="__coins__" --numindexerworkers=$NUM_INDEXER_WORKERS --indexeraccesstoken=$INDEXER_ACCESS_TOKEN $IS_FULL_VALIDATION --limitfee $LIMIT_FEE --discoverpeers --discoverpeersaddress $BOOTNODE_IP --privatekey $PRIVATEKEY --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" --loglevel "$LOG_LEVEL" > cmd.sh
    ./incognito  --relayshards "$RELAY_SHARD" $BOOTSTRAP_ADD --usecoindata --coindatapre="__coins__" --numindexerworkers=$NUM_INDEXER_WORKERS --indexeraccesstoken=$INDEXER_ACCESS_TOKEN $IS_FULL_VALIDATION --limitfee $LIMIT_FEE --discoverpeers --discoverpeersaddress $BOOTNODE_IP --privatekey $PRIVATEKEY --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT"  --loglevel "$LOG_LEVEL" 2>/data/error.log | cronolog /data/$CONTRACT_IP-%Y-%m-%d.log
elif [ -n "$MININGKEY" ]; then
    echo ./incognito --relayshards "$RELAY_SHARD" $BOOTSTRAP_ADD --usecoindata --coindatapre="__coins__" --numindexerworkers=$NUM_INDEXER_WORKERS --indexeraccesstoken=$INDEXER_ACCESS_TOKEN $IS_FULL_VALIDATION --limitfee $LIMIT_FEE --discoverpeers --discoverpeersaddress $BOOTNODE_IP --miningkeys $MININGKEY --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" --loglevel "$LOG_LEVEL" > cmd.sh
    ./incognito --relayshards "$RELAY_SHARD" $BOOTSTRAP_ADD --usecoindata --coindatapre="__coins__" --numindexerworkers=$NUM_INDEXER_WORKERS --indexeraccesstoken=$INDEXER_ACCESS_TOKEN $IS_FULL_VALIDATION --limitfee $LIMIT_FEE --discoverpeers --discoverpeersaddress $BOOTNODE_IP --miningkeys $MININGKEY --datadir "/data" --listen "0.0.0.0:$NODE_PORT" --externaladdress "$PUBLIC_IP:$NODE_PORT" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "12345678" --walletautoinit --rpclisten "0.0.0.0:$RPC_PORT" --rpcwslisten "0.0.0.0:$WS_PORT" --loglevel "$LOG_LEVEL" 2>/data/error.log | cronolog /data/$CONTRACT_IP-%Y-%m-%d.log
fi
