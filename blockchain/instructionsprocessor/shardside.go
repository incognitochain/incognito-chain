package instructionsprocessor

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

type SInsProcessor struct {
	ShardDB map[byte]incdb.Database
}

func (sP *SInsProcessor) FromInstructionsToInstructions(inses []instruction.Instruction) []instruction.Instruction {

	return nil
}

func (sP *SInsProcessor) BuildTransactionsFromInstructions(
	inses []instruction.Instruction,
	txStateDB *statedb.StateDB,
	producerPrivateKey *privacy.PrivateKey, shardID byte,
) []metadata.Transaction {
	res := []metadata.Transaction{}
	for _, ins := range inses {
		switch ins.GetType() {
		case instruction.RETURN_ACTION:
			insStake, ok := ins.(*instruction.ReturnStakeIns)
			if !ok {
				//TODO find the correctly way to handle error here, panic or continue or do something else?
				continue
			}
			tx, err := returnStakingFromIns(*insStake, producerPrivateKey, sP.ShardDB[shardID], txStateDB)
			if err != nil {
				res = append(res, tx)
			}
		}

	}
	return nil
}
