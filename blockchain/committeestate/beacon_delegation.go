package committeestate

import (
	"fmt"
	"log"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/pkg/errors"
)

func (s *StakerInfo) getScore() uint64 {
	return s.StakingAmount + uint64(s.TotalDelegators)*config.Param().StakingAmountShard
}

//when validator return staking, it will end its delegation reward
//also update delegator number for the delegated beacon
func (s *BeaconCommitteeStateV4) ProcessDelegateRewardForReturnValidator(env ProcessContext) ([][]string, error) {
	if !firstBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}

	for _, delegator := range env.RemovedStaker {
		//calculate beacon delegators
		shardStakerInfo, exists, _ := statedb.GetShardStakerInfo(s.stateDB, delegator)
		if (!exists) || (shardStakerInfo == nil) {
			Logger.log.Errorf("Cannot find delegator %v in statedb", delegator)
			panic(1)
		}
		oldDelegate := shardStakerInfo.GetDelegate()
		oldDelegateUID := shardStakerInfo.GetDelegateUID()
		if oldDelegate != "" {
			if stakerInfo := s.getBeaconStakerInfo(oldDelegate); stakerInfo != nil {
				if uID, err := s.GetBeaconCandidateUID(oldDelegate); (err != nil) || (uID != oldDelegateUID) {
					Logger.log.Error(errors.Errorf("Can not get request UID for old delegate %v of delegator %v, uid found in db %v", oldDelegate, oldDelegateUID, uID))
				} else {
					if err = s.removeDelegators(oldDelegate, 1); err != nil {
						return nil, err
					}
				}
			} else {
				Logger.log.Error(errors.Errorf("Cannot find staker cpk %v in statedb", oldDelegate))
			}
		}

		//end its delegation reward
		affectEpoch := env.Epoch
		receiver := shardStakerInfo.RewardReceiver()
		err := statedb.StoreDelegationReward(s.stateDB, receiver.Pk, receiver, delegator, int(affectEpoch), "", config.Param().StakingAmountShard)
		if err != nil {
			Logger.log.Error(err)
			panic(err)
			continue
		}
	}

	return nil, nil
}

// Process add stake amount
func (s *BeaconCommitteeStateV4) ProcessBeaconRedelegateInstruction(env ProcessContext) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.RE_DELEGATE {
			reDelegateInst := instruction.ImportReDelegateInstructionFromString(inst)
			for i, delegator := range reDelegateInst.CommitteePublicKeys {
				shardStakerInfo, exists, _ := statedb.GetShardStakerInfo(s.stateDB, delegator)
				if (!exists) || (shardStakerInfo == nil) {
					Logger.log.Errorf("Cannot find delegator %v in statedb", delegator)
					continue
				}
				newDelegate := reDelegateInst.DelegateList[i]
				newDelegateUID, err := common.Hash{}.NewHashFromStr(reDelegateInst.DelegateUIDList[i])
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				if stakerInfo := s.getBeaconStakerInfo(newDelegate); stakerInfo != nil {
					if uID, err := s.GetBeaconCandidateUID(newDelegate); (err != nil) || (uID != newDelegateUID.String()) {
						Logger.log.Error(errors.Errorf("Can not get request UID for new delegate %v of delegator %v, uid found in db %v", newDelegate, newDelegateUID, uID))
						continue
					}
				} else {
					Logger.log.Error(errors.Errorf("Cannot find staker cpk %v in statedb", newDelegate))
					continue
				}

				oldDelegate := shardStakerInfo.GetDelegate()
				oldDelegateUID := shardStakerInfo.GetDelegateUID()
				log.Println("oldDelegate", oldDelegate)
				if oldDelegate != "" {
					if stakerInfo := s.getBeaconStakerInfo(oldDelegate); stakerInfo != nil {
						if uID, err := s.GetBeaconCandidateUID(oldDelegate); (err != nil) || (uID != oldDelegateUID) {
							Logger.log.Error(errors.Errorf("Can not get request UID for old delegate %v of delegator %v, uid found in db %v", oldDelegate, oldDelegateUID, uID))
						} else {
							log.Println("removeDelegators", oldDelegate)
							if err = s.removeDelegators(oldDelegate, 1); err != nil {
								return nil, err
							}
						}
					} else {
						Logger.log.Error(errors.Errorf("Cannot find staker cpk %v in statedb", oldDelegate))
					}
				}

				shardStakerInfo.SetDelegate(newDelegate, newDelegateUID.String(), env.BeaconHeight)

				//update delegation reward
				affectEpoch := env.Epoch + 1
				receiver := shardStakerInfo.RewardReceiver()
				err = statedb.StoreDelegationReward(s.stateDB, receiver.Pk, receiver, delegator, int(affectEpoch), newDelegateUID.String(), config.Param().StakingAmountShard)
				if err != nil {
					Logger.log.Error(err)
					panic(err)
					continue
				}
				if err = s.addDelegators(shardStakerInfo.GetDelegate(), 1); err != nil {
					return nil, err
				}
				if err = statedb.StoreStakerInfoV2(s.stateDB, reDelegateInst.CommitteePublicKeysStruct[i], shardStakerInfo); err != nil {
					return nil, err
				}
			}
		}
	}

	return nil, nil
}

