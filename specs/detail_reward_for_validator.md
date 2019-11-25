
# Privacy (PRV) Mining & Distribution
## Total Supply
A strict limit of 100M Privacy (PRV) will be minted. This number will never increase.

## 95M Fair-Mining
Proof-of-stake networks often heavily pre-allocate tokens for the team, investors, and ICO.  Incognito takes a different approach.  Incognito is a proof-of-stake blockchain that designs its block rewards similarly to proof-of-work blockchains like Bitcoin.

95M of the 100M PRV total supply is mined through block rewards.  The total block reward for the first year is 8,751,970 PRV.  Block rewards are reduced by 9% for every subsequent year.  PRV will be fully mined after 40 years.  
## 5M Pre-Mining
5M PRV is pre-mined and purchased by the founding team for the aforementioned $1M, and has been used to cover salaries, server costs, operational expenses, and marketing activities.
## Block Reward Split
95% of PRV total supply is minted through block rewards.  Block rewards are split between the validators and Incognito DAO, a decentralized autonomous organization designed to fund  protocol development and network growth.  Incognito DAO collects a gradually reducing percentage of the block rewards, from 10% to 3%.  

With this income, Incognito DAO will fuel the growth of the network, fund interesting projects, and give the project longevity.

Incognito DAO’s funds are initially managed by the founding team.  Management responsibilities will be gradually distributed to the community. 

Year | Validators | Incognito DAO
-- | -- | --
1 | 90% | 10%
2 | 91% | 9%
3 | 92% | 8%
4 | 93% | 7%
5 | 94% | 6%
6 | 95% | 5%
7 | 96% | 4%
8+ | 97% | 3%

## Detailed reward for validator
Let ![equation](https://latex.codecogs.com/gif.latex?T) be the total reward (PRV) and transaction fee (PRV, pETH, pBTC...) earned by all shards.
Let ![equation](https://latex.codecogs.com/gif.latex?R_i) be block_reward + transaction fee earned by shard ![equation](https://latex.codecogs.com/gif.latex?i);
and ![equation](https://latex.codecogs.com/gif.latex?s) be the number of shards

![equation](https://latex.codecogs.com/gif.latex?\text{Then,&space;}T&space;=&space;\sum_{i=1}^{s}&space;R_i)

Reward of beacon = 2* reward of a single shard
Let ![equation](https://latex.codecogs.com/gif.latex?x) be the reward percentage of IncognitoDAO at the current year

![equation](https://latex.codecogs.com/gif.latex?\text{Incognito&space;DAO&space;earn:&space;}&space;x*T)

![equation](https://latex.codecogs.com/gif.latex?\text{Beacon&space;chain&space;earn:&space;}(1-x)*T&space;*\frac{2}{s&plus;2})

![equation](https://latex.codecogs.com/gif.latex?\text{Shard&space;i&space;earn:&space;}(1-x)*R_i&space;*\frac{s}{s&plus;2})

The reward of beacon/shard is splitted to each validator equally at the beginning of the next epoch. Slashing will be applied in the future version.
Epoch length is 350 blocks ~ 4 hours. The shard's epoch is synced with the epoch of the beacon chain. Thus, a shard epoch maybe shorter or longer than 350 blocks.

Initially, Incognito chain start with 8 shards and 1 beacon, the reward for each block at any shard is 1.38666 PRV, this reward is reduced by 9% every 788,400 blocks (~ 1 year).