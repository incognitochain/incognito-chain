<img src="https://i.postimg.cc/CMzvFrhs/Screen-Shot-2021-01-20-at-12-29-04-PM.png" width="150">


## What is Incognito? 

We now live in a token economy. Everything is being tokenized. Fiats. Gold. Bond. Storage. Computing Power. Digital Art. Collectibles. Buildings. More.

We believe the biggest problem to this new token economy is the lack of privacy. Current blockchains display all transactions publicly — so anyone can see your balances or track your activity. Given the choice, we strongly believe that very few people will willingly disclose their crypto financials to the entire world. Incognito offers anyone the option to turn on privacy mode in this new token economy.

That's why we work to create solutions that make privacy widely accessible and incredibly simple for anyone, anywhere on earth, to exercise their right to that choice.

Incognito is live. Give it a try!

* Power Incognito: [Get a Node Device](https://incognito.org/) or [Host a Virtual Node](https://we.incognito.org/t/how-to-host-a-virtual-node/194)

* Use Incognito: Get the Incognito Wallet on [iOS](https://apps.apple.com/us/app/incognito-crypto-wallet/id1475631606?ls=1) or [Android](https://play.google.com/store/apps/details?id=com.incognito.wallet)

## Github branches and running environments

<img src="https://i.postimg.cc/d0C5Bxpg/Screen-Shot-2021-01-20-at-12-21-50-PM.png" width="600" height="400">

In Incognito’s Git version control, there are three main branches, first for core team development, second for release and third for production:

* **development** (on Testnet 1): this is a branch that developers would branch from for building feature X and create pull requests to merge into. The code on this branch is usually deployed onto Testnet 1.


* **release** (on Testnet 2): After finishing testing on Testnet 1, the changes from development branch would be merged into this branch then deployed onto Testnet 2. The code on this branch should be almost the same as production’s.


* **production** (on Mainnet): After finishing testing on Testnet 2, the changes from release branch would be merged into this branch then deployed onto Mainnet. In an emergency case, hot fixes would also be branched directly from this branch and merged into it through PR(s). The fixes would be deployed onto Mainnet and then cherry-picked into release & development branches.

## Build the code

### By golang programming language

The below instructions will get you up and running on your local machine for development and testing purposes. Building Incognito requires [Go](http://golang.org/doc/install). Once Go is installed, clone this project to your local GOPATH and build it.

```shell
go build -o incognito
```

Then, run an Incognito Node

```shell
./incognito -n incognito --testnet false --discoverpeers --discoverpeersaddress 51.91.72.45:9330 --miningkeys "your validator key" --nodemode "auto" --datadir "/path/to/data" --listen "0.0.0.0:9334" --externaladdress "0.0.0.0:9334" --norpcauth --enablewallet --wallet "incognito" --walletpassphrase "your wallet passphrase" --walletautoinit --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:9336" --loglevel "info"
```

### By make package

You can simply build and run the code with a few simple command with make package:

1. Clear everything with `make clean` (optional for clearing database)
1. Build with command: `make build`
1. Run with command: `make mainnet`

**Optional**

When running Incognito node with docker environment you can update image by a third party. we recommend https://github.com/containrrr/watchtower by running below command

```
docker run -d \
    --name watchtower \
    -v /var/run/docker.sock:/var/run/docker.sock \
    containrrr/watchtower \
    --cleanup \
    --remove-volumes \
    --include-restarting \
    incognito-mainnet
```

## Read the code

There is more than 1 million lines of code in the Incognito codebase. Below is a quick overview of Incognito architecture and its key components that will help you navigate through the code.

<img src="https://user-images.githubusercontent.com/4691695/67144758-ab0fab80-f2a4-11e9-881d-75432653fd24.png" width="450">

* **P2P Networking**

  * **Peer Management**. Peer management handles peer-to-peer communications such as finding peers, connecting to them, sending and receiving transactions, blocks, and messages. Its code is in the [connmanager](https://github.com/incognitochain/incognito-chain/tree/production/connmanager) package and [addrmanager](https://github.com/incognitochain/incognito-chain/tree/production/addrmanager) package.

  * **NetSync**. NetSync is a mediator that receives incoming messages, parses them, and routes the messages to the right components. Its code is in [netsync](https://github.com/incognitochain/incognito-chain/tree/production/netsync) package.

  * **Highway**. Highway is a new network topology design that speeds up P2P communications. Its code is in [highway](https://github.com/incognitochain/incognito-highway) repo.

* **Blockchain**

  * **Shards**. Shards are subchains. A subchain is a Proof-of-Stake blockchain with its own committee of N nodes. A shard's job is to produces new block via a Practical Byzantine Fault Toloerance (pBFT) consenus algorithm. Its code is in the [blockchain](https://github.com/incognitochain/incognito-chain/tree/production/blockchain) package.

  * **Beacon**. Beacon is also a subchain. A beacon's job is to coordinates the shards and maintain the global state of the network. Its code is in the [blockchain](https://github.com/incognitochain/incognito-chain/tree/production/blockchain) package.

  * **Synker**. Synker makes sure the node is up to date among its peers and also broadcasts the node status to its peers. Its code is in the [blockchain](https://github.com/incognitochain/incognito-chain/tree/production/blockchain) package.

  * **Mempool**. Mempool (memory pool) is a collection of transactions and blocks that have been verified but are not yet confirmed. Its code is in the [mempool](https://github.com/incognitochain/incognito-chain/tree/production/mempool) package.

  * **Wallet**. Software that holds all your Incognito keys. Use it to send and receive your Incognito tokens. Its code is in the [wallet](https://github.com/incognitochain/incognito-chain/tree/production/wallet) package.

  * **Database**. Incognito uses LevelDB to store block data. Its code is in the [drawdbv2](https://github.com/incognitochain/incognito-chain/tree/production/dataaccessobject/rawdbv2) package and [statedb](https://github.com/incognitochain/incognito-chain/tree/production/dataaccessobject/statedb) package.

* **Core**

  * **Consensus**

    * **pBFT**. For consensus algorithm, Incognito implements pBFT (Practical Byzantine Fault Tolerance). Its code is in the [blsbft](https://github.com/incognitochain/incognito-chain/tree/production/consensus/blsbft) package.

    * **BLS**. For multi-signature agregation, Incognito implements BLS Multi-Signatures. Its code is in the [blsmultisig](https://github.com/incognitochain/incognito-chain/tree/production/consensus/signatureschemes/blsmultisig) package.

    * **RNG**. For random number generator, Incognito currently uses Bitcoin block hash. We'll explore other RNG solutions in the future. Its code is in the [btc](https://github.com/incognitochain/incognito-chain/tree/production/blockchain/btc) package.

  * **Privacy**

    * **RingCT**. For privacy, Incognito implements RingCT (Ring Confidential Transaction) with ring signatures, stealth addresses, and confidential transactions. Its code is in the [privacy](https://github.com/incognitochain/incognito-chain/tree/production/privacy) package.

    * **Confidential Asset**. RingCT hides the amount of the transaction, but it doesn't hide the type of asset being sent. Confidential Asset solves that. It's under development under the [new-privacy-dev](https://github.com/incognitochain/incognito-chain/tree/new-privacy-dev) branch and will be merged into the master branch in April 2021.

    * **Mobile ZKP**. Incognito implements Zero-Knowledge Proofs (ZKP) Generation on mobile. Private transactions can be sent on any regular phone under 15 seconds. Its code is in the [wasm](https://github.com/incognitochain/incognito-chain/tree/production/privacy/wasm) package and the [zeroknowledge](https://github.com/incognitochain/incognito-chain/tree/production/privacy/zeroknowledge) package.

  * **Bridges**

    * **Ethereum**. Incognito implements a trustless two-way bridge between Incognito and Ethereum to let anyone send and receive ETH & ERC20 privately. Its code is in the [bridge-eth](https://github.com/incognitochain/bridge-eth) repository.

    * **Bitcoin**.  Incognito is working on a trustless two-way bridge between Incognito and Bitcoin to let anyone send and receive BTC privately. Here is its proposal (TBD). Estimated ship date: April 2021.

  * **pDEX**

    Incognito implements a first privacy-protecting decentralized exchange to let anyone freely trade without disclosing their identity with high throughput and low latency. Its code is in the [blockchain](https://github.com/incognitochain/incognito-chain/tree/production/blockchain) package.

* **Developer Tools**

  * **RPC**. RPC lets developers interact with Incognito via your own programs. Its code is in the [rpcserver](https://github.com/incognitochain/incognito-chain/tree/production/rpcserver) package.

  * **WebSocket**. WebSocket is another way for developers to interact with Incognito via your own programs. Its code is in the [rpcserver](https://github.com/incognitochain/incognito-chain/tree/production/rpcserver) package.

  * **SDK**. Incognito is working on Developer SDKs to make it even easier to build on top of Incognito. Estimated ship date (TBD).

* **Apps**

  * **Mobile Apps**. It's easy to build your own mobile apps on top of Incognito, once the SDK is available. Here is an example: [Mobile Wallet](https://github.com/incognitochain/incognito-chain-wallet-client-app).

  * **Web Apps**. It's easy to build your web apps on top of Incognito, once the SDK is available. Here are some examples: [Web Wallet](https://github.com/incognitochain/incognito-chain-wallet-client-2) or a [Desktop Network Monitor](https://github.com/incognitochain/incognito-monitor).

  * **Hardware Devices**. It's easy to build your own hardware on top of Incognito, once the SDK is available. Here is an example: [Node Device](https://github.com/incognitochain/incognito-node).

## Contribution

Incognito is and will always be 100% open-source. Anyone can participate, everyone can join, no one can ever restrict or control Incognito. There are many ways to participate.

* Report a bug via [Github Issues](https://github.com/incognitochain/incognito-chain/issues)

* Suggest a new feature via [Github Issues](https://github.com/incognitochain/incognito-chain/issues)

* Want to build a feature or fix a bug? Please send a [Pull Request](https://github.com/incognitochain/incognito-chain/pulls) for the maintainers to review your code and merge into the main codebase.

* Write tests. *We'll provide instructions soon how to setup local test environment.*

## License

Incognito is released under the terms of the MIT license. See [COPYING](https://github.com/incognitochain/incognito-chain/blob/master/COPYING) for more information or see [LICENSE](https://opensource.org/licenses/MIT)
