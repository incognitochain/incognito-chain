echo "check exist and install brew libs"

if [ $(which brew) == "brew not found" ]; then
  echo "Please install homebrew first: https://brew.sh/"
  exit 1
else
  echo "brew has been installed"
fi

if !(brew ls --versions boost > /dev/null); then
  # install boost
  echo "install boost ... "
  echo "brew install boost"
  brew install boost
else
  echo "boost has been installed"
fi

if !(brew ls --versions libsodium > /dev/null); then
  # install lib sodium
  echo "install libsodium ... "
  echo "brew install libsodium"
  brew install libsodium
else
  echo "libsodium has been installed"
fi

if !(brew ls --versions gmp > /dev/null); then
  # install gmp
  echo "install gmp ... "
  echo "brew install gmp"
  brew install gmp
  brew link gmp
else
  echo "gmp has been installed"
fi

if !(brew ls --versions protobuf > /dev/null); then
  # install lib protobuf
  echo "install protobuf ... "
  echo "brew install protobuf"
  brew install protobuf
else
  echo "protobuf has been installed"
fi

if !(brew ls --versions protobuf-c > /dev/null); then
  # install lib protobuf-c
  echo "install protobuf-c ... "
  echo "brew install protobuf-c version"
  brew install protobuf-c
else
  echo "protobuf-c has been installed"
fi

if !(brew ls --versions grpc > /dev/null); then
  # install grpc
  echo "install grpc ... "
  echo "brew install grpc"
  brew install grpc
else
  echo "grpc has been installed"
fi

if !(brew ls --versions cmake > /dev/null); then
  # install lib cmake
  echo "install cmake ... "
  echo "brew install cmake"
  brew install cmake
else
  echo "cmake has been installed"
fi

# # install lib libomp
# echo "install libomp ... "
# echo "brew install cmake"
# brew install libomp

BUILD_DIR="build"
if [ ! -d "$BUILD_DIR" ]; then
  mkdir build
fi

cd build

cmake ..
make

file="./proving.key"
if [ -f "$file" ]; then
	echo "$file found."
else
	wget https://github.com/ninjadotorg/cash/releases/download/zkpp-v0.0.3/proving.key
fi

file="./verifying.key"
if [ -f "$file" ]; then
	echo "$file found."
else
	wget https://github.com/ninjadotorg/cash/releases/download/zkpp-v0.0.3/verifying.key
fi

echo "privacy has been built"
