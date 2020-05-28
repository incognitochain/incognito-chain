package transaction

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"github.com/incognitochain/incognito-chain/wallet"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

type TxVersion2 struct{}

func (*TxVersion2) CheckAuthorizedSender(tx *Tx, publicKey []byte) (bool, error) {
	if !tx.Metadata.ShouldSignMetaData() {
		Logger.Log.Error("Check authorized sender failed because tx.Metadata is not appropriate")
		return false, errors.New("Check authorized sender failed because tx.Metadata is not appropriate")
	}
	meta, ok := tx.Metadata.(*metadata.StopAutoStakingMetadata)
	if !ok {
		Logger.Log.Error("Check authorized sender failed because tx.Metadata is not correct type")
		return false, errors.New("Check authorized sender failed because tx.Metadata is not correct type")
	}

	if meta.Sig == nil {
		Logger.Log.Error("CheckAuthorizedSender: should have sig for metadata to verify")
		return false, errors.New("CheckAuthorizedSender should have sig for metadata to verify")
	}
	/****** verify Schnorr signature *****/
	verifyKey := new(privacy.SchnorrPublicKey)
	metaSigPublicKey, err := new(privacy.Point).FromBytesS(publicKey)
	if err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(DecompressSigPubKeyError, err)
	}
	verifyKey.Set(metaSigPublicKey)

	signature := new(privacy.SchnSignature)
	if err := signature.SetBytes(meta.Sig); err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(InitTxSignatureFromBytesError, err)
	}
	return verifyKey.Verify(signature, tx.HashWithoutMetadataSig()[:]), nil
}

func txSigPubKeyToBytes(indexes [][]*big.Int) ([]byte, error) {
	n := len(indexes)
	if n == 0 {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is empty")
	}
	if n > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many rows")
	}
	m := len(indexes[0])
	if m > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many columns")
	}
	for i := 1; i < n; i += 1 {
		if len(indexes[i]) != m {
			return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is not a rectangle array")
		}
	}

	b := make([]byte, 0)
	b = append(b, byte(n))
	b = append(b, byte(m))
	for i := 0; i < n; i += 1 {
		for j := 0; j < m; j += 1 {
			currentByte := indexes[i][j].Bytes()
			lengthByte := len(currentByte)
			if lengthByte > MaxSizeByte {
				return nil, errors.New("TxSigPublicKeyVer2.ToBytes: IndexesByte is too large")
			}
			b = append(b, byte(lengthByte))
			b = append(b, currentByte...)
		}
	}
	return b, nil
}

func txSigPubKeyFromBytes(b []byte) ([][]*big.Int, error) {
	if len(b) < 2 {
		return nil, errors.New("txSigPubKeyFromBytes: cannot parse length of indexes, length of input byte is too small")
	}
	n := int(b[0])
	m := int(b[1])
	offset := 2
	indexes := make([][]*big.Int, n)
	for i := 0; i < n; i += 1 {
		row := make([]*big.Int, m)
		for j := 0; j < m; j += 1 {
			if offset >= len(b) {
				return nil, errors.New("txSigPubKeyFromBytes: cannot parse byte length of index[i][j], length of input byte is too small")
			}
			byteLength := int(b[offset])
			offset += 1
			if offset+byteLength > len(b) {
				return nil, errors.New("txSigPubKeyFromBytes: cannot parse big int index[i][j], length of input byte is too small")
			}
			currentByte := b[offset : offset+byteLength]
			offset += byteLength
			row[j] = new(big.Int).SetBytes(currentByte)
		}
		indexes[i] = row
	}

	return indexes, nil
}

