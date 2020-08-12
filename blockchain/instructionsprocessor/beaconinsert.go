package instructionsprocessor

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

func insertInstructionsSwap(bcDB *statedb.StateDB, ins instruction.Instruction) error {
	swI, ok := ins.(*instruction.SwapInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	if swI.IsReplace {
		//TODO Merge code replace
		return nil
	}
	if swI.ChainID == instruction.BEACON_CHAIN_ID {
		err := statedb.StoreBeaconCommittee(bcDB, swI.InPublicKeyStructs)
		if err != nil {
			return err
		}
		err = statedb.DeleteBeaconSubstituteValidator(bcDB, swI.InPublicKeyStructs)
		if err != nil {
			return err
		}
		return statedb.DeleteBeaconCommittee(bcDB, swI.OutPublicKeyStructs)
	}
	err := statedb.StoreOneShardCommittee(bcDB, byte(swI.ChainID), swI.InPublicKeyStructs)
	if err != nil {
		return err
	}
	err = statedb.DeleteOneShardSubstitutesValidator(bcDB, byte(swI.ChainID), swI.InPublicKeyStructs)
	if err != nil {
		return err
	}
	return statedb.DeleteOneShardCommittee(bcDB, byte(swI.ChainID), swI.OutPublicKeyStructs)
}

func insertInstructionsStopAutoStake(sDB *statedb.StateDB, ins instruction.Instruction) error {
	saI, ok := ins.(*instruction.StopAutoStakeInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	pkStructs, err := incognitokey.CommitteeBase58KeyListToStruct(saI.PublicKeys)
	if err != nil {
		return err
	}
	// TODO:
	// Instead of preprocessing for create input for storeStakerInfo, create function update stakerinfo
	asMap := map[string]bool{}
	for _, pk := range saI.PublicKeys {
		asMap[pk] = true
	}
	return statedb.StoreStakerInfo(
		sDB,
		pkStructs,
		map[string]privacy.PaymentAddress{}, //Empty map cuz we just update auto staking flag
		asMap,
		map[string]common.Hash{},
	)
}

func insertInstructionsStake(sDB *statedb.StateDB, ins instruction.Instruction) error {
	sI, ok := ins.(*instruction.StakeInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	rrMap := map[string]privacy.PaymentAddress{}
	asMap := map[string]bool{}
	tsMap := map[string]common.Hash{}
	for i, pk := range sI.PublicKeys {
		rrMap[pk] = sI.RewardReceiverStructs[i]
		asMap[pk] = sI.AutoStakingFlag[i]
		tsMap[pk] = sI.TxStakeHashes[i]
	}
	err := statedb.StoreStakerInfo(
		sDB,
		sI.PublicKeyStructs,
		rrMap,
		asMap,
		tsMap,
	)
	if err != nil {
		return err
	}
	//TODO Replace from Next Epoch ---> Common Pool
	if sI.Chain == "beacon" {
		return statedb.StoreNextEpochBeaconCandidate(
			sDB,
			sI.PublicKeyStructs,
			rrMap,
			asMap,
			tsMap,
		)
	}
	return statedb.StoreNextEpochShardCandidate(
		sDB,
		sI.PublicKeyStructs,
		rrMap,
		asMap,
		tsMap,
	)
}

func insertInstructionsAssign(sDB *statedb.StateDB, ins instruction.Instruction) error {
	aI, ok := ins.(*instruction.AssignInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	candidates, err := incognitokey.CommitteeBase58KeyListToStruct(aI.ShardCandidates)
	if err != nil {
		return err
	}
	if aI.ChainID == instruction.BEACON_CHAIN_ID {
		err = statedb.StoreBeaconSubstituteValidator(sDB, candidates)
		if err != nil {
			return err
		}
	}
	err = statedb.StoreOneShardSubstitutesValidator(sDB, byte(aI.ChainID), candidates)
	if err != nil {
		return err
	}
	return nil
}

func insertInstructionsSwapv2(bcDB *statedb.StateDB, ins instruction.Instruction) error {
	swI, ok := ins.(*instruction.SwapInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	if swI.IsReplace {
		//TODO Merge code replace
		return nil
	}
	if swI.ChainID == instruction.BEACON_CHAIN_ID {
		err := statedb.StoreBeaconCommittee(bcDB, swI.InPublicKeyStructs)
		if err != nil {
			return err
		}
		err = statedb.DeleteMembersAtBeaconPool(bcDB, swI.InPublicKeyStructs)
		if err != nil {
			return err
		}
		return statedb.DeleteBeaconCommittee(bcDB, swI.OutPublicKeyStructs)
	}
	err := statedb.StoreOneShardCommittee(bcDB, byte(swI.ChainID), swI.InPublicKeyStructs)
	if err != nil {
		return err
	}
	err = statedb.DeleteMembersAtShardPool(bcDB, byte(swI.ChainID), swI.InPublicKeyStructs)
	if err != nil {
		return err
	}
	return statedb.DeleteOneShardCommittee(bcDB, byte(swI.ChainID), swI.OutPublicKeyStructs)
}

func insertInstructionsStopAutoStakev2(sDB *statedb.StateDB, ins instruction.Instruction) error {
	saI, ok := ins.(*instruction.StopAutoStakeInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	pkStructs, err := incognitokey.CommitteeBase58KeyListToStruct(saI.PublicKeys)
	if err != nil {
		return err
	}
	// TODO:
	// Instead of preprocessing for create input for storeStakerInfo, create function update stakerinfo
	asMap := map[string]bool{}
	for _, pk := range saI.PublicKeys {
		asMap[pk] = true
	}
	return statedb.StoreStakerInfo(
		sDB,
		pkStructs,
		map[string]privacy.PaymentAddress{}, //Empty map cuz we just update auto staking flag
		asMap,
		map[string]common.Hash{},
	)
}

func insertInstructionsStakev2(sDB *statedb.StateDB, ins instruction.Instruction) error {
	sI, ok := ins.(*instruction.StakeInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	rrMap := map[string]privacy.PaymentAddress{}
	asMap := map[string]bool{}
	tsMap := map[string]common.Hash{}
	for i, pk := range sI.PublicKeys {
		rrMap[pk] = sI.RewardReceiverStructs[i]
		asMap[pk] = sI.AutoStakingFlag[i]
		tsMap[pk] = sI.TxStakeHashes[i]
	}
	err := statedb.StoreStakerInfo(
		sDB,
		sI.PublicKeyStructs,
		rrMap,
		asMap,
		tsMap,
	)
	if err != nil {
		return err
	}
	//TODO Replace from Next Epoch ---> Common Pool
	if sI.Chain == "beacon" {
		return statedb.StoreMembersAtCommonBeaconPool(
			sDB,
			sI.PublicKeyStructs,
		)
	}
	return statedb.StoreMembersAtCommonShardPool(
		sDB,
		sI.PublicKeyStructs,
	)
}

func insertInstructionsAssignv2(sDB *statedb.StateDB, ins instruction.Instruction) error {
	aI, ok := ins.(*instruction.AssignInstruction)
	if !ok {
		return errors.New("Can not parse this instruction to AssignInstructions")
	}
	candidates, err := incognitokey.CommitteeBase58KeyListToStruct(aI.ShardCandidates)
	if err != nil {
		return err
	}
	if aI.ChainID == instruction.BEACON_CHAIN_ID {
		err = statedb.StoreMembersAtBeaconPool(sDB, candidates)
		if err != nil {
			return err
		}
	}
	err = statedb.StoreMembersAtShardPool(sDB, byte(aI.ChainID), candidates)
	if err != nil {
		return err
	}
	return nil
}
