# Validator Random Substitution Mechanism Design

Designers: @0xmerman, @duyhtq, @thaibao-incognito

Developers: @0xmerman

## Design goals

Validator random substitution is an important component to keep the network secure. The design should satisfy these system goals.

* **Randomness**. Randomness is a crucial building block to avoid the validators being compromised.

* **Liveliness**. The substitutions should not violate the consensus engine.

* **Data availability**. New validators should have complete block data to start working with existing validators right away.

* **Bandwidth**.  The substitutions should not significantly increase the message throughput peer connections.

Additionally, for the validator's own benefits, this design should also consider these user goals.

* **Idle time**. The shorter a user has to wait to become a validator, the better the validator experience is. It's also more efficient to make use of their capital.

* **Staking amount**. The smaller the staking amount is, as long as the system is still secure, the better it is for the majority of the users of the network.

## Code change notes

The implementation will mostly affect only the beacon chain. The code for the beacon chain is in the [blockchain](https://github.com/incognitochain/incognito-chain/tree/master/blockchain) package.

## The parameters

Parameter fine-tuning is an important part of Incognito mechanism design. Parameter fine-tuning responsibility is initially handled by the core team, but will be and gradually transferred to the community once self-governance is implemented in early 2020.

| Parameter | Description |
| --------- | ----------- |
| X       | the required staking amount |
| N       | the number of shards |
| M       | the maximum number of validators per shards |
| K       | the maximum number of substitutions per shard in an epoch |
| P       | the maximum number of substitutes per shard |
| T       | the number of blocks in an epoch |
| A       | the ratio between substitutes and substitutions |

## The validator role

The Incognito blockchain is composed of N+1 subchains: 1 beacon subchain and N shard subchains.

<img src="https://i.postimg.cc/nrvqWYzz/0-Lrtbq-R3rm-Laured-N.png" width=600>

Each subchain is a Proof-of-Stake blockchain with its own committee of M validators.

* The validators produce new blocks via a Practical Byzantine Fault Tolerance (pBFT) consensus algorithm.

* Via a round-robin setup, each validator takes turns to become the block producer and proposes a new block to the committee. If at least (2/3)*M+1 validators confirm the validity of the block, the block will be added to the subchain.

* Every epoch (or T blocks), the committee will be shuffled. K random validators in the current committee will be substituted by new validators.

## The life cycle of a validator

The life cycle of a validator is as follow:

1. The **user** stakes X amount of PRV to become a candidate.

2. The **candidate** is randomly selected to become a substitute for a specific shard S.

   * If the number of current substitutes is less than P, the candidate automatically becomes a new substitute for shard S.

   * Otherwise, the candidate will wait until the next epoch for the next random selection.

3. The **substitute** must sync all block data of shard S in advance.

   * If the number of current validators in shard S is less than M, the substitute automatically becomes a new validator for shard S at the next epoch.

   * Otherwise, the substitute will replace an existing validator in shard S at the next epoch.

4. The **new validator** starts producing blocks for shard S.

5. The **ex-validator** becomes a normal user and can manually stake again (go back to step 1). Note that there is an "auto re-stake" option for validators who don't want to manually re-stake.

![Incognito Validator Life Cycle](https://i.postimg.cc/KcCZ1cVf/image.png)

## What should be the value of K?

Because Incognito implements pBFT, it requires at least (2/3)*M+1 honest validators in a shard at all time. If we make too many substitutions at a time, it may significantly reduce the security of the shard.

The initial value of K is set as (1/10)*M.

## What should be the value of P?

Because there are K substitutions per shard every epoch, P can't be smaller than K.

Additionally, we feel that there should be a buffer so that in case a substitute is offline or some substitutes need more time to complete block data syncing.

P = A*K

The initial value of A is set as 2.

## Future work

### Weighted random selection

Weighted random selection could further improve the design. We will explore various weighting mechanisms such as reputation-based (the number of blocks a validator has produced)or time-based (how long has a candidate been waiting to become a validator).

### Random substitution

While the current solution select substitutes randomly from the candidate pool, the existing validators are being replaced as first-in first-out. Random substitution could further improve the design by selecting existing validators randomly.

A side note, this design could hurt validator experience as a user had been waiting for 10 epochs, just started working for 1 epoch, and now is being substituted. One idea is to set a threshold H so that only validators who work more than H epochs are part of the random substitution.
