package transaction

import (
	"fmt"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	// "github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"github.com/incognitochain/incognito-chain/wallet"
)


// for multi-CA : take in an array of tokenIDs
func createPrivKeyMlsagCA(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, outputSharedSecrets []*operation.Point, params *TxPrivacyInitParams, shardID byte) ([]*operation.Scalar, error) {
	senderSK := params.senderSK
	// db := params.stateDB
	tokenID := params.tokenID
	sumRand := new(operation.Scalar).FromUint64(0)
	// for _, in := range inputCoins {
	// 	sumRand.Add(sumRand, in.GetRandomness())
	// }
	// for _, out := range outputCoins {
	// 	sumRand.Sub(sumRand, out.GetRandomness())
	// }

	privKeyMlsag := make([]*operation.Scalar, len(inputCoins)+2)
	sumInputAssetTagBlinders := new(operation.Scalar).FromUint64(0)
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
		rehashed := operation.HashToPoint(tokenID[:])
		isUnblinded := operation.IsPointEqual(rehashed, inputCoin_specific.GetAssetTag())
		if isUnblinded{
			Logger.Log.Infof("Signing TX : processing an unblinded input coin")
		}

		_, indexForShard, err := inputCoin_specific.GetTxRandomDetail()
		if err != nil {
			Logger.Log.Errorf("Cannot retrieve tx random detail : %v", err)
			return nil, err
		}
		bl := new(operation.Scalar).FromUint64(0)
		if !isUnblinded{
			bl, err = coin.ComputeAssetTagBlinder(sharedSecret, indexForShard)
			if err != nil {
				return nil, err
			}
		}
		
		Logger.Log.Infof("CA-MLSAG : processing input asset tag %s\n", string(inputCoin_specific.GetAssetTag().MarshalText()))
		Logger.Log.Debugf("Shared secret is %s\n", string(sharedSecret.MarshalText()))
		Logger.Log.Debugf("Blinder is %s\n", string(bl.MarshalText()))
		v := inputCoin_specific.GetAmount()
		Logger.Log.Debugf("Value is %d\n",v.ToUint64Little())
		effectiveRCom := new(operation.Scalar).Mul(bl,v)
		effectiveRCom.Add(effectiveRCom, inputCoin_specific.GetRandomness())

		sumInputAssetTagBlinders.Add(sumInputAssetTagBlinders, bl)
		sumRand.Add(sumRand, effectiveRCom)
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
		Logger.Log.Infof("CA-MLSAG : processing output asset tag %s\n", string(oc.GetAssetTag().MarshalText()))
		Logger.Log.Debugf("Shared secret is %s\n", string(outputSharedSecrets[i].MarshalText()))
		Logger.Log.Debugf("Blinder is %s\n", string(bl.MarshalText()))
		
		v := oc.GetAmount()
		Logger.Log.Debugf("Value is %d\n",v.ToUint64Little())
		effectiveRCom := new(operation.Scalar).Mul(bl,v)
		effectiveRCom.Add(effectiveRCom, oc.GetRandomness())
		sumOutputAssetTagBlinders.Add(sumOutputAssetTagBlinders, bl)
		sumRand.Sub(sumRand, effectiveRCom)
	}
	sumOutputAssetTagBlinders.Mul(sumOutputAssetTagBlinders, numOfInputs)

	// 2 final elements in `private keys` for MLSAG
	assetSum := new(operation.Scalar).Sub(sumInputAssetTagBlinders, sumOutputAssetTagBlinders)
	temp1 := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenRandomnessIndex], assetSum)
	temp2 := new(operation.Point).ScalarMult(operation.PedCom.G[operation.PedersenRandomnessIndex], sumRand)

	Logger.Log.Debugf("Last 2 private keys will correspond to points %s and %s\n", temp1.MarshalText(), temp2.MarshalText())

	privKeyMlsag[len(inputCoins)] 	= assetSum
	privKeyMlsag[len(inputCoins)+1]	= sumRand
	return privKeyMlsag, nil
}

