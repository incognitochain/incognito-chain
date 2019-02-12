#!/usr/bin/env bash

kill -9 $(pgrep -u root bootnode)

cd ~/go/src/github.com/ninjadotorg/constant/bootnode
git pull
/usr/local/go/bin/go build -o bootnode *.go
./bootnode
cd ~/go/src/github.com/ninjadotorg/constant/bin
