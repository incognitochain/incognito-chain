package transaction

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"github.com/incognitochain/incognito-chain/wallet"
)



func createPrivKeyMlsagCA(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, outputSharedSecrets []*operation.Point, senderSK *key.PrivateKey) ([]*operation.Scalar, error) {
	sumRand := new(operation.Scalar).FromUint64(0)
	for _, in := range inputCoins {
		sumRand.Add(sumRand, in.GetRandomness())
	}
	for _, out := range outputCoins {
		sumRand.Sub(sumRand, out.GetRandomness())
	}

	privKeyMlsag := make([]*operation.Scalar, len(inputCoins)+2)
	sumInputAssetTagBlinders := new(operation.Scalar).FromUint64(0)
	// TODO : change to lcm div
	numOfInputs := new(operation.Scalar).FromUint64(uint64(len(inputCoins)))
	numOfOutputs := new(operation.Scalar).FromUint64(uint64(len(outputCoins)))
	mySkBytes := (*senderSK)[:]
	for i := 0; i < len(inputCoins); i += 1 {
		var err error
		privKeyMlsag[i], err = inputCoins[i].ParsePrivateKeyOfCoin(*senderSK)
		if err != nil {
			Logger.Log.Errorf("Cannot parse private key of coin %v", err)
			return nil, err
		}

		inputCoin_specific, ok := inputCoins[i].(*coin.CoinV2)
		if !ok{
			return nil, errors.New("Cannot cast a coin as v2")
		}
		sharedSecret, err := inputCoin_specific.RecomputeSharedSecret(mySkBytes)
		if err != nil {
			Logger.Log.Errorf("Cannot recompute shared secret : %v", err)
			return nil, err
		}

		_, indexForShard, err := inputCoin_specific.GetTxRandomDetail()
		if err != nil {
			Logger.Log.Errorf("Cannot retrieve tx random detail : %v", err)
			return nil, err
		}
		bl, err := coin.ComputeAssetTagBlinder(sharedSecret, indexForShard)
		sumInputAssetTagBlinders.Add(sumInputAssetTagBlinders, bl)
	}
	sumInputAssetTagBlinders.Mul(sumInputAssetTagBlinders, numOfOutputs)

	sumOutputAssetTagBlinders := new(operation.Scalar).FromUint64(0)
	for i, oc := range outputCoins{
		_, indexForShard, err := oc.GetTxRandomDetail()
		if err != nil {
			Logger.Log.Errorf("Cannot retrieve tx random detail : %v", err)
			return nil, err
		}
		bl, err := coin.ComputeAssetTagBlinder(outputSharedSecrets[i], indexForShard)
		sumOutputAssetTagBlinders.Add(sumOutputAssetTagBlinders, bl)
	}
	sumOutputAssetTagBlinders.Mul(sumOutputAssetTagBlinders, numOfInputs)

	// 2 final elements in `private keys` for MLSAG
	privKeyMlsag[len(inputCoins)] 	= sumRand
	privKeyMlsag[len(inputCoins)+1] = new(operation.Scalar).Sub(sumInputAssetTagBlinders, sumOutputAssetTagBlinders)

	return privKeyMlsag, nil
}


func generateMlsagRingWithIndexesCA(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, params *TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, error) {
	// TODO : fork into newer function which adds a column for real/fake commitments to zero for CFA
	// then edit flow in MLSAG package to handle 2 final columns instead of 1
	
	lenOTA, err := txDatabaseWrapper.getOTACoinLength(params.stateDB, *params.tokenID, shardID)
	if err != nil || lenOTA == nil {
		Logger.Log.Errorf("Getting length of commitment error, either database length ota is empty or has error, error = %v", err)
		return nil, nil, err
	}
	sumOutputsWithFee := calculateSumOutputsWithFee(coin.CoinV2ArrayToCoinArray(outputCoins), params.fee)
	inCount := new(operation.Scalar).FromUint64(uint64(len(inputCoins)))
	outCount := new(operation.Scalar).FromUint64(uint64(len(outputCoins)))
	sumInputAssetTags := new(operation.Point).Identity()

	sumOutputAssetTags := new(operation.Point).Identity()
	for _, oc := range outputCoins{
		sumOutputAssetTags.Add(sumOutputAssetTags, oc.GetAssetTag())
	}
	sumOutputAssetTags.ScalarMult(sumOutputAssetTags, inCount)

	indexes := make([][]*big.Int, ringSize)
	ring := make([][]*operation.Point, ringSize)
	for i := 0; i < ringSize; i += 1 {
		sumInputs := new(operation.Point).Identity()
		sumInputs.Sub(sumInputs, sumOutputsWithFee)

		row := make([]*operation.Point, len(inputCoins))
		rowIndexes := make([]*big.Int, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j += 1 {
				row[j] = inputCoins[j].GetPublicKey()
				publicKeyBytes := inputCoins[j].GetPublicKey().ToBytesS()
				if rowIndexes[j], err = txDatabaseWrapper.getOTACoinIndex(params.stateDB, *params.tokenID, publicKeyBytes); err != nil {
					Logger.Log.Errorf("Getting commitment index error %v ", err)
					return nil, nil, err
				}
				sumInputs.Add(sumInputs, inputCoins[j].GetCommitment())
				inputCoin_specific, ok := inputCoins[j].(*coin.CoinV2)
				if !ok{
					return nil, nil, errors.New("Cannot cast a coin as v2")
				}
				sumInputAssetTags.Add(sumInputAssetTags, inputCoin_specific.GetAssetTag())
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				rowIndexes[j], _ = common.RandBigIntMaxRange(lenOTA)
				coinBytes, err := txDatabaseWrapper.getOTACoinByIndex(params.stateDB, *params.tokenID, rowIndexes[j].Uint64(), shardID)
				if err != nil {
					Logger.Log.Errorf("Get coinv2 by index error %v ", err)
					return nil, nil, err
				}
				coinDB := new(coin.CoinV2)
				if err := coinDB.SetBytes(coinBytes); err != nil {
					Logger.Log.Errorf("Cannot parse coinv2 byte error %v ", err)
					return nil, nil, err
				}
				row[j] = coinDB.GetPublicKey()
				sumInputs.Add(sumInputs, coinDB.GetCommitment())
				sumInputAssetTags.Add(sumInputAssetTags, coinDB.GetAssetTag())
			}
		}
		sumInputAssetTags.ScalarMult(sumInputAssetTags, outCount)

		row = append(row, sumInputs)
		row = append(row, sumInputAssetTags.Sub(sumInputAssetTags, sumOutputAssetTags))
		ring[i] = row
		indexes[i] = rowIndexes
	}
	return mlsag.NewRing(ring), indexes, nil
}

