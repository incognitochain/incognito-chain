# prototype

Pre: install glide https://github.com/Masterminds/glide

- Clone Project
- Run "glide install"
- Run "go build ./"
- node 1 "./cash-prototype"
- node 2 "./cash-prototype -listen “127.0.0.1:9555” --norpc --datadir “data1" --connect “/ip4/127.0.0.1/tcp/9333/ipfs/QmawrS2w63oXTq9dS8sFYk6ebttLPpdKm7eosTUPx4YGu8” --generate --miningaddr “mgnUx4Ah4VBvtaL7U1VXkmRjKUk3h8pbst”"