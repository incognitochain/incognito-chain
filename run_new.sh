#!/usr/bin/env bash
if [ "$1" == "s1" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12YVmQc5Lx6De1JonnEnqogAFrTVmaz2isWRVx2DTHXiz7cBYGN" --nodemode "auto" --datadir "data/1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "s2" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1i8dHfnbrHL1ccFdaRKsoHcabxTkj4vdocd4eb4jug9kyNqSQ8" --nodemode "auto" --datadir "data/2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpc
fi
if [ "$1" == "s3" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12GFisa1rgM2khcut4ZjioPtBLDb75zKmrufJnNR2bUt8v266V4" --nodemode "auto" --datadir "data/3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpc
fi
if [ "$1" == "s4" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12pHqTGawG2YXh1MLR97QkrWQ5ZFQL7NrAXogmjhFKk9P1oj4yJ" --nodemode "auto" --datadir "data/4" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpc
fi
if [ "$1" == "b1" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1X6xHXfbtPKF7uoJ1spHttDpYNUMVZQRPWJzAkycQKmvJ7rpRW" --nodemode "auto" --datadir "data/b1" --listen "0.0.0.0:9335" --externaladdress "0.0.0.0:9335" --norpcauth --rpclisten "0.0.0.0:9036"
fi
if [ "$1" == "b2" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12bdqh5Qe9NXfR5Q1Bay43VJ4vM1DYe3mUo6VavGpU9tg3VCS4g" --nodemode "auto" --datadir "data/b2" --listen "0.0.0.0:9336" --externaladdress "0.0.0.0:9336" --norpc
fi
if [ "$1" == "b3" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1mErSm6LC96VUvwETcHMoKzXG2b93bMXJxUwBAG7y7CYu8oToS" --nodemode "auto" --datadir "data/b3" --listen "0.0.0.0:9337" --externaladdress "0.0.0.0:9337" --norpc
fi
if [ "$1" == "b4" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1Su7WUi7UmjjkVMaDB65czif1CsPuJc27NZbhHwKRvRWxnft1v" --nodemode "auto" --datadir "data/b4" --listen "0.0.0.0:9338" --externaladdress "0.0.0.0:9338" --norpc
fi
