package transaction

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"github.com/incognitochain/incognito-chain/wallet"
	"math/big"
	"strconv"
	"time"

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

type TxVersion2 struct {
	TxBase
}

// ========== GET FUNCTION ===========

func (tx TxVersion2) GetTxMintData() (bool, []byte, []byte, uint64, *common.Hash, error) {
	publicKeys, txRandoms, amounts, err := tx.GetReceiverData()
	if err != nil {
		Logger.Log.Error("GetTxMintData: Cannot get receiver data")
		return false, nil, nil, 0, nil, err
	}
	if len(publicKeys) != 1 {
		Logger.Log.Error("GetTxMintData : Should only have one receiver")
		return false, nil, nil, 0, nil, errors.New("Error Tx mint has more than one receiver")
	}
	if inputCoins := tx.Proof.GetInputCoins(); len(inputCoins) > 0 {
		return false, nil, nil, 0, nil, errors.New("Error this is not Tx mint")
	}
	if txRandoms == nil {
		return true, publicKeys[0].ToBytesS(), nil, amounts[0], &common.PRVCoinID, nil
	} else {
		return true, publicKeys[0].ToBytesS(), txRandoms[0].Bytes(), amounts[0], &common.PRVCoinID, nil
	}
}

func (tx *TxVersion2) GetReceiverData() ([]*privacy.Point, []*coin.TxRandom, []uint64, error) {
	publicKeys := make([]*privacy.Point, 0)
	txRandoms := make([]*coin.TxRandom, 0)
	amounts := []uint64{}

	if tx.Proof != nil && len(tx.Proof.GetOutputCoins()) > 0 {
		outputCoins := tx.Proof.GetOutputCoins()
		for i:= 0; i < len(outputCoins); i++ {
			coin := outputCoins[i].(*coin.CoinV2)
			publicKey := coin.GetPublicKey()
			txRandom := coin.GetTxRandom()
			added := false
			for j :=0; j < len(publicKeys); j ++ {
				if bytes.Equal(publicKey.ToBytesS(), publicKeys[j].ToBytesS()) {
					amounts[j] += coin.GetValue()
					added = true
					break
				}
			}
			if !added {
				publicKeys = append(publicKeys, publicKey)
				amounts = append(amounts, coin.GetValue())
				txRandoms = append(txRandoms, txRandom)
			}
		}
	}
	return publicKeys, txRandoms, amounts, nil
}

func (tx TxVersion2) GetTxBurnData(retriever metadata.ChainRetriever, blockHeight uint64) (bool, []byte, uint64, *common.Hash, error) {
	pubkeys, _, amounts, err := tx.GetReceiverData()
	if err != nil {
		Logger.Log.Errorf("Cannot get receiver data, error %v", err)
		return false, nil, 0, nil, err
	}
	if len(pubkeys) > 2 {
		Logger.Log.Error("GetAndCheckBurning receiver: More than 2 receivers")
		return false, nil, 0, nil, err
	}

	burnAccount, err := wallet.Base58CheckDeserialize(retriever.GetBurningAddress(blockHeight))
	if err != nil {
		return false, nil, 0, nil, err
	}
	burnPaymentAddress := burnAccount.KeySet.PaymentAddress

	isBurned, pubkey, amount := false, []byte{}, uint64(0)
	for i, pk := range pubkeys {
		if bytes.Equal(burnPaymentAddress.Pk, pk.ToBytesS()) {
			isBurned = true
			pubkey = pk.ToBytesS()
			amount += amounts[i]
		}
	}
	return isBurned, pubkey, amount, &common.PRVCoinID, nil
}

// ========== CHECK FUNCTION ===========

