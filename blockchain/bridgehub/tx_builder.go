package bridgehub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/metadata/bridgehub"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

type TxBuilder struct {
}

func (txBuilder TxBuilder) Build(
	metaType int,
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	var err error
	switch metaType {
	case metadataCommon.StakePRVRequestMeta:
		if inst[2] != "rejected" {
			// continue
			return nil, nil
		}
		if len(inst) != 4 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 4, len(inst))
		}
		tx, err = buildIssuingResponse(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.ShieldingBTCRequestMeta:
		if len(inst) >= 4 && inst[2] == "accepted" {
			fmt.Println("0xcrypto got into tx builder")
			//metadataCommon.ShieldingBTCResponse
			tx, err = buildBridgeHubShieldingRequest(inst, producerPrivateKey, shardID, transactionStateDB)
		}
	}

	return tx, err
}

func buildIssuingResponse(
	content []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	inst := metadataCommon.NewInstruction()
	if err := inst.FromStringSlice(content); err != nil {
		return nil, err
	}
	if inst.ShardID != shardID {
		return nil, nil
	}
	Logger.log.Info("[BridgeHub] Starting...")
	// decode failed content
	instData := metadataCommon.RejectContent{}
	err := instData.FromString(inst.Content)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while decoding reject content: ", err)
		return nil, nil
	}

	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while decoding content string of Reshield accepted issuance instruction: ", err)
		return nil, nil
	}
	var bridgeHubStakeFailed bridgehub.StakePRVRequestContentInst
	err = json.Unmarshal(instData.Data, &bridgeHubStakeFailed)
	if err != nil {
		Logger.log.Warn("WARNING: an error occurred while unmarshaling stake failed issuance instruction: ", err)
		return nil, nil
	}

	if bridgeHubStakeFailed.TokenID != common.PRVCoinID {
		return nil, fmt.Errorf("only can stake with prv for now")
	}

	txHash := &common.Hash{}
	txHash, _ = txHash.NewHashFromStr(bridgeHubStakeFailed.TxReqID)
	issuingReshieldRes := bridgehub.NewBridgeHubStakingResponse(
		*txHash,
		bridgeHubStakeFailed.StakeAmount,
		bridgeHubStakeFailed.TokenID,
		bridgeHubStakeFailed.Staker,
		metadataCommon.BridgeHubStakeResponse,
	)

	var recv = bridgeHubStakeFailed.Staker
	txParam := transaction.TxSalaryOutputParams{Amount: bridgeHubStakeFailed.StakeAmount, ReceiverAddress: nil, PublicKey: recv.PublicKey, TxRandom: &recv.TxRandom, TokenID: &bridgeHubStakeFailed.TokenID}
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, func(c privacy.Coin) metadataCommon.Metadata { return issuingReshieldRes })
}

func buildBridgeHubShieldingRequest(
	content []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	fmt.Println("0xcrypto got into tx builder 2")
	var otaReceiver privacy.OTAReceiver
	inst := metadataCommon.NewInstruction()
	if err := inst.FromStringSlice(content); err != nil {
		return nil, err
	}
	if inst.ShardID != shardID {
		return nil, nil
	}
	contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
	if err != nil {
		return nil, err
	}
	acceptedContent := bridgehub.ShieldingBTCAcceptedInst{}
	err = json.Unmarshal(contentBytes, &acceptedContent)
	if err != nil {
		return nil, err
	}
	txParam := transaction.TxSalaryOutputParams{
		Amount:          acceptedContent.IssuingAmount,
		ReceiverAddress: nil,
		PublicKey:       otaReceiver.PublicKey,
		TxRandom:        &otaReceiver.TxRandom,
		TokenID:         &acceptedContent.IncTokenID,
		Info:            []byte{},
	}
	fmt.Println("0xcrypto got into tx builder 3")
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB,
		func(c privacy.Coin) metadataCommon.Metadata {
			return bridgehub.NewIssuingBTCResponse(acceptedContent.TxReqID, acceptedContent.UniqTx, metadataCommon.ShieldingBTCResponse)
		},
	)
}
