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

# Ethereum ➔ Incognito
To convert ETH/ERC20 tokens on Ethereum to private ETH/ERC20 (pETH/pERC20) tokens on Incognito, users simply need to deposit their tokens to the Incognito Wallet. Under the hood, users are in effect sending their ETH or ERC20 tokens to a decentralized smart contract on Ethereum then sending the proof of that transaction to Incognito.

The Incognito chain verifies that proof and extracts information from it in order to mint private tokens to the predefined address.
The whole process is illustrated in the following figure.

![Mint private ETH/ERC20](https://i.postimg.cc/tCBcrnbS/0-2-Es-ULaqrtob-I-k-Mp.png)

### Deposit Ether/ERC20 to smart contract
Our decentralized smart contract on Ethereum performs two pretty simple functions in the aforementioned deposit process — one for Ether and one for ERC20.

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