func (tx *TxVersion2) CheckAuthorizedSender(publicKey []byte) (bool, error) {
	if !tx.Metadata.ShouldSignMetaData() {
		Logger.Log.Error("Check authorized sender failed because tx.Metadata is not appropriate")
		return false, errors.New("Check authorized sender failed because tx.Metadata is not appropriate")
	}
	//meta, ok := tx.Metadata.(*metadata.StopAutoStakingMetadata)
	//if !ok {
	//	Logger.Log.Error("Check authorized sender failed because tx.Metadata is not correct type")
	//	return false, errors.New("Check authorized sender failed because tx.Metadata is not correct type")
	//}
	metaSig := tx.Metadata.GetSig()
	fmt.Println("Metadata Sig", metaSig)
	if metaSig == nil || len(metaSig) == 0 {
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
	if err := signature.SetBytes(metaSig); err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(InitTxSignatureFromBytesError, err)
	}
	return verifyKey.Verify(signature, tx.HashWithoutMetadataSig()[:]), nil
}

// ========== NORMAL INIT FUNCTIONS ==========

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

func (tx *TxVersion2) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxPrivacyInitParams)
	if !ok {
		return errors.New("params of tx Init is not TxPrivacyInitParam")
	}

	Logger.Log.Debugf("CREATING TX........\n")
	if err := validateTxParams(params); err != nil {
		return err
	}

	// Init tx and params (tx and params will be changed)
	if err := tx.initializeTxAndParams(params); err != nil {
		return err
	}

	// Check if this tx is nonPrivacyNonInput
	// Case 1: tx ptoken transfer with ptoken fee
	// Case 2: tx Reward
	// If it is non privacy non input then return
	if check, err := tx.isNonPrivacyNonInput(params); check {
		return err
	}

	if err := tx.prove(params); err != nil {
		return err
	}
	return nil
}

func (tx *TxVersion2) signOnMessage(inp []coin.PlainCoin, out []*coin.CoinV2, params *TxPrivacyInitParams, hashedMessage []byte) error {
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
	mlsagSignature, err := sag.Sign(hashedMessage)
	if err != nil {
		Logger.Log.Errorf("Cannot signOnMessage mlsagSignature, error %v ", err)
		return err
	}
	// inputCoins already hold keyImage so set to nil to reduce size
	mlsagSignature.SetKeyImages(nil)
	tx.Sig, err = mlsagSignature.ToBytes()

	return err
}

func (tx *TxVersion2) signMetadata(privateKey *privacy.PrivateKey) error {
	// signOnMessage meta data
	metaSig := tx.Metadata.GetSig()
	if metaSig != nil && len(metaSig) > 0 {
		return NewTransactionErr(UnexpectedError, errors.New("meta.Sig should be empty or nil"))
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
	fmt.Println("Set Signature ")
	fmt.Println("Set Signature ")
	fmt.Println("Set Signature ")
	tx.Metadata.SetSig(signature.Bytes())
	fmt.Println("Signature Detail", tx.Metadata.GetSig())
	return nil
}

func (tx *TxVersion2) prove(params *TxPrivacyInitParams) error {
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
		if err := tx.signMetadata(params.senderSK); err != nil {
			Logger.Log.Error("Cannot signOnMessage txMetadata in shouldSignMetadata")
			return err
		}
	}
	err = tx.signOnMessage(inputCoins, outputCoins, params, tx.Hash()[:])
	return err
}

func (tx *TxVersion2) proveASM(params *TxPrivacyInitParamsForASM) error {
	return tx.prove(&params.txParam)
}

