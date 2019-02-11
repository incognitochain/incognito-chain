#!/usr/bin/env bash

kill -9 $(pgrep -u cashadmin bootnode)

cd ~/go/src/github.com/ninjadotorg/constant/bootnode
git pull
/usr/local/go/bin/go build main.go -o ./bootnode
./bootnode
