# Minimum Transaction Fees

**Designers:** @0xankylosaurus, @0xmerman @duyhtq

**Developers:** @0xankylosaurus, @0xmerman

This document details how transaction fees work in Incognito.

## Pay with PRV

When you make a transaction on Incognito, you set your own transaction fee - how much you're willing to pay for your transaction. The higher the transaction fee is, the earlier your transaction will be processed. 

The MIN_TX_FEE is 0.00005 PRV.

## Pay with pToken

Incognito lets you pay for transaction fees in the tokens you're sending. This is great UX because the user doesn't need to care about the network token.

Incognito implements this via the in-built Incognito Decentralized Exchange (Incognito DEX). Specifically, Incognito DEX is used as an oracle price feed of PRV/pToken to calculate how much pToken is needed.

Here is how it works:

1. Alice is sending 1000 pDAI to Bob.

2. Alice sets the transaction fee she's willing to pay as 3 pDAI.

3. The validator gets the exchange rate R of PRV/pDAI from Incognito DEX.

   * If R is not available, the transaction is rejected. *The reasoning for the rejection is that there is a slight chance that pDAI is worthless on the Incognito network and it's not fair to pass on the risk to the validators. The users can always choose to pay with PRV*.

   * Otherwise,

     * If 3 pDAI * R < MIN_TX_FEE, the transaction is rejected. *The reasoning for the rejection is that 3 pDAI does not meet the minimum transaction fee requirement if we swap 3 pDAI for PRV*.

     * Otherwise, the transaction validation process continues.
