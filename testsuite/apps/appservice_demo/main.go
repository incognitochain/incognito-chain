package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

var (
	shouldSubmitKey        bool
	shouldStakeShard       bool
	shouldStakeBeacon      bool
	shouldStop             bool
	shouldUnstakeBeacon    bool
	shouldAddStakingBeacon bool
	shouldWatchValidator   bool
	shouldWatchOnly        bool
	watchBeaconIndex       int
	shardValidators        map[string]*Validator
	beaconValidators       map[string]*Validator
	keys                   []Key
)

func main() {
	readData()
	fullnode := flag.String("h", "http://localhost:8334/", "Fullnode Endpoint")
	flag.Parse()

	app := devframework.NewAppService(*fullnode, true)
	var lastCs *jsonresult.CommiteeState

	bState, err := app.GetBeaconBestState()
	if err != nil {
		panic(err)
	}
	bHeight := bState.BeaconHeight + 5
	if bHeight < 15 {
		bHeight = 15
	}
	epochBlockTime := uint64(10)
	submitkeyHeight := bHeight
	convertTxHeight := bHeight + 10
	sendFundsHeight := bHeight + 20

	log.Println("Will be listening to beacon height:", bHeight)
	var startStakingHeight uint64
	if shouldSubmitKey {
		startStakingHeight = bHeight + 40
	} else {
		startStakingHeight = bHeight
	}
	startStakingBeaconHeight := startStakingHeight + epochBlockTime + 5
	log.Println("Will be starting shard staking on beacon height:", startStakingHeight)
	log.Println("Will be starting beacon staking on beacon height:", startStakingBeaconHeight)

	app.OnBeaconBlock(bHeight, func(blk types.BeaconBlock) {
		if shouldSubmitKey {
			submitkeys(
				blk.GetBeaconHeight(),
				submitkeyHeight, convertTxHeight, sendFundsHeight,
				shardValidators, beaconValidators, app,
			)
		}
		if shouldStakeShard {
			if blk.GetBeaconHeight() == startStakingHeight {
				if shouldStop {
					v := beaconValidators[bIndexes[watchBeaconIndex]]
					if err := v.ShardStaking(app); err != nil {
						panic(err)
					}
					panic("Finish staking shard")
				}
				//Stake each nodes
				for _, v := range shardValidators {
					if err := v.ShardStaking(app); err != nil {
						panic(err)
					}
				}
				for _, v := range beaconValidators {
					if err := v.ShardStaking(app); err != nil {
						panic(err)
					}
				}
			}
		}
		if shouldStakeBeacon {
			if blk.GetBeaconHeight() >= startStakingBeaconHeight {
				if shouldStop {
					v := beaconValidators[bIndexes[watchBeaconIndex]]
					v.BeaconStaking(app)
					panic("Finish staking beacon")
				}
				cs, err := getCSByHeight(blk.GetBeaconHeight(), app)
				if err != nil {
					panic(err)
				}
				//Stake beacon nodes
				for _, v := range beaconValidators {
					if !v.HasStakedBeacon {
						var shouldStake bool
						for _, committee := range cs.Committee {
							for _, c := range committee {
								miningPublicKey := shortKey(v.MiningPublicKey)
								if c == miningPublicKey {
									shouldStake = true
								}
							}
						}
						if shouldStake {
							v.BeaconStaking(app)
						}
					} else {
					}
				}
			}
		}
		if shouldUnstakeBeacon {
			v := beaconValidators[bIndexes[watchBeaconIndex]]
			if v.Role == BeaconCommitteeRole {
				if shouldStop {
					//v := beaconValidators[bIndexes[watchBeaconIndex]]
					//v.ShardStaking(app)
					panic("Finish staking shard")
				}
				if _, found := v.ActionsIndex[unstakingBeaconArg]; !found {
					app.Unstaking(v.PrivateKey, v.PaymentAddress, v.MiningKey)
					v.ActionsIndex[unstakingBeaconArg] = blk.GetBeaconHeight()
				}
			}
		}
		if shouldAddStakingBeacon {
			v := beaconValidators[bIndexes[watchBeaconIndex]]
			if v.Role == BeaconPendingRole || v.Role == BeaconWaitingRole {
				if shouldStop {
					panic("Finish add staking beacon")
				}
				if _, found := v.ActionsIndex[addStakingBeaconArg]; !found {
					resp, err := app.AddStaking(v.PrivateKey, v.MiningKey, v.PaymentAddress, 175000000000000)
					if err != nil {
						panic(err)
					}
					v.ActionsIndex[addStakingBeaconArg] = blk.GetBeaconHeight()
					fmt.Println(resp)
				}
			}
		}
		if shouldWatchValidator {
			v := beaconValidators[bIndexes[watchBeaconIndex]]
			if err := v.watch(blk.GetBeaconHeight(), epochBlockTime); err != nil {
				panic(err)
			}
		}
		cs, err := getCSByHeight(blk.GetBeaconHeight(), app)
		if err != nil {
			panic(err)
		}
		if cs.IsDiffFrom(lastCs) {
			if lastCs == nil {
				if err = updateRole(shardValidators, beaconValidators, cs, true); err != nil {
					panic(err)
				}
			} else {
				if err = updateRole(shardValidators, beaconValidators, cs, false); err != nil {
					panic(err)
				}
			}
			lastCs = new(jsonresult.CommiteeState)
			*lastCs = *cs
			cs.Print()
		}
	})

	select {}
}
