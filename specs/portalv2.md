# Portal: Bring incognito mode to every token

## Introduction

Incognito is a privacy-protecting blockchain. It's interoperable with other blockchains, allowing for secure two-way transfers of crypto whenever privacy is needed. So you can privately send, receive, and store your crypto - like BTC, ETH, BNB, and more.

Current blockchain interoperability solutions are mostly building ad-hoc bridges ([Cosmos](https://cosmos.network/), [Polkadot](https://polkadot.network/)). As the number of blockchains is increasing, doing it ad-hoc is no longer an option. We just can't build N<sup>2</sup> bridges.

Incognito Portal takes a different approach: **build once, work with any blockchain**. Portal is a *general bridge design* that connects Incognito with as many blockchains as possible if not all.

Portal is especially helpful for creating interoperability with blockchains that don't support smart contracts like Bitcoin and Binance Chain. Its design is public.  The code is open-source. We hope that the crypto community will find our work helpful and create more portals to connect more blockchains together.

## pTokens

Privacy-protecting tokens (pTokens) - like pBTC and pDAI - are tokens on the Incognito blockchain that are backed 1:1 by cryptoassets on other blockchains. For example, pBTC is 1:1 backed by BTC, and pDAI is 1:1 backed by DAI. You can always redeem 1 pBTC for 1 BTC and vice versa.

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


## Becoming a custodian

Anyone can become a custodian just by supplying some collaterals. The collaterals could be either in PRV or any pTokens such as pBTC, pETH, and pDAI.

In addition to miners/validators, the custodians have a crucial role in the Incognito network as they make cross-chain communication between the two blockchains possible and enable the choice of “incognito mode” for transfers of crypto assets. In return, custodians will “earn” porting fees and PRVs reward just similar to what miners/validators are receiving (of course, in different formula/amount).


## Proof verification

In either Porting or Redeeming processes, there will be needs to verify a transaction has been confirmed on “public” blockchains (Bitcoin/Binance/Ethereum/etc). Fortunately, almost popular blockchains have supported Simplified Payment Verification (SPV for short) regardless of differences between those blockchains as Bitcoin’s using Merkel tree, Ethereum’s using Merkle Patricia tree, Binance’s using IAVL+ tree and so on.

In order to verify the validity of proof on Incognito, we need to:

(1) Make sure that each Incognito node maintains a valid “public” blockchains’ header chain.

(2) Verify the validity of the proof (Merkle branches) with the “public” blockchain header’s Merkle root. This is made possible via a Merkle tree from all transactions in a block.

Here are a few ways to achieve (1):

(a) Someone continuously relays public blockchain’s headers into Incognito and Incognito itself verifies the header validity.

(b) Each Incognito validator node keeps a popular public blockchain light client. That light client also continuously syncs, verifies the validity of block headers in a similar way to (1a)

In either (1a) or (1b), Incognito must obtain a valid header in order to support the verification in (2). The team opted for (1b) for now, a solution sufficient for moving forward based on the following assumption: as long as at least ⅓ of the number of validators in the Incognito committee run honest light clients, all is well. A transaction may only be comprised when a malicious light client is run by at least (⅔ + 1) of the number of validators. (Note: Incognito’s consensus is Proof of Stake)

The verification process is illustrated in the following figure:

[![image.png](https://i.postimg.cc/1Xh56L7t/proof-verf.png)](https://postimg.cc/Hczg1PgD)


## Porting public tokens into pTokens

To port public tokens (BTC/BNB/ETH/etc) on mainstream public blockchains out there to private tokens (pBTC/pBNB/pETH/etc) on Incognito, users simply submit a porting registration to Incognito along with the needed info (Unique registration id, Incognito address, private token id and amount). Incognito chain will use this info in order to mint ptokens a 1:1 ratio and send them to the users. Besides, the info will also be used to prevent front-running attacks that can occur during the process.

Any custodian, who wants to take advantage of the porting (fees and PRVs reward), picks an existing porting registration and deposits collaterals corresponding with the amount in the registered info. Each custodian has its own set of custodian addresses, such as BTC address and BNB address, for receiving deposits. The collateral deposit also encloses these addresses as its metadata.

Once the porting registration and collaterals deposit are matched together by unique registration id on Incognito, the process will start over: users need to send public tokens (BTC/BNB/ETH/etc) to the provided custodian’s address then extract and submit the proof of that transaction to Incognito. The Incognito chain verifies the proof validity with the aforementioned process and mints private tokens to the predefined users’ Incognito address.

The whole process of porting public tokens into pTokens is illustrated in the following figure:

[![image.png](https://i.postimg.cc/XJyYrV8m/porting.png)](https://postimg.cc/tZpjfG8N)

## Over-collateralization

The total deposits made by users across all custodian addresses of a custodian should never exceed the total collaterals provided by that custodian. Over-collateralization ensures that the custodian does not run away with the deposits.

We also introduce another variable &alpha;, which is initially set as 150%. &alpha; is effectively the Deposit-to-Value ratio, which makes sure that the total deposits never exceeds the total collaterals even if there is a significant drop on collateral value.

## Redeeming pTokens for public tokens

Redeeming a pToken is pretty straightforward. The user inits a redeem transaction, which burns the pToken and instructs the custodian to send the public token back to the user by a deadline. The deadline is initially set within 12 hours.

Once the custodian finishes sending public tokens (BTC/BNB/ETH/etc) to the user’s address that is included in redeem instruction, he can get redeem proof from those “public” blockchains and submit it to Incognito as Request collateral back transaction. The Incognito chain will verify redeem proof’s validity with the aforementioned process and then release collaterals to the custodian.

The redeem process is illustrated as the following figure:

[![image.png](https://i.postimg.cc/gjVcKN1K/redeeming.png)](https://postimg.cc/jnjrxQxW)

## Auto-liquidation

What if the custodian doesn't send the public token back by the deadline or doesn't send it back at all? If that happens, their collaterals are automatically liquidated to pay the users back.

Auto-liquidation also kicks in if the collaterals value drop significantly below &alpha x &sum;Deposit<sub>i</sub>. The custodian must add more collaterals to avoid auto-liquidation.

## Fees

The initial fee structure is simple - with a fixed deposit fee of 0.01% and a withdrawal fee of 0.01%.

It could later allow users to set their fees so that we can have a market-driven pricing structure or a more complex fee structure as a combination of deposit, withdraw, and custodial time.

## Mobile app

The Incognito Wallet plays a crucial role in making it as easy as possible for users to deposit and withdraw, as well as for custodian to add collaterals and generate custodian addresses.  We will detail the UI/UX in another post.

## Conclusion

Once Portal ships in Q1 2020, anyone will be able to send more crypto privately. We cannot wait to see some cool use cases. Do drop us any questions, comments, and feedback — we are always glad to hear them.
