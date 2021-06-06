# pDEX: The first privacy-protecting decentralized exchange

## Introduction

The first generation of exchanges includes centralized exchanges like Binance and Coinbase. The second generation of exchanges comprises of decentralized exchanges (DEX) like Bancor, Kyber, and Uniswap. pDEX is an upgraded DEX. Like decentralized exchanges, it is trustless. It also implements additional features such as privacy via zero-knowledge proofs, high throughput via sharding, low latency via automated market making, and inter-blockchain trading via interoperable bridges.

|                   | DEX   | pDEX  |
| ---------         | ------| --    |
| Trustless         | YES   | YES   |
| Privacy           | NO    | YES   |
| High throughput   | NO    | YES   |
| Inter-blockchain  | NO    | YES   |

## Liquidity providers

Liquidity providers play an essential role in pDEX. They provide liquidity to various pools on pDEX and earn trading fees.

The main idea is to replace the traditional order book with a bonding curve mechanism known as constant product. On a typical exchange such as Coinbase or Binance, market makers supply liquidity at various price points. pDEX takes everyone's bids and asks and pool them into two giant buckets. Market makers no longer specify at which prices they are willing to buy or sell. Instead, pDEX automatically makes markets based on a [Automated Market Making algorithm](https://github.com/runtimeverification/verified-smart-contracts/blob/uniswap/uniswap/x-y-k.pdf).

<img src="https://i.ibb.co/Wk3vyq2/Screen-Shot-2021-06-03-at-4-16-15-PM.png" alt="drawing" style="width:500px;height:500px"/>

### Conentrate liquidity

LP’s (Liquidity Providers) can concentrate their capital within custom price ranges, providing greater amounts of liquidity at desired prices. In doing so, LPs construct individualized price curves that reflect their own preferences.

<img src="https://i.ibb.co/hmv7jPs/Screen-Shot-2021-06-03-at-3-04-52-PM.png" alt="drawing" style="width:500px;height:500px"/>

For example, an LP in the ETH/DAI pool may choose to allocate $100 to the price ranges $1,000-$2,000 and an additional $50 to the ranges $1,500-$1,750.

### Capital Efficiency

By concentrating their liquidity, LPs can provide the same liquidity depth as v2 within specified price ranges while putting far less capital at risk.

For example:

- Alice and Bob both want to provide liquidity in an ETH/DAI pool on Uniswap v3. They each have $1m. The current price of ETH is 1,500 DAI.

- Alice decides to deploy her capital across the entire price range (as she would have in Uniswap v2). She deposits 500,000 DAI and 333.33 ETH (worth a total of $1m).

- Bob instead creates a concentrated position, depositing only within the price range from 1,000 to 2,250. He deposits 91,751 DAI and 61.17 ETH, worth a total of about $183,500. He keeps the other $816,500 himself, investing it however he prefers.

- While Alice has put down 5.44x as much capital as Bob, they earn the same amount of fees, as long as the ETH/DAI price stays within the 1,000 to 2,250 range.

<img src="https://i.ibb.co/Y8dRkbh/Screen-Shot-2021-06-04-at-4-45-32-PM.png" alt="drawing">

### Active Liquidity

If market prices move outside an LP’s specified price range, their liquidity is effectively removed from the pool and is no longer earning fees

### Range Orders

LPs can deposit a single token in a custom price range above or below the current price: if the market price enters into their specified range, they sell one asset for another along a smooth curve while earning swap fees in the process.

## How to trade on pDEX?

`Below is trading process if only new price P won't cross to new initialized tick`

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

`In the case trading cross to new initialized tick, we need to calculate the liquidity and trading fee by the new tick`

## Trading fees 

1. Trading fees will be calcualted by trading volume in a period of time (But user can set it greater than the calculated value)

1. Trading fees will be calculate by formula: `Y = y - f(x + X * (1 - fee - Z))` which `x` and `y` is the current amount of 2 tokens will be traded, `fee` is base fee,  `Z` is the factor depend on average volume of AMM during a time period

1. Block producers sort orders by Trading FeeTrade Amount ratio, from highest to lowest, and process orders one by one. (Fee at least similar to trading fees which has been calculated above)

1. The trading fees go directly to liquidity providers provide in current tick.

Unlike other exchanges which offer fixed trading fee structures, pDEX gives users complete control on how they want to set their trading fees.  It’s entirely market-driven.  The higher you set your trading fees, the more likely your orders will be processed first at the prices you want.  The lower you set your trading fees, you’ll pay less but your orders will be more likely to fail due to price movements.

Unlike other exchanges which take all of the trading fees, on pDEX, the entirety of trading fees go directly to the liquidity providers.

We think that it’s a fairer way to build a crypto exchange. 

## Workflow

### Deposit

Deposit process in a nutshell

<img src="https://i.ibb.co/PM2qbRf/pdex-v3.png" alt="drawing"/>

### Trade

Trade process in a nutshell
<img src="https://i.ibb.co/2gZPWzr/pdex-v3-trade-2.png" alt="drawing"/>

### Withdraw

Withdraw process in a nutshell

<img src="https://i.ibb.co/BC9HKSn/pdex-v3-withdraw.png" alt="drawing"/>

## Architect

### Overview

<img src="https://i.ibb.co/KFLPTJg/pdex-v3-architect.png" alt="drawing"/>

### Detail

<img src="https://i.ibb.co/PzKzMty/pdex-v3-class-diagram.png" alt="drawing"/>

## References

1. Kyber Dynamic AMM (https://files.kyber.network/DMM-Feb21.pdf)
1. Uniswap v3 core (https://uniswap.org/whitepaper-v3.pdf)