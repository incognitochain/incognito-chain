package transaction

import (
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

func parseLastByteSender(senderFullKey *incognitokey.KeySet) byte {
	return senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]
}

func parseSenderFullKey(params *TxPrivacyInitParams) (*incognitokey.KeySet, error) {
	senderFullKey := incognitokey.KeySet{}
	err := senderFullKey.InitFromPrivateKey(params.senderSK)
	if err != nil {
		Logger.log.Error(errors.New(fmt.Sprintf("Can not import Private key for sender keyset from %+v", params.senderSK)))
		return nil, NewTransactionErr(PrivateKeySenderInvalidError, err)
	}
	// get public key last byte of sender
	return &senderFullKey, nil
}

func parseCommitments(params *TxPrivacyInitParams, shardID byte) (*[]uint64, *[]uint64, error) {
	var commitmentIndexs []uint64   // array index random of commitments in db
	var myCommitmentIndexs []uint64 // index in array index random of commitment in db

	if params.hasPrivacy {
		randomParams := NewRandomCommitmentsProcessParam(params.inputCoins, privacy.CommitmentRingSize, params.db, shardID, params.tokenID)
		commitmentIndexs, myCommitmentIndexs, _ = RandomCommitmentsProcess(randomParams)

		// Check number of list of random commitments, list of random commitment indices
		if len(commitmentIndexs) != len(params.inputCoins)*privacy.CommitmentRingSize {
			return nil, nil, NewTransactionErr(RandomCommitmentError, nil)
		}

		if len(myCommitmentIndexs) != len(params.inputCoins) {
			return nil, nil, NewTransactionErr(RandomCommitmentError, errors.New("number of list my commitment indices must be equal to number of input coins"))
		}
	}
	return &commitmentIndexs, &myCommitmentIndexs, nil
}

func parseOverBalance(params *TxPrivacyInitParams) (int64, error) {
	// Calculate sum of all output coins' value
	sumOutputValue := uint64(0)
	for _, p := range params.paymentInfo {
		sumOutputValue += p.Amount
	}

	// Calculate sum of all input coins' value
	sumInputValue := uint64(0)
	for _, coin := range params.inputCoins {
		sumInputValue += coin.CoinDetails.GetValue()
	}
	Logger.log.Debugf("sumInputValue: %d\n", sumInputValue)

	overBalance := int64(sumInputValue - sumOutputValue - params.fee)
	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return 0, NewTransactionErr(WrongInputError, errors.New(fmt.Sprintf("input value less than output value. sumInputValue=%d sumOutputValue=%d fee=%d", sumInputValue, sumOutputValue, params.fee)))
	}

	return overBalance, nil
}

func parseSndOut(params *TxPrivacyInitParams) *[]*privacy.Scalar {
	ok := true
	sndOuts := make([]*privacy.Scalar, 0)
	for ok {
		for i := 0; i < len(params.paymentInfo); i++ {
			sndOut := privacy.RandomScalar()
			for {
				ok1, err := CheckSNDerivatorExistence(params.tokenID, sndOut, params.db)
				if err != nil {
					Logger.log.Error(err)
				}
				// if sndOut existed, then re-random it
				if ok1 {
					sndOut = privacy.RandomScalar()
				} else {
					break
				}
			}
			sndOuts = append(sndOuts, sndOut)
		}

		// if sndOuts has two elements that have same value, then re-generates it
		ok = privacy.CheckDuplicateScalarArray(sndOuts)
		if ok {
			sndOuts = make([]*privacy.Scalar, 0)
		}
	}
	return &sndOuts
}

func parseOutputCoins(params *TxPrivacyInitParams) (*[]*privacy.OutputCoin, error) {
	sndOuts := *parseSndOut(params)
	outputCoins := make([]*privacy.OutputCoin, len(params.paymentInfo))
	for i, pInfo := range params.paymentInfo {
		outputCoins[i] = new(privacy.OutputCoin)
		outputCoins[i].CoinDetails = new(privacy.Coin)
		outputCoins[i].CoinDetails.SetValue(pInfo.Amount)
		if len(pInfo.Message) > 0 {
			if len(pInfo.Message) > privacy.MaxSizeInfoCoin {
				return nil, NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
			}
		}
		outputCoins[i].CoinDetails.SetInfo(pInfo.Message)

		PK, err := new(privacy.Point).FromBytesS(pInfo.PaymentAddress.Pk)
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("can not decompress public key from %+v", pInfo.PaymentAddress)))
			return nil, NewTransactionErr(DecompressPaymentAddressError, err, pInfo.PaymentAddress)
		}
		outputCoins[i].CoinDetails.SetPublicKey(PK)
		outputCoins[i].CoinDetails.SetSNDerivator(sndOuts[i])
	}
	return &outputCoins, nil
}
