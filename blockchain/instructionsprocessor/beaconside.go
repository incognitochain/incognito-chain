package instructionsprocessor

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
)

type BInsProcessor struct {
	CommitteeEngine BeaconCommitteeEngine
}

func (bP *BInsProcessor) FromInstructionsToInstructions(inses []instruction.Instruction) ([]instruction.Instruction, error) {
	res := []instruction.Instruction{}
	for _, ins := range inses {
		switch ins.GetType() {
		case instruction.RANDOM_ACTION:
			rI, ok := ins.(*instruction.RandomInstruction)
			if !ok {
				//TODO handle error
				continue
			}
			newInses := buildAssignInstructionFromRandom(rI, bP.CommitteeEngine)
			res = append(res, newInses...)
		case instruction.SWAP_ACTION:
		}
	}
	return res, nil
}

func (bP *BInsProcessor) BuildInstructionsFromTransactions(txs []metadata.Transaction) []instruction.Instruction {
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
			//case metadata.UnstakingMeta
			//TODO @tin
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

func (bP *BInsProcessor) StoreInfoFromInstructions(stDB *statedb.StateDB, inses []instruction.Instruction) error {
	//TODO:
	//Should we using this func for filter instructions, like create empty stateDB, insert and check if they return error? If error == nil, we append it into instruction array
	//It mean this func will be called twice, once when create and filter instructions, once when store block?
	//This code implement below is just for store block, not for filter
	for _, ins := range inses {
		switch ins.GetType() {
		case instruction.ASSIGN_ACTION:
			insertInstructionsAssign(stDB, ins)
		case instruction.STAKE_ACTION:
			insertInstructionsStake(stDB, ins)
		case instruction.STOP_AUTO_STAKE_ACTION:
			insertInstructionsStopAutoStake(stDB, ins)
		case instruction.SWAP_ACTION:
			insertInstructionsSwap(stDB, ins)
		}
	}
	return nil
}

func (bP *BInsProcessor) BuildEpochInstructions() []instruction.Instruction {
	//Build Swap/ReturnStaking/RewardInstructions
	return []instruction.Instruction{}
}

// func ()
