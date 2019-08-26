#!/usr/bin/env bash
if [ "$1" == "s01" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1E8JuGoey6ddhPwtkSF4fRDymyHjiFCyZbvrvxc8KtjvZBmRFE" --nodemode "auto" --datadir "data/s01" --listen "0.0.0.0:8334" --externaladdress "0.0.0.0:8334" --norpcauth --rpclisten "0.0.0.0:9334" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "s02" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1NLdnvXktVuV1Xy8mQPoQRjPHmQNTd6iKteM22ocKYkhFTjdJC" --nodemode "auto" --datadir "data/s02" --listen "0.0.0.0:8335" --externaladdress "0.0.0.0:8335" --rpclisten "0.0.0.0:9335" --norpc
fi
if [ "$1" == "s03" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12uBM6EUCLyStZUzqC8CSQPCybvibiQS52ZMnFS1woUBv86f5FF" --nodemode "auto" --datadir "data/s03" --listen "0.0.0.0:8336" --externaladdress "0.0.0.0:8336" --rpclisten "0.0.0.0:9336" --norpc
fi
if [ "$1" == "s04" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1bNcBGmmAfdqB1Nx4xSfUAvLRG3qRcHuk3jbSpqnUs6j32gYKt" --nodemode "auto" --datadir "data/s04" --listen "0.0.0.0:8337" --externaladdress "0.0.0.0:8337" --rpclisten "0.0.0.0:9337" --norpc
fi

if [ "$1" == "s11" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:17GfWKosbMAqfPafEjmjLqix4Fa5crsEqsrVG2Y2VHkq41QBwj" --nodemode "auto" --datadir "data/s11" --listen "0.0.0.0:8344" --externaladdress "0.0.0.0:8344" --norpcauth --rpclisten "0.0.0.0:9344" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "s12" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1yMnajT3rmW1dmvPPVMWvwusiL8oC177QDHQwjHT1eAeaMTQCp" --nodemode "auto" --datadir "data/s12" --listen "0.0.0.0:8345" --externaladdress "0.0.0.0:8345" --rpclisten "0.0.0.0:9345" --norpc
fi
if [ "$1" == "s13" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12psUXJ1cxyxYn4nBNGrwykNmuDowQGfmzrH8u7ag8bjLSr7ski" --nodemode "auto" --datadir "data/s13" --listen "0.0.0.0:8346" --externaladdress "0.0.0.0:8346" --rpclisten "0.0.0.0:9346" --norpc
fi
if [ "$1" == "s14" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:122Tsv1QQ5zm5ntuSJrLyB8HbucT17c9dK9JVt5kR7gY9SybvbJ" --nodemode "auto" --datadir "data/s14" --listen "0.0.0.0:8347" --externaladdress "0.0.0.0:8347" --rpclisten "0.0.0.0:9347" --norpc
fi

if [ "$1" == "s21" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12Cb6a85KDuSU3KudG3cak6LSdDfHrBUap1DUnFDjht5h1DvKTX" --nodemode "auto" --datadir "data/s21" --listen "0.0.0.0:8354" --externaladdress "0.0.0.0:8354" --norpcauth --rpclisten "0.0.0.0:9354" --enablewallet --wallet "wallet1" --walletpassphrase "12345678" --walletautoinit
fi
if [ "$1" == "s22" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1aYmLwnmkhqT2tDgKYgZ8jpgKbdtkckBshKu54zr9j8FJaB6vW" --nodemode "auto" --datadir "data/s22" --listen "0.0.0.0:8355" --externaladdress "0.0.0.0:8355" --rpclisten "0.0.0.0:9355" --norpc
fi
if [ "$1" == "s23" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1bbsenjzGGxhr3MHrAh5gwAceFFbrNRnXL4Mk4VCi22X4CsvSr" --nodemode "auto" --datadir "data/s23" --listen "0.0.0.0:8356" --externaladdress "0.0.0.0:8356" --rpclisten "0.0.0.0:9356" --norpc
fi
if [ "$1" == "s24" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1JSjpQW8g5ZvbJSPxne1GyXXJe8jeMsEJTvVz7nbvZHwFUB1C8" --nodemode "auto" --datadir "data/s24" --listen "0.0.0.0:8357" --externaladdress "0.0.0.0:8357" --rpclisten "0.0.0.0:9357" --norpc
fi

if [ "$1" == "b1" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:12u5h8om1cj8c13HJfgSutRVHxiviV352Ma4WrpJL8c3pbZYTTK" --nodemode "auto" --datadir "data/b1" --listen "0.0.0.0:8434" --externaladdress "0.0.0.0:8434" --norpcauth --rpclisten "0.0.0.0:9434"
fi
if [ "$1" == "b2" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1gagNeWHJMEHUhrzM5qbV3fHPXoreaFKBBmghsdnhYXmTjHE6Y" --nodemode "auto" --datadir "data/b2" --listen "0.0.0.0:8435" --externaladdress "0.0.0.0:8435" --rpclisten "0.0.0.0:9435" --norpc
fi
if [ "$1" == "b3" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1D4VB7gWVEBFjW17ZDeh4TZNW37rbdiV7oPTxtHsinjHTKnBMZ" --nodemode "auto" --datadir "data/b3" --listen "0.0.0.0:8436" --externaladdress "0.0.0.0:8436" --rpclisten "0.0.0.0:9436" --norpc
fi
if [ "$1" == "b4" ]; then
go run *.go --discoverpeersaddress "0.0.0.0:9330" --miningkeys "bls:1G2XoWV8baWzZWRsGD1sHMMPMQzb8biAwj5ex629nxRusFiTBC" --nodemode "auto" --datadir "data/b4" --listen "0.0.0.0:8437" --externaladdress "0.0.0.0:8437" --rpclisten "0.0.0.0:9437" --norpc
fi
if [ "$1" == "3" ]; then
go run *.go --miningkeys "bls:12v9jokoYDWFQ7731mR6QjjkvUG1Q" --nodemode "auto" --datadir "data/3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337"
fi
