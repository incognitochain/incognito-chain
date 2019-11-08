# pDEX: The world's first privacy-protecting decentralized exchange

**Designers:** @0xankylosaurus, @0xgazeka, @hieu013, @duyhtq

**Developers:** @0xankylosaurus

## Introduction

The first generation of exchanges is centralized exchanges like Binance and Coinbase. The second generation of exchanges is decentralized exchanges (DEX) like Bancor, Kyber, and Uniswap. pDEX is an upgraded version of DEX. Like DEX, it's trustless. And it provides additional features like privacy via zero-knowledge proofs, high throughput via sharding, and inter-blockchain trading via interoperable bridges.

|                   | DEX   | pDEX  |
| ---------         | ------| --    |
| Trustless         | YES   | YES   |
| Privacy           | NO    | YES   |
| High throughput   | NO    | YES   |
| Inter-blockchain  | NO    | YES   |

## Trustless 

pDEX is code ([beaconpdeproducer.go](https://github.com/incognitochain/incognito-chain/blob/dev/master/blockchain/beaconpdeproducer.go), [beaconpdeprocess.go](https://github.com/incognitochain/incognito-chain/blob/dev/master/blockchain/beaconpdeprocess.go)) deployed to thousands of Nodes that power the Incognito network. It runs entirely on-chain, completely decentralized.

It takes some time to get used to the Automated Market Making mechanism of pDEX. But once you understand it, you'll see that it has some major advantages over traditional exchanges. pDEX borrows heavily from Nick Johnson's [reddit post](https://www.reddit.com/r/ethereum/comments/54l32y/euler_the_simplest_exchange_and_currency/) in 2016, Vitalik Buterin's [reddit post](https://www.reddit.com/r/ethereum/comments/55m04x/lets_run_onchain_decentralized_exchanges_the_way/) in 2016, Hayden Adam's [Uniswap implementation](https://github.com/Uniswap/contracts-vyper/blob/master/contracts/uniswap_exchange.vy) in 2018. 

pDEX does not use an order book.  Instead, it implements a novel Automated Market Making algorithm that provides instant matching, no matter how large the order size is or how tiny the liquidity pool is.  

### How to trade on pDEX?

The main idea is to replace the traditional order book with a bonding curve mechanism known as constant product. On a typical exchange such as Coinbase or Binance, market makers supply liquidity at various price points. pDEX takes everyone's bids and asks and pool them into two giant buckets. Market makers no longer specify at which prices they are willing to buy or sell. Instead, pDEX automatically makes markets based on a [Automated Market Making algorithm](https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf). 

[![1-gh35s-RDa-Nmn-VY7-S5avb42-A.png](https://i.postimg.cc/sfKZK5sv/1-gh35s-RDa-Nmn-VY7-S5avb42-A.png)](https://postimg.cc/V5r6krS8)


Let's go through a simple example. There are 100,000 pDAI and 10 pBTC currently in the pDAI/pBTC pool. No matter how much trading activity occurs, the goal is to keep this product equal to 1,000,000.

```
pDAI pool   : 100,000
pBTC pool   : 10
Constant    : 10 * 100,000 = 1,000,000
```

A buyer sends 10,000 pDAI to buy pBTC. A fee (in this example, 0.1% or 10 pDAI) is taken out for the liquidity providers, and the remaining 9,990 is added to the pDAI pool. The constant product is divided by the new amount of pDAI in the liquidity pool to determine the new pBTC pool. The buyer receives the remaining pBTC.

```
Buyer sends : 10,000 pDAI
Fee         : 10,000 * 0.1% = 10 pDAI
pDAI pool   : 100,000 + 10,000 - 10 = 109,990
pBTC pool   : 1,000,000 / 109,990 = 9.091735
Buyer gets  : 10 - 9.091735 = 0.908254 pBTC
```

Note that:

* The fee is added to the liquidity pool. pDAI pool is now 109,990 + 10 = 110,000. Liquidity providers collect fees when they exit the pool.

* Because the fee is added, the Constant slightly increases after each trade. In this example, it is now 110,000 * 9.091735 = 1,000,090.85

* The price has shifted. If a buyer makes a trade in the same direction, they will get a slightly worse pBTC/pDAI rate. However, if a buyer makes a trade in the opposite direction, they will get a slightly better pDAI/pBTC rate.

### How to become a liquidity provider and earn trading fees?

Liquidity providers play an essential role in pDEX. They provide liquidity to various pools on pDEX and earn trading fees.

Let's walk through it. In our example above, there are 100,000 pDAI and 10 pBTC. The ratio represents the price of the trading pair on pDEX: 100,000 pDAI / 10 pBTC = 10,000. If the price of BTC on another exchange (say, Binance) is not 10,000, there would be an arbitrage opportunity between pDEX and Binance.

The ratio between pDAI and pBTC should be maintained. Therefore, liquidity providers must provide liquidity to both sides of the pair. For example, if Alice wants to supply 1 pBTC, she must also supply 10,000 pDAI.

pDEX records liquidity contribution by issuing shares representing ownership of the liquidity provider in the pool. In our example, Alice owns 9.09% of the pool. Not only this represents what she contributes, but it also includes her earning: 9.09% of the total trading fees the pool has been making.

```
Alice supplies  : 10,000 pDAI and 1 pBTC
pDAI pool       : 100,000 + 10,000 = 110,000
pBTC pool       : 10 + 1 = 11
Alice owns      : 1 pBTC / 11 pBTC = 9.09% of the pool
```

### Trading fees 

All trading fees go directly to the liquidity providers. The trading fees do not go to the Incognito network or the founding team. 

The trading fee is calculated as follow:

```
F = min(B + P * V, C)
```

The core team sets the initial values of these parameters, but in the future, this parameter adjustment responsibility will gradually be transferred to the community.

| Parameter | Description | 
| --------- | ----------- | 
| F       | the trading fee| 
| B       | the base fee| 0 |
| P       | the percentage on the value of the trade|
| V       | the value of the trade|
| C       | the capped amount on the trading fee |

## Inter-Blockchain Liquidity

pDEX leverages Incognito's bridges to the other blockchains - like Ethereum and Bitcoin - to enable inter-blockchain trading. So you can trade BTC and ERC20 tokens privately on pDEX.

<img src="https://i.postimg.cc/cHBQGfcK/image.png" width=600>

As of November 2019, the [Incognito <> Ethereum Bridge](https://github.com/incognitochain/bridge-eth) is the only trustless bridge on the Mainnet. Other bridges are trusted and will migrate to a trustless implementation in early 2020.

| Bridges | Current Implementation | 
| --------- | ----------- | 
| Ethereum <> Incognito       | Trustless |
| Bitcoin <> Incognito      | Trusted (Going trustless Q1 2020) |
| Binance <> Incognito       | Trusted (Going trustless Q1 2020) | 
| EOS <> Incognito       | Trusted (Going trustless Q2 2020) | 

## Privacy

The solution for privacy is relatively simple. Within the Incognito Wallet, there is a unique address reserved for pDEX activities.  Each user has a pDEX address. A user can transfer funds from any address to the pDEX address.  This transfer is private by default on the Incognito network. No one can see the connection between the sending address and the pDEX address.

Once the pDEX address is funded, the user can freely trade without disclosing their identity.

## High Throughput

pDEX leverages Incognito's sharded architecture to deliver high throughput. The workload is shared among shards.  Trades are processed in parallel.  Incognito's throughput scales linearly by adding more shards, so it can continue to grow to support large transaction volume.

<img src="https://camo.githubusercontent.com/906df0b49559cb1cef57655b94379f7acb9175ff/68747470733a2f2f692e706f7374696d672e63632f6e72767157597a7a2f302d4c727462712d5233726d2d4c61757265642d4e2e706e67" width=800>

One of benefits of Automated Market Making is **instant matching**. If you enter an order, it will match right away. The current block time of Incognito is 40s. Once the block is produced on a shard, it is sent to the beacon chain, which takes another 40s. A trade will complete within 80s. Incognito plans to reduce the block time to 10s, once transaction size is further optimized.

## Conclusion

Not just being trustless, pDEX combines several advanced technologies like zero-knowledge proofs for privacy, sharding for high throughput and interoperable bridges for inter-blockchain trading. We hope that it could be a helpful product for the crypto community.