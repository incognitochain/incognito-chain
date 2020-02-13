package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

// beacon build new instruction from instruction received from ShardToBeaconBlock
func buildCustodianDepositAcceptedInst(
	custodianAddressStr string,
	depositedAmount uint64,
	remoteAddresses map[string]string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	custodianDepositContent := metadata.PortalCustodianDepositContent{
		IncogAddressStr: custodianAddressStr,
		RemoteAddresses:remoteAddresses,
		DepositedAmount: depositedAmount,
		TxReqID:               txReqID,
	}
	custodianDepositContentBytes, _ := json.Marshal(custodianDepositContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		common.PortalCustodianDepositAcceptedChainStatus,
		string(custodianDepositContentBytes),
	}
}

func (blockchain *BlockChain) buildInstructionsForCustodianDeposit(
	contentStr string,
	shardID byte,
	metaType int,
	currentPDEState *CurrentPortalState,
	beaconHeight uint64,
) ([][]string, error) {


	// todo: validate instruction
	// case 1: custodian deposit accepted instruction
	//inst := buildCustodianDepositAcceptedInst ()

	// case 2: custodian deposit accepted instruction
	return [][]string{}, nil
}
