# Design Spec
https://drive.google.com/file/d/1gI_dwf2h8irzAGefHfuX_79ASWItdqch/view?usp=sharing

## Key
key is 32 bytes length with:
- 10 bytes prefix
- 20 bytes key

## Value
value is depend on type of state object

## State Object
1. Committee 
Used for beacon and all shards, distinguish between shards and beacon by prefix. Each shard or beacon has different prefix value
- key: first 12 bytes of `hash(committee-shardID-prefix)` with first 20 bytes of `hash(committee-key-bytes)`
- value: committee state:
    * shard id: beacon is -1, shard range from 0 to 256
    * role: candidate, substitute, validator
    * committee public key: base 58 of incognitokey.CommitteePublicKey struct
    * reward receiver payment address: base 58 string of reward receiver
    * auto staking: yes or no
    
2. Committee Reward
- key: first 12 bytes of `hash(committee-reward-prefix)` with first 20 bytes of `hash(incognito-public-key-bytes)`
- value: committee state:
    * reward: map token id => reward amount
    * incognito public key: 33 bytes public key encoded as base 58 string
3. Reward Request
- key: first 12 bytes of `hash(reward-request-prefix + epoch)` with first 20 bytes of `hash(shardID + tokenID)`
- value: reward request state:
    * epoch
    * shardID
    * tokenID
    * amount
4. Black List Producer:
- key: first 12 bytes of `hash(black-list-producer-prefix)` with first 20 bytes of `hash(committee-publickey-base58-string)`
- value: black list producer state:
    * committee public key base58 string 
    * punished epoch (punished duration left)
    * beacon height at which this state is calculated
5. Serial Number:
- key: first 12 bytes of `hash(serial-number-prefix + tokenID + shardID)` with first 20 bytes of `hash(serial-number-bytes)`
- value: serial number state:
    * tokenID
    * shardID
    * serial number value
6. Commitment:
3 type of object: commitment, commitment index, commitment length
a. Commitment: store commitment value
- key: first 12 bytes of `hash(commitment-prefix + tokenID + shardID)` with first 20 bytes of `hash(commitment-bytes)`
- value: commitment state:
    * tokenID
    * shardID
    * serial number value   
b. Commitment index: store key to commitment value
- key: first 12 bytes of `hash(commitment-index-prefix + tokenID + shardID)` with first 20 bytes of `hash(commitment-index-bytes)`
- value: key of commitment (a)
c. Commitment Length: Number of commitment in one of shard of a token
- key: first 12 bytes of `hash(commitment-length-prefix)` with first 20 bytes of `hash(tokenID + shardID)`
- value: current number of commitment in this shard of a token
7. Output Coin:
- key: first 12 bytes of `hash(output-coin-prefix + tokenID + shardID)` with first 20 bytes of `hash(incognito-public-key-bytes)`
- value: output coin state:
    * tokenID
    * shardID
    * publicKey
    * outputCoins: list of output coin
8. SNDerivator:
- key: first 12 bytes of `hash(SNDerivator-prefix + tokenID)` with first 20 bytes of `hash(snd-bytes)`
- value: snderivator state:
    * tokenID
    * snd
9. PDE:
a. Waiting Contribution
- key: first 12 bytes of `hash(waiting-pde-contribution-prefix + beacon-height-bytes)` with 20 bytes of `hash(pairID)`
- value: waiting pde contribution
    * beacon height 
    * pairID
    * contributor address
    * tokenID
    * amount
    * transaction request id
b. Pool Pair
- key: first 12 bytes of `hash(pool-pari-prefix + beacon-height-bytes)` with 20 bytes of `hash(token1ID + token2ID)`
- value: pool pair state
    * beaconHeight
    * token1ID
    * token1PoolValue
    * token2ID
    * token2PoolValue
c. Share
- key: first 12 bytes of `hash(share-prefix + beacon-height-bytes)` with 20 bytes of `hash(token1ID + token2ID + contributor-address)`
- value: share state
    * beaconHeight
    * token1ID
    * token2ID
    * contributor address
    * amount
d. Share
- key: first 12 bytes of `hash(pde-status-prefix)` with 20 bytes of `hash(statusType + statusSuffix)`
- value: status state
    * statusType
    * statusSuffix
    * statusContent
10. Bridge
a. Eth tx
- key: first 12 bytes of `hash(bridge-ethtx-prefix)` with 20 bytes of `hash(unique-eth-tx-id)`
- value: bridge eth state
    * unique eth tx
b. Status 
- key: first 12 bytes of `hash(bridge-status-prefix)` with 20 bytes of `hash(tx-request-id)`
- value: bridge status state
    * tx request id
    * status
c. Token Info 
- key: first 12 bytes of `hash(bridge-token-info-prefix + centralized-or-decentralized-prefix)` with 20 bytes of `hash(incoginto-token-id)`
- value: bridge token info state
    * incTokenID
    * externalTokenID
    * amount
    * network
    * isCentralized
11. Burning Confirm
- key: first 12 bytes of `hash(burning-confirm-prefix)` with 20 bytes of `hash(txid)`
- value: burning confirm state
    * txID
    * height