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

// Process add stake amount
func (s *BeaconCommitteeStateV4) ProcessBeaconRedelegateInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {

	delegateState, _, err := statedb.GetBeaconReDelegateState(s.stateDB)
	if err != nil {
		return nil, err
	}

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
				receiver := shardStakerInfo.RewardReceiver()
				err = statedb.StoreDelegationReward(s.stateDB, receiver.Pk, receiver, cpk, int(affectEpoch), newDelegateUID.String(), config.Param().StakingAmountShard)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
			}
		}
	}

	if err := statedb.StoreBeaconReDelegateState(s.stateDB, delegateState); err != nil {
		return nil, err
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
	info, ok := delegateInfoM[delegateeUID]
	if ok {
		if info.beaconPK == delegatee {
			info.delegators = append(info.delegators, delegator)
		}
	} else {
		info = RedelegateInfo{
			beaconPK:   delegatee,
			delegators: []string{delegator},
		}
	}
	delegateInfoM[delegateeUID] = info
}

func (s *BeaconCommitteeStateV4) ProcessAcceptNextDelegate(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	if !lastBlockEpoch(env.BeaconHeight) {
		return nil, nil
	}
	delegateState, _, err := statedb.GetBeaconReDelegateState(s.stateDB)
	if err != nil {
		return nil, err
	}
	oldDelegateM := map[string]RedelegateInfo{}
	newDelegateM := map[string]RedelegateInfo{}
	for k, v := range delegateState.NextEpochDelegate {
		oldDelegate := v.Old
		oldDelegateUID := v.OldUID
		addReDelegateInfo(oldDelegateM, oldDelegateUID, k, oldDelegate)
		newDelegate := v.New
		newDelegateUID := v.NewUID
		addReDelegateInfo(newDelegateM, newDelegateUID, k, newDelegate)
	}
	for k, v := range oldDelegateM {
		Logger.log.Infof("BEFOREREMOVE: Beacon %v - %v has current %v delegator, need to remove %v delegator", v.beaconPK, k, len(v.delegators))
	}
	if err := processUpdateDelegate(oldDelegateM, s, s.removeDelegator); err != nil {
		return nil, err
	}
	for k, v := range oldDelegateM {
		if detailInfo := s.getStakerInfo(v.beaconPK); detailInfo != nil {
			Logger.log.Infof("AFTERREMOVE: Beacon %v - %v has current %v delegator", v.beaconPK, k, detailInfo.TotalDelegators)
		}
	}
	for k, v := range newDelegateM {
		Logger.log.Infof("BEFOREADD: Beacon %v - %v has current %v delegator, need to add %v delegator", v.beaconPK, k, len(v.delegators))
	}
	if err := processUpdateDelegate(newDelegateM, s, s.addDelegator); err != nil {
		return nil, err
	}
	for k, v := range newDelegateM {
		if detailInfo := s.getStakerInfo(v.beaconPK); detailInfo != nil {
			Logger.log.Infof("AFTERADD: Beacon %v - %v has current %v delegator", v.beaconPK, k, detailInfo.TotalDelegators)
		}
	}
	if len(delegateState.NextEpochDelegate) != 0 {
		newDelegate := statedb.NewBeaconDelegateState()
		if err := statedb.StoreBeaconReDelegateState(s.stateDB, newDelegate); err != nil {
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

func (s *BeaconCommitteeStateV4) GetBeaconCandidateUID(cpk string) (string, error) {
	info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, cpk)
	if !exist {
		return "", fmt.Errorf("Cannot find cpk %v in statedb", cpk)
	}
	hash := common.HashH([]byte(fmt.Sprintf("%v-%v", cpk, info.GetBeaconConfirmTime())))
	return hash.String(), nil
}

func (s *BeaconCommitteeStateV4) addDelegator(beaconStakerPK string) error {

	if stakerInfo := s.getStakerInfo(beaconStakerPK); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, beaconStakerPK)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", beaconStakerPK)
		}
		info.AddDelegator()
		stakerInfo.TotalDelegators++
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", beaconStakerPK)
}

func (s *BeaconCommitteeStateV4) removeDelegator(oldBeaconStakerPK string) error {
	if stakerInfo := s.getStakerInfo(oldBeaconStakerPK); stakerInfo != nil {
		info, exist, _ := statedb.GetBeaconStakerInfo(s.stateDB, oldBeaconStakerPK)
		if !exist {
			return fmt.Errorf("Cannot find cpk %v in statedb", oldBeaconStakerPK)
		}
		info.RemoveDelegator()
		stakerInfo.TotalDelegators--
		return statedb.StoreBeaconStakerInfo(s.stateDB, stakerInfo.cpkStruct, info)
	}
	return fmt.Errorf("Cannot find cpk %v in memstate", oldBeaconStakerPK)
}