func (s *BeaconCommitteeStateV4) ProcessBeaconSharePrice(env ProcessContext) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		//share price update instruction
		if inst[0] == instruction.SHARE_PRICE {
			sharePrice, err := instruction.ValidateAndImportSharePriceInstructionFromString(inst)
			if err != nil {
				return nil, err
			}
			newSharePrice := sharePrice.GetValue()
			for k, v := range newSharePrice {
				statedb.StoreBeaconSharePrice(s.stateDB, k, v)
			}

		}
	}
	return nil, nil
}

func (s *BeaconCommitteeStateV4) GetBeaconCandidateUID(cpk string) (string, error) {
	info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
	if !exist {
		return "", fmt.Errorf("Cannot find cpk %v in statedb", cpk)
	}
	hash := common.HashH([]byte(fmt.Sprintf("%v-%v", cpk, info.GetBeaconConfirmTime())))
	return hash.String(), nil
}

func (s *BeaconCommitteeStateV4) addDelegators(beaconStakerPK string, total uint64) error {
	log.Println("add delegator to", beaconStakerPK)
	if stakerInfo := s.getBeaconStakerInfo(beaconStakerPK); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, beaconStakerPK)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", beaconStakerPK)
		}
		info.AddDelegator(total)
		stakerInfo.TotalDelegators += uint64(total)
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", beaconStakerPK)
}

func (s *BeaconCommitteeStateV4) removeDelegators(oldBeaconStakerPK string, total uint64) error {
	if stakerInfo := s.getBeaconStakerInfo(oldBeaconStakerPK); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, oldBeaconStakerPK)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", oldBeaconStakerPK)
		}
		err := info.RemoveDelegator(total)
		if err != nil {
			return err
		}
		stakerInfo.TotalDelegators -= uint64(total)
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", oldBeaconStakerPK)
}

func (s *BeaconCommitteeStateV4) ProcessWithdrawDelegationReward(env ProcessContext) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		switch inst[0] {
		case instruction.MINT_DREWARD_ACTION:
			mintDRewardInstruction, err := instruction.ValidateAndImportMintDelegationRewardInstructionFromString(inst)
			if err != nil {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			for i, _ := range mintDRewardInstruction.PaymentAddressesStruct {
				currentEpoch := env.Epoch
				//if this is first block of epoch, we need to update the affect epoch is previous one.
				//We are not count the previous epoch in delegation reward (i.e share price is not updated yet)
				if firstBlockEpoch(env.BeaconHeight) {
					currentEpoch--
				}
				err := statedb.ResetDelegationReward(s.stateDB, mintDRewardInstruction.PaymentAddressesStruct[i].Pk, mintDRewardInstruction.PaymentAddressesStruct[i], int(currentEpoch))
				if err != nil {
					err := errors.Errorf("Can not reset delegation reward for payment %v, tx request %v, reward in inst %v", mintDRewardInstruction.PaymentAddresses[i], mintDRewardInstruction.TxRequestIDs[i], mintDRewardInstruction.RewardAmount[i])
					Logger.log.Error(err)
					return nil, err
				}
			}
		default:
			continue
		}
	}
	return [][]string{}, nil
}
