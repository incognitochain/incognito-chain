# Design Spec
link: https://drive.google.com/file/d/1gI_dwf2h8irzAGefHfuX_79ASWItdqch/view?usp=sharing

## Key
key is 32 bytes length with:
- 10 bytes prefix
- 20 bytes key

## Value
value is depend on type of state object

## State Object
- Assume that hash(something) return 32 bytes value
1. All Shard Committee (deprecate)
- key: first 12 bytes of `hash(shard-committee-prefix)` with first 20 bytes of `hash(beaconheight)`
- value: 
    * temporary value: 32 bytes random ID
    * real value: all shard committee
- key -> temporary value -> real value. Beacause shard committee not change so frequently

2. Committee (deprecate)
Used for beacon and all shards, distinguish between shards and beacon by prefix. Each shard or beacon has different prefix value
- key: first 12 bytes of `hash(committee-shardID-prefix)` with first 20 bytes of `hash(committee-key-bytes)`
- value: committee state( shardID and Committee Key)

3. Reward Receiver (deprecate)
Used for reward receiver in beacon only
- key: first 12 bytes of `hash(reward-receiver-prefix)` with first 20 bytes of `hash(incognito-key-bytes)`
- value: reward receiver state( incognito public key and reward receiver payment address)

4. Stop Auto Staking (deprecate)
Used for reward receiver in beacon only
- key: first 12 bytes of `hash(stop-auto-staking-prefix)` with first 20 bytes of `hash(incognito-key-bytes)`
- value: reward receiver state( incognito public key and reward receiver payment address)

5. Committee 
Used for beacon and all shards, distinguish between shards and beacon by prefix. Each shard or beacon has different prefix value
- key: first 12 bytes of `hash(committee-shardID-prefix)` with first 20 bytes of `hash(committee-key-bytes)`
- value: committee state:
    * shard id: beacon is -1, shard range from 0 to 256
    * role: candidate, substitute, validator
    * committee public key: base 58 of incognitokey.CommitteePublicKey struct
    * reward receiver payment address: base 58 string of reward receiver
    * auto staking: yes or no
    
6. Committee Reward
- key: first 12 bytes of `hash(committee-reward-prefix)` with first 20 bytes of `hash(incognito-public-key-bytes)`
- value: committee state:
    * reward: map token id => reward amount
    * incognito public key: 33 bytes public key encoded as base 58 string
7. Reward Request
- key: first 12 bytes of `hash(reward-request-prefix)` with first 20 bytes of `hash(epoch + shardID + tokenID)`
- value: reward request state:
    * epoch
    * shardID
    * tokenID
    * amount
8. Black List Producer:
- key: first 12 bytes of `hash(black-list-producer-prefix)` with first 20 bytes of `hash(committee-publickey-base58-string)`
- value: black list producer state:
    * committee public key base58 string 
    * punished epoch (punished duration left)
    * beacon height at which this state is calculated
9. Serial Number:
- key: first 12 bytes of `hash(serial-number-prefix + tokenID + shardID)` with first 20 bytes of `hash(serial-number-bytes)`
- value: serial number state:
    * tokenID
    * shardID
    * serial number value
10. Commitment:
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
11. Output Coin:
- key: first 12 bytes of `hash(output-coin-prefix + tokenID + shardID)` with first 20 bytes of `hash(incognito-public-key-bytes)`
- value: output coin state:
    * tokenID
    * shardID
    * publicKey
    * outputCoins: list of output coin
12. SNDerivator:
- key: first 12 bytes of `hash(SNDerivator-prefix + tokenID)` with first 20 bytes of `hash(snd-bytes)`
- value: snderivator state:
    * tokenID
    * snd
13. PDE:
- key: first 12 bytes of `hash(waiting-pde-contribution-prefix)` with 20 bytes of `hash(beacon-height-bytes)`
- value: waiting pde contribution
    * beacon height 
    * pairID
    * contributor address
    * tokenID
    * amount
    * transaction request id
    