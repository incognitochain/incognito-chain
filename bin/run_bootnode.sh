#!/bin/sh

if [ -z $NODE_PORT ]; then
    NODE_PORT=9433;
fi

/bootnode --rpcport $NODE_PORT
