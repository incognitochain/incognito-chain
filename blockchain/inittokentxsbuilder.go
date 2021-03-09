package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/transaction"
)

//buildInstructionsForTokenInitReq builds an instruction indicating whether a tokenInit request is accepted or rejected
//following 2 steps:
//	1. Verify if the token has existed
//		1.1 In the current block
//		1.2 In the db (all previous blocks)
//	2. Build a rejected/accepted instruction for shard. If accepted, add the token to the accumulated values.
func (blockchain *BlockChain) buildInstructionsForTokenInitReq(
	beaconBestState *BeaconBestState,
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	Logger.log.Info("Starting building instructions for token init requests...")
	instructions := [][]string{}
	initTokenReqAction, err := metadata.ParseInitTokenInstContent(contentStr)
	if err != nil {
		Logger.log.Warn("ParseInitTokenInstContent: ", err)
		return nil, nil
	}

	initTokenReq := initTokenReqAction.Meta
	tokenID := initTokenReqAction.TokenID
	tokenName := initTokenReq.TokenName
	tokenSymbol := initTokenReq.TokenSymbol
	rejectedInst := buildInstruction(metaType, shardID, "rejected", initTokenReqAction.TxReqID.String())

	// check existence in the current block (on mem)
	if !ac.CanProcessTokenInit(tokenID) {
		Logger.log.Warnf("tokenID %v has already existed in the current block\n", tokenID.String())
		return append(instructions, rejectedInst), nil
	}

	// check existence in previous blocks (on blockchain's db)
	privacyTokenExisted, err := blockchain.PrivacyTokenIDExistedInAllShards(beaconBestState, tokenID)
	if err != nil {
		Logger.log.Warnf("checking tokenID existed error: %v\n", err)
		return append(instructions, rejectedInst), nil
	}
	if privacyTokenExisted {
		Logger.log.Warnf("tokenID %v has already existed in the db\n", tokenID.String())
		return append(instructions, rejectedInst), nil
	}

	otaBytes, _, err := base58.Base58Check{}.Decode(initTokenReq.OTAStr)
	if err != nil {
		Logger.log.Warnf("cannot decode OTAStr (%v): %v", initTokenReq.OTAStr, otaBytes)
		return append(instructions, rejectedInst), nil
	}
	lastByte := otaBytes[len(otaBytes)-1]
	receivingShardID := common.GetShardIDFromLastByte(lastByte)

	initTokenAcceptedInst := metadata.InitTokenAcceptedInst{
		OTAStr:        initTokenReq.OTAStr,
		TxRandomStr:   initTokenReq.TxRandomStr,
		Amount:        initTokenReq.Amount,
		TokenID:       tokenID,
		TokenName:     tokenName,
		TokenSymbol:   tokenSymbol,
		TokenType:     statedb.InitToken,
		ShardID:       receivingShardID,
		RequestedTxID: initTokenReqAction.TxReqID,
	}
	initTokenAcceptedInstBytes, err := json.Marshal(initTokenAcceptedInst)
	if err != nil {
		Logger.log.Warnf("marshal a tokenInit instruction error: %v\n", err)
		return append(instructions, rejectedInst), nil
	}

	ac.InitTokens = append(ac.InitTokens, &tokenID)
	returnedInst := buildInstruction(metaType, shardID, "accepted", base64.StdEncoding.EncodeToString(initTokenAcceptedInstBytes))
	return append(instructions, returnedInst), nil
}

func (blockGenerator *BlockGenerator) buildTokenInitAcceptedTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
) (metadata.Transaction, error) {
	acceptedContent, err := parseInitTokenAcceptedContent(contentStr)
	if err != nil {
		return nil, nil
	}

	if shardID != acceptedContent.ShardID {
		return nil, nil
	}

	meta := metadata.NewInitTokenResponse(
		acceptedContent.RequestedTxID,
		metadata.InitTokenResponseMeta,
	)

	resTx, err := buildTokenInitTx(
		acceptedContent.OTAStr,
		acceptedContent.TxRandomStr,
		acceptedContent.Amount,
		acceptedContent.TokenID.String(),
		producerPrivateKey,
		shardView.GetCopiedTransactionStateDB(),
		meta,
	)
	if err != nil {
		Logger.log.Errorf("buildTokenInitTx failed: %v\n", err)
		return nil, nil
	}
	Logger.log.Info("[Token Init] Create accepted tx ok.")
	return resTx, nil
}

func buildTokenInitTx(
	otaStr string,
	txRandomStr string,
	receiveAmt uint64,
	tokenIDStr string,
	producerPrivateKey *privacy.PrivateKey,
	transactionStateDB *statedb.StateDB,
	meta metadata.Metadata,
) (metadata.Transaction, error) {

	tokenID, err := common.Hash{}.NewHashFromStr(tokenIDStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while converting tokenid to hash: %+v", err)
		return nil, err
	}

	if len(txRandomStr) == 0 || len(otaStr) == 0 {
		Logger.log.Errorf("txRandomStr or otaStr is empty\n")
		return nil, fmt.Errorf("txRandomStr (%v) or otaStr(%) is empty", txRandomStr, otaStr)
	}

	publicKey, txRandom, err := coin.ParseOTAInfoFromString(otaStr, txRandomStr)
	if err != nil {
		Logger.log.Errorf("ParseOTAInfoFromString error: %v\n", err)
		return nil, err
	}

	txParam := transaction.TxSalaryOutputParams{Amount: receiveAmt, ReceiverAddress: nil, PublicKey: publicKey, TxRandom: txRandom, TokenID: tokenID, Info: []byte{}}
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, func(c privacy.Coin) metadata.Metadata {
		return meta
	})
}

func parseInitTokenAcceptedContent(
	contentStr string,
) (*metadata.InitTokenAcceptedInst, error) {
	contentBytes := []byte(contentStr)
	var acceptedContent metadata.InitTokenAcceptedInst
	err := json.Unmarshal(contentBytes, &acceptedContent)
	if err != nil {
		Logger.log.Errorf("parsing tokenInit accepted instruction failed: %v\n", err)
		return nil, err
	}
	return &acceptedContent, nil
}
