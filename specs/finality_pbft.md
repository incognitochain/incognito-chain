## System Assumption
* 	There are 2/3+1 correctness members in the committee
*	Member can be online/offline, disconnect, or timeout
## Goal
*	Safety: block is proposed by a correctness validator and all validator agree on it
*	Liveness: all correctness validators decide finality block
*	Non-blocking: validators keep producing & adding new blocks into chain if more than 2/3 correct  are working
## PBFT in general
![Picture1](https://user-images.githubusercontent.com/37530661/67836086-3d724380-fb1e-11e9-9b89-6576cdff1060.png)
PROPOSE PHASE
A proposer proposes a block, and broadcast to validator

VOTE PHASE
Each validator validate and broadcast its vote for the block

COMMIT PHASE
If a validator or proposer collects enough 2/3+1 votes, it commits on the block

### Finality property

A block is finality means that it is voted by more than 2/3 validators and irreversible.

## State transition of a committee member

 
<img width="755" alt="Screen Shot 2019-10-30 at 14 08 06" src="https://user-images.githubusercontent.com/37530661/67836275-c25d5d00-fb1e-11e9-91e7-0ec6a230b59f.png">

A committee member maybe in one of the states {New, Wait, Abort, Commit}

Concurrency set: Let si denote the state of the committee member i. The set of all the states that may be concurrent with it is concurrency set (C(si)).

We assume the time is synced by all committee members. 
*	This means that when a member is in Waiting state the other maybe in {Waiting, Commit} state, it cannot be in Abort state due to time out. 
*	When a member in is Abort state the other member maybe in {Abort, Commit} state.

**Lemma 1.** *If the concurrency set of a validator has both abort and commit state. Then some validators may commit and the other may abort.*

*Proof*: On a proposed block, some validators get enough votes will go to commit. The other get time out will go to abort. 

**Claim 1**. *In case the proposer is fixed and correctness, the chain can get finality on block n iff the proposer commit on block n.*

*Proof*: when a proposer commit on block n, i.e. it collected > 2/3 votes, but validator may commit or abort on block n (see Lemma 1).  However, when proposer proposes block n+1, validators which have aborted block n could download and finalize block n. In case proposer is byzantine, it proposes an invalid block, it cannot collect enough 2/3 votes in order to commit the block.

In practice, the fixed proposer may cause the chain blocked when proposer crash or it’s byzantine.  To overcome this issue, rotation proposer role in committee is a popular solution.

**Claim 2**. *In case proposer is rotated, a committee member can commit two different blocks at the same height, i.e. the chain is forked.*
Proof: 1. let proposer A propose block A_n at height n, more than 2/3 committee members vote for this block but only few nodes collected enough votes. Then, in the next time slot, proposer B proposes block B_n at height n, since the majority of committee members have not commit on block A_n, they again vote for B_n. Some of committee members which already committed on block A_n could again collect more than 2/3 votes and commit on block B_n.

*Discussion*
To avoid fork chain:
*	Option 1. A validator can only vote 1 block at a specific height. This could lead to deadlock case. E.g. firstly 1/2 validators voted for block A_n at height n, this block is not committed due to lack of votes. Lately, another ½ validators, which don’t know about block A_n, vote for block B_n at height n too. None of proposed blocks get committed, the chain get blocked.
*	Option 2. A validator can vote for more than 1 block at a specific height. This again lead to fork chain. See proof of Lemma 2.

**Claim 3**. *A node with two forked chains, if one chain is one block longer than the other, then it cannot ensure the longer chain is the finalized chain.*

*Proof*:
 
<img width="361" alt="Screen Shot 2019-10-30 at 14 28 36" src="https://user-images.githubusercontent.com/37530661/67837455-a27b6880-fb21-11e9-902f-8031d27bdead.png">

Fig 1. One chain is one-block longer than the other

<img width="234" alt="Screen Shot 2019-10-30 at 14 28 57" src="https://user-images.githubusercontent.com/37530661/67837525-d35b9d80-fb21-11e9-8b88-6d2830520b08.png">

Fig 2. Two chains are equal in length

Assume a node is in the state as Fig 1. While, the majority group of nodes may in state as in Fig 2.
At this state, nodes may vote for a proposed block B’(n+2) which appended to B’(n+1). This means some node will get the forked chain with two blocks each chain. 

*Discussion*: to avoid block added to multiple fork chains, one can set the rule that if a node voted for a block at one forked chain, then not to vote for any other block at any other chain. However, the chain may get deadlock as option 1 in the proof of Claim 2.

**Claim 4**. *In the worst case, two forked chains may grow to infinity.*

Let N be the size of committee.

Let consider the case that network is highly congested, a major group, >=(2/3)N, failed to propose new block but can receive new block, vote for it, but not collect enough votes to commit the block. While a smaller group, <(1/3)N, can commit and propose new blocks. Moreover, this  group is divided into two forked chains. The next proposer is alternatively picked from each forked chain. In this case, the two forked chains can produce block and append to it chain continuously.

**Claim 5.**  *In rotated proposer mode, a node with two forked chains, there is one chain is one-block longer than the other as in Fig 1, then the chain can not be blocked if the note is not vote for the shorter chain.*

*Proof*: we will proof that the chain will not be blocked, even nodes do not vote for some old blocks.
*	If only a minor group of nodes is in this state, then the major group (as in Fig2) could produce block on the shorter chain without blocking.
*	If > 1/3 nodes are in this state (Fig1), both group in state as in Fig 1. and Fig 2. cannot make progress alone. But when one propose the new block with new height B(n+3), nodes in the state of shorter chain (Fig 2) could update the longer chain and vote for it.

**Claim 6.** *A node with two forked chains, if one chain is two-block longer than the other, then the shorter chain will be obsoleted.*

*Proof*: 
<img width="553" alt="Screen Shot 2019-10-30 at 14 35 50" src="https://user-images.githubusercontent.com/37530661/67837860-98a63500-fb22-11e9-8487-45b614a4cfc7.png">
Pic 3. One chain is two-block longer than the other

If a node commits on block B(n+3), this means majority have committed on block B(n+2). Then majority won’t vote for any block B’(n+2).  As a result, the shorter chain will be obsoleted.

##  Rules for vote and propose a block

As above analysis, the chain can be blocked or be forked. For forked case, the consensus could process the blockchain base on the rule of longer chain.

General rules: 
*	Avoid to propose multiple block at the same height
*	Follow the longest chain
<img width="978" alt="Screen Shot 2019-10-30 at 14 54 02" src="https://user-images.githubusercontent.com/37530661/67839288-e2444f00-fb25-11e9-9140-45be8551fe61.png">
<img width="1137" alt="Screen Shot 2019-10-30 at 14 54 58" src="https://user-images.githubusercontent.com/37530661/67839312-ee301100-fb25-11e9-87ce-684a803199d8.png">
<img width="1112" alt="Screen Shot 2019-10-30 at 14 55 22" src="https://user-images.githubusercontent.com/37530661/67839323-f720e280-fb25-11e9-8bc5-dcee2b45f10e.png">

### Analysis

Let N be the size of committee.

The worst case scenario is as in **Claim 4**.

The proposers must in minor group (<= 1/3 N), only this minor group can propose and commit on new block. 
This group is divided into two smaller groups, called A & , each one have (1/6)N members, which holds chain A and B, respectively. 

The probability that a proposer be in group A or B is 1/6.
The probability that the chain is forked into two chains with length 1-block per chain is 1/6.
The probability that the chain is forked into two chains with length 2-block per chain is satisfied:
* 	The first block forked, probability 1/6 
* 	The next block proposer is in small group, probability 1/6
* 	The next block proposer is on the other chain, probability 1/2
Thus, this case happens with probability (1/6)*(1/2)*(1/6) = (1/6) * (1/12)

Generally, the probability that the chain is forked into two chains with length n-block per chain is:  (1/6)*(1/12)^n.