package statedb

import (
	"fmt"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

func storeCommittee(stateDB *StateDB, shardID int, role int, committees []incognitokey.CommitteePublicKey) error {
	enterTime := time.Now().UnixNano()
	for _, committee := range committees {
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		value := NewCommitteeState()
		has := false
		value, has, err = stateDB.getCommitteeState(key)
		if err != nil {
			return err
		}
		if !has {
			value = NewCommitteeStateWithValueAndTime(shardID, role, committee, enterTime)
		}
		err = stateDB.SetStateObject(CommitteeObjectType, key, value)
		if err != nil {
			return err
		}
		enterTime++
	}
	return nil
}

func StoreBeaconCommittee(stateDB *StateDB, beaconCommittees []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconShardID, CurrentValidator, beaconCommittees)
	if err != nil {
		return NewStatedbError(StoreBeaconCommitteeError, err)
	}
	return nil
}

func StoreOneShardCommittee(stateDB *StateDB, shardID byte, shardCommittees []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, int(shardID), CurrentValidator, shardCommittees)
	if err != nil {
		return NewStatedbError(StoreShardCommitteeError, err)
	}
	return nil
}
func StoreAllShardCommittee(stateDB *StateDB, allShardCommittees map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardCommittees {
		err := storeCommittee(stateDB, int(shardID), CurrentValidator, committee)
		if err != nil {
			return NewStatedbError(StoreAllShardCommitteeError, err)
		}
	}
	return nil
}

func StoreNextEpochShardCandidate(
	stateDB *StateDB,
	candidate []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]privacy.PaymentAddress,
	// funderAddress map[string]privacy.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
	// amountStaking map[string]uint64,
) error {
	err := storeStakerInfo(stateDB, candidate, rewardReceiver, autoStaking, stakingTx)
	if err != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err)
	}
	err = storeCommittee(stateDB, CandidateShardID, NextEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err)
	}
	return nil
}

func StoreCurrentEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, CandidateShardID, CurrentEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(StoreCurrentEpochCandidateError, err)
	}
	return nil
}

func StoreNextEpochBeaconCandidate(
	stateDB *StateDB,
	candidate []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]privacy.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
) error {
	err := storeStakerInfo(stateDB, candidate, rewardReceiver, autoStaking, stakingTx)
	if err != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err)
	}
	err = storeCommittee(stateDB, BeaconShardID, NextEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(StoreNextEpochCandidateError, err)
	}
	return nil
}

func StoreCurrentEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconShardID, CurrentEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(StoreCurrentEpochCandidateError, err)
	}
	return nil
}

func StoreAllShardSubstitutesValidator(stateDB *StateDB, allShardSubstitutes map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardSubstitutes {
		err := storeCommittee(stateDB, int(shardID), SubstituteValidator, committee)
		if err != nil {
			return NewStatedbError(StoreNextEpochCandidateError, err)
		}
	}
	return nil
}

func StoreOneShardSubstitutesValidator(stateDB *StateDB, shardID byte, shardSubstitutes []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, int(shardID), SubstituteValidator, shardSubstitutes)
	if err != nil {
		return NewStatedbError(StoreOneShardSubstitutesValidatorError, err)
	}
	return nil
}

func StoreBeaconSubstituteValidator(stateDB *StateDB, beaconSubstitute []incognitokey.CommitteePublicKey) error {
	err := storeCommittee(stateDB, BeaconShardID, SubstituteValidator, beaconSubstitute)
	if err != nil {
		return NewStatedbError(StoreBeaconSubstitutesValidatorError, err)
	}
	return nil
}

func GetBeaconCommittee(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.getAllValidatorCommitteePublicKey(CurrentValidator, []int{BeaconShardID})
	tempBeaconCommitteeStates := m[BeaconShardID]
	sort.Slice(tempBeaconCommitteeStates, func(i, j int) bool {
		return tempBeaconCommitteeStates[i].EnterTime() < tempBeaconCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempBeaconCommitteeState := range tempBeaconCommitteeStates {
		list = append(list, tempBeaconCommitteeState.CommitteePublicKey())
	}
	return list
}

func GetBeaconSubstituteValidator(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	m := stateDB.getAllValidatorCommitteePublicKey(SubstituteValidator, []int{BeaconShardID})
	tempBeaconCommitteeStates := m[BeaconShardID]
	sort.Slice(tempBeaconCommitteeStates, func(i, j int) bool {
		return tempBeaconCommitteeStates[i].EnterTime() < tempBeaconCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempBeaconCommitteeState := range tempBeaconCommitteeStates {
		list = append(list, tempBeaconCommitteeState.CommitteePublicKey())
	}
	return list
}

func GetOneShardCommittee(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	tempShardCommitteeStates := stateDB.getByShardIDCurrentValidatorState(int(shardID))
	sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
		return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempShardCommitteeState := range tempShardCommitteeStates {
		list = append(list, tempShardCommitteeState.CommitteePublicKey())
	}
	return list
}

func GetOneShardSubstituteValidator(stateDB *StateDB, shardID byte) []incognitokey.CommitteePublicKey {
	tempShardCommitteeStates := stateDB.getByShardIDSubstituteValidatorState(int(shardID))
	sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
		return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempShardCommitteeState := range tempShardCommitteeStates {
		list = append(list, tempShardCommitteeState.CommitteePublicKey())
	}
	return list
}

func GetAllShardCommittee(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	tempM := stateDB.getAllValidatorCommitteePublicKey(CurrentValidator, shardIDs)
	m := make(map[int][]incognitokey.CommitteePublicKey)
	for shardID, tempShardCommitteeStates := range tempM {
		sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
			return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
		})
		list := []incognitokey.CommitteePublicKey{}
		for _, tempShardCommitteeState := range tempShardCommitteeStates {
			list = append(list, tempShardCommitteeState.CommitteePublicKey())
		}
		m[shardID] = list
	}
	return m
}

func GetAllShardSubstituteValidator(stateDB *StateDB, shardIDs []int) map[int][]incognitokey.CommitteePublicKey {
	tempM := stateDB.getAllValidatorCommitteePublicKey(SubstituteValidator, shardIDs)
	m := make(map[int][]incognitokey.CommitteePublicKey)
	for shardID, tempShardCommitteeStates := range tempM {
		sort.Slice(tempShardCommitteeStates, func(i, j int) bool {
			return tempShardCommitteeStates[i].EnterTime() < tempShardCommitteeStates[j].EnterTime()
		})
		list := []incognitokey.CommitteePublicKey{}
		for _, tempShardCommitteeState := range tempShardCommitteeStates {
			list = append(list, tempShardCommitteeState.CommitteePublicKey())
		}
		m[shardID] = list
	}
	return m
}

func GetNextEpochCandidate(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	tempNextEpochShardCandidateStates := stateDB.getAllCandidateCommitteePublicKey(NextEpochShardCandidate)
	sort.Slice(tempNextEpochShardCandidateStates, func(i, j int) bool {
		return tempNextEpochShardCandidateStates[i].EnterTime() < tempNextEpochShardCandidateStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempNextEpochShardCandidateStates := range tempNextEpochShardCandidateStates {
		list = append(list, tempNextEpochShardCandidateStates.CommitteePublicKey())
	}
	return list
}

func GetCurrentEpochCandidate(stateDB *StateDB) []incognitokey.CommitteePublicKey {
	tempCurrentEpochShardCandidateStates := stateDB.getAllCandidateCommitteePublicKey(CurrentEpochShardCandidate)
	sort.Slice(tempCurrentEpochShardCandidateStates, func(i, j int) bool {
		return tempCurrentEpochShardCandidateStates[i].EnterTime() < tempCurrentEpochShardCandidateStates[j].EnterTime()
	})
	list := []incognitokey.CommitteePublicKey{}
	for _, tempCurrentEpochShardCandidateState := range tempCurrentEpochShardCandidateStates {
		list = append(list, tempCurrentEpochShardCandidateState.CommitteePublicKey())
	}
	return list
}

func GetAllCommitteeState(stateDB *StateDB, shardIDs []int) map[int][]*CommitteeState {
	return stateDB.getShardsCommitteeState(shardIDs)
}

func GetAllCommitteeStakeInfo(stateDB *StateDB, shardIDs []int) map[int][]*StakerInfo {
	return stateDB.getShardsCommitteeInfo(shardIDs)
}

func GetMapStakingTx(bcDB *StateDB, sdb incdb.Database, shardIDs []int, shardID int) map[string]string {
	res, err := bcDB.getMapStakingTx(shardIDs)
	if err != nil {
		panic(err)
	}
	for k, v := range res {
		var txHash = &common.Hash{}
		err := (&common.Hash{}).Decode(txHash, v)
		if err != nil {
			incdb.Logger.Log.Error(err)
			delete(res, k)
			continue
		}
		_, _, err = rawdbv2.GetTransactionByHash(sdb, *txHash)
		if err != nil {
			incdb.Logger.Log.Warn(err)
			delete(res, k)
			continue
		}
	}
	return res
}

func GetStakerInfo(stateDB *StateDB, stakerPubkey string) (*StakerInfo, bool, error) {
	pubKey := incognitokey.NewCommitteePublicKey()
	err := pubKey.FromString(stakerPubkey)
	if err != nil {
		return nil, false, err
	}
	pubKeyBytes, _ := pubKey.RawBytes()
	key := GetStakerInfoKey(pubKeyBytes)
	return stateDB.getStakerInfo(key)
}

func deleteCommittee(stateDB *StateDB, shardID int, role int, committees []incognitokey.CommitteePublicKey) error {
	for _, committee := range committees {
		key, err := GenerateCommitteeObjectKeyWithRole(role, shardID, committee)
		if err != nil {
			return err
		}
		stateDB.MarkDeleteStateObject(CommitteeObjectType, key)
	}
	return nil
}

func DeleteBeaconCommittee(stateDB *StateDB, beaconCommittees []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, CurrentValidator, beaconCommittees)
	if err != nil {
		return NewStatedbError(DeleteBeaconCommitteeError, err)
	}
	return nil
}

func DeleteOneShardCommittee(stateDB *StateDB, shardID byte, shardCommittees []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, int(shardID), CurrentValidator, shardCommittees)
	if err != nil {
		return NewStatedbError(DeleteOneShardCommitteeError, err)
	}
	return nil
}
func DeleteAllShardCommittee(stateDB *StateDB, allShardCommittees map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardCommittees {
		err := deleteCommittee(stateDB, int(shardID), CurrentValidator, committee)
		if err != nil {
			return NewStatedbError(DeleteAllShardCommitteeError, err)
		}
	}
	return nil
}

func DeleteNextEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, CandidateShardID, NextEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteNextEpochShardCandidateError, err)
	}
	return nil
}

func DeleteCurrentEpochShardCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, CandidateShardID, CurrentEpochShardCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteCurrentEpochShardCandidateError, err)
	}
	return nil
}

func DeleteNextEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, NextEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteNextEpochBeaconCandidateError, err)
	}
	return nil
}

func DeleteCurrentEpochBeaconCandidate(stateDB *StateDB, candidate []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, CurrentEpochBeaconCandidate, candidate)
	if err != nil {
		return NewStatedbError(DeleteCurrentEpochBeaconCandidateError, err)
	}
	return nil
}

func DeleteAllShardSubstitutesValidator(stateDB *StateDB, allShardSubstitutes map[byte][]incognitokey.CommitteePublicKey) error {
	for shardID, committee := range allShardSubstitutes {
		err := deleteCommittee(stateDB, int(shardID), SubstituteValidator, committee)
		if err != nil {
			return NewStatedbError(DeleteAllShardSubstitutesValidatorError, err)
		}
	}
	return nil
}

func DeleteOneShardSubstitutesValidator(stateDB *StateDB, shardID byte, shardSubstitutes []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, int(shardID), SubstituteValidator, shardSubstitutes)
	if err != nil {
		return NewStatedbError(DeleteAllShardSubstitutesValidatorError, err)
	}
	return nil
}

func DeleteBeaconSubstituteValidator(stateDB *StateDB, beaconSubstitute []incognitokey.CommitteePublicKey) error {
	err := deleteCommittee(stateDB, BeaconShardID, SubstituteValidator, beaconSubstitute)
	if err != nil {
		return NewStatedbError(DeleteBeaconSubstituteValidatorError, err)
	}
	return nil
}

func storeStakerInfo(
	stateDB *StateDB,
	committees []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]privacy.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
) error {
	for _, committee := range committees {
		keyBytes, err := committee.RawBytes()
		if err != nil {
			return err
		}
		key := GetStakerInfoKey(keyBytes)
		committeeString, err := committee.ToBase58()
		if err != nil {
			return err
		}
		value := NewStakerInfo()
		has := false
		value, has, err = stateDB.getStakerInfo(key)
		if err != nil {
			return err
		}
		autoStakingValue, ok := autoStaking[committeeString]
		if !ok {
			return fmt.Errorf("auto staking of %+v not found", committeeString)
		}
		if !has {
			rewardReceiverPaymentAddress, ok := rewardReceiver[committee.GetIncKeyBase58()]
			if !ok {
				return fmt.Errorf("reward receiver of %+v not found", committeeString)
			}
			txStakingID, ok := stakingTx[committeeString]
			if !ok {
				return fmt.Errorf("txStakingID of %+v not found", committeeString)
			}
			value = NewStakerInfoWithValue(rewardReceiverPaymentAddress, autoStakingValue, txStakingID)
		} else {
			value.autoStaking = autoStakingValue
		}
		err = stateDB.SetStateObject(StakerObjectType, key, value)
		if err != nil {
			return err
		}
		delete(autoStaking, committeeString)
		if _, ok := stakingTx[committeeString]; ok {
			delete(stakingTx, committeeString)
		}
		if _, ok := rewardReceiver[committee.GetIncKeyBase58()]; ok {
			delete(stakingTx, committee.GetIncKeyBase58())
		}
	}
	return nil
}

func StoreStakerInfo(
	stateDB *StateDB,
	committees []incognitokey.CommitteePublicKey,
	rewardReceiver map[string]privacy.PaymentAddress,
	autoStaking map[string]bool,
	stakingTx map[string]common.Hash,
) error {
	return storeStakerInfo(stateDB, committees, rewardReceiver, autoStaking, stakingTx)
}
