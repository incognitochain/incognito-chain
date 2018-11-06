if [ -z "$1" ]
then
    echo "Please enter Key ./run <key>"
    exit 0
fi

KEY=$1

cd ~/go/src/github.com/ninjadotorg/constant
cd privacy/server/build
sudo ./main > privacy.log &
echo "Started privacy..."

cd ../../../
/usr/local/go/bin/go build
./cash-prototype --discoverpeers --generate --producerkeyset $KEY
