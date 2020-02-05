# Portal: Bring incognito mode to every token

**Designers:** @duyhtq, @0xankylosaurus, @0xkraken

**Developers:** @0xankylosaurus, @0xkraken

Also thanks to @corollari for really useful feedback.

## Introduction

Portal focuses on three main properties:

**Privacy**: Incognito is a privacy-protecting blockchain. It's interoperable with other blockchains, allowing for secure two-way transfers of crypto whenever privacy is needed. So you can privately send, receive, and store your crypto - like BTC, ETH, BNB, and more.

**Scale-out**: The total amount of private tokens (pTokens for short) available for circulation increases with the total amount of collaterals which are locked up in two supporting custodial vaults (Incognito and Ethereum smart contract). In other words, Portal can leverage both security and user set (along with a huge market cap) of each platform.

**Compatibility**: Portal does not rely on a single cryptocurrency implementation with a set of specific features so it is especially helpful for creating interoperability with blockchains that don't support smart contracts like Bitcoin and Binance Chain.

Its design is public.  The code is open-source. We hope that the crypto community will find our work helpful and create more portals to connect more blockchains together.

## pTokens

Privacy-protecting tokens (pTokens) - like pBTC and pDAI - are tokens on the Incognito blockchain that are backed 1:1 by crypto-assets on other blockchains. For example, pBTC is 1:1 backed by BTC, and pDAI is 1:1 backed by DAI. You can always redeem 1 pBTC for 1 BTC and vice versa.

pTokens have the best of both worlds: the value of the original tokens and privacy. You can send, receive, and store them with total privacy.

Through Portal, you can port your public tokens - like BTC and DAI - into its mirrored private tokens - pBTC and pDAI.

## Trustless custodians

Portal is secured by a group of custodians.

