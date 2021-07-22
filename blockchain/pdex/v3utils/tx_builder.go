package v3utils

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instructionPdexv3 "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
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
