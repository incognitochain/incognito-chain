# Minimum Transaction Fees

**Designers:** @0xankylosaurus, @0xmerman @duyhtq

**Developers:** @0xankylosaurus, @0xmerman

This document details how transaction fees work in Incognito.

## Pay in PRV

When you make a transaction on Incognito, you set your own transaction fee - how much you're willing to pay for that transaction. The higher the transaction fee is, the earlier your transaction will be processed by validators.

Transaction fees are required to avoid flood attacks and compensate validators for maintaining the network.

The MIN_TX_FEE is 1 NANO (0.000000001 PRV).

## Pay in pToken

Incognito lets you pay for transaction fees in the same tokens you're sending. This is great UX because you don't have to carry the native token, PRV, to pay for the transaction fees.

Incognito implements this via the in-built Incognito Decentralized Exchange (Incognito DEX).  Incognito DEX is used as an oracle price feed of PRV/pToken to calculate how much pToken is needed for transaction fees. Note that Incognito DEX is solely used as an oracle price feed, there is no token swap between PRV and pToken on Incognito DEX.

Here is an example of how it works:

1. Alice is sending 1000 pDAI to Bob.

2. Alice sets the transaction fee she's willing to pay as 3 pDAI.

3. Validators get the exchange rate R of PRV/pDAI from Incognito DEX.

   * If R is not available, the transaction is rejected. *The reasoning is that there is a slight chance that pDAI is worthless on the Incognito network and it's not fair to pass the risk onto validators. The users can always choose to pay in PRV*.

   * If R is available,

     * If 3 pDAI * R < MIN_TX_FEE, the transaction is rejected. *The reasoning is that 3 pDAI does not meet the minimum transaction fee requirement*.

     * If 3 pDAI * R >= MIN_TX_FEE, the transaction passes the minimum transaction fee validation.

## Issues

### PRV/pToken exchange rate fluctuation

There is a time lag between when the user sends the transaction into the network and when the validators actually receives their pToken. While there is a risk that pToken value goes down during that window, because transaction fees collected during the time lag are relatively small, we feel that the risk is acceptable.

### pToken liquidity

Similar to the above issue, pToken could be a lot less liquid by the time validators receive their pToken earnings. Because transaction fees collected during the time lag are relatively small, we feel that the risk is acceptable.