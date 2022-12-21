package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
)

var (
	shouldSubmitKey           bool
	shouldStakeShard          bool
	shouldStakeBeacon         bool
	shouldStopAutoStakeBeacon bool
	shardValidators           map[string]*Validator
	beaconValidators          map[string]*Validator
	keys                      []Key
)

func init() {
	shardValidators = map[string]*Validator{
		sKey0:  {},
		sKey1:  {},
		sKey2:  {},
		sKey3:  {},
		sKey4:  {},
		sKey5:  {},
		sKey6:  {},
		sKey7:  {},
		sKey8:  {},
		sKey9:  {},
		sKey10: {},
		sKey11: {},
	}

	beaconValidators = map[string]*Validator{
		bKey0: {},
		bKey1: {},
		bKey2: {},
		bKey3: {},
		bKey4: {},
		bKey5: {},
	}

	args := os.Args

	if len(args) > 1 {
		t := args[1:]
		for _, v := range t {
			if v == submitkeyArg {
				shouldSubmitKey = true
			} else if v == stakingShardArg {
				shouldStakeShard = true
			} else if v == stakingBeaconArg {
				shouldStakeBeacon = true
			} else if v == stopAutoStakeBeaconArg {
				shouldStopAutoStakeBeacon = true
			}
		}
	} else {
		shouldSubmitKey = true
		shouldStakeShard = true
		shouldStakeBeacon = true
	}

	data, err := ioutil.ReadFile("accounts.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &keys); err != nil {
		panic(err)
	}

	for _, k := range keys {
		if _, found := shardValidators[k.MiningKey]; found {
			shardValidators[k.MiningKey] = &Validator{Key: k, HasStakedShard: false, HasStakedBeacon: false}
		}
		if _, found := beaconValidators[k.MiningKey]; found {
			beaconValidators[k.MiningKey] = &Validator{Key: k, HasStakedShard: false, HasStakedBeacon: false}
		}
	}
}

func main() {

	fullnode := flag.String("h", "http://localhost:8334/", "Fullnode Endpoint")
	flag.Parse()

	app := devframework.NewAppService(*fullnode, true)
	lastCs := &jsonresult.CommiteeState{}

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
	convertTxHeight := bHeight + 5
	sendFundsHeight := bHeight + 15

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
				//Stake each nodes
				for _, v := range shardValidators {
					v.ShardStaking(app)
				}
				for _, v := range beaconValidators {
					v.ShardStaking(app)
				}
			}
		}
		if shouldStakeBeacon {
			if blk.GetBeaconHeight() >= startStakingBeaconHeight {
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
							v.StakeBeaconFromHeight = blk.GetBeaconHeight()
						}
					} else {
						if blk.GetBeaconHeight() >= v.StakeBeaconFromHeight+3 {
							if v.Role == ShardCommitteeRole {
								v.shouldInBeaconWaiting(cs)
							}
						}
					}
				}
			}
		}
		if shouldStopAutoStakeBeacon {
			v := beaconValidators[bKey0]
			resp, err := app.StopAutoStaking(v.PrivateKey, v.PaymentAddress, v.MiningKey)
			if err != nil {
				panic(err)
			}
			fmt.Println(resp)
		}
		cs, err := getCSByHeight(blk.GetBeaconHeight(), app)
		if err != nil {
			panic(err)
		}
		if cs.IsDiffFrom(lastCs) {
			lastCs = new(jsonresult.CommiteeState)
			*lastCs = *cs
			updateRole(shardValidators, beaconValidators, cs)
			cs.Print()
		}
	})

	select {}
}
