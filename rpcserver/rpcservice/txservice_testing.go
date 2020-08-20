package rpcservice

import (
	"errors"
	"fmt"
	// "os"

	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
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

	inputCoins, realFee, err1 = txService.chooseOutsCoinByKeyset(
		params.PaymentInfos, params.EstimateFeeCoinPerKb, 0,
		params.SenderKeySet, params.ShardIDSender, params.HasPrivacyCoin,
		meta, nil)
	if err1 != nil {
		return nil, err1
	}

	// try to send some back to self
	params.PaymentInfos[0].PaymentAddress = params.SenderKeySet.PaymentAddress
	txPrivacyParams = transaction.NewTxPrivacyInitParams(
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
		clonedCoin.SetCommitment(operation.RandomPoint())
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
	if tx.GetVersion()==1{
		transaction.TestResignTxV1(tx)
	}
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

func (txService TxService) TestBuildDoubleSpendingTokenTransaction(params interface{}, metaData metadata.Metadata) ([]transaction.TransactionToken, *RPCError) {
	txParam, errParam := bean.NewCreateRawPrivacyTokenTxParam(params)
	if errParam != nil {
		return nil, NewRPCError(RPCInvalidParamsError, errParam)
	}
	tokenParamsRaw := txParam.TokenParamsRaw
	tokenParams, err := txService.BuildTokenParam(tokenParamsRaw, txParam.SenderKeySet, txParam.ShardIDSender)
	if err != nil {
		return nil, err
	}
	// fmt.Fprintf(os.Stderr,"Wanted amount is %d\n",tokenParams.Receiver[0].Amount)
	tokenParamsCloned, err := txService.BuildTokenParam(tokenParamsRaw, txParam.SenderKeySet, txParam.ShardIDSender)
	if err != nil {
		return nil, err
	}

	if tokenParams == nil {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("can not build token params for request"))
	}
	/******* START choose output native coins(PRV), which is used to create tx *****/
	var inputCoins []coin.PlainCoin
	var mat [][]coin.PlainCoin
	var realFeeArr []uint64
	beaconView := txService.BlockChain.BeaconChain.GetFinalViewState()
	mat, realFeeArr, err = txService.chooseOutsCoinByKeysetTwice(txParam.PaymentInfos,
		txParam.EstimateFeeCoinPerKb, beaconView.BeaconHeight, txParam.SenderKeySet,
		txParam.ShardIDSender, txParam.HasPrivacyCoin, nil, tokenParams)
	inputCoins = mat[0]
	realFeePRV := realFeeArr[0]
	if err != nil {
		return nil, err
	}

	if len(txParam.PaymentInfos) == 0 && realFeePRV == 0 {
		txParam.HasPrivacyCoin = false
	}
	/******* END GET output coins native coins(PRV), which is used to create tx *****/


	txTokenParams := transaction.NewTxTokenParams(&txParam.SenderKeySet.PrivateKey,
		txParam.PaymentInfos,
		inputCoins,
		realFeePRV,
		tokenParams,
		txService.BlockChain.GetBestStateShard(txParam.ShardIDSender).GetCopiedTransactionStateDB(),
		metaData,
		txParam.HasPrivacyCoin,
		txParam.HasPrivacyToken,
		txParam.ShardIDSender, txParam.Info,
		beaconView.GetBeaconFeatureStateDB())

	tx, errTx := transaction.NewTransactionTokenFromParams(txTokenParams)
	if errTx != nil {
		Logger.log.Errorf("Cannot create new transaction token from params, err %v", err)
		return nil, NewRPCError(CreateTxDataError, errTx)
	}
	errTx = tx.Init(txTokenParams)
	if errTx != nil {
		return nil, NewRPCError(CreateTxDataError, errTx)
	}
	isValidTxItself, errItself := tx.ValidateTxByItself(tx.IsPrivacy(), txService.BlockChain.GetBestStateShard(txParam.ShardIDSender).GetCopiedTransactionStateDB(), beaconView.GetBeaconFeatureStateDB(), nil, txParam.ShardIDSender, false, nil, nil)
	// fmt.Fprintf(os.Stderr,"Valid by itself ? %v\n",isValidTxItself)
	if !isValidTxItself{
		// fmt.Fprintf(os.Stderr,"Error : %v\n",errItself)
	}
	// fmt.Fprintf(os.Stderr,"Height is %d\n",int64(beaconView.BeaconHeight))
	// _, _, errMp := txService.TxMemPool.MaybeAcceptTransaction(tx, int64(beaconView.BeaconHeight))
	// fmt.Fprintf(os.Stderr,"Through mempool : %v\n",errMp)


	inputCoins = mat[1]
	realFeePRV = realFeeArr[1]
	tokenParams = tokenParamsCloned
	tokenParams.Receiver[0].PaymentAddress = txParam.SenderKeySet.PaymentAddress
	txTokenParams = transaction.NewTxTokenParams(&txParam.SenderKeySet.PrivateKey,
		txParam.PaymentInfos,
		inputCoins,
		realFeePRV,
		tokenParams,
		txService.BlockChain.GetBestStateShard(txParam.ShardIDSender).GetCopiedTransactionStateDB(),
		metaData,
		txParam.HasPrivacyCoin,
		txParam.HasPrivacyToken,
		txParam.ShardIDSender, txParam.Info,
		beaconView.GetBeaconFeatureStateDB())

	tx2, errTx := transaction.NewTransactionTokenFromParams(txTokenParams)
	if errTx != nil {
		Logger.log.Errorf("Cannot create new transaction token from params, err %v", err)
		return nil, NewRPCError(CreateTxDataError, errTx)
	}
	errTx = tx2.Init(txTokenParams)
	if errTx != nil {
		return nil, NewRPCError(CreateTxDataError, errTx)
	}
	// outs := tx.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	// for _,c := range outs{
	// 	dc,err := c.Decrypt(txParam.SenderKeySet)
	// 	if err != nil{
	// 		continue
	// 	}
	// 	fmt.Fprintf(os.Stderr,"Token amount is %d\n",dc.GetValue())
	// }
	// outs = tx2.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()
	// for _,c := range outs{
	// 	dc,err := c.Decrypt(txParam.SenderKeySet)
	// 	if err != nil{
	// 		continue
	// 	}
	// 	fmt.Fprintf(os.Stderr,"Token amount is %d\n",dc.GetValue())
	// }

	return []transaction.TransactionToken{tx, tx2}, nil
}

func (txService TxService) chooseOutsCoinByKeysetTwice(
	paymentInfos []*privacy.PaymentInfo,
	unitFeeNativeToken int64, numBlock uint64, keySet *incognitokey.KeySet, shardIDSender byte,
	hasPrivacy bool,
	metadataParam metadata.Metadata,
	privacyCustomTokenParams *transaction.TokenParam,
	) ([][]coin.PlainCoin, []uint64, *RPCError) {
	// estimate fee according to 8 recent block
	if numBlock == 0 {
		numBlock = 1000
	}
	// calculate total amount to send
	totalAmmount := uint64(0)
	for _, receiver := range paymentInfos {
		totalAmmount += receiver.Amount
	}
	// get list outputcoins tx
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	plainCoins, err := txService.BlockChain.GetListDecryptedOutputCoinsByKeyset(keySet, shardIDSender, prvCoinID, uint64(0))
	if err != nil {
		return nil, nil, NewRPCError(GetOutputCoinError, err)
	}
	// remove out coin in mem pool
	plainCoins, err = txService.filterMemPoolOutcoinsToSpent(plainCoins)
	if err != nil {
		return nil, nil, NewRPCError(GetOutputCoinError, err)
	}
	if len(plainCoins) == 0 && totalAmmount > 0 {
		return nil, nil, NewRPCError(GetOutputCoinError, errors.New("not enough output coin"))
	}
	var realFeeArray []uint64
	var result [][]coin.PlainCoin
	for i:=0;i<2;i++{
		// Use Knapsack to get candiate output coin
		candidatePlainCoins, outCoins, candidateOutputCoinAmount, err := txService.chooseBestOutCoinsToSpent(plainCoins, totalAmmount)
		plainCoins = outCoins
		if err != nil {
			return nil, nil, NewRPCError(GetOutputCoinError, err)
		}
		// fmt.Fprintf(os.Stderr,"Already have %d in %d coins\n",candidateOutputCoinAmount, len(candidatePlainCoins))
		// refund out put for sender
		overBalanceAmount := candidateOutputCoinAmount - totalAmmount
		changedPaymentInfos := paymentInfos
		if overBalanceAmount > 0 {
			// add more into output for estimate fee
			changedPaymentInfos = append(changedPaymentInfos, &privacy.PaymentInfo{
				PaymentAddress: keySet.PaymentAddress,
				Amount:         overBalanceAmount,
			})
		}
		// check real fee(nano PRV) per tx
		beaconState, err := txService.BlockChain.GetClonedBeaconBestState()
		if err != nil {
			return nil, nil, NewRPCError(GetOutputCoinError, err)
		}
		beaconHeight := beaconState.BeaconHeight
		realFee, _, _, err := txService.EstimateFee(unitFeeNativeToken, false, candidatePlainCoins,
			changedPaymentInfos, shardIDSender, numBlock, hasPrivacy,
			metadataParam,
			privacyCustomTokenParams, int64(beaconHeight))
		if err != nil {
			return nil, nil, NewRPCError(RejectInvalidTxFeeError, err)
		}
		if totalAmmount == 0 && realFee == 0 {
			if metadataParam != nil {
				metadataType := metadataParam.GetType()
				switch metadataType {
				case metadata.WithDrawRewardRequestMeta:
					{
						return nil, nil, nil
					}
				}
				return nil, nil, NewRPCError(RejectInvalidTxFeeError, fmt.Errorf("totalAmmount: %+v, realFee: %+v", totalAmmount, realFee))
			}
			if privacyCustomTokenParams != nil {
				// for privacy token
				return nil, nil, nil
			}
		}
		needToPayFee := int64((totalAmmount + realFee) - candidateOutputCoinAmount)
		// if not enough to pay fee
		// fmt.Fprintf(os.Stderr,"Need to pay %d\n",needToPayFee)
		if needToPayFee > 0 {
			if len(outCoins) > 0 {
				candidateOutputCoinsForFee, remaining, _, err1 := txService.chooseBestOutCoinsToSpent(outCoins, uint64(needToPayFee))
				if err != nil {
					return nil, nil, NewRPCError(GetOutputCoinError, err1)
				}
				candidatePlainCoins = append(candidatePlainCoins, candidateOutputCoinsForFee...)
				plainCoins = remaining
			}
		}
		result = append(result,candidatePlainCoins)
		realFeeArray = append(realFeeArray,realFee)
	}
	return result, realFeeArray, nil
}