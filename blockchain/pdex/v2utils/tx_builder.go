package v2utils

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instructionPdexv3 "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

func TradeRefundTx(
	acn instructionPdexv3.Action,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadataCommon.Transaction, error) {
	if shardID != acn.ShardID() {
		return nil, nil
	}
	refundedTrade, ok := acn.Content.(*metadataPdexv3.RefundedTrade)
	if !ok {
		return nil, fmt.Errorf("Incorrect metadata type. Expected Refunded Trade")
	}
	md := metadataPdexv3.TradeResponse{acn.GetStatus(), acn.RequestTxID, metadataCommon.MetadataBase{acn.GetType()}}

	txParam := transaction.TxSalaryOutputParams{Amount: refundedTrade.Amount, ReceiverAddress: nil, PublicKey: &refundedTrade.Receiver.PublicKey, TxRandom: &refundedTrade.Receiver.TxRandom, TokenID: &refundedTrade.TokenToSell, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB,
		func(c privacy.Coin) metadataCommon.Metadata { return &md },
	)
}

func TradeAcceptTx(
	acn instructionPdexv3.Action,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadataCommon.Transaction, error) {
	if shardID != acn.ShardID() {
		return nil, nil
	}
	acceptedTrade, ok := acn.Content.(*metadataPdexv3.AcceptedTrade)
	if !ok {
		return nil, fmt.Errorf("Incorrect metadata type. Expected Accepted Trade")
	}
	md := metadataPdexv3.TradeResponse{acn.GetStatus(), acn.RequestTxID, metadataCommon.MetadataBase{acn.GetType()}}

	txParam := transaction.TxSalaryOutputParams{Amount: acceptedTrade.Amount, ReceiverAddress: nil, PublicKey: &acceptedTrade.Receiver.PublicKey, TxRandom: &acceptedTrade.Receiver.TxRandom, TokenID: &acceptedTrade.TokenToBuy, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB,
		func(c privacy.Coin) metadataCommon.Metadata { return &md },
	)
}

func OrderRefundTx(
	acn instructionPdexv3.Action,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadataCommon.Transaction, error) {
	if shardID != acn.ShardID() {
		return nil, nil
	}
	refundedOrder, ok := acn.Content.(*metadataPdexv3.RefundedAddOrder)
	if !ok {
		return nil, fmt.Errorf("Incorrect metadata type. Expected Refunded Order")
	}
	md := metadataPdexv3.AddOrderResponse{acn.GetStatus(), acn.RequestTxID, metadataCommon.MetadataBase{acn.GetType()}}

	txParam := transaction.TxSalaryOutputParams{Amount: refundedOrder.Amount, ReceiverAddress: nil, PublicKey: &refundedOrder.Receiver.PublicKey, TxRandom: &refundedOrder.Receiver.TxRandom, TokenID: &refundedOrder.TokenToSell, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB,
		func(c privacy.Coin) metadataCommon.Metadata { return &md },
	)
}

func MintPDEXGenesis(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadataCommon.Transaction, error) {
	if instStatus != metadataPdexv3.RequestAcceptedChainStatus {
		return nil, fmt.Errorf("Pdex v3 mint PDEX token genesis: Not support status %v", instStatus)
	}

	contentBytes := []byte(contentStr)
	var instContent metadataPdexv3.MintPDEXGenesisContent
	err := json.Unmarshal(contentBytes, &instContent)
	if err != nil {
		return nil, nil
	}

	if instContent.ShardID != shardID {
		return nil, nil
	}

	meta := metadataPdexv3.NewPdexv3MintPDEXGenesisResponse(
		metadataCommon.Pdexv3MintPDEXGenesisMeta,
		instContent.MintingPaymentAddress,
		instContent.MintingAmount,
	)

	keyWallet, err := wallet.Base58CheckDeserialize(instContent.MintingPaymentAddress)
	if err != nil {
		return nil, nil
	}
	// in case the returned currency is privacy custom token
	receiver := &privacy.PaymentInfo{
		Amount:         instContent.MintingAmount,
		PaymentAddress: keyWallet.KeySet.PaymentAddress,
	}

	tokenID := common.PDEXCoinID
	txParam := transaction.TxSalaryOutputParams{Amount: receiver.Amount, ReceiverAddress: &receiver.PaymentAddress, TokenID: &tokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			meta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return meta
	}
	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, makeMD)
}

func WithdrawLPFee(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadataCommon.Transaction, error) {
	contentBytes := []byte(contentStr)
	var instContent metadataPdexv3.WithdrawalLPFeeContent
	err := json.Unmarshal(contentBytes, &instContent)
	if err != nil {
		return nil, nil
	}

	if instContent.ShardID != shardID || instContent.Receiver.Amount == 0 {
		return nil, nil
	}

	meta := metadataPdexv3.NewPdexv3WithdrawalLPFeeResponse(
		metadataCommon.Pdexv3WithdrawLPFeeResponseMeta,
		instContent.TokenType,
		instContent.TxReqID,
	)

	receiverAddress := privacy.OTAReceiver{}
	err = receiverAddress.FromString(instContent.Receiver.AddressStr)
	if err != nil {
		return nil, nil
	}
	if !receiverAddress.IsValid() {
		return nil, nil
	}

	txParam := transaction.TxSalaryOutputParams{
		Amount:          instContent.Receiver.Amount,
		ReceiverAddress: nil,
		PublicKey:       &receiverAddress.PublicKey,
		TxRandom:        &receiverAddress.TxRandom,
		TokenID:         &instContent.Receiver.TokenID, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, func(c privacy.Coin) metadata.Metadata {
		return meta
	})
}

func WithdrawProtocolFee(
	instStatus string,
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (metadataCommon.Transaction, error) {
	contentBytes := []byte(contentStr)
	var instContent metadataPdexv3.WithdrawalProtocolFeeContent
	err := json.Unmarshal(contentBytes, &instContent)
	if err != nil {
		return nil, nil
	}

	if instContent.ShardID != shardID || instContent.Receiver.Amount == 0 {
		return nil, nil
	}

	meta := metadataPdexv3.NewPdexv3WithdrawalProtocolFeeResponse(
		metadataCommon.Pdexv3WithdrawProtocolFeeResponseMeta,
		instContent.TokenType,
		instContent.TxReqID,
	)

	receiverAddress := privacy.OTAReceiver{}
	err = receiverAddress.FromString(instContent.Receiver.AddressStr)
	if err != nil {
		return nil, nil
	}
	if !receiverAddress.IsValid() {
		return nil, nil
	}

	txParam := transaction.TxSalaryOutputParams{
		Amount:          instContent.Receiver.Amount,
		ReceiverAddress: nil,
		PublicKey:       &receiverAddress.PublicKey,
		TxRandom:        &receiverAddress.TxRandom,
		TokenID:         &instContent.Receiver.TokenID, Info: []byte{}}

	return txParam.BuildTxSalary(producerPrivateKey, transactionStateDB, func(c privacy.Coin) metadata.Metadata {
		return meta
	})
}
