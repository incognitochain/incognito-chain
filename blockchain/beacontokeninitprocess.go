package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"strconv"
)

func (blockchain *BlockChain) processTokenInitInstructions(tokenInitStateDB *statedb.StateDB, block *types.BeaconBlock) {
	for _, inst := range block.Body.Instructions {
		if len(inst) != 4 {
			continue
		}
		switch inst[0] {
		case strconv.Itoa(metadata.InitTokenRequestMeta):
			blockchain.processTokenInitReq(tokenInitStateDB, inst)
		}
	}
}

func (blockchain *BlockChain) processTokenInitReq(tokenInitStateDB *statedb.StateDB, instruction []string) {
	if len(instruction) != 4 {
		return // skip the instruction
	}
	if instruction[2] == "rejected" {
		Logger.log.Info("token init rejected!")
		return
	}
	contentBytes, err := base64.StdEncoding.DecodeString(instruction[3])
	if err != nil {
		Logger.log.Warnf("decode instruction content failed: %v\n", err)
		return
	}

	var acceptedInst metadata.InitTokenAcceptedInst
	err = json.Unmarshal(contentBytes, &acceptedInst)
	if err != nil {
		Logger.log.Warnf("unmarshal tokenInit accepted instruction fail: %v\n", err)
		return
	}

	err = statedb.StorePrivacyToken(
		tokenInitStateDB,
		acceptedInst.TokenID,
		acceptedInst.TokenName,
		acceptedInst.TokenSymbol,
		acceptedInst.TokenType,
		false,
		acceptedInst.Amount,
		[]byte{},
		acceptedInst.RequestedTxID,
	)
	if err != nil {
		Logger.log.Warnf("store privacy token failed: %v\n", err)
	}
}
