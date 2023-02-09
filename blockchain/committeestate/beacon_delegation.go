package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/pkg/errors"
)

type RedelegateInfo struct {
	beaconPK   string
	delegators []string
}

func (s *StakerInfo) getScore() uint64 {
	return s.StakingAmount + uint64(s.TotalDelegators)*config.Param().StakingAmountShard
}

func (s *BeaconCommitteeStateV4) getDelegateState() (*statedb.BeaconDelegateState, error) {
	if s.delegateState == nil {
		res, _, err := statedb.GetBeaconReDelegateState(s.stateDB)
		if err != nil {
			return nil, err
		}
		s.delegateState = res
	}
	return s.delegateState, nil
}

// Process add stake amount
func (s *BeaconCommitteeStateV4) ProcessBeaconRedelegateInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	delegateState, err := s.getDelegateState()
	changed := false
	if err != nil {
		Logger.log.Error(err)
		delegateState = statedb.NewBeaconDelegateState()
	}
	for _, inst := range env.BeaconInstructions {
		if inst[0] == instruction.RE_DELEGATE {
			reDelegateInst := instruction.ImportReDelegateInstructionFromString(inst)
			for i, cpk := range reDelegateInst.CommitteePublicKeys {
				shardStakerInfo, exists, _ := statedb.GetStakerInfo(s.stateDB, cpk)
				if (!exists) || (shardStakerInfo == nil) {
					Logger.log.Errorf("Cannot find delegator %v in statedb", cpk)
					continue
				}
				oldDelegate := shardStakerInfo.GetDelegate()
				oldDelegateUID := shardStakerInfo.GetDelegateUID()
				newDelegate := reDelegateInst.DelegateList[i]
				newDelegateUID, err := common.Hash{}.NewHashFromStr(reDelegateInst.DelegateUIDList[i])
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				if reInfo, ok := delegateState.NextEpochDelegate[cpk]; ok {
					reInfo.New = newDelegate
					reInfo.NewUID = newDelegateUID.String()
				} else {
					delegateState.AddReDelegateInfo(cpk, statedb.ReDelegateInfo{
						Old:    oldDelegate,
						OldUID: oldDelegateUID,
						New:    newDelegate,
						NewUID: newDelegateUID.String(),
					})
				}
				//update delegation reward
				affectEpoch := env.Epoch + 1
				if lastBlockEpoch(env.BeaconHeight) {
					affectEpoch++
				}
				receiver := shardStakerInfo.RewardReceiver()
				err = statedb.StoreDelegationReward(s.stateDB, receiver.Bytes(), cpk, int(affectEpoch), newDelegateUID.String(), 1750*1e9)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
			}
		}
	}

	if changed {
		s.delegateState = delegateState
		if err := statedb.StoreBeaconReDelegateState(s.stateDB, delegateState); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func processUpdateDelegate(delegateM map[string]RedelegateInfo, s *BeaconCommitteeStateV4, updateFunc func(bcDelegator string) error) error {
	for uid, info := range delegateM {
		if info.beaconPK == "" {
			continue
		}
		bcInfor, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, info.beaconPK)
		if !exist {
			err := errors.Errorf("Cannot find cpk %v in statedb", info.beaconPK)
			Logger.log.Error(err)
		}
		curBeaconUID := common.HashH([]byte(fmt.Sprintf("%v-%v", info.beaconPK, bcInfor.BeaconConfirmHeight())))
		if uid == curBeaconUID.String() {
			if err := updateFunc(info.beaconPK); err != nil {
				return err
			}
		}
	}
	return nil
}

func addReDelegateInfo(delegateInfoM map[string]RedelegateInfo, delegateeUID string, delegator string, delegatee string) {
	if info, ok := delegateInfoM[delegateeUID]; ok {
		info.delegators = append(info.delegators, delegator)
	} else {
		delegateInfoM[delegateeUID] = RedelegateInfo{
			beaconPK:   delegatee,
			delegators: []string{delegator},
		}
	}
}

func (s *BeaconCommitteeStateV4) ProcessAcceptNextDelegate(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	if !lastBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}
	delegateState, err := s.getDelegateState()
	if err != nil {
		Logger.log.Error(err)
		delegateState = statedb.NewBeaconDelegateState()
	}
	oldDelegateM := map[string]RedelegateInfo{}
	newDelegateM := map[string]RedelegateInfo{}
	for k, v := range delegateState.NextEpochDelegate {
		oldDelegate := v.Old
		oldDelegateUID := v.OldUID
		if info := s.getStakerInfo(oldDelegate); info != nil {
			uid, err := s.GetBeaconCandidateUID(info.CPK)
			if err != nil {
				panic(1) //todo: remove this
				continue
			}
			if uid == oldDelegateUID {
				addReDelegateInfo(oldDelegateM, oldDelegateUID, k, oldDelegate)
			}
		}
		newDelegate := v.New
		newDelegateUID := v.NewUID
		if info := s.getStakerInfo(newDelegate); info != nil {
			uid, err := s.GetBeaconCandidateUID(info.CPK)
			if err != nil {
				panic(1) //todo: remove this
				continue
			}
			if uid == newDelegateUID {
				addReDelegateInfo(newDelegateM, newDelegateUID, k, newDelegate)
			}
		}
	}
	if err := processUpdateDelegate(oldDelegateM, s, s.removeDelegator); err != nil {
		return nil, err
	}
	if err := processUpdateDelegate(newDelegateM, s, s.addDelegator); err != nil {
		return nil, err
	}
	if len(delegateState.NextEpochDelegate) != 0 {
		s.delegateState = statedb.NewBeaconDelegateState()
		if err := statedb.StoreBeaconReDelegateState(s.stateDB, s.delegateState); err != nil {
			return nil, err
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
