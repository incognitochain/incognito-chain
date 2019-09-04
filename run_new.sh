#!/usr/bin/env bash
if [ "$1" == "s01" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1oHW8buNXixuCZoNLxkyHfdJP6iJosFDJUgbDiMgVc2amfj6H7" --nodemode "auto" --datadir "data/s01" --listen "0.0.0.0:8334" --externaladdress "0.0.0.0:8334" --norpcauth --rpclisten "0.0.0.0:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "s02" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12UXBbiNwmrBPHQZr9oKGCd5ZxTd4dzj6SmKFKw3mcsJ7X1YdK5" --nodemode "auto" --datadir "data/s02" --listen "0.0.0.0:8335" --externaladdress "0.0.0.0:8335" --rpclisten "0.0.0.0:9335" --norpc
fi
if [ "$1" == "s03" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12AY9ACUVXNWCUv6KTERsvPigdgXqKyet3HW6dUjmsfyn7CkEhM" --nodemode "auto" --datadir "data/s03" --listen "0.0.0.0:8336" --externaladdress "0.0.0.0:8336" --rpclisten "0.0.0.0:9336" --norpc
fi
if [ "$1" == "s04" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1Xm6FiWKXcaamPgXz5gJGF25c4BtmbrGYmc1iqkDr4FTJ24EdU" --nodemode "auto" --datadir "data/s04" --listen "0.0.0.0:8337" --externaladdress "0.0.0.0:8337" --rpclisten "0.0.0.0:9337" --norpc
fi

if [ "$1" == "s11" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1TGvoYkXLNCSEgTjQotkzPWN3W6kFwVj2s2aMnfDjU8ei3tCha" --nodemode "auto" --datadir "data/s11" --listen "0.0.0.0:8344" --externaladdress "0.0.0.0:8344" --norpcauth --rpclisten "0.0.0.0:9344" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "s12" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12G6S27uaYe1d53awSX5Rkn9p5p3T7sunTL758mPY7CEc9SEPeK" --nodemode "auto" --datadir "data/s12" --listen "0.0.0.0:8345" --externaladdress "0.0.0.0:8345" --rpclisten "0.0.0.0:9345" --norpc
fi
if [ "$1" == "s13" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1YVKsf8gdaKSNYbAndFpEyve9EKFkhBpcLpuKvkjZLkGxHjYx3" --nodemode "auto" --datadir "data/s13" --listen "0.0.0.0:8346" --externaladdress "0.0.0.0:8346" --rpclisten "0.0.0.0:9346" --norpc
fi
if [ "$1" == "s14" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1R9x3TBwYrKd8PmQjufd5W4Pko1UGsfTqzPUtiCzGazAFSUWMa" --nodemode "auto" --datadir "data/s14" --listen "0.0.0.0:8347" --externaladdress "0.0.0.0:8347" --rpclisten "0.0.0.0:9347" --norpc
fi

if [ "$1" == "s21" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1Sfx1VzWsZEfhTgC2umSki7SKQQQ1dL8F8xNUvB65HaHA9uc5c" --nodemode "auto" --datadir "data/s21" --listen "0.0.0.0:8354" --externaladdress "0.0.0.0:8354" --norpcauth --rpclisten "0.0.0.0:9354" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "s22" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12ThTo7fVqpqHwgyhTERqRNd8YqPHsVgNawbNBQGJMRqLpAD6yW" --nodemode "auto" --datadir "data/s22" --listen "0.0.0.0:8355" --externaladdress "0.0.0.0:8355" --rpclisten "0.0.0.0:9355" --norpc
fi
if [ "$1" == "s23" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12GgJqDQzATnG2CMCTAu6Bnkq8zX6bimuqdy3kN7K8rRTcgU3X6" --nodemode "auto" --datadir "data/s23" --listen "0.0.0.0:8356" --externaladdress "0.0.0.0:8356" --rpclisten "0.0.0.0:9356" --norpc
fi
if [ "$1" == "s24" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12wP2tDBK8dYq8UcKzwxZtS19JhyjUQZS1eE5uFbtckeSQbZBfg" --nodemode "auto" --datadir "data/s24" --listen "0.0.0.0:8357" --externaladdress "0.0.0.0:8357" --rpclisten "0.0.0.0:9357" --norpc
fi
if [ "$1" == "st1" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12WQy3WYgtdp1G52hcdpv9zhW2wRg9DnzcM5pZM1cFhNo3VA6pB" --nodemode "auto" --datadir "data/st1" --listen "0.0.0.0:8364" --externaladdress "0.0.0.0:8364" --rpclisten "0.0.0.0:9364" --norpc
fi
if [ "$1" == "st2" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1UCU4M3U6XXJTgjcpT7e5fY9swXnM74sWSoP6Qyx9CD4GKSKGG" --nodemode "auto" --datadir "data/st2" --listen "0.0.0.0:8365" --externaladdress "0.0.0.0:8365" --rpclisten "0.0.0.0:9365" --norpc
fi
if [ "$1" == "b1" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1Gmyq2e1nYqkAjkyFcjwmgPHSPTBBuh7PUEQsgEbwdiD1J2Rwp" --nodemode "auto" --datadir "data/b1" --listen "0.0.0.0:8434" --externaladdress "0.0.0.0:8434" --norpcauth --rpclisten "0.0.0.0:9434"
fi
if [ "$1" == "b2" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1g2N8sSA1S4yJRca8nhaqjnxmjziRuF1oXaDPBAqHMQfZLfrNf" --nodemode "auto" --datadir "data/b2" --listen "0.0.0.0:8435" --externaladdress "0.0.0.0:8435" --rpclisten "0.0.0.0:9435" --norpc
fi
if [ "$1" == "b3" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12UR2apXz6bTzaAUiCj83YB5F7c9k9Toe4JQwmtySSWWLU4MWRH" --nodemode "auto" --datadir "data/b3" --listen "0.0.0.0:8436" --externaladdress "0.0.0.0:8436" --rpclisten "0.0.0.0:9436" --norpc
fi
if [ "$1" == "b4" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1dhmiTWb7M6KPqqPdXJbyJZRwu9pYxT1d1YD7bh1EpuVqHsWi2" --nodemode "auto" --datadir "data/b4" --listen "0.0.0.0:8437" --externaladdress "0.0.0.0:8437" --rpclisten "0.0.0.0:9437" --norpc
fi
if [ "$1" == "3" ]; then
go run *.go --miningkeys "bls:12v9jokoYDWFQ7731mR6QjjkvUG1Q" --nodemode "auto" --datadir "data/3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi
