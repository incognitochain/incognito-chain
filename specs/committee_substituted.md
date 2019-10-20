# Committee Random Substitution

Committee random substitution is an important part of Incognito design to increase network security.

## Committees

The Incognito blockchain is composed of $M+1$ subchains: 1 beacon subchain and $M$ shard subchains.

<img src="https://i.postimg.cc/nrvqWYzz/0-Lrtbq-R3rm-Laured-N.png" width=600>

Each subchain is a Proof-of-Stake blockchain with its own committee of $N$ nodes.

* The committee produces new blocks via a Practical Byzantine Fault Tolerance (pBFT) consensus algorithm.

* Nodes are both block producers and validators.

* Via a round-robin setup, each node takes turns to become the block producer and proposes a new block to the committee. If at least $\frac{2}{3}N+1$ nodes validate the block and confirm that it is valid, the block will be added to the subchain.

* Every $T$ blocks, the committee will be shuffled. **$K$ random nodes in the current committee will be substituted by new nodes**.

## The Life Cycle

The life cycle of a validator is as follow:

1. The **user** stake 1750 PRV to become a candidate.

2. The **candidate** is randomly chosen to become a pending validator.

3. The **pending validator** prepares to join the committee, first by syncing all block data.

4. The **validator** is now in the committee and starts producing blocks.

5. Every $T$ blocks, the **validator** may be substituted out of the committee.

6. Once being substituted out, the user can choose to stake again (go back to step 1).

![Incognito Validator Life Cycle](https://i.postimg.cc/nVS1wjJV/image.png)

## The Parameters

These are the key parameters.  These parameters are dynamic.  They can be adjusted, initially by the core team.  Parameter adjustment responsibility will be and gradually transferred to the community once self-governance (propose & vote) is implemented in early 2020.

| Parameter | Description |
| --------- | ----------- |
| $M$       | the number of shards |
| $N$       | the number of validators per shards (committee size) |
| $T$       | the number of blocks in an epoch |
| $K$       | the number of validators being substituted every epoch |
| $P$       | the number of pending validators preparing to join the committee|
| $\alpha$  | alpha |

## What should be the value of $K$?

Because Incognito implements pBFT, it requires at least $\frac{2}{3}N+1$ honest nodes. Because of that, K should never exceed $\frac{1}{3}N$.

Initially, $K$ is set as $\frac{1}{10}N$.

## What should be the value of $P$?

Because $K$ validators are being substituted every epoch, $P$ can't be smaller than $K$.

In fact, we feel that there should be a buffer 
$P=\alpha*K$