func generateMlsagRingWithIndexes(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, params *TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, error) {
	//listUsableCommitments := make(map[common.Hash][]byte)
	//for _, in := range inputCoins {
	//	usableCommitment := in.GetCommitment().ToBytesS()
	//	commitmentInHash := common.HashH(usableCommitment)
	//	listUsableCommitments[commitmentInHash] = usableCommitment
	//}
	lenOTA, err := statedb.GetOTACoinLength(params.stateDB, *params.tokenID, shardID)
	if err != nil {
		Logger.Log.Errorf("Getting length of commitment error %v ", err)
		return nil, nil, err
	}
	if lenOTA == nil {
		Logger.Log.Error(errors.New("OnetimeAddress database is empty"))
		return nil, nil, errors.New("OnetimeAddress database is empty")
	}

	outputCommitments := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		outputCommitments.Add(outputCommitments, outputCoins[i].GetCommitment())
	}
	feeCommitment := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenValueIndex],
		new(operation.Scalar).FromUint64(params.fee),
	)

	// The indexes array is for validator recheck
	indexes := make([][]*big.Int, ringSize)
	ring := make([][]*operation.Point, ringSize)
	for i := 0; i < ringSize; i += 1 {
		sumInputs := new(operation.Point).Identity()
		sumInputs.Sub(sumInputs, feeCommitment)
		sumInputs.Sub(sumInputs, outputCommitments)

		row := make([]*operation.Point, len(inputCoins))
		rowIndexes := make([]*big.Int, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j += 1 {
				row[j] = inputCoins[j].GetPublicKey()
				publicKeyBytes := inputCoins[j].GetPublicKey().ToBytesS()
				if rowIndexes[j], err = statedb.GetOTACoinIndex(params.stateDB, *params.tokenID, publicKeyBytes); err != nil {
					Logger.Log.Errorf("Getting commitment index error %v ", err)
					return nil, nil, err
				}
				sumInputs.Add(sumInputs, inputCoins[j].GetCommitment())
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				rowIndexes[j], _ = common.RandBigIntMaxRange(lenOTA)
				fmt.Println("Length OTA is")
				fmt.Println("Length OTA is")
				fmt.Println(lenOTA)
				fmt.Println(lenOTA)
				fmt.Println(lenOTA)
				fmt.Println(lenOTA)
				if ok, err := statedb.HasOTACoinIndex(params.stateDB, *params.tokenID, rowIndexes[j].Uint64(), shardID); !ok || err != nil {
					Logger.Log.Errorf("Has commitment index error %v ", err)
					return nil, nil, err
				}
				coinBytes, err := statedb.GetOTACoinByIndex(params.stateDB, *params.tokenID, rowIndexes[j].Uint64(), shardID)
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
			}
		}
		row = append(row, sumInputs)
		ring[i] = row
		indexes[i] = rowIndexes
	}
	return mlsag.NewRing(ring), indexes, nil
}

func createPrivKeyMlsag(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, senderSK *key.PrivateKey) ([]*operation.Scalar, error) {
	sumRand := new(operation.Scalar).FromUint64(0)
	for _, in := range inputCoins {
		sumRand.Add(sumRand, in.GetRandomness())
	}
	for _, out := range outputCoins {
		sumRand.Sub(sumRand, out.GetRandomness())
	}

	privKeyMlsag := make([]*operation.Scalar, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		var err error
		privKeyMlsag[i], err = inputCoins[i].ParsePrivateKeyOfCoin(*senderSK)
		if err != nil {
			Logger.Log.Errorf("Cannot parse private key of coin %v", err)
			return nil, err
		}
	}
	privKeyMlsag[len(inputCoins)] = sumRand
	return privKeyMlsag, nil
}

func txSignatureToBytes(mlsagSignature *mlsag.MlsagSig) ([]byte, error) {
	var b []byte
	b = append(b, mlsag.MlsagPrefix)
	b = append(b, mlsagSignature.GetC().ToBytesS()...)

	r := mlsagSignature.GetR()
	n := len(r)
	if n == 0 {
		return nil, errors.New("Cannot parse byte of txSignature: length of r is 0")
	}
	if n > MaxSizeByte {
		return nil, errors.New("Cannot parse byte of txSignature: length row of r is larger than 256")
	}
	m := len(r[0])
	if m > MaxSizeByte {
		return nil, errors.New("Cannot parse byte of txSignature: length column of r is larger than 256")
	}
	for i := 1; i < n; i += 1 {
		if len(r[i]) != m {
			return nil, errors.New("Cannot parse byte of txSignature: r is not a proper rectangle")
		}
	}
	b = append(b, byte(n))
	b = append(b, byte(m))
	for i := 0; i < n; i += 1 {
		for j := 0; j < m; j += 1 {
			b = append(b, r[i][j].ToBytesS()...)
		}
	}
	return b, nil
}

func txSignatureFromBytesAndKeyImages(b []byte, keyImages []*operation.Point) (*mlsag.MlsagSig, error) {
	if len(b) == 0 {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: Length of byte is 0")
	}
	if b[0] != mlsag.MlsagPrefix {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: first byte is not mlsagPrefix")
	}
	offset := 1
	if offset+operation.Ed25519KeySize > len(b) {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: length of b too small to parse C")
	}
	c := new(operation.Scalar).FromBytesS(b[offset : offset+operation.Ed25519KeySize])
	offset += operation.Ed25519KeySize
	if offset+2 > len(b) {
		return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: length of b too small to parse n and m")
	}
	n := int(b[offset])
	m := int(b[offset+1])
	offset += 2
	r := make([][]*operation.Scalar, n)
	for i := 0; i < n; i += 1 {
		row := make([]*operation.Scalar, m)
		for j := 0; j < m; j += 1 {
			if offset+operation.Ed25519KeySize > len(b) {
				return nil, errors.New("Error in txSignatureFromBytesAndKeyImages: length of b too small to parse value of R")
			}
			currentBytes := b[offset : offset+operation.Ed25519KeySize]
			offset += operation.Ed25519KeySize
			row[j] = new(operation.Scalar).FromBytesS(currentBytes)
		}
		r[i] = row
	}
	return mlsag.NewMlsagSig(c, keyImages, r)
}

// signTx - signs tx
func signTxVer2(inp []coin.PlainCoin, out []*coin.CoinV2, tx *Tx, params *TxPrivacyInitParams) error {
	if tx.Sig != nil {
		return NewTransactionErr(UnexpectedError, errors.New("input transaction must be an unsigned one"))
	}

	ringSize := privacy.RingSize
	if !params.hasPrivacy {
		ringSize = 1
	}

	var pi int = common.RandIntInterval(0, ringSize-1)
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)

	ring, indexes, err := generateMlsagRingWithIndexes(inp, out, params, pi, shardID, ringSize)
	if err != nil {
		Logger.Log.Errorf("generateMlsagRingWithIndexes got error %v ", err)
		return err
	}

	privKeysMlsag, err := createPrivKeyMlsag(inp, out, params.senderSK)
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

	tx.SigPubKey, err = txSigPubKeyToBytes(indexes)
	if err != nil {
		Logger.Log.Errorf("tx.SigPubKey cannot parse from Bytes, error %v ", err)
		return err
	}

	message := tx.Hash()
	fmt.Println("Proving txHash =", tx.Hash())
	fmt.Println("Proving txHash =", tx.Hash())
	fmt.Println("Proving txHashWithoutSig =", tx.HashWithoutMetadataSig())
	fmt.Println("Proving txHashWithoutSig =", tx.HashWithoutMetadataSig())

	mlsagSignature, err := sag.Sign(message[:])
	if err != nil {
		Logger.Log.Errorf("Cannot sign mlsagSignature, error %v ", err)
		return err
	}

	tx.Sig, err = txSignatureToBytes(mlsagSignature)
	return err
}

func newCoinUniqueOTABasedOnPaymentInfo(paymentInfo *privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) (*coin.CoinV2, error) {
	for {
		fmt.Println("Generating new coin from payment info")
		c, err := coin.NewCoinFromPaymentInfo(paymentInfo)
		if err != nil {
			Logger.Log.Errorf("Cannot parse coin based on payment info err: %v", err)
			return nil, err
		}
		// If previously created coin is burning address
		if wallet.IsPublicKeyBurningAddress(c.GetPublicKey().ToBytesS()) {
			return c, nil // No need to check db
		}
		// Onetimeaddress should be unique
		publicKeyBytes := c.GetPublicKey().ToBytesS()
		found, err := statedb.HasOnetimeAddress(stateDB, *tokenID, publicKeyBytes)
		if err != nil {
			Logger.Log.Errorf("Cannot check public key existence in DB, err %v", err)
			return nil, err
		}
		if !found {
			return c, nil
		}
	}
}

func newCoinArrayBasedOnPaymentInfoArray(paymentInfo []*privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) ([]*coin.CoinV2, error) {
	outputCoins := make([]*coin.CoinV2, len(paymentInfo))
	for index, info := range paymentInfo {
		fmt.Println("Checking index", index)
		var err error
		outputCoins[index], err = newCoinUniqueOTABasedOnPaymentInfo(info, tokenID, stateDB)
		if err != nil {
			Logger.Log.Errorf("Cannot create coin with unique OTA, error: %v", err)
			return nil, err
		}
	}
	return outputCoins, nil
}

func verifyTxMetadata(tx *Tx) (bool, error) {
	// check input transaction
	meta := tx.GetMetadata().(*metadata.StopAutoStakingMetadata)
	if tx.Metadata == nil || meta.Sig == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}

	var err error
	res := false

	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verifyKey := new(privacy.SchnorrPublicKey)
	sigPublicKey, err := new(privacy.Point).FromBytesS(tx.SigPubKey)

	if err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(DecompressSigPubKeyError, err)
	}
	verifyKey.Set(sigPublicKey)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	err = signature.SetBytes(tx.Sig)
	if err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(InitTxSignatureFromBytesError, err)
	}

	// verify signature
	/*Logger.log.Debugf(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash()[:])
	if tx.Proof != nil {
		Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX Proof bytes before verifing the signature: %v\n", tx.Proof.Bytes())
	}
	Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX meta: %v\n", tx.Metadata)*/
	res = verifyKey.Verify(signature, tx.Hash()[:])

	return res, nil
}

func signTxMetadata(tx *Tx, privateKey *privacy.PrivateKey) error {
	// sign meta data
	meta, okType := tx.Metadata.(*metadata.StopAutoStakingMetadata)
	if !okType {
		return NewTransactionErr(UnexpectedError, errors.New("meta is not StopAutoStakingMetadata although ShouldSignMetaData() = true"))
	}
	if meta.Sig != nil {
		return NewTransactionErr(UnexpectedError, errors.New("meta.Sig should be empty"))
	}

	/****** using Schnorr signature *******/
	// sign with sigPrivKey
	// prepare private key for Schnorr
	sk := new(operation.Scalar).FromBytesS(*privateKey)
	r := new(operation.Scalar).FromUint64(0)
	sigKey := new(schnorr.SchnorrPrivateKey)
	sigKey.Set(sk, r)

	// signing
	signature, err := sigKey.Sign(tx.Hash()[:])
	if err != nil {
		return err
	}

	// convert signature to byte array
	meta.Sig = signature.Bytes()
	tx.Metadata = meta
	return nil
}

func (*TxVersion2) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	fmt.Println("About to prove")
	fmt.Println("About to prove")
	fmt.Println("About to prove")
	outputCoins, err := newCoinArrayBasedOnPaymentInfoArray(params.paymentInfo, params.tokenID, params.stateDB)
	if err != nil {
		Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
		return err
	}

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.inputCoins

	fmt.Println("About to prove proof")
	fmt.Println("About to prove proof")
	fmt.Println("About to prove proof")
	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, params.paymentInfo)
	if err != nil {
		Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	if tx.ShouldSignMetaData() {
		fmt.Println("It should sign meta data")
		fmt.Println("It should sign meta data")
		fmt.Println("It should sign meta data")
		fmt.Println("It should sign meta data")

		if err := signTxMetadata(tx, params.senderSK); err != nil {
			Logger.Log.Error("Cannot sign txMetadata when shouldSignMetadata")
			return err
		}
	}
	err = signTxVer2(inputCoins, outputCoins, tx, params)
	return err
}

func (txVer2 *TxVersion2) ProveASM(tx *Tx, params *TxPrivacyInitParamsForASM) error {
	return txVer2.Prove(tx, &params.txParam)
}

func getRingFromIndexesWithDatabase(tx *Tx, indexes [][]*big.Int, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (*mlsag.Ring, error) {
	var err error
	n := len(indexes)
	m := len(indexes[0])

	if n == 0 {
		return nil, errors.New("Cannot get ring from indexes: Indexes is empty")
	}

	// This is txver2 so outputCoin should be in txver2 format
	outputCoins := tx.Proof.GetOutputCoins()
	outputCommitments := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		commitment := outputCoins[i].GetCommitment()
		outputCommitments.Add(outputCommitments, commitment)
	}
	feeCommitment := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenValueIndex],
		new(operation.Scalar).FromUint64(tx.Fee),
	)

	ring := make([][]*operation.Point, n)
	for i := 0; i < n; i += 1 {
		sumCommitment := new(operation.Point).Identity()
		sumCommitment.Sub(sumCommitment, feeCommitment)

		row := make([]*operation.Point, m+1)
		for j := 0; j < m; j += 1 {
			index := indexes[i][j]
			ok, err := statedb.HasOTACoinIndex(transactionStateDB, *tokenID, index.Uint64(), shardID)
			if !ok || err != nil {
				Logger.Log.Errorf("HasOTACoinIndex error %v ", err)
				return nil, err
			}
			randomCoinBytes, err := statedb.GetOTACoinByIndex(transactionStateDB, *tokenID, index.Uint64(), shardID)
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
		}
		sumCommitment.Sub(sumCommitment, outputCommitments)
		byteCommitment := sumCommitment.ToBytesS()
		row[m], err = new(operation.Point).FromBytesS(byteCommitment)
		if err != nil {
			Logger.Log.Errorf("Getting last column commitment fromBytesS got error %v ", err)
			return nil, err
		}
		ring[i] = row
	}
	return mlsag.NewRing(ring), nil
}

// verifySigTx - verify signature on tx
func verifySigTxVer2(tx *Tx, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (bool, error) {
	// check input transaction
	fmt.Println("About to verify")
	fmt.Println("About to verify")
	fmt.Println("About to verify")
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}
	var err error

	indexes, err := txSigPubKeyFromBytes(tx.SigPubKey)
	if err != nil {
		Logger.Log.Errorf("Error when parsing bytes to txSigPubKey: %v ", err)
		return false, err
	}
	ring, err := getRingFromIndexesWithDatabase(tx, indexes, transactionStateDB, shardID, tokenID, isNewTransaction)
	if err != nil {
		Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	inputCoins := tx.Proof.GetInputCoins()
	keyImages := make([]*operation.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = operation.RandomPoint()

	txSig, err := txSignatureFromBytesAndKeyImages(tx.Sig, keyImages)
	if err != nil {
		return false, err
	}

	message := tx.Hash()
	fmt.Println("Verifying txHash =", tx.Hash())
	fmt.Println("Verifying txHash =", tx.Hash())
	fmt.Println("Verifying txHashWithoutSig =", tx.HashWithoutMetadataSig())
	fmt.Println("Verifying txHashWithoutSig =", tx.HashWithoutMetadataSig())
	return mlsag.Verify(txSig, ring, message[:])
}

func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var valid bool
	var err error

	tokenID, err = parseTokenID(tokenID)
	if err != nil {
		return false, err
	}

	if valid, err := verifySigTxVer2(tx, transactionStateDB, shardID, tokenID, isNewTransaction); !valid {
		if err != nil {
			Logger.Log.Errorf("Error verifying signature ver2 with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String())
		return false, NewTransactionErr(VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String()))
	}

	// Check input coins' commitment is exists in cm list (Database)
	inputCoins := tx.Proof.GetInputCoins()
	for i := 0; i < len(inputCoins); i++ {
		ok, err := tx.CheckCMExistence(inputCoins[i].GetCommitment().ToBytesS(), transactionStateDB, shardID, tokenID)
		if !ok || err != nil {
			if err != nil {
				Logger.Log.Error(err)
			}
			return false, NewTransactionErr(InputCommitmentIsNotExistedError, err)
		}
	}

	if tx.Proof == nil {
		return true, nil
	}

	// Verify the payment proof
	txProofV2 := tx.Proof.(*privacy.ProofV2)
	valid, err = txProofV2.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, shardID, tokenID, isBatch, nil)
	if !valid {
		if err != nil {
			Logger.Log.Error(err)
		}
		Logger.Log.Error("FAILED VERIFICATION PAYMENT PROOF VER 2")
		err1, ok := err.(*privacy.PrivacyError)
		if ok {
			// parse error detail
			if err1.Code == privacy.ErrCodeMessage[errhandler.VerifyOneOutOfManyProofFailedErr].Code {
				if isNewTransaction {
					return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
				} else {
					// for old txs which be get from sync block or validate new block
					if tx.LockTime <= ValidateTimeForOneoutOfManyProof {
						// only verify by sign on block because of issue #504(that mean we should pass old tx, which happen before this issue)
						return true, nil
					} else {
						return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
					}
				}
			}
		}
		return false, NewTransactionErr(TxProofVerifyFailError, err, tx.Hash().String())
	}
	Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil
}
