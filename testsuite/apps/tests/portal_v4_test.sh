#!/bin/bash -e

function cleanup {
  docker rm -f bitcoinlocal
}
trap cleanup EXIT

docker run -d -p 18443:18443 -p 18444:18444 --name bitcoinlocal ruimarinho/bitcoin-core -printtoconsole -regtest=1 -rpcallowip=172.17.0.0/16 -rpcbind=0.0.0.0 -rpcport=18443 -port=18444 -server=1 -rpcuser=admin -rpcpassword=123123AZ -fallbackfee=1 -maxtxfee=1000000 -bytespersigop=1000 -datacarriersize=1000

echo "Init environment ..."
sleep 100s



