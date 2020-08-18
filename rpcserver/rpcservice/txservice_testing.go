package rpcservice

import (
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/privacy/coin"
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

func (txService TxService) TestBuildDuplicateInputTransaction(params *bean.CreateRawTxParam, meta metadata.Metadata) ([]metadata.Transaction, *RPCError) {
	Logger.log.Infof("Build Raw Transaction Params: \n %+v", params)
	// get output coins to spend and real fee
	inputCoins, realFee, err1 := txService.chooseOutsCoinByKeyset(
		params.PaymentInfos, params.EstimateFeeCoinPerKb, 0,
		params.SenderKeySet, params.ShardIDSender, params.HasPrivacyCoin,
		meta, nil)
	if err1 != nil {
		return nil, err1
	}

	var clonedCoin coin.PlainCoin
	if inputCoins[0].GetVersion()==1{
		clonedCoin = &coin.PlainCoinV1{}
		clonedCoin.SetBytes(inputCoins[0].Bytes())
	}else{
		clonedCoin = &coin.CoinV2{}
		clonedCoin.SetBytes(inputCoins[0].Bytes())
	}
	inputCoins = append(inputCoins,clonedCoin)

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
	return result, nil
}

func (txService TxService) TestBuildOutGtInTransaction(params *bean.CreateRawTxParam, meta metadata.Metadata) ([]metadata.Transaction, *RPCError) {
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
	realFee = tx.GetTxFee()
	tx.SetTxFee(realFee + 1000)
	result = append(result,tx)
	return result, nil
}

func (txService TxService) TestBuildReceiverExistsTransaction(params *bean.CreateRawTxParam, meta metadata.Metadata) ([]metadata.Transaction, *RPCError) {
	Logger.log.Infof("Build Raw Transaction Params: \n %+v", params)
	params.PaymentInfos[0].PaymentAddress = params.SenderKeySet.PaymentAddress
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
	switch txSpecific := tx.(type){
	case *transaction.TxVersion2:
		if err := txSpecific.InitTestOldOTA(txPrivacyParams); err != nil {
			return nil, NewRPCError(CreateTxDataError, err)
		}
	default:
		if err := txSpecific.Init(txPrivacyParams); err != nil {
			return nil, NewRPCError(CreateTxDataError, err)
		}
	}

	// realFee = tx.GetTxFee()
	// tx.SetTxFee(realFee + 1000)
	result = append(result,tx)
	return result, nil
}