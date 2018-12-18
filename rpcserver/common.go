package rpcserver

import (
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wallet"
	"errors"
	"github.com/ninjadotorg/constant/common/base58"
)

// buildRawCustomTokenTransaction ...
func (self RpcServer) buildRawCustomTokenTransaction(
	params interface{},
) (*transaction.TxCustomToken, error) {
	// all params
	arrayParams := common.InterfaceSlice(params)

	// param #1: private key of sender
	senderKeyParam := arrayParams[0]
	senderKey, err := wallet.Base58CheckDeserialize(senderKeyParam.(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	senderKey.KeySet.ImportFromPrivateKey(&senderKey.KeySet.PrivateKey)
	lastByte := senderKey.KeySet.PaymentAddress.Pk[len(senderKey.KeySet.PaymentAddress.Pk)-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	// param #2: estimation fee coin per kb
	estimateFeeCoinPerKb := int64(arrayParams[1].(float64))

	// param #3: estimation fee coin per kb by numblock
	numBlock := uint64(arrayParams[2].(float64))

	// param #4: token params
	tokenParamsRaw := arrayParams[3].(map[string]interface{})
	tokenParams := &transaction.CustomTokenParamTx{
		PropertyID:     tokenParamsRaw["TokenID"].(string),
		PropertyName:   tokenParamsRaw["TokenName"].(string),
		PropertySymbol: tokenParamsRaw["TokenSymbol"].(string),
		TokenTxType:    int(tokenParamsRaw["TokenTxType"].(float64)),
		Amount:         uint64(tokenParamsRaw["TokenAmount"].(float64)),
		Receiver:       transaction.CreateCustomTokenReceiverArray(tokenParamsRaw["TokenReceivers"]),
	}
	switch tokenParams.TokenTxType {
	case transaction.CustomTokenTransfer:
		{
			tokenID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			unspentTxTokenOuts, err := self.config.BlockChain.GetUnspentTxCustomTokenVout(senderKey.KeySet, tokenID)
			Logger.log.Info("buildRawCustomTokenTransaction ", unspentTxTokenOuts)
			if err != nil {
				return nil, NewRPCError(ErrUnexpected, err)
			}
			if len(unspentTxTokenOuts) == 0 {
				return nil, NewRPCError(ErrUnexpected, errors.New("Balance of token is zero"))
			}
			txTokenIns := []transaction.TxTokenVin{}
			txTokenInsAmount := uint64(0)
			for _, out := range unspentTxTokenOuts {
				item := transaction.TxTokenVin{
					PaymentAddress:  out.PaymentAddress,
					TxCustomTokenID: out.GetTxCustomTokenID(),
					VoutIndex:       out.GetIndex(),
				}
				// create signature by keyset -> base58check.encode of txtokenout double hash
				signature, err := senderKey.KeySet.Sign(out.Hash()[:])
				if err != nil {
					return nil, NewRPCError(ErrUnexpected, err)
				}
				// add signature to TxTokenVin to use token utxo
				item.Signature = base58.Base58Check{}.Encode(signature, 0)
				txTokenIns = append(txTokenIns, item)
				txTokenInsAmount += out.Value
			}
			tokenParams.SetVins(txTokenIns)
			tokenParams.SetVinsAmount(txTokenInsAmount)
		}
	case transaction.CustomTokenInit:
		{
			if tokenParams.Receiver[0].Value != tokenParams.Amount { // Init with wrong max amount of custom token
				return nil, NewRPCError(ErrUnexpected, errors.New("Init with wrong max amount of property"))
			}
		}
	}

	totalAmmount := estimateFeeCoinPerKb

	// list unspent tx for estimation fee
	estimateTotalAmount := totalAmmount
	outCoins, _ := self.config.BlockChain.GetListOutputCoinsByKeyset(&senderKey.KeySet, chainIdSender)
	candidateOutputCoins := make([]*privacy.OutputCoin, 0)
	for _, note := range outCoins {
		amount := note.CoinDetails.Value
		candidateOutputCoins = append(candidateOutputCoins, note)
		estimateTotalAmount -= int64(amount)
		if estimateTotalAmount <= 0 {
			break
		}
	}

	// check real fee per TxNormal
	var realFee uint64
	if int64(estimateFeeCoinPerKb) == -1 {
		temp, _ := self.config.FeeEstimator[chainIdSender].EstimateFee(numBlock)
		estimateFeeCoinPerKb = int64(temp)
	}
	estimateFeeCoinPerKb += int64(self.config.Wallet.Config.IncrementalFee)
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateOutputCoins, nil)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)

	// list unspent tx for create tx
	totalAmmount += int64(realFee)
	estimateTotalAmount = totalAmmount
	candidateOutputCoins = make([]*privacy.OutputCoin, 0)
	if totalAmmount > 0 {
		for _, note := range outCoins {
			amount := note.CoinDetails.Value
			candidateOutputCoins = append(candidateOutputCoins, note)
			estimateTotalAmount -= int64(amount)
			if estimateTotalAmount <= 0 {
				break
			}
		}
	}

	// get list custom token
	listCustomTokens, err := self.config.BlockChain.ListCustomToken()

	inputCoins := transaction.ConvertOutputCoinToInputCoin(candidateOutputCoins)
	tx := &transaction.TxCustomToken{}
	err = tx.Init(
		&senderKey.KeySet.PrivateKey,
		nil,
		inputCoins,
		realFee,
		tokenParams,
		listCustomTokens,
	)

	return tx, err
}

func (self RpcServer) EstimateFee(defaultFee int64, candidateOutputCoins []*privacy.OutputCoin, paymentInfos []*privacy.PaymentInfo, chainID byte, numBlock uint64) uint64 {
	// check real fee(nano constant) per tx
	var realFee uint64
	estimateFeeCoinPerKb := uint64(0)
	if defaultFee == -1 {
		temp, _ := self.config.FeeEstimator[chainID].EstimateFee(numBlock)
		estimateFeeCoinPerKb = uint64(temp)
	}
	if estimateFeeCoinPerKb == 0 {
		estimateFeeCoinPerKb = self.config.BlockChain.GetFeePerKbTx()
	}
	estimateFeeCoinPerKb += uint64(self.config.Wallet.Config.IncrementalFee)
	estimateTxSizeInKb := transaction.EstimateTxSize(candidateOutputCoins, nil)
	realFee = uint64(estimateFeeCoinPerKb) * uint64(estimateTxSizeInKb)
	return realFee
}
