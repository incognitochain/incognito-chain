package manager

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
)

type BeaconManager struct {
	CommitteeEngine BeaconCommitteeEngine
}

func (bM *BeaconManager) FromInstructionsToInstructions(inses []instruction.Instruction) ([]instruction.Instruction, error) {
	res := []instruction.Instruction{}
	for _, ins := range inses {
		switch ins.GetType() {
		case instruction.RANDOM_ACTION:
			rI, ok := ins.(*instruction.RandomInstruction)
			if !ok {
				//TODO handle error
				continue
			}
			newInses := buildAssignInstructionFromRandom(rI, bM.CommitteeEngine)
			res = append(res, newInses...)
		}
	}
	return res, nil
}

func (bM *BeaconManager) BuildInstructionsFromTransactions(txs []metadata.Transaction) []instruction.Instruction {
	res := []instruction.Instruction{}
	stakeInsMap := map[string]*instruction.StakeInstruction{}
	stopIns := instruction.NewStopAutoStakeInstruction()
	for _, tx := range txs {
		switch tx.GetMetadataType() {
		case metadata.ShardStakingMeta, metadata.BeaconStakingMeta:
			err := stakeInsFromTx(
				tx.GetSenderAddrLastByte(),
				tx.GetMetadata(),
				tx.Hash().String(),
				stakeInsMap,
			)
			if err != nil {
				//TODO handle error
				continue
			}
		case metadata.StopAutoStakingMeta:
			err := stopInsFromTx(
				tx.GetSenderAddrLastByte(),
				tx.GetMetadata(),
				stopIns,
			)
			if err != nil {
				//TODO
				continue
			}
		case metadata.UnStakingMeta:
			unstakeIns, err := unstakeInsFromTx(
				tx.GetSenderAddrLastByte(),
				tx.GetMetadata())
			if err != nil {
				continue
			}
			res = append(res, unstakeIns)
		}
	}
	for _, v := range stakeInsMap {
		if v != nil {
			res = append(res, v)
		}
	}
	if len(stopIns.PublicKeys) != 0 {
		res = append(res, stopIns)
	}
	return nil
}

func (bM *BeaconManager) StoreInfoFromInstructions(sDB *statedb.StateDB, inses []instruction.Instruction) error {
	//TODO:
	//Should we using this func for filter instructions, like create empty stateDB, insert and check if they return error? If error == nil, we append it into instruction array
	//It mean this func will be called twice, once when create and filter instructions, once when store block?
	//This code implement below is just for store block, not for filter
	for _, ins := range inses {
		err := ins.InsertIntoStateDB(sDB)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bM *BeaconManager) BuildEpochInstructions() {}