// Retrieve ring from database using sigpubkey and last column commitment (last column = sumOutputCoinCommitment + fee)
func getRingFromSigPubKeyAndLastColumnCommitment(sigPubKey []byte, sumOutputsWithFee *operation.Point, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (*mlsag.Ring, error) {
	txSigPubKey := new(TxSigPubKeyVer2)
	if err := txSigPubKey.SetBytes(sigPubKey); err != nil {
		errStr := fmt.Sprintf("Error when parsing bytes of txSigPubKey %v", err)
		return nil, NewTransactionErr(UnexpectedError, errors.New(errStr))
	}
	indexes := txSigPubKey.indexes

	n := len(indexes)
	m := len(indexes[0])
	if n == 0 {
		return nil, errors.New("Cannot get ring from indexes: Indexes is empty")
	}

	ring := make([][]*operation.Point, n)
	for i := 0; i < n; i += 1 {
		sumCommitment := new(operation.Point).Identity()
		sumCommitment.Sub(sumCommitment, sumOutputsWithFee)
		row := make([]*operation.Point, m+1)
		for j := 0; j < m; j += 1 {
			index := indexes[i][j]
			if ok, err := txDatabaseWrapper.hasOTACoinIndex(transactionStateDB, *tokenID, index.Uint64(), shardID); !ok || err != nil {
				Logger.Log.Errorf("HasOTACoinIndex error %v ", err)
				return nil, err
			}
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
		}
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

// ========== NORMAL VERIFY FUNCTIONS ==========

func generateMlsagRingWithIndexes(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, params *TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, error) {
	lenOTA, err := txDatabaseWrapper.getOTACoinLength(params.stateDB, *params.tokenID, shardID)
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
				if rowIndexes[j], err = txDatabaseWrapper.getOTACoinIndex(params.stateDB, *params.tokenID, publicKeyBytes); err != nil {
					Logger.Log.Errorf("Getting commitment index error %v ", err)
					return nil, nil, err
				}
				sumInputs.Add(sumInputs, inputCoins[j].GetCommitment())
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				rowIndexes[j], _ = common.RandBigIntMaxRange(lenOTA)
				if ok, err := txDatabaseWrapper.hasOTACoinIndex(params.stateDB, *params.tokenID, rowIndexes[j].Uint64(), shardID); !ok || err != nil {
					Logger.Log.Errorf("Has commitment index error %v ", err)
					return nil, nil, err
				}
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
			}
		}
		row = append(row, sumInputs)
		ring[i] = row
		indexes[i] = rowIndexes
	}
	return mlsag.NewRing(ring), indexes, nil
}

func getMLSAGSigFromTxSigAndKeyImages(txSig []byte, keyImages []*operation.Point) (*mlsag.MlsagSig, error) {
	mlsagSig, err := new(mlsag.MlsagSig).FromBytes(txSig)
	if err != nil {
		Logger.Log.Errorf("Has error when converting byte to mlsag signature, err: %v", err)
		return nil, err
	}

	return mlsag.NewMlsagSig(mlsagSig.GetC(), keyImages, mlsagSig.GetR())
}

