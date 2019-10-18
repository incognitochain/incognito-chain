# Incognito mode for Ethereum

Enabling private ETH/ERC20 transactions via a fully decentralized bridge between Incognito and Ethereum

## Introduction

Incognito is a high-throughput proof-of-stake sidechain, made possible by the implementation of state sharding. Incognito takes a practical approach in designing and implementing its consensus mechanism, based on previous research and existing engineering by OmniLedger, Bitcoin, Ethereum 2.0, and Zilliqa.

The Incognito sidechain can be attached to any blockchain to conduct confidential asset transfer — in this case, Ethereum. The Incognito sidechain runs parallel to main blockchains, allowing for secure two-way transfers of crypto assets whenever privacy is needed.

This post will detail the means by which Incognito enables 100% confidential transfers of ETH and ERC20 tokens.

## Incognito mode for Ethereum

In this post, we present the Incognito-Ethereum bridge — designed for fully decentralized cross-chain interoperability. Implementations will enable cross-chain communication between the two blockchains, and enable the choice of “incognito mode” for transfers of crypto assets (ETH and ERC20 tokens). Using this bridge, anyone can turn on privacy for their tokens and shield their balances and activity.

![Ethereum + Incognito Trustless Birdge](https://i.postimg.cc/W4ygMzmh/0-5-Noou-G41-Mp2m-TUl.png)

In the following sections, we will define and explain the functionality of the bridge as well as the mechanism by which such requirements are implemented.

## Cross-chain asset transfer

A bridge can facilitate the transfer of crypto assets (e.g. tokens) between two blockchains by implementing locking, minting and burning mechanisms on each blockchain. When tokens are sent to a locking contract on Ethereum, Incognito needs to verify that the “locking” transaction was indeed confirmed on Ethereum, and upon the submission of the token lock transaction on Ethereum, proceed to mint a corresponding amount of privacy tokens (e.g., private Ether or private ERC20 tokens). When the private tokens are burned, the locking contract on Ethereum will verify the validity and unlock it upon submission of proof. This effectively maintains a 1:1 ratio between the private token on Incognito and the original token on Ethereum.

## Ethereum ➔ Incognito

To convert ETH/ERC20 tokens on Ethereum to private ETH/ERC20 (pETH/pERC20) tokens on Incognito, users simply need to deposit their tokens to the Incognito Wallet. Under the hood, users are in effect sending their ETH or ERC20 tokens to a decentralized smart contract on Ethereum then sending the proof of that transaction to Incognito.

The Incognito chain verifies that proof and extracts information from it in order to mint private tokens to the predefined address.
The whole process is illustrated in the following figure.

![Mint private ETH/ERC20](https://i.postimg.cc/tCBcrnbS/0-2-Es-ULaqrtob-I-k-Mp.png)

### Deposit Ether/ERC20 to smart contract

Our decentralized smart contract on Ethereum performs two pretty simple functions in the aforementioned deposit process — one for Ether and one for ERC20.

```js
function deposit(string memory incognitoAddress) public payable {
        require(msg.value + address(this).balance <= 10 ** 27);
        emit Deposit(ETH_TOKEN, incognitoAddress, msg.value);
}

function depositERC20(address token, uint amount, string memory incognitoAddress) public payable {
        IERC20 erc20Interface = IERC20(token);
        uint tokenBalance = erc20Interface.balanceOf(address(this));
        require(amount + tokenBalance <= 10 ** 18);
        require(erc20Interface.transferFrom(msg.sender, address(this), amount));
        emit Deposit(token, incognitoAddress, amount);
}
```

When those functions are called, two important pieces of information (incognito address and deposit amount) will be logged in anticipation of the related transactions being confirmed on the Ethereum network. In the following section, we explain how we use this information in minting private tokens.

### Verify deposit proof

After the deposit transaction has been confirmed on the Ethereum chain, the proof can be obtained and a special transaction containing that proof can be submitted to the Incognito chain for verification, in order to mint the corresponding private tokens. The metadata of this special transaction contains the following fields:

* Incognito private token id
* Ethereum block hash
* Ethereum transaction index
* Merkle path (or proof) for a receipt relating to the deposit transaction in the Ethereum block

In order to verify the validity of an Ethereum proof on Incognito, we need to:

* (1) Make sure that each Incognito node maintains a valid Ethereum header chain.
* (2) Verify the validity of the proof (receipt’s Merkle path) with Ethereum header’s receipts root field (Merkle root). This is made possible via a receipt Trie (a special Ethereum data structure) from all receipts of transactions in a block.

Since Ethereum’s consensus is Proof of Work, here are some ways to achieve (1):

- (a) Someone continuously relays Ethereum headers into Incognito and Incognito itself calculates the proof-of-work difficulty of the header and maintains a tree of all submitted headers. A block header is valid if it resides on the longest branch of the tree (a.k.a longest chain)
- (b) Each Incognito validator node keeps a popular Ethereum light node (e.g. Parity or Geth). That light node also continuously syncs, verifies the proof-of-work difficulty and resolves the fork situation of block headers in a similar way to (1a)

In either (1a) or (1b), Incognito must obtain a valid Ethereum header in order to support the verification in (3). The team opted for (1b), a solution sufficient for moving forward based on the following assumption: as long as at least ⅓ of the number of validators in the Incognito committee run honest Ethereum light nodes, all is well. A transaction may only be comprised when when a malicious Ethereum light node is run by at least (⅔ + 1) of the number of validators. (Note: Incognito’s consensus is Proof of Stake)

The verification process is illustrated in the following figure.

![Verify an Ethereum Deposit](https://i.postimg.cc/6pcpsPyK/0-Ut1-GXHQ36-DFJYs-Pj.png)

### Mint private tokens

Along with the verification of Merkle path, the aforementioned process also extracts the content of the receipt belonging to a deposit transaction. From that receipt, the required information (incognito address and deposit amount) can be parsed for minting private tokens.
In other words, both deposit amount and incognito address of the receiver are stored in the proof so that no one can obtain the proof, do a front-running attack or steal any private tokens.

## Incognito ➔ Ethereum

Withdrawing is the process of converting pETH (or pERC20) on Incognito to ETH (or ERC20) on Ethereum. On Ethereum, assets (ETH and ERC20) are held in a smart contract called the Incognito Vault. When someone wants to withdraw their tokens back to the main Ethereum blockchain, they need to supply the Incognito Vault with proof that sufficient tokens have been burned (i.e. destroyed) on Incognito. The private asset must be burned to maintain the 1:1 peg between the private asset on Incognito and the base asset.

The following diagram shows the whole process.

![Incognito to Ethereum](https://i.postimg.cc/cHwqWKX3/0-f-VP7-HNv8p-Blt-XQh-F.png)

### Burn private tokens

At any time, users can create a special transaction on Incognito called BurningRequest to initiate the withdrawal process. This tx sends an arbitrary amount of private tokens to the BurningAddress. Assets locked in this address cannot be moved, effectively destroying them. Along with the number of tokens to be withdrawn, users also provide a valid receiving address for the corresponding asset on Ethereum.

```go
type BurningRequest struct {
        BurnerAddress privacy.PaymentAddress
        BurningAmount uint64 
        TokenID       common.Hash
        TokenName     string
        RemoteAddress string
        MetadataBase
}
```

Here, TokenID is the private token id on Incognito chain, which will be used to derive the corresponding Ethereum asset (ETH or ERC20 contract address). RemoteAddress is the aforementioned address that will receive the Ethereum asset.

The tx is then mined and instructions generated and stored in the same block. This instruction will be included in a **Burn Proof**, a cryptographic proof (signed by Incognito’s validators) that someone destroyed some amount of private tokens. Since the proof is stored onchain, it is viewable for anyone wishing to assess validity. The amount and token receiver (on Ethereum) is stored in the proof so that again, no one can get your proof, do a front-running attack and steal your tokens.

The proof is only considered valid when more than ⅔ of the current number of validators have signed it. On the Ethereum side, the list of validators is stored and continually updated to mirror Incognito’s.

The full code of BurningRequest tx can be found here: https://github.com/incognitochain/incognito-chain/blob/master/metadata/burningrequest.go

### Withdraw tokens with burn proof

After a burn proof has been generated, it can be obtained from the Incognito chain and submitted to the Vault contract. The proof contains the following metadata (burn instruction in the diagram below):

* Type of token to withdraw (ETH or ERC20 address)
* Amount of token
* Token receiving address

Additionally, the contract also needs to verify that this proof is valid and signed by Incognito’s validators. We can split this into 3 steps:

* Verify that the burn instruction is in a merkle tree with some root X
* Merkle tree with root X is in a block with block hash Y
* Validators signed block Y

These steps are illustrated in the diagram below.

![Burn Proof](https://i.postimg.cc/1zDcdX2V/0-f-QAcg-LYYt-Kz-6-XIx.png)

The code below shows the first step: verifying that the instruction (with hash instHash) is in merkle tree (with root instRoot), given the merkle path (instPath).

```js
require(instructionInMerkleTree(
        instHash,
        instRoot,
        instPath,
        instPathIsLeft
));
```

For the 2nd step, we need to recompute the hash of the block. On Incognito, block hash is computed by hashing the root of the merkle tree (storing all instructions) with the hash of the block body (containing all transactions). With this data, the contract can easily verify that merkle root X is in a block with hash Y.

```js
bytes32 blk = keccak256(abi.encodePacked(keccak256(abi.encodePacked(blkData, instRoot)))); 
```

Finally, the aggregated signature helps to prove that block hash Y has been approved by a group of validators Z. Note that since the contract already stores validator public keys, there is no need to provide them when validating a burn proof. The code below shows how we check if a sufficient number of validators (at least ⅔ + 1) signed the block.

```js
if (sigIdx.length <= signers.length * 2 / 3) {
        return false;
}

require(verifySig(signers, blk, sigV, sigR, sigS));
```

The Incognito Vault contract is open-source here: https://github.com/incognitochain/bridge-eth/blob/master/bridge/contracts/vault.sol

### Withdraw example on Kovan

We deployed the Incognito Vault contract on Ethereum’s Kovan testnet and configured it with the committee from the Incognito testnet (https://test-explorer.incognito.org/). The contract can be seen here.

As an example, we deposited some tokens (with this tx) to the Vault and withdrew some of it by providing a burn proof (in this tx). To obtain this proof, we simply need to make an RPC call to an Incognito fullnode (http://test-node.incognito.org/). The result of the RPC call can be viewed here and is partly shown in the image below.

![Kovan view](https://i.postimg.cc/1z6PmYz4/0-9-NILdi-EJa-Es-KYlw2.png)

### Swapping committee

We have left an important question unanswered so far: Incognito’s validators change all the time, how can the Incognito Vault smart contract keep track of them?

This is fairly simple to do by utilizing the tools discussed above. Every time a new validator joins the list or an existing one is removed (swapped), they collectively create and sign a SwapConfirm instruction. This instruction stores the pubkey of all validators and is validated in the exact same way as a burn instruction. Each committee effectively “hands over” to the next, and the chain of instructions ensure the list of validators is correct.

The following diagram illustrates this process.

![Swap committee](https://i.postimg.cc/qBQrvKpc/0-H3zy-J-DT5-I-F7-YSX.png)

## Conclusion

Once the bridge is officially released along with the Incognito mainnet (scheduled for end October 2019), Ethereum users will be able to explore some really cool features we’re excited about — privacy and high-throughput among others. We’re looking forward to it and hope you are too.

In the meantime, please feel free to cross the bridge and play around with sending your testnet ETH/ERC20 tokens privately. The MVP wallet app is available for iOS and Android.

Do drop us any questions, comments and feedback — we’re always glad to hear them.