func generateMlsagRingWithIndexesCA(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, params *TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, error) {
	// TODO : fork into newer function which adds a column for real/fake commitments to zero for CFA
	// then edit flow in MLSAG package to handle 2 final columns instead of 1
	
	lenOTA, err := txDatabaseWrapper.getOTACoinLength(params.stateDB, common.ConfidentialAssetID, shardID)
	if err != nil || lenOTA == nil {
		Logger.Log.Errorf("Getting length of commitment error, either database length ota is empty or has error, error = %v", err)
		return nil, nil, err
	}
	sumOutputsWithFee := calculateSumOutputsWithFee(coin.CoinV2ArrayToCoinArray(outputCoins), params.fee)
	inCount := new(operation.Scalar).FromUint64(uint64(len(inputCoins)))
	outCount := new(operation.Scalar).FromUint64(uint64(len(outputCoins)))

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
		sumInputAssetTags := new(operation.Point).Identity()

		row := make([]*operation.Point, len(inputCoins))
		rowIndexes := make([]*big.Int, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j += 1 {
				row[j] = inputCoins[j].GetPublicKey()
				publicKeyBytes := inputCoins[j].GetPublicKey().ToBytesS()
				if rowIndexes[j], err = txDatabaseWrapper.getOTACoinIndex(params.stateDB, common.ConfidentialAssetID, publicKeyBytes); err != nil {
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
				coinBytes, err := txDatabaseWrapper.getOTACoinByIndex(params.stateDB, common.ConfidentialAssetID, rowIndexes[j].Uint64(), shardID)
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

		assetSum := new(operation.Point).Sub(sumInputAssetTags, sumOutputAssetTags)
		row = append(row, assetSum)
		row = append(row, sumInputs)
		if i==pi{
			Logger.Log.Debugf("Last 2 columns in ring are %s and %s\n", assetSum.MarshalText(), sumInputs.MarshalText())
		}
		
		ring[i] = row
		indexes[i] = rowIndexes
	}
	return mlsag.NewRing(ring), indexes, nil
}

func (tx *TxVersion2) proveCA(params *TxPrivacyInitParams) error {
	var err error
	var outputCoins 	[]*coin.CoinV2
	var sharedSecrets 	[]*operation.Point
	// fmt.Printf("tokenID is %v\n",params.tokenID)
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

	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, sharedSecrets, params.hasPrivacy, params.paymentInfo)
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
	privKeysMlsag, err := createPrivKeyMlsagCA(inp, out, outputSharedSecrets, params, shardID)
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

func reconstructRingCA(sigPubKey []byte, sumOutputsWithFee , sumOutputAssetTags *operation.Point, numOfOutputs *operation.Scalar, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (*mlsag.Ring, error) {
	txSigPubKey := new(TxSigPubKeyVer2)
	if err := txSigPubKey.SetBytes(sigPubKey); err != nil {
		errStr := fmt.Sprintf("Error when parsing bytes of txSigPubKey %v", err)
		return nil, NewTransactionErr(UnexpectedError, errors.New(errStr))
	}
	indexes := txSigPubKey.Indexes
	n := len(indexes)
	if n == 0 {
		return nil, errors.New("Cannot get ring from Indexes: Indexes is empty")
	}

	m := len(indexes[0])

	ring := make([][]*operation.Point, n)
	for i := 0; i < n; i += 1 {
		sumCommitment := new(operation.Point).Identity()
		sumCommitment.Sub(sumCommitment, sumOutputsWithFee)
		sumAssetTags := new(operation.Point).Identity()
		sumAssetTags.Sub(sumAssetTags, sumOutputAssetTags)
		row := make([]*operation.Point, m+2)
		for j := 0; j < m; j += 1 {
			index := indexes[i][j]
			randomCoinBytes, err := txDatabaseWrapper.getOTACoinByIndex(transactionStateDB, *tokenID, index.Uint64(), shardID)
			if err != nil {
				Logger.Log.Errorf("Get random onetimeaddresscoin error %v ", err)
				return nil, err
			}
			randomCoin := new(coin.CoinV2)
			if err := randomCoin.SetBytes(randomCoinBytes); err != nil {
				Logger.Log.Errorf("Set coin Byte error %v ", err)
				return nil, err
			}
			row[j] = randomCoin.GetPublicKey()
			sumCommitment.Add(sumCommitment, randomCoin.GetCommitment())
			temp := new(operation.Point).ScalarMult(randomCoin.GetAssetTag(), numOfOutputs)
			sumAssetTags.Add(sumAssetTags, temp)
		}

		row[m] 	 = new(operation.Point).Set(sumAssetTags)
		row[m+1] = new(operation.Point).Set(sumCommitment)
		ring[i] = row
	}
	return mlsag.NewRing(ring), nil
}

func (tx *TxVersion2) verifySigCA(transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be a signed one"))
	}
	var err error

	// Reform Ring
	sumOutputsWithFee := calculateSumOutputsWithFee(tx.Proof.GetOutputCoins(), tx.Fee)
	sumOutputAssetTags := new(operation.Point).Identity()
	for _, oc := range tx.Proof.GetOutputCoins(){
		output_specific, ok := oc.(*coin.CoinV2)
		if !ok{
			Logger.Log.Errorf("Error when casting coin as v2")
			return false, errors.New("Error when casting coin as v2")
		}
		sumOutputAssetTags.Add(sumOutputAssetTags, output_specific.GetAssetTag())
	}
	inCount := new(operation.Scalar).FromUint64(uint64(len(tx.GetProof().GetInputCoins())))
	outCount := new(operation.Scalar).FromUint64(uint64(len(tx.GetProof().GetOutputCoins())))
	sumOutputAssetTags.ScalarMult(sumOutputAssetTags, inCount)

	// fmt.Printf("Token id is %v\n",tokenID)
	ring, err := reconstructRingCA(tx.SigPubKey, sumOutputsWithFee, sumOutputAssetTags, outCount, transactionStateDB, shardID, tokenID)
	if err != nil {
		Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	// Reform MLSAG Signature
	inputCoins := tx.Proof.GetInputCoins()
	keyImages := make([]*operation.Point, len(inputCoins)+2)
	for i := 0; i < len(inputCoins); i += 1 {
		if inputCoins[i].GetKeyImage()==nil {
			Logger.Log.Errorf("Error when reconstructing mlsagSignature: missing keyImage")
			return false, err
		}
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = operation.RandomPoint()
	keyImages[len(inputCoins)+1] = operation.RandomPoint()
	mlsagSignature, err := getMLSAGSigFromTxSigAndKeyImages(tx.Sig, keyImages)
	if err != nil {
		return false, err
	}

	return mlsag.VerifyConfidentialAsset(mlsagSignature, ring, tx.Hash()[:])
}

func createUniqueOTACoinCA(paymentInfo *privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) (*coin.CoinV2, *operation.Point, error) {
	for i:=coin.MAX_TRIES_OTA;i>0;i--{
		c, sharedSecret, err := coin.GenerateOTACoinAndSharedSecret(paymentInfo, tokenID)
		Logger.Log.Infof("Created a new coin with tokenID %s, shared secret %s, asset tag %s\n", tokenID.String(), sharedSecret.MarshalText(), c.GetAssetTag().MarshalText())
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