Current custodian solutions are mostly centralized. Not only depositors have to trust a third party ([Bitgo](https://www.bitgo.com/), [Coinbase Custody](https://custody.coinbase.com/)), but the custodian fees are also often expensive, and the deposit/withdrawal process is cumbersome. Additionally, Incognito cannot use a centralized custodian because that leaks the user's private information to third parties.

Portal's custodians are entirely trustless. Here is a comparison between Portal and trusted custodian solutions.

|                     |     Incognito Portal       |       Trusted Custodian  |
|---------------------| -------------------------- | ------------------------ |
|  Single point of failure  |     No, fully decentralized       |        No              |
|  User privacy |           Yes              |          No              |
|  Trustless          |           Yes, just code              |          No              |
|  Safety         |           Backed by over-collateralization |          Backed by nothing     |
|  Processing time         |           Minutes or hours     |          Days              |
|  Deposit fees               |           Low              |          High            |
|  Withdrawal fees               |           Low              |          High            |
|  Setup fees               |           Zero              |          Expensive            |


## Custodial vaults & custodian

Like mention earlier, Portal is a practical and secure system to construct cryptocurrency-backed assets without trusted intermediaries but it still needs to store the deposited cryptoassets somewhere in a decentralized manner. We're introducing two options: Incognito itself and Ethereum smart contract as custodial vaults.

Anyone can become a custodian just by supplying collaterals. The collaterals can be either in PRV (for Incognito vault) or Ether/ERC20 (for Ethereum smart contract vault).

In addition to miners/validators, the custodians have a crucial role in the Incognito network as they make cross-chain communication between the two blockchains possible and enable the choice of “incognito mode” for transfers of crypto assets. In return, custodians will “earn” porting fees and PRVs reward just similar to what miners/validators are receiving. The PRVs reward for custodians is a proportion of Community fund that has been estimated about 10% of total PRV supply.


## Proof verification

In either Porting or Redeeming processes, there will be needs to verify a transaction has been confirmed on “public” blockchains (Bitcoin/Binance/Ethereum/etc). Fortunately, most of the popular blockchains have supported Simplified Payment Verification (SPV for short) regardless of slight differences between those blockchains as Bitcoin’s using Merkel tree, Ethereum’s using Merkle Patricia tree, Binance’s using IAVL+ tree and so on.

In order to verify the validity of proof on Incognito, we need to:

(1) Make sure that each Incognito node maintains a validated “public” blockchains’ header chain. In our approach, anyone can continuously relay public blockchain’s headers into Incognito and Incognito itself verifies the header validity before extending the header chain.

(2) Verify the validity of the proof (Merkle branches) with the “public” blockchain header’s Merkle root. This is made possible via a Merkle tree from all transactions in a block.

The verification process is illustrated in the following figure:

[![image.png](https://i.postimg.cc/1Xh56L7t/proof-verf.png)](https://postimg.cc/Hczg1PgD)



## Incognito as a custodial vault

### Porting public tokens into pTokens

Any custodian, who wants to take advantage of the porting (fees and PRVs reward), supplies collaterals (eg., PRVs) to a pool on Incognito. Each custodian has its own set of custodian addresses, such as BTC address and BNB address, for receiving deposits. The collateral deposit also encloses these addresses as its metadata.

To port public tokens (BTC/BNB/ETH/etc) on mainstream public blockchains out there to private tokens (pBTC/pBNB/pETH/etc) on Incognito, users simply submit a porting registration to Incognito along with the needed info (Unique registration id, Incognito address, private token id and amount). Incognito chain will use this info in order to mint pTokens a 1:1 ratio and send them to the users. Besides, the info will also be used to prevent front-running attacks that can occur during the process.

Once the porting registration and custodians are matched together on Incognito, the process will start over: a user needs to send public tokens (BTC/BNB/ETH/etc) to the provided custodians’ addresses then extracts and submits proof of that transaction to Incognito. The Incognito chain verifies the proof validity with the aforementioned process and mints private tokens to the predefined user’s Incognito address.

The whole process of porting public tokens into pTokens with Incognito vault is illustrated in the following figure:

[![image.png](https://i.postimg.cc/SsJKZjC3/prv-porting.png)](https://postimg.cc/vgwMTYPt)


### Redeeming pTokens for public tokens

Redeeming a pToken is pretty straightforward. A user inits a redeem transaction, which burns the pToken.

The system chooses custodians in the custodian list and asks them to send public tokens back to the user (Incognito will also provide a client tool in order for custodians to make the process execute automatically). Once the custodian finishes sending public tokens (BTC/BNB/ETH/etc) to the user’s address that is included in a redeem instruction, he can get a BTC/BNB/etc proof from those “public” blockchains and submit it to Incognito. The Incognito chain verifies that proof’s validity with the aforementioned process and then unlocks collaterals that will be available for either withdrawal or serving incoming porting requests.

In case the user doesn't still receive enough his original assets, the collaterals of custodians, who didn't return public tokens to the user, will be liquidated.

The redeem process with Incognito vault is illustrated as the following figure:

[![image.png](https://i.postimg.cc/pdB00LRf/prv-redeeming.png)](https://postimg.cc/bGsHr8Gr)



## Ethereum Smart Contract as a custodial vault

### Porting public tokens into pTokens

Quite similar to Incognito vault's above, custodians supply collaterals (eg., Ether/ERC20) to a smart contract. Each custodian has its own set of custodian addresses, such as BTC address and BNB address, for receiving deposits. The collateral deposit also encloses these addresses as its metadata.

A user, who wants to port public tokens to private tokens on Incognito, needs to send a porting registration to a provided smart contract along with the needed info (Unique registration ID, Incognito address, private token ID and amount). The Bond smart contract selects trustless custodians for the public coins and provides the user the custodians’ deposit addresses.

Once the deposit is confirmed on the cryptonetwork of the public coins, the contract verifies and produces an <em>accepted proof</em> if the submitted <em>deposit proof</em> by the user is valid and the user then can initiates a transaction on Incognito along with the <em>accepted proof</em> to request minting pTokens.

Incognito validators verify the transaction and the <em>accepted proof</em> inside it in particular by using the aforementioned process and new privacy coins are minted at a 1:1 ratio the predefined user’s Incognito address.

The whole process of porting public tokens into pTokens with Ethereum Smart Contract vault is illustrated in the following figure:

[![image.png](https://i.postimg.cc/hPW73jLS/eth-porting2.png)](https://postimg.cc/Btp6jJ7y)

### Redeeming pTokens for public tokens

Redeeming a pToken with Smart Contract vault is slightly different from Incognito vault's. A user inits a redeem transaction on Incognito, which burns the pToken. Then he can extract redeem proof and submit it to a provided smart contract.

The system chooses custodians in the custodian list and asks them to return public tokens to the user (Incognito will also provide a client tool in order for custodians to make the process execute automatically). Once the custodians finish sending public tokens (BTC/BNB/ETH/etc) to the user’s address that is included in redeem instruction, he can extract BTC/BNB/etc proof from those “public” blockchains and submit it to the smart contract. The smart contract verifies that proof’s validity with the aforementioned process and then unlocks collaterals that will be available for either withdrawal or serving incoming porting requests.

In case the user doesn't still receive enough his original assets, the collaterals of custodians, who didn't return public tokens to the user, will be liquidated.

The redeem process with Ethereum Smart Contract vault is illustrated as the following figure:

[![image.png](https://i.postimg.cc/C5vRN1h6/eth-redeeming.png)](https://postimg.cc/vxV88G6W)



## Over-collateralization

The total deposits made by users across all custodian addresses of a custodian should never exceed the total collaterals provided by that custodian. Over-collateralization ensures that the custodian does not run away with the deposits.

We also introduce another variable &alpha;, which is initially set as 150%. &alpha; is effectively the Deposit-to-Value ratio, which makes sure that the total deposits never exceeds the total collaterals even if there is a significant drop on collateral value.

## Auto-liquidation

What if the custodian doesn't send the public tokens back? If that happens, their collaterals are automatically liquidated to pay the users back.

Auto-liquidation also kicks in if the collaterals value drop significantly below &alpha x &sum;Deposit<sub>i</sub>. The custodian must add more collaterals to avoid auto-liquidation.

## Fees

The initial fee structure is simple - with a fixed deposit fee of 0.01% and a withdrawal fee of 0.01%.

It could later allow users to set their fees so that we can have a market-driven pricing structure or a more complex fee structure as a combination of deposit, withdrawal, and custodial time.

## Mobile app

The Incognito Wallet plays a crucial role in making it as easy as possible for users to deposit and withdraw, as well as for custodian to add collaterals and generate custodian addresses.  We will detail the UI/UX in another post.

## Conclusion

Once Portal ships in Q1 2020, anyone will be able to send more crypto privately. We cannot wait to see some cool use cases. Do drop us any questions, comments, and feedback — we are always glad to hear them.


