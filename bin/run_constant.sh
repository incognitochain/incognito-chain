#!/usr/bin/env bash

KEY=""

kill -9 $(pgrep -u cashadmin constant)

cd ~/go/src/github.com/ninjadotorg/constant
git pull
/usr/local/go/bin/go build -o ./constant
./constant