func (tx *TxVersion2) proveCA(params *TxPrivacyInitParams) error {
	var err error
	var outputCoins 	[]*coin.CoinV2
	var sharedSecrets 	[]*operation.Point
	for _,inf := range params.paymentInfo{
		c, ss, err := createUniqueOTACoinCA(inf, params.tokenID, params.stateDB)
		if err != nil {
			Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
			return err
		}
		outputCoins 	= append(outputCoins, c)
		sharedSecrets 	= append(sharedSecrets, ss)
	}
	// outputCoins, err := newCoinV2ArrayFromPaymentInfoArray(params.paymentInfo, params.tokenID, params.stateDB)

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.inputCoins

	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, params.paymentInfo)
	if err != nil {
		Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	if tx.ShouldSignMetaData() {
		if err := tx.signMetadata(params.senderSK); err != nil {
			Logger.Log.Error("Cannot signOnMessage txMetadata in shouldSignMetadata")
			return err
		}
	}
	err = tx.signCA(inputCoins, outputCoins, sharedSecrets, params, tx.Hash()[:])
	return err
}

func (tx *TxVersion2) signCA(inp []coin.PlainCoin, out []*coin.CoinV2, outputSharedSecrets []*operation.Point, params *TxPrivacyInitParams, hashedMessage []byte) error {
	if tx.Sig != nil {
		return NewTransactionErr(UnexpectedError, errors.New("input transaction must be an unsigned one"))
	}
	ringSize := privacy.RingSize
	if !params.hasPrivacy {
		ringSize = 1
	}

	// Generate Ring
	piBig,piErr := common.RandBigIntMaxRange(big.NewInt(int64(ringSize)))
	if piErr!=nil{
		return piErr
	}
	var pi int = int(piBig.Int64())
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	ring, indexes, err := generateMlsagRingWithIndexesCA(inp, out, params, pi, shardID, ringSize)
	if err != nil {
		Logger.Log.Errorf("generateMlsagRingWithIndexes got error %v ", err)
		return err
	}

	// Set SigPubKey
	txSigPubKey := new(TxSigPubKeyVer2)
	txSigPubKey.Indexes = indexes
	tx.SigPubKey, err = txSigPubKey.Bytes()
	if err != nil {
		Logger.Log.Errorf("tx.SigPubKey cannot parse from Bytes, error %v ", err)
		return err
	}

	// Set sigPrivKey
	privKeysMlsag, err := createPrivKeyMlsagCA(inp, out, outputSharedSecrets, params.senderSK)
	if err != nil {
		Logger.Log.Errorf("Cannot create private key of mlsag: %v", err)
		return err
	}
	sag := mlsag.NewMlsag(privKeysMlsag, ring, pi)
	tx.sigPrivKey, err = privacy.ArrayScalarToBytes(&privKeysMlsag)
	if err != nil {
		Logger.Log.Errorf("tx.SigPrivKey cannot parse arrayScalar to Bytes, error %v ", err)
		return err
	}

	// Set Signature
	mlsagSignature, err := sag.SignConfidentialAsset(hashedMessage)
	if err != nil {
		Logger.Log.Errorf("Cannot signOnMessage mlsagSignature, error %v ", err)
		return err
	}
	// inputCoins already hold keyImage so set to nil to reduce size
	mlsagSignature.SetKeyImages(nil)
	tx.Sig, err = mlsagSignature.ToBytes()

	return err
}

func createUniqueOTACoinCA(paymentInfo *privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) (*coin.CoinV2, *operation.Point, error) {
	for i:=coin.MAX_TRIES_OTA;i>0;i--{
		c, sharedSecret, err := coin.GenerateOTACoinAndSharedSecret(paymentInfo, tokenID)
		if err != nil {
			Logger.Log.Errorf("Cannot parse coin based on payment info err: %v", err)
			return nil, nil, err
		}
		// If previously created coin is burning address
		if wallet.IsPublicKeyBurningAddress(c.GetPublicKey().ToBytesS()) {
			return c, nil, nil // No need to check db
		}
		// Onetimeaddress should be unique
		publicKeyBytes := c.GetPublicKey().ToBytesS()
		// here tokenID should always be TokenConfidentialAssetID (for db storage)
		found, err := txDatabaseWrapper.hasOnetimeAddress(stateDB, common.ConfidentialAssetID, publicKeyBytes)
		if err != nil {
			Logger.Log.Errorf("Cannot check public key existence in DB, err %v", err)
			return nil, nil, err
		}
		if !found {
			return c, sharedSecret, nil
		}
	}
	// MAX_TRIES_OTA could be exceeded if the OS's RNG or the statedb is corrupted
	Logger.Log.Errorf("Cannot create unique OTA after %d attempts", coin.MAX_TRIES_OTA)
	return nil, nil, errors.New("Cannot create unique OTA")
}

