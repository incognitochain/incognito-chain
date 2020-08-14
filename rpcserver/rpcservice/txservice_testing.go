package rpcservice

import (
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
)

func (txService TxService) TestBuildDoubleSpendingTransaction(params *bean.CreateRawTxParam, meta metadata.Metadata) ([]metadata.Transaction, *RPCError) {
	Logger.log.Infof("Build Raw Transaction Params: \n %+v", params)
	// get output coins to spend and real fee
	inputCoins, realFee, err1 := txService.chooseOutsCoinByKeyset(
		params.PaymentInfos, params.EstimateFeeCoinPerKb, 0,
		params.SenderKeySet, params.ShardIDSender, params.HasPrivacyCoin,
		meta, nil)
	if err1 != nil {
		return nil, err1
	}

	txPrivacyParams := transaction.NewTxPrivacyInitParams(
		&params.SenderKeySet.PrivateKey,
		params.PaymentInfos,
		inputCoins,
		realFee,
		params.HasPrivacyCoin,
		txService.BlockChain.GetBestStateShard(params.ShardIDSender).GetCopiedTransactionStateDB(),
		nil, // use for prv coin -> nil is valid
		meta,
		params.Info,
	)
	var result []metadata.Transaction
	tx, err := transaction.NewTransactionFromParams(txPrivacyParams)
	if err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}
	if err := tx.Init(txPrivacyParams); err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}
	result = append(result,tx)

	tx, err = transaction.NewTransactionFromParams(txPrivacyParams)
	if err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}
	if err := tx.Init(txPrivacyParams); err != nil {
		return nil, NewRPCError(CreateTxDataError, err)
	}
	result = append(result,tx)
	return result, nil
}