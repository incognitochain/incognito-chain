package bridgeagg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
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
	case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
		if len(inst) != 4 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 4, len(inst))
		}
		tx, err = buildBridgeAggConvertTokenUnifiedTokenResponse(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.IssuingUnifiedTokenRequestMeta:
		if len(inst) != 4 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 4, len(inst))
		}
		tx, err = buildShieldUnifiedTokenResponse(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.BurningUnifiedTokenRequestMeta:
		if len(inst) != 4 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 4, len(inst))
		}
		tx, err = buildUnshieldUnifiedTokenResponse(inst, producerPrivateKey, shardID, transactionStateDB)
	}
	return tx, err
}

func buildBridgeAggConvertTokenUnifiedTokenResponse(
	content []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	var txReqID, tokenID common.Hash
	var otaReceiver privacy.OTAReceiver
	var amount uint64
	inst := metadataCommon.NewInstruction()
	if err := inst.FromStringSlice(content); err != nil {
		return tx, err
	}
	if inst.ShardID != shardID {
		return nil, nil
	}
	switch inst.Status {
	case common.AcceptedStatusStr:
		contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
		if err != nil {
			return nil, err
		}
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{}
		err = json.Unmarshal(contentBytes, &acceptedContent)
		if err != nil {
			return nil, err
		}
		amount = acceptedContent.MintAmount
		tokenID = acceptedContent.UnifiedTokenID
		otaReceiver = acceptedContent.Receiver
		txReqID = acceptedContent.TxReqID
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return nil, err
		}
		txReqID = rejectContent.TxReqID
		var rejectedConvertTokenToUnifiedToken metadataBridge.RejectedConvertTokenToUnifiedToken
		if err := json.Unmarshal(rejectContent.Data, &rejectedConvertTokenToUnifiedToken); err != nil {
			return nil, err
		}
		amount = rejectedConvertTokenToUnifiedToken.Amount
		tokenID = rejectedConvertTokenToUnifiedToken.TokenID
		otaReceiver = rejectedConvertTokenToUnifiedToken.Receiver
	}
	md := metadataBridge.NewBridgeAggConvertTokenToUnifiedTokenResponseWithValue(inst.Status, txReqID)
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: nil, PublicKey: &otaReceiver.PublicKey, TxRandom: &otaReceiver.TxRandom, TokenID: &tokenID, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB,
		func(c privacy.Coin) metadataCommon.Metadata { return md },
	)
}

func buildUnshieldUnifiedTokenResponse(
	content []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadata.Transaction, error) {
	var tx metadata.Transaction
	var txReqID, tokenID common.Hash
	var amount uint64
	var otaReceiver privacy.OTAReceiver
	inst := metadataCommon.NewInstruction()
	if err := inst.FromStringSlice(content); err != nil {
		return tx, err
	}
	if inst.ShardID != shardID {
		return nil, nil
	}
	switch inst.Status {
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return nil, err
		}
		txReqID = rejectContent.TxReqID
		var rejectedUnshieldRequest metadataBridge.RejectedUnshieldRequest
		if err := json.Unmarshal(rejectContent.Data, &rejectedUnshieldRequest); err != nil {
			return nil, err
		}
		amount = rejectedUnshieldRequest.Amount
		tokenID = rejectedUnshieldRequest.UnifiedTokenID
		otaReceiver = rejectedUnshieldRequest.Receiver
	default:
		return nil, nil
	}
	md := metadataBridge.NewUnshieldResponseWithValue(inst.Status, txReqID)
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: nil, PublicKey: &otaReceiver.PublicKey, TxRandom: &otaReceiver.TxRandom, TokenID: &tokenID, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB,
		func(c privacy.Coin) metadataCommon.Metadata { return md },
	)
}

func buildShieldUnifiedTokenResponse(
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
	if inst.Status != common.AcceptedStatusStr {
		return nil, nil
	}

	contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
	if err != nil {
		return nil, err
	}
	acceptedContent := metadataBridge.AcceptedInstShieldRequest{}
	err = json.Unmarshal(contentBytes, &acceptedContent)
	if err != nil {
		return nil, err
	}

	// calculate total shield amount and reward
	var shieldResponseDatas []metadataBridge.ShieldResponseData
	var shieldAmount, reward uint64
	for _, data := range acceptedContent.Data {
		shieldAmount += data.ShieldAmount
		reward += data.Reward
		shieldResponseData := metadataBridge.ShieldResponseData{
			ExternalTokenID: data.ExternalTokenID,
			UniqTx:          data.UniqTx,
			IncTokenID:      data.IncTokenID,
		}
		shieldResponseDatas = append(shieldResponseDatas, shieldResponseData)
	}
	mintAmount := shieldAmount + reward
	if mintAmount == 0 {
		return nil, nil
	}

	md := metadataBridge.NewShieldResponseWithValue(metadataCommon.IssuingUnifiedTokenResponseMeta, shieldAmount, reward, shieldResponseDatas, acceptedContent.TxReqID, nil)
	txParam := transaction.TxSalaryOutputParams{Amount: mintAmount, ReceiverAddress: &acceptedContent.Receiver, TokenID: &acceptedContent.UnifiedTokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			md.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return md
	}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}