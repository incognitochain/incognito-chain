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
		acceptedInst := metadataBridge.AcceptedConvertTokenToUnifiedToken{}
		err = json.Unmarshal(contentBytes, &acceptedInst)
		if err != nil {
			return nil, err
		}
		amount = acceptedInst.Amount
		tokenID = acceptedInst.UnifiedTokenID
		otaReceiver = acceptedInst.Receivers[tokenID]
		txReqID = acceptedInst.TxReqID
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
	var address privacy.PaymentAddress
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
		//mdData, _ := rejectContent.Meta.(*metadataBridge.BurningRequest)
		/*amount = mdData.BurningAmount*/
		/*tokenID = mdData.TokenID*/
		/*address = mdData.BurnerAddress*/
	default:
		return nil, nil
	}
	md := metadataBridge.NewUnshieldResponseWithValue(inst.Status, txReqID, nil)
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: &address, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			md.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return md
	}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
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
	case common.RejectedStatusStr:
		rejectContent := metadataCommon.NewRejectContent()
		if err := rejectContent.FromString(inst.Content); err != nil {
			return nil, err
		}
		txReqID = rejectContent.TxReqID
		//mdData, _ := rejectContent.Meta.(*metadataBridge.BurningRequest)
		/*amount = mdData.BurningAmount*/
		/*tokenID = mdData.TokenID*/
		/*address = mdData.BurnerAddress*/
	default:
		return nil, nil
	}
	md := metadataBridge.NewUnshieldResponseWithValue(inst.Status, txReqID, nil)
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: &address, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			md.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return md
	}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}
