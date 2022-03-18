package bridgeagg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridgeAgg "github.com/incognitochain/incognito-chain/metadata/bridgeagg"
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
	case metadataCommon.BurningUnifiedTokenRequestMeta:
		if len(inst) != 4 {
			return tx, fmt.Errorf("Length of instruction is invalid expect equal or greater than %v but get %v", 4, len(inst))
		}
		tx, err = buildBurningUnifiedTokenResponse(inst, producerPrivateKey, shardID, transactionStateDB)
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
		acceptedInst := metadataBridgeAgg.AcceptedConvertTokenToUnifiedToken{}
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
		mdData, _ := rejectContent.Meta.(*metadataBridgeAgg.ConvertTokenToUnifiedTokenRequest)
		amount = mdData.Amount
		tokenID = mdData.TokenID
		otaReceiver = mdData.Receivers[tokenID]
	}
	md := metadataBridgeAgg.NewBridgeAggConvertTokenToUnifiedTokenResponseWithValue(inst.Status, txReqID)
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: nil, PublicKey: &otaReceiver.PublicKey, TxRandom: &otaReceiver.TxRandom, TokenID: &tokenID, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB,
		func(c privacy.Coin) metadataCommon.Metadata { return md },
	)
}

func buildBurningUnifiedTokenResponse(
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
		mdData, _ := rejectContent.Meta.(*metadata.BurningRequest)
		amount = mdData.BurningAmount
		tokenID = mdData.TokenID
		address = mdData.BurnerAddress
	default:
		return nil, nil
	}
	md := metadata.NewBuringResponseWithValue(inst.Status, txReqID)
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: &address, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			md.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return md
	}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}
