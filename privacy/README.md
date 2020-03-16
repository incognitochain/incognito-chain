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

    And, output must be coin_v2