# Trustless bridges between Incognito and other blockchains

## Introduction

Incognito lets you send, receive, and store your crypto - Bitcoin, Ethereum, Binance, and more - privately. Incognito itself is a privacy-protecting blockchain. It comes with a set of trustless bridges that allow for secure two-way transfers of cryptoassets whenever privacy is needed.

This spec will detail how trustless bridges are constructed.

## pTokens

Privacy-protecting tokens (pTokens) - like pBTC and pDAI - are tokens on the Incognito blockchain that are backed 1:1 by cryptoassets on other public blockchains. For example, pBTC is 1:1 backed by BTC, and pDAI is 1:1 backed by DAI. You can always redeem one pBTC for one BTC and vice versa.

pTokens have the best of both worlds. They have the value of the original tokens. And they are private. You can send, receive, and store them with total privacy.

## The very first trustless bridge

The very first trustless bridge is the [Incognito and Ethereum bridge](https://github.com/incognitochain/incognito-chain/blob/master/specs/ethereum_bridge.md). The very first pTokens are from Ethereum, including pETH, pDAI, and pOMG.

The second bridge, third bridge, and beyond, which we will cover in this specs, are built on top of the first trustless bridge.

## Trustless custodians

Our trustless bridges are built on top of a custodial solution.

The current custodian solution is entirely centralized. Not only depositors have to trust a custodian, but the custodian fees are also often expensive, and the deposit/withdraw process is cumbersome. 

It is also important to note that Incognito cannot use a centralized custodian solution because that will leak user information to third parties and immediately go against the goal of the project.

Trustless, fully-decentralized custodians, are essential for building these trustless bridges.

## Becoming a custodian

Anyone can become a custodian by simply supplying some collaterals. The collaterals could be either in PRV or pTokens such as pETH, pDAI, and pUSDC.

Custodians can add more collaterals at any time they want.

## Taking deposits & minting pTokens

Each custodian has their own set of custodian addresses, such as BTC address and BNB address. Users, who want to convert their public tokens into pTokens, deposit their public tokens into these custodian addresses. Once the custodian receives these deposits, the custodian will mint pTokens back at a 1:1 ratio and send them to the users.

[![image.png](https://i.postimg.cc/CxdCnHGH/image.png)](https://postimg.cc/XrSBT5Dp)

## Over-collateralization

The total deposits supplied by users across all custodian addresses of a custodian should never exceed the total collaterals provided by that custodian. Over-collateralization ensures that the custodian will not run away with the deposits.

We also introduce another variable &alpha; which is initially set as 150%. &alpha; is effectively the Deposit-to-Value ratio, which makes sure that the total deposits will never exceed the total collaterals even if there is a significant drop on collateral value.

&alpha; x &sum;Deposit<sub>i</sub> &le; &sum;Collateral<sub>i</sub>

## Redeeming pTokens & returning deposits

Redeeming pTokens is pretty straightforward. The user inits a redeem transaction, which burns the pTokens and instructs the custodian to send the cryptoasset back to the user within a time window (such as 12 hours). 

[![image.png](https://i.postimg.cc/jjvr3735/image.png)](https://postimg.cc/TyySpp6M)

## Auto-liquidation

What if the custodian doesn't send the cryptoasset back in time or doesn't send it back at all? Their collaterals will be automatically liquidated to pay the users back.

Auto-liquidation also kicks in if the collaterals value drop significantly below &alpha x &sum;Deposit<sub>i</sub>. The custodian must add more collaterals to avoid auto-liquidation.

## Fees

The initial fee structure is kept simple - with fixed deposit and withdraw fee at 0.01%. 

It could later allow users to set their fees so that we can have a market-driven pricing structure or a more complex fee structure as a combination of deposit, withdraw, and custodial time. 

## Conclusion

Once the trustless bridge implementation is completed in Q1 2020, any crypto user will be able to send any of their current crypto privately. We cannot wait to see some really cool use cases like privacy-protecting crypto trading or crypto payroll. We are looking forward to it and hope you are too.

In the meantime, please feel free to cross the first decentralized bridge, Ethereum, and play around with sending your ETH/ERC20 tokens privately. The Incognito wallet app is available for iOS and Android.

Do drop us any questions, comments, and feedback — we are always glad to hear them.