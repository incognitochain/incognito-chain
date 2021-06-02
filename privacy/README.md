# Privacy README

This file holds notes for privacy core team.

coins: 
    - Get coin commitments
    - .Init()
    - .Bytes()
    - 

coin object:
    - coin.ParseCoinObjectToInputCoin

Proofs:
    - Get output coins
    - Get input coins

Fix packages for v2:
    RPC, incognitokey, blockchain, transaction, privacy.

.ConvertOutputCoinToInputCoin()

When integrating v2 into v1 there are 3 scenarios that can happen:

    - 1: inputs only coin_v1

    - 2: inputs only coin_v2

    - 3: inputs have coin_v1 with coin_v2

   - output must be coin_v2
    
Things need to be done when coding new version:

    - 1: Code new version in transaction package (txprivacy.go, txversion_interface.go, txver1.go, txver2.go, ...).
    
    - 2: Code new version for proof (if change)
    
    - 3: Code new version for rangeproof (if change)
    
    - If have 2, 3 then: Check batchtransaction.go file and change
    