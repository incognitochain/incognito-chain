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

// Process add stake amount
func (s *BeaconCommitteeStateV4) ProcessBeaconRedelegateInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.RE_DELEGATE {
			reDelegateInst := instruction.ImportReDelegateInstructionFromString(inst)
			for i, delegator := range reDelegateInst.CommitteePublicKeys {
				shardStakerInfo, exists, _ := statedb.GetStakerInfo(s.stateDB, delegator)
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
				if stakerInfo := s.getStakerInfo(newDelegate); stakerInfo != nil {
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
					if stakerInfo := s.getStakerInfo(oldDelegate); stakerInfo != nil {
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

				shardStakerInfo.SetDelegate(newDelegate)
				shardStakerInfo.SetDelegateUID(newDelegateUID.String())

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

func (s *BeaconCommitteeStateV4) ProcessBeaconSharePrice(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
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

func (s *BeaconCommitteeStateV4) addDelegators(beaconStakerPK string, total uint) error {
	log.Println("add delegator to", beaconStakerPK)
	if stakerInfo := s.getStakerInfo(beaconStakerPK); stakerInfo != nil {
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

func (s *BeaconCommitteeStateV4) removeDelegators(oldBeaconStakerPK string, total uint) error {
	if stakerInfo := s.getStakerInfo(oldBeaconStakerPK); stakerInfo != nil {
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
