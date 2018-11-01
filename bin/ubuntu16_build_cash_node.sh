sudo apt update
sudo apt -y upgrade

echo "Install packages"
sudo apt install git

echo "Install golang..."
wget https://dl.google.com/go/go1.10.2.linux-amd64.tar.gz
sudo tar -xvf go1.10.2.linux-amd64.tar.gz
sudo mv go /usr/local

echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.bashrc

source ~/.bashrc

mkdir ~/go/bin -p
mkdir ~/go/src/github.com/ninjadotorg -p

echo "Install glide..."
curl https://glide.sh/get | sh

echo "Clone cash-prototype..."
cd go/src/github.com/ninjadotorg
git clone https://github.com/ninjadotorg/constant -b Parallel-PoS-Privacy

echo "Install cash-prototype packages..."
cd cash-prototype
glide install

echo "Build privacy..."
cd privacy/server
sudo bash ./build_linux.sh
