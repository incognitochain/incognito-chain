echo "install Lib"


# install boost
echo "install boost ... "
echo "brew install boost"
brew install boost

# install lib sodium
echo "install libsodium ... "
echo "brew install libsodium"
brew install libsodium

# install gmp
echo "install gmp ... "
echo "brew install gmp"
brew install gmp
brew link gmp

# install lib protobuf
echo "install protobuf ... "
echo "brew install protobuf"
brew install protobuf

# install lib protobuf-c
echo "install protobuf-c ... "
echo "brew install protobuf-c version"
brew install protobuf-c

# install grpc
echo "install grpc ... "
echo "brew install grpc"
brew install grpc

# install lib cmake
echo "install cmake ... "
echo "brew install cmake"
brew install cmake

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
	wget https://github.com/ninjadotorg/cash-prototype/releases/download/zkpp-v0.0.1/proving.key
fi

file="./verifying.key"
if [ -f "$file" ]
then
	echo "$file found."
else
	wget https://github.com/ninjadotorg/cash-prototype/releases/download/zkpp-v0.0.1/verifying.key
fi

./main
