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
	"github.com/incognitochain/incognito-chain/wallet"
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
	case metadataCommon.ShieldUnifiedTokenRequestMeta:
		if len(inst) != 4 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 4, len(inst))
		}
		tx, err = buildShieldUnifiedTokenResponse(inst, producerPrivateKey, shardID, transactionStateDB)
	case metadataCommon.UnshieldUnifiedTokenRequestMeta:
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
		amount = acceptedContent.Amount
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
		tokenID = rejectedUnshieldRequest.TokenID
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
	var tx metadata.Transaction
	var txReqID, tokenID common.Hash
	var amount uint64
	var address privacy.PaymentAddress
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
		acceptedInst := metadataBridge.AcceptedShieldRequest{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			return nil, err
		}
		key, err := wallet.Base58CheckDeserialize(acceptedInst.Receiver)
		if err != nil {
			Logger.log.Warn("WARNING: an error occurred while deserializing receiver address string: ", err)
			return nil, err
		}
		address = key.KeySet.PaymentAddress
		for _, data := range acceptedInst.Data {
			amount += data.IssuingAmount
		}
		tokenID = acceptedInst.IncTokenID
		txReqID = acceptedInst.TxReqID
	default:
		return nil, nil
	}

	md := metadataBridge.NewUnshieldResponseWithValue(inst.Status, txReqID)
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: &address, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			md.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return md
	}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}
