# Incognito Mainnet v0 Configuration

**Designers:** @0xbarmanu, @0xmerman, @0xgazeka, @thaibao-incognito, @hieu013, @0xkumi, @duyhtq

**Developers:** @0xmerman, @thaibao-incognito

This document details the key parameters for Incognito Mainnet v0. It is the beginning of a better, safer internet. We also provide the reasons we pick the initial values for these parameters.

Some of these parameters will likely be fine-tuned over time. Parameter fine-tuning is a crucial part of Incognito mechanism design. Parameter fine-tuning responsibility is initially handled by the core team, but will be and gradually transferred to the community once self-governance is implemented in early 2020.

## GENESIS_TIME

The initial value is 1572480000 or Thu 31 Oct 2019 12:00:00 AM UTC.

## NUMBER_OF_SHARDS

The initial value is 8.

We choose the initial value as 8, because the Incognito Testnet performs most stably with 8 shards from July 2019 - October 2019. Incognito will scale the number of shards over time to increase throughput of the network. The goal is to reach 64 shards in 2020 with our Highway implementation (to be deployed late 2019).

## NUMBER_OF_VALIDATORS  

The initial value is 32 validators per shard.

This is the maximum number of validators that can produce blocks for a shard at a time. The goal is to scale up to 256 validators in 2020 with our Highway implementation (to be deployed late 2019).

## NUMBER_OF_SUBSTITUTES

The initial value is 8 substitutes per shard.

This is the maximum number of substitutes that are waiting to substitute existing validators in the next epoch. We feel that this number should be around 15%-20% of NUMBER_OF_VALIDATORS.

## NUMBER OF SUBSTITUTIONS

The initial value is 4 substitutions per shard per epoch.

This is the maximum number of substitutions that a shard will make in an epoch. We feel that this number should be around 50% of NUMBER_OF_SUBSTITUTES, in case some substitutes are offline or haven't fully synced the shard block data yet.

## NUMBER OF CANDIDATES

There is no limit on the number of candidates.  Anyone can stake to become a candidate.

## BLOCK_SIZE

The initial value is 2MB.

## BLOCK_LATENCY

The initial value is 40s.

There are multiple discussions on block latency. While block latency sometimes could be really fast (between 2s-10s), it could be slow in many other cases due to the number of nodes and the number of transactions. At this stage, we opt to go with a safer number for latency and gradually optimize the block latency.

## EPOCH_SIZE

The initial value is 350 blocks or ~4 hours.

Epoch size determines when substitutions happen. If the epoch size is too big, it may affect the security of the network. If the epoch size is too small, too many substitutions happen and data availability could be a problem. We feel that 350 blocks is a reasonable window, considering the number of nodes at the beginning.

## STAKING_AMOUNT

The initial value is 1750 PRV.

Picking the staking amount is both art and science. Our heuristic is to go with the ultimate network configuration with 64 subchains (1 beacon and 64 shards) and 256 validators per subchain.  If each validator stakes 1750 PRV, the total staking amount of the entire network is:

65 * 256 * 1,750 PRV = 29,120,000 PRV

The total supply of the network is 100,000,000 PRV.  So the total staking amount is 29% of the total supply. If someone wants to take control of the network, they need to stake at least 20,000,000 PRV (exceeds the pBFT 2/3+1 signature threshold). We feel that 29% is a reasonably safe staking pool. We also don't want to lock up more than 29M PRV as that could begin to hurt the utility of PRV.

## FIRST_YEAR_BLOCK_REWARDS

The total block rewards for the first year are 8,751,974 PRV.

Block rewards are reduced by 9% for every subsequent year. 100% of PRV is fully mined after ~40 years.

## Limitations & future work

### Fixed committee setup

Note that for Mainnet v0, 22 out of 32 validators are operated by the core team. These 22 validators are fixed and are not substitutable; only the other 10 validator slots are substitutable. The reason is that in anticipation of potential issues in the early stage of the network the core team wants to make sure the issues can be resolved as soon as possible. The core team will gradually release their validator spots to the community once the network is stable.

We expect to remove this fixed committee setup within 3-6 months since Mainnet launch.

### Fixed block producer setup

Via a round-robin setup, each validator takes turns to become the block producer and proposes a new block to the committee. If at least (2*N/3) + 1 validators confirm that the block is valid, the block will be added to the subchain.

This currently causes a problem. Every 40 seconds, a new block producer is selected. For various reasons — such as network delay, CPU usage, or insufficient memory — the last block at height H may not reach other validators in the committee, including the new block producer, within those 10 seconds. The new block producer, now still at block height H-1, starts producing a new block at the same height H as the last block — creating a fork of the subchain.

For Mainnet v0, we're taking a short-term solution which clearly separates the responsibilities between block producers and block validators.

* There is only 1 block producer per subchain. Its only job is to produce blocks and send them to the other N-1 nodes in the committee.

* The other N-1 nodes in the committee are block validators. A validator’s only job is to verify blocks. It receives a new block from the block producer, verifies it, and collects signatures from other validators on the same block. If it collects at least (2*N/3) + 1 signatures from other validators, it will append the block to its local copy of the subchain.

This solution is not perfect by any means but is a good-enough temporary solution. We're working on a long-term solution.

We expect to remove this fixed block producer setup within 3 months since Mainnet launch.