func (tx *TxVersion2) verifySig(transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be a signed one"))
	}
	var err error

	// Reform Ring
	sumOutputsWithFee := calculateSumOutputsWithFee(tx.Proof.GetOutputCoins(), tx.Fee)
	ring, err := getRingFromSigPubKeyAndLastColumnCommitment(tx.SigPubKey, sumOutputsWithFee, transactionStateDB, shardID, tokenID)
	if err != nil {
		Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	// Reform MLSAG Signature
	inputCoins := tx.Proof.GetInputCoins()
	keyImages := make([]*operation.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i += 1 {
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = operation.RandomPoint()
	mlsagSignature, err := getMLSAGSigFromTxSigAndKeyImages(tx.Sig, keyImages)
	if err != nil {
		return false, err
	}

	return mlsag.Verify(mlsagSignature, ring, tx.Hash()[:])
}

func (tx *TxVersion2) Verify(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var err error
	if tokenID, err = parseTokenID(tokenID); err != nil {
		return false, err
	}
	if valid, err := tx.verifySig(transactionStateDB, shardID, tokenID, isNewTransaction); !valid {
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
						// only verify by signOnMessage on block because of issue #504(that mean we should pass old tx, which happen before this issue)
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

func (tx TxVersion2) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []metadata.Transaction, txsUsed []int, insts [][]string, instsUsed []int, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if tx.IsPrivacy() {
		return true, nil
	}
	proof := tx.GetProof()
	meta := tx.GetMetadata()

	inputCoins := make([]coin.PlainCoin, 0)
	outputCoins := make([]coin.Coin, 0)
	if tx.GetProof() != nil {
		inputCoins = tx.GetProof().GetInputCoins()
		outputCoins = tx.GetProof().GetOutputCoins()
	}
	if proof != nil && len(inputCoins) == 0 && len(outputCoins) > 0 { // coinbase tx
		if meta == nil {
			return false, nil
		}
		if !meta.IsMinerCreatedMetaType() {
			return false, nil
		}
	}
	if meta != nil {
		ok, err := meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, &tx, bcr, accumulatedValues, nil, nil)
		if err != nil {
			Logger.Log.Error(err)
			return false, NewTransactionErr(VerifyMinerCreatedTxBeforeGettingInBlockError, err)
		}
		return ok, nil
	}
	return true, nil
}

func (tx TxVersion2) ValidateTxByItself(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	ok, err := tx.ValidateTransaction(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, prvCoinID, false, isNewTransaction)
	if !ok {
		return false, err
	}
	meta := tx.GetMetadata()
	if meta != nil {
		if hasPrivacy {
			return false, errors.New("Metadata can not exist in not privacy tx")
		}
		validateMetadata := meta.ValidateMetadataByItself()
		if validateMetadata {
			return validateMetadata, nil
		} else {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
	}
	return true, nil
}

// ========== SALARY FUNCTIONS: INIT AND VALIDATE  ==========

func (tx *TxVersion2) InitTxSalary(otaCoin *coin.CoinV2, privateKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata) error {
	tx.Version = TxVersion2Number
	tx.Type = common.TxRewardType
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	tempOutputCoin := make([]coin.Coin, 1)
	tempOutputCoin[0] = otaCoin

	proof := new(privacy.ProofV2)
	proof.Init()
	proof.SetOutputCoins(tempOutputCoin)
	tx.Proof = proof

	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	tx.PubKeyLastByteSender = publicKeyBytes[len(publicKeyBytes) - 1]

	// signOnMessage Tx using ver1 schnorr
	tx.sigPrivKey = *privateKey
	tx.SetMetadata(metaData)

	var err error
	if tx.Sig, tx.SigPubKey, err = signNoPrivacy(privateKey, tx.Hash()[:]); err != nil {
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func (tx *TxVersion2) ValidateTxSalary(db *statedb.StateDB) (bool, error) {
	// verify signature
	if valid, err := verifySigNoPrivacy(tx.Sig, tx.SigPubKey, tx.Hash()[:]); !valid {
		if err != nil {
			Logger.Log.Debugf("Error verifying signature of tx: %+v", err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		return false, nil
	}
	// check whether output coin's input exists in input list or not
	tokenID := &common.Hash{}
	if err := tokenID.SetBytes(common.PRVCoinID[:]); err != nil {
		return false, NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
	}

	// Check commitment
	outputCoins := tx.Proof.GetOutputCoins()
	if len(outputCoins) != 1 {
		return false, NewTransactionErr(UnexpectedError, errors.New("length outputCoins of proof is not 1"))
	}
	outputCoin := outputCoins[0].(*coin.CoinV2)
	cmpCommitment := operation.PedCom.CommitAtIndex(outputCoin.GetAmount(), outputCoin.GetRandomness(), operation.PedersenValueIndex)
	if !operation.IsPointEqual(cmpCommitment, outputCoin.GetCommitment()) {
		return false, NewTransactionErr(UnexpectedError, errors.New("check output coin's coin commitment isn't calculated correctly"))
	}

	// Check shardID
	coinShardID, errShard := outputCoin.GetShardID()
	if errShard != nil {
		errStr := fmt.Sprintf("error when getting coin shardID, err: %v", errShard)
		return false, NewTransactionErr(UnexpectedError, errors.New(errStr))
	}
	if coinShardID != tx.PubKeyLastByteSender {
		return false, NewTransactionErr(UnexpectedError, errors.New("output coin's shardID is different from tx pubkey last byte"))
	}

	// Check database for ota
	found, err := txDatabaseWrapper.hasOnetimeAddress(db, *tokenID, outputCoin.GetPublicKey().ToBytesS())
	if err != nil {
		Logger.Log.Errorf("Cannot check public key existence in DB, err %v", err)
		return false, err
	}
	if found {
		Logger.Log.Error("ValidateTxSalary got error: found onetimeaddress in database")
		return false, errors.New("found onetimeaddress in database")
	}
	return true, nil
}

func (tx TxVersion2) StringWithoutMetadataSig() string {
	record := strconv.Itoa(int(tx.Version))
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		record += base64.StdEncoding.EncodeToString(tx.Proof.Bytes())
	}
	if tx.Metadata != nil {
		metadataHash := tx.Metadata.HashWithoutSig()
		record += metadataHash.String()
	}
	record += string(tx.Info)
	return record
}

func (tx *TxVersion2) HashWithoutMetadataSig() *common.Hash {
	inBytes := []byte(tx.StringWithoutMetadataSig())
	hash := common.HashH(inBytes)
	//tx.cachedHash = &hash
	return &hash
}

// ========== VALIDATE FUNCTIONS ============
func (tx TxVersion2) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	meta := tx.GetMetadata()
	Logger.Log.Debugf("\n\n\n START Validating sanity data of metadata %+v\n\n\n", meta)
	if meta != nil {
		Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, &tx)
		Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	Logger.Log.Debugf("\n\n\n END sanity data of metadata%+v\n\n\n")
	//check version
	if tx.GetVersion() > TxVersion2Number {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version <= %d", tx.GetVersion(), currentTxVersion))
	}
	// check LockTime before now
	if int64(tx.GetLockTime()) > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.GetLockTime()))
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	// check sanity of Proof
	if tx.GetProof() != nil {
		ok, err := tx.GetProof().ValidateSanity()
		if !ok || err != nil {
			s := ""
			if !ok {
				s = "ValidateSanity Proof got error: ok = false\n"
			} else {
				s = fmt.Sprintf("ValidateSanity Proof got error: error %s\n", err.Error())
			}
			return false, errors.New(s)
		}
	}

	// check Type is normal or salary tx
	switch tx.GetType() {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.GetType()))
	}

	//if txN.Type != common.TxNormalType && txN.Type != common.TxRewardType && txN.Type != common.TxCustomTokenType && txN.Type != common.TxCustomTokenPrivacyType { // only 1 byte
	//	return false, errors.New("wrong tx type")
	//}

	// check info field
	info := tx.GetInfo()
	if len(info) > 512 {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(info), 512))
	}

	return true, nil
}

