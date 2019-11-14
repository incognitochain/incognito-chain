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

Custodians can add more collaterals at any time they want.

## Porting public tokens into pTokens

Each custodian has their own set of custodian addresses, such as BTC address and BNB address, for receiving deposits. Users, who want to port their public tokens into pTokens, deposit their public tokens into these custodian addresses. Once the custodian receives these deposits, the custodian mints pTokens a 1:1 ratio and sends them to the users.

[![image.png](https://i.postimg.cc/CxdCnHGH/image.png)](https://postimg.cc/XrSBT5Dp)

## Over-collateralization

The total deposits made by users across all custodian addresses of a custodian should never exceed the total collaterals provided by that custodian. Over-collateralization ensures that the custodian does not run away with the deposits.

We also introduce another variable &alpha;, which is initially set as 150%. &alpha; is effectively the Deposit-to-Value ratio, which makes sure that the total deposits never exceeds the total collaterals even if there is a significant drop on collateral value.

&alpha; x &sum;Deposit<sub>i</sub> &le; &sum;Collateral<sub>i</sub>

## Redeeming pTokens for public tokens

Redeeming a pToken is pretty straightforward. The user inits a redeem transaction, which burns the pToken and instructs the custodian to send the public token back to the user by a deadline. The deadline is initially set *within 12 hours*.

[![image.png](https://i.postimg.cc/jjvr3735/image.png)](https://postimg.cc/TyySpp6M)

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