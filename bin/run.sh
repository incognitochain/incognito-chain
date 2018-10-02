if [ -z "$1" ]
then
    echo "Please enter How many node(s) to start"
    exit 0
fi

KEY=$1

cd ~/go/src/github.com/ninjadotorg/cash-prototype
cd privacy/server/build
./main > privacy.log &
cd ../../../
go build
./cash-prototype --discoverpeers --generate --sealerkeyset $KEY
