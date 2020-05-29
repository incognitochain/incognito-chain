package transaction

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

// TxSigPubKey of ver2 is array of indexes in database
type TxSigPubKeyVer2 struct {
	indexes [][]*big.Int
}

func (sigPub TxSigPubKeyVer2) Bytes() ([]byte, error) {
	n := len(sigPub.indexes)
	if n == 0 {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is empty")
	}
	if n > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many rows")
	}
	m := len(sigPub.indexes[0])
	if m > MaxSizeByte {
		return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is too large, too many columns")
	}
	for i := 1; i < n; i += 1 {
		if len(sigPub.indexes[i]) != m {
			return nil, errors.New("TxSigPublicKeyVer2.ToBytes: Indexes is not a rectangle array")
		}
	}

	b := make([]byte, 0)
	b = append(b, byte(n))
	b = append(b, byte(m))
	for i := 0; i < n; i += 1 {
		for j := 0; j < m; j += 1 {
			currentByte := sigPub.indexes[i][j].Bytes()
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

func (sigPub *TxSigPubKeyVer2) SetBytes(b []byte) error {
	if len(b) < 2 {
		return errors.New("txSigPubKeyFromBytes: cannot parse length of indexes, length of input byte is too small")
	}
	n := int(b[0])
	m := int(b[1])
	offset := 2
	indexes := make([][]*big.Int, n)
	for i := 0; i < n; i += 1 {
		row := make([]*big.Int, m)
		for j := 0; j < m; j += 1 {
			if offset >= len(b) {
				return errors.New("txSigPubKeyFromBytes: cannot parse byte length of index[i][j], length of input byte is too small")
			}
			byteLength := int(b[offset])
			offset += 1
			if offset+byteLength > len(b) {
				return errors.New("txSigPubKeyFromBytes: cannot parse big int index[i][j], length of input byte is too small")
			}
			currentByte := b[offset : offset+byteLength]
			offset += byteLength
			row[j] = new(big.Int).SetBytes(currentByte)
		}
		indexes[i] = row
	}
	if sigPub == nil {
		sigPub = new(TxSigPubKeyVer2)
	}
	sigPub.indexes = indexes
	return nil
}

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

func generateMlsagRingWithIndexes(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, params *TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, error) {
	lenOTA, err := statedb.GetOTACoinLength(params.stateDB, *params.tokenID, shardID)
	if err != nil || lenOTA == nil {
		Logger.Log.Errorf("Getting length of commitment error, either database length ota is empty or has error, error = %v", err)
		return nil, nil, err
	}
	sumOutputsWithFee := calculateSumOutputsWithFee(coin.CoinV2ArrayToCoinArray(outputCoins), params.fee)
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
				if rowIndexes[j], err = statedb.GetOTACoinIndex(params.stateDB, *params.tokenID, publicKeyBytes); err != nil {
					Logger.Log.Errorf("Getting commitment index error %v ", err)
					return nil, nil, err
				}
				sumInputs.Add(sumInputs, inputCoins[j].GetCommitment())
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				rowIndexes[j], _ = common.RandBigIntMaxRange(lenOTA)
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

func getMLSAGSignatureFromTx(tx *Tx) (*mlsag.MlsagSig, error) {
	inputCoins := tx.Proof.GetInputCoins()
	keyImages := make([]*operation.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = operation.RandomPoint()

	mlsagSig, err := new(mlsag.MlsagSig).FromBytes(tx.Sig)
	if err != nil {
		Logger.Log.Errorf("Has error when converting byte to mlsag signature, err: %v", err)
		return nil, err
	}

	return mlsag.NewMlsagSig(mlsagSig.GetC(), keyImages, mlsagSig.GetR())
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

	// Generate Ring
	var pi int = common.RandIntInterval(0, ringSize-1)
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	ring, indexes, err := generateMlsagRingWithIndexes(inp, out, params, pi, shardID, ringSize)
	if err != nil {
		Logger.Log.Errorf("generateMlsagRingWithIndexes got error %v ", err)
		return err
	}

	// Set SigPubKey
	txSigPubKey := new(TxSigPubKeyVer2)
	txSigPubKey.indexes = indexes
	tx.SigPubKey, err = txSigPubKey.Bytes()
	if err != nil {
		Logger.Log.Errorf("tx.SigPubKey cannot parse from Bytes, error %v ", err)
		return err
	}

	// Set sigPrivKey
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

	// Set Signature
	mlsagSignature, err := sag.Sign(tx.Hash()[:])
	if err != nil {
		Logger.Log.Errorf("Cannot sign mlsagSignature, error %v ", err)
		return err
	}
	// inputCoins already hold keyImage so set to nil to reduce size
	mlsagSignature.SetKeyImages(nil)
	tx.Sig, err = mlsagSignature.ToBytes()
	return err
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
	outputCoins, err := newCoinV2ArrayFromPaymentInfoArray(params.paymentInfo, params.tokenID, params.stateDB)
	if err != nil {
		Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
		return err
	}

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.inputCoins

	tx.Proof, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy, params.paymentInfo)
	if err != nil {
		Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	if tx.ShouldSignMetaData() {
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

func getRingFromTxWithDatabase(tx *Tx, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (*mlsag.Ring, error) {
	txSigPubKey := new(TxSigPubKeyVer2)
	if err := txSigPubKey.SetBytes(tx.SigPubKey); err != nil {
		errStr := fmt.Sprintf("Error when parsing bytes of txSigPubKey %v", err)
		return nil, NewTransactionErr(UnexpectedError, errors.New(errStr))
	}
	indexes := txSigPubKey.indexes

	n := len(indexes)
	m := len(indexes[0])
	if n == 0 {
		return nil, errors.New("Cannot get ring from indexes: Indexes is empty")
	}

	sumOutputsWithFee := calculateSumOutputsWithFee(tx.Proof.GetOutputCoins(), tx.Fee)
	ring := make([][]*operation.Point, n)
	for i := 0; i < n; i += 1 {
		sumCommitment := new(operation.Point).Identity()
		row := make([]*operation.Point, m+1)
		for j := 0; j < m; j += 1 {
			index := indexes[i][j]
			if ok, err := statedb.HasOTACoinIndex(transactionStateDB, *tokenID, index.Uint64(), shardID); !ok || err != nil {
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
		sumCommitment.Sub(sumCommitment, sumOutputsWithFee)
		byteCommitment := sumCommitment.ToBytesS()
		var err error
		if row[m], err = new(operation.Point).FromBytesS(byteCommitment); err != nil {
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
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be a signed one"))
	}
	var err error

	ring, err := getRingFromTxWithDatabase(tx, transactionStateDB, shardID, tokenID)
	if err != nil {
		Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}
	mlsagSignature, err := getMLSAGSignatureFromTx(tx)
	if err != nil {
		return false, err
	}

	return mlsag.Verify(mlsagSignature, ring, tx.Hash()[:])
}

func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var err error
	if tokenID, err = parseTokenID(tokenID); err != nil {
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
	if tx.Proof == nil {
		return true, nil
	}
	if valid, err := tx.Proof.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, shardID, tokenID, isBatch, nil); !valid {
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
