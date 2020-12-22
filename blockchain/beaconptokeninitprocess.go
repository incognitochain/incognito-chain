package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) processPTokenInitInstructions(pTokenInitStateDB *statedb.StateDB, block *BeaconBlock) {
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		switch inst[0] {
		case strconv.Itoa(metadata.InitPTokenRequestMeta):
			blockchain.processPTokenInitReq(pTokenInitStateDB, inst)
		}
	}
}

func (blockchain *BlockChain) processPTokenInitReq(pTokenInitStateDB *statedb.StateDB, instruction []string) {
	if len(instruction) != 4 {
		return // skip the instruction
	}
	if instruction[2] == "rejected" {
		Logger.log.Info("pToken init rejected!")
		return
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while decoding content string of accepted ptoken init instruction: ", err)
		return
	}

	var initPTokenAcceptedInst metadata.InitPTokenAcceptedInst
	err = json.Unmarshal(contentBytes, &initPTokenAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while unmarshaling accepted ptoken init instruction: ", err)
		return
	}

	err = statedb.StorePTokenInit(
		pTokenInitStateDB,
		initPTokenAcceptedInst.IncTokenID.String(),
		initPTokenAcceptedInst.IncTokenName,
		initPTokenAcceptedInst.IncTokenSymbol,
		initPTokenAcceptedInst.Amount,
	)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while storing ptoken intialization to leveldb: ", err)
	}
}
