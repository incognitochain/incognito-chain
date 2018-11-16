echo "Pre-require install"
echo "apt install build-essential autoconf libtool pkg-config unzip -y"
sudo apt install build-essential autoconf libtool pkg-config unzip -y

echo "install Lib"

# install boost
echo "install boost ... "
echo "apt install libboost* -y"
sudo apt install libboost* -y

# install lib sodium
echo "install libsodium ... "
echo "apt install libsodium-dev -y"
sudo apt install libsodium-dev -y

# install gmp
echo "install gmp ... "
echo "apt install libgmp3-dev -y"
sudo apt install libgmp3-dev -y

# install grpc
echo "install grpc ... "
git clone -b $(curl -L http://grpc.io/release) https://github.com/grpc/grpc
cd grpc
git submodule update --init
sudo make -j4
sudo make install
cd ..

# install lib protobuf
echo "install protobuf ... "
sudo curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protobuf-cpp-3.6.1.zip
sudo unzip protobuf-cpp-3.6.1.zip
cd protobuf-3.6.1
sudo ./configure
sudo make -j4
sudo make check
sudo make install
sudo ldconfig # refresh shared library cache.
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
sudo wget http://www.cmake.org/files/v3.5/cmake-3.5.2.tar.gz
sudo tar xf cmake-3.5.2.tar.gz
cd cmake-3.5.2
sudo ./configure
sudo make -j4
sudo make install
cd ..

BUILD_DIR="build"
if [ ! -d "$BUILD_DIR" ]; then
  mkdir build
fi

cd build

sudo cmake ..
sudo make -j4

file="./proving.key"
if [ -f "$file" ]
then
	echo "$file found."
else
	sudo wget https://github.com/ninjadotorg/constant/releases/download/zkpp-v0.0.3/proving.key
fi

file="./verifying.key"
if [ -f "$file" ]
then
	echo "$file found."
else
	sudo wget https://github.com/ninjadotorg/constant/releases/download/zkpp-v0.0.3/verifying.key
fi