func (tx TxVersion2) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	if tx.GetType() == common.TxRewardType || tx.GetType() == common.TxReturnStakingType {
		return nil
	}
	meta := tx.GetMetadata()
	if meta != nil {
		isContinued, err := meta.ValidateTxWithBlockChain(&tx, chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)
		fmt.Printf("[transactionStateDB] validate metadata with blockchain: %d %h %t %v\n", tx.GetMetadataType(), tx.Hash(), isContinued, err)
		if err != nil {
			Logger.Log.Errorf("[db] validate metadata with blockchain: %d %s %t %v\n", tx.GetMetadataType(), tx.Hash().String(), isContinued, err)
			return NewTransactionErr(RejectTxMedataWithBlockChain, fmt.Errorf("validate metadata of tx %s with blockchain error %+v", tx.Hash().String(), err))
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
}

func (tx TxVersion2) ValidateTransaction(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	if tx.GetType() == common.TxRewardType {
		return tx.ValidateTxSalary(transactionStateDB)
	}
	if tx.GetType() == common.TxReturnStakingType {
		return tx.ValidateTxReturnStaking(transactionStateDB), nil
	}
	if tx.Version == TxConversionVersion12Number {
		return validateConversionVer1ToVer2(&tx, transactionStateDB, shardID, tokenID)
	}
	return tx.Verify(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}