#!/usr/bin/env bash

KEY=""
BOOTNODE=""

kill -9 $(pgrep -u root constant)

cd ~/go/src/github.com/ninjadotorg/constant
git pull
/usr/local/go/bin/go build -o constant
./constant
