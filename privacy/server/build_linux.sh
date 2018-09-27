echo "Pre-require install"
echo "apt install build-essential autoconf libtool pkg-config unzip -y"
apt install build-essential autoconf libtool pkg-config unzip -y

echo "install Lib"

# install boost
echo "install boost ... "
echo "apt install libboost* -y"
apt install libboost* -y

# install lib sodium
echo "install libsodium ... "
echo "apt install libsodium-dev -y"
apt install libsodium-dev -y

# install gmp
echo "install gmp ... "
echo "apt install libgmp3-dev -y"
apt install libgmp3-dev -y

# install grpc
echo "install grpc ... "
git clone -b $(curl -L http://grpc.io/release) https://github.com/grpc/grpc
cd grpc
git submodule update --init
make
make install
cd ..

# install lib protobuf
echo "install protobuf ... "
curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protobuf-cpp-3.6.1.zip
unzip protobuf-cpp-3.6.1.zip
cd protobuf-3.6.1
./configure
make
make check
make install
ldconfig # refresh shared library cache.
cd ..

# install lib protobuf-compiler
#echo "install protobuf-compiler... "
#echo "apt install protobuf-compiler -y"
#apt install protobuf-compiler -y

# install lib protobuf-c
#echo "install protobuf-c ... "
#echo "apt install protobuf-c-compiler -y"
#apt install protobuf-c-compiler -y

# install lib cmake
echo "install cmake ... "
echo "apt install cmake -y"
apt install cmake -y

BUILD_DIR="build"
if [ ! -d "$BUILD_DIR" ]; then
  mkdir build
fi

cd build

cmake ..
make

file="./proving.key"
if [ -f "$file" ]
then
	echo "$file found."
else
	wget https://github.com/ninjadotorg/cash-prototype/releases/download/zkpp-v0.0.2/proving.key
fi

file="./verifying.key"
if [ -f "$file" ]
then
	echo "$file found."
else
	wget https://github.com/ninjadotorg/cash-prototype/releases/download/zkpp-v0.0.2/verifying.key
fi

./main
