#!/bin/bash -e
BITCOIND_PATH=/home/david/Downloads/bitcoin-0.21.1
BUILDPROOF_PATH=/home/david/WORK/Autonomous/buildBTCMerkleProof

export PATH=$PATH:$BITCOIND_PATH/bin:$BUILDPROOF_PATH

function cleanup {
  if [ -n "$bitcoindPID" ] && ps -p $bitcoindPID > /dev/null; then
    echo "Kill bitcoin pid $bitcoindPID"
    kill -9 $bitcoindPID
    rm -rf ./regtest
  fi
}
trap cleanup EXIT

if [ "$1" = "send" ]; then
  sendAddress=$3
  sendTx=`bitcoin-cli -rpcport="18443" -rpcuser="test" -rpcpassword="pass" sendtoaddress $sendAddress $2 "" "" true`
  bitcoin-cli -rpcport="18443" -rpcuser="test" -rpcpassword="pass" -generate 5 > /dev/null;
  buildProof -host "localhost:18443" -username "test" -password "pass" -txhash $sendTx;

elif [ "$1" = "sendrawtransaction" ]; then
  bitcoin-cli -rpcport=18443 -rpcuser="test" -rpcpassword="pass" sendrawtransaction $2
  bitcoin-cli -rpcport="18443" -rpcuser="test" -rpcpassword="pass" -generate 5 > /dev/null;
elif [ "$1" = "listunspent" ]; then
  bitcoin-cli -rpcport=18443 -rpcuser=test -rpcpassword=pass importaddress $2 "" true
  bitcoin-cli -rpcport=18443 -rpcuser=test -rpcpassword=pass listunspent 0 999999 [\"$2\"]
else
  rm -rf ./regtest
  echo "Start bitcoind ..."
  bitcoind -regtest -rpcport=18443 -datadir=./ -printtoconsole -server -rpcuser=test -rpcpassword=pass -fallbackfee=1 -maxtxfee=1000000 -bytespersigop=1000 -datacarriersize=1000 > test_bitcoind.log &
  bitcoindPID=$!
  sleep 1s
  bitcoin-cli -rpcport=18443 -rpcuser=test -rpcpassword=pass createwallet mywallet > /dev/null
  newAddress=`bitcoin-cli -rpcport=18443 -rpcuser=test -rpcpassword=pass getnewaddress`
  bitcoin-cli -rpcport=18443 -rpcuser=test -rpcpassword=pass generatetoaddress 100 $newAddress > /dev/null
  bitcoin-cli -rpcport=18443 -rpcuser=test -rpcpassword=pass -generate 100 > /dev/null
  echo "Check init balance ..."
  bitcoin-cli -rpcport=18443 -rpcuser=test -rpcpassword=pass getbalances
  go run *.go
fi



#go run *.go

