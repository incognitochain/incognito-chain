if [ -z "$1" ]
then
    echo "Please enter Key ./run <key>"
    exit 0
fi

KEY=$1

cd ~/go/src/github.com/ninjadotorg/cash-prototype
cd privacy/server/build
sudo ./main > privacy.log &
cd ../../../
go build
./cash-prototype --discoverpeers --generate --sealerkeyset $KEY
