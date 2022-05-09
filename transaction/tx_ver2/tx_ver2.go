package tx_ver2

import (
	"encoding/json"
	"fmt"
	"sort"

	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"

	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

// SigPubKey defines the public key to sign ring signatures in version 2. It is an array of coin indexes.
type SigPubKey struct {
	Indexes [][]*big.Int
}

// Tx struct for version 2 contains the same fields as before
type Tx struct {
	tx_generic.TxBase
}

func (sigPub SigPubKey) Bytes() ([]byte, error) {
	n := len(sigPub.Indexes)
	if n == 0 {
		return nil, fmt.Errorf("txSigPublicKeyVer2.ToBytes: Indexes is empty")
	}
	if n > utils.MaxSizeByte {
		return nil, fmt.Errorf("txSigPublicKeyVer2.ToBytes: Indexes is too large, too many rows")
	}
	m := len(sigPub.Indexes[0])
	if m > utils.MaxSizeByte {
		return nil, fmt.Errorf("txSigPublicKeyVer2.ToBytes: Indexes is too large, too many columns")
	}
	for i := 1; i < n; i++ {
		if len(sigPub.Indexes[i]) != m {
			return nil, fmt.Errorf("txSigPublicKeyVer2.ToBytes: Indexes is not a rectangle array")
		}
	}

	b := make([]byte, 0)
	b = append(b, byte(n))
	b = append(b, byte(m))
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			currentByte := sigPub.Indexes[i][j].Bytes()
			lengthByte := len(currentByte)
			if lengthByte > utils.MaxSizeByte {
				return nil, fmt.Errorf("txSigPublicKeyVer2.ToBytes: IndexesByte is too large")
			}
			b = append(b, byte(lengthByte))
			b = append(b, currentByte...)
		}
	}
	return b, nil
}

func (sigPub *SigPubKey) SetBytes(b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("txSigPubKeyFromBytes: cannot parse length of Indexes, length of input byte is too small")
	}
	n := int(b[0])
	m := int(b[1])
	offset := 2
	indexes := make([][]*big.Int, n)
	for i := 0; i < n; i++ {
		row := make([]*big.Int, m)
		for j := 0; j < m; j++ {
			if offset >= len(b) {
				return fmt.Errorf("txSigPubKeyFromBytes: cannot parse byte length of index[i][j], length of input byte is too small")
			}
			byteLength := int(b[offset])
			offset++
			if offset+byteLength > len(b) {
				return fmt.Errorf("txSigPubKeyFromBytes: cannot parse big int index[i][j], length of input byte is too small")
			}
			currentByte := b[offset : offset+byteLength]
			offset += byteLength
			row[j] = new(big.Int).SetBytes(currentByte)
		}
		indexes[i] = row
	}
	if sigPub == nil {
		sigPub = new(SigPubKey)
	}
	sigPub.Indexes = indexes
	return nil
}

// ========== GET FUNCTION ===========

// GetReceiverData returns output coins for this function
func (tx *Tx) GetReceiverData() ([]privacy.Coin, error) {
	if tx.Proof != nil && len(tx.Proof.GetOutputCoins()) > 0 {
		return tx.Proof.GetOutputCoins(), nil
	}
	return nil, nil
}

// ========== NORMAL INIT FUNCTIONS ==========

func createPrivKeyMlsag(inputCoins []privacy.PlainCoin, outputCoins []*privacy.CoinV2, senderSK *privacy.PrivateKey, commitmentToZero *privacy.Point) ([]*privacy.Scalar, error) {
	sumRand := new(privacy.Scalar).FromUint64(0)
	for _, in := range inputCoins {
		sumRand.Add(sumRand, in.GetRandomness())
	}
	for _, out := range outputCoins {
		sumRand.Sub(sumRand, out.GetRandomness())
	}

	privKeyMlsag := make([]*privacy.Scalar, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i++ {
		var err error
		privKeyMlsag[i], err = inputCoins[i].ParsePrivateKeyOfCoin(*senderSK)
		if err != nil {
			utils.Logger.Log.Errorf("Cannot parse private key of coin %v", err)
			return nil, err
		}
	}
	commitmentToZeroRecomputed := new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenRandomnessIndex], sumRand)
	match := privacy.IsPointEqual(commitmentToZeroRecomputed, commitmentToZero)
	if !match {
		return nil, utils.NewTransactionErr(utils.SignTxError, fmt.Errorf("error : asset tag sum or commitment sum mismatch"))
	}
	privKeyMlsag[len(inputCoins)] = sumRand
	return privKeyMlsag, nil
}

// Init uses the information in parameter to create a valid, signed Tx.
func (tx *Tx) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*tx_generic.TxPrivacyInitParams)
	if !ok {
		return fmt.Errorf("params of tx Init is not TxPrivacyInitParam")
	}

	jsb, _ := json.Marshal(params)
	utils.Logger.Log.Infof("Create TX v2 with params %s", string(jsb))
	if err := tx_generic.ValidateTxParams(params); err != nil {
		return err
	}

	// Init tx and params (tx and params will be changed)
	if err := tx.InitializeTxAndParams(params); err != nil {
		return err
	}

	// Check if this tx is nonPrivacyNonInput
	// Case 1: tx ptoken transfer with ptoken fee
	// Case 2: tx Reward
	// If it is non privacy non input then return
	if check, err := tx.IsNonPrivacyNonInput(params); check {
		return err
	}
	if err := tx.prove(params); err != nil {
		return err
	}
	jsb, _ = json.Marshal(tx)
	utils.Logger.Log.Infof("TX Creation complete ! The resulting transaction is: %v, %s", tx.Hash().String(), string(jsb))
	txSize := tx.GetTxActualSize()
	if txSize > common.MaxTxSize {
		return utils.NewTransactionErr(utils.ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	return nil
}

func (tx *Tx) signOnMessage(inp []privacy.PlainCoin, out []*privacy.CoinV2, params *tx_generic.TxPrivacyInitParams, hashedMessage []byte) error {
	if tx.Sig != nil {
		return utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("input transaction must be an unsigned one"))
	}
	ringSize := privacy.RingSize

	// Generate Ring
	piBig, piErr := common.RandBigIntMaxRange(big.NewInt(int64(ringSize)))
	if piErr != nil {
		return piErr
	}
	var pi int = int(piBig.Int64())
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	ring, indexes, commitmentToZero, err := generateMlsagRingWithIndexes(inp, out, params, pi, shardID, ringSize)
	if err != nil {
		utils.Logger.Log.Errorf("generateMlsagRingWithIndexes got error %v ", err)
		return err
	}

	// Set SigPubKey
	txSigPubKey := new(SigPubKey)
	txSigPubKey.Indexes = indexes
	tx.SigPubKey, err = txSigPubKey.Bytes()
	if err != nil {
		utils.Logger.Log.Errorf("tx.SigPubKey cannot parse from Bytes, error %v ", err)
		return err
	}

	// Set sigPrivKey
	privKeysMlsag, err := createPrivKeyMlsag(inp, out, params.SenderSK, commitmentToZero)
	if err != nil {
		utils.Logger.Log.Errorf("Cannot create private key of mlsag: %v", err)
		return err
	}
	sag := mlsag.NewMlsag(privKeysMlsag, ring, pi)
	sk, err := privacy.ArrayScalarToBytes(&privKeysMlsag)
	if err != nil {
		utils.Logger.Log.Errorf("tx.SigPrivKey cannot parse arrayScalar to Bytes, error %v ", err)
		return err
	}
	tx.SetPrivateKey(sk)

	// Set Signature
	mlsagSignature, err := sag.Sign(hashedMessage)
	if err != nil {
		utils.Logger.Log.Errorf("Cannot signOnMessage mlsagSignature, error %v ", err)
		return err
	}
	// inputCoins already hold keyImage so set to nil to reduce size
	mlsagSignature.SetKeyImages(nil)
	tx.Sig, err = mlsagSignature.ToBytes()

	return err
}

func (tx *Tx) prove(params *tx_generic.TxPrivacyInitParams) error {
	var senderKeySet incognitokey.KeySet
	_ = senderKeySet.InitFromPrivateKey(params.SenderSK)
	b := senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]
	outputCoins, err := utils.NewCoinV2ArrayFromPaymentInfoArray(params.PaymentInfo, int(common.GetShardIDFromLastByte(b)), params.TokenID, params.StateDB)
	if err != nil {
		utils.Logger.Log.Errorf("Cannot parse outputCoinV2 to outputCoins, error %v ", err)
		return err
	}

	// inputCoins is plainCoin because it may have coinV1 with coinV2
	inputCoins := params.InputCoins

	tx.Proof, err = privacy.ProveV2(inputCoins, outputCoins, nil, false, params.PaymentInfo)
	if err != nil {
		utils.Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}

	if tx.GetMetadata() != nil {
		if err := tx.GetMetadata().Sign(params.SenderSK, tx); err != nil {
			utils.Logger.Log.Error("Cannot signOnMessage txMetadata in shouldSignMetadata")
			return err
		}
	}

	err = tx.signOnMessage(inputCoins, outputCoins, params, tx.Hash()[:])
	return err
}

// ========== NORMAL VERIFY FUNCTIONS ==========

func generateMlsagRingWithIndexes(inputCoins []privacy.PlainCoin, outputCoins []*privacy.CoinV2, params *tx_generic.TxPrivacyInitParams, pi int, shardID byte, ringSize int) (*mlsag.Ring, [][]*big.Int, *privacy.Point, error) {
	lenOTA, err := statedb.GetOTACoinLength(params.StateDB, *params.TokenID, shardID)
	if err != nil || lenOTA == nil {
		utils.Logger.Log.Errorf("Getting length of commitment error, either database length ota is empty or has error, error = %v", err)
		return nil, nil, nil, err
	}
	outputCoinsAsGeneric := make([]privacy.Coin, len(outputCoins))
	for i := 0; i < len(outputCoins); i++ {
		outputCoinsAsGeneric[i] = outputCoins[i]
	}
	sumOutputsWithFee := tx_generic.CalculateSumOutputsWithFee(outputCoinsAsGeneric, params.Fee)
	indexes := make([][]*big.Int, ringSize)
	ring := make([][]*privacy.Point, ringSize)
	var commitmentToZero *privacy.Point
	attempts := 0
	for i := 0; i < ringSize; i++ {
		sumInputs := new(privacy.Point).Identity()
		sumInputs.Sub(sumInputs, sumOutputsWithFee)

		row := make([]*privacy.Point, len(inputCoins))
		rowIndexes := make([]*big.Int, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j++ {
				row[j] = inputCoins[j].GetPublicKey()
				publicKeyBytes := inputCoins[j].GetPublicKey().ToBytesS()
				if rowIndexes[j], err = statedb.GetOTACoinIndex(params.StateDB, *params.TokenID, publicKeyBytes); err != nil {
					utils.Logger.Log.Errorf("Getting commitment index error %v ", err)
					return nil, nil, nil, err
				}
				sumInputs.Add(sumInputs, inputCoins[j].GetCommitment())
			}
		} else {
			for j := 0; j < len(inputCoins); j++ {
				coinDB := new(privacy.CoinV2)
				for attempts < privacy.MaxPrivacyAttempts { // The chance of infinite loop is negligible
					rowIndexes[j], _ = common.RandBigIntMaxRange(lenOTA)
					coinBytes, err := statedb.GetOTACoinByIndex(params.StateDB, *params.TokenID, rowIndexes[j].Uint64(), shardID)
					if err != nil {
						utils.Logger.Log.Errorf("Get coinv2 by index error %v ", err)
						return nil, nil, nil, err
					}

					if err = coinDB.SetBytes(coinBytes); err != nil {
						utils.Logger.Log.Errorf("Cannot parse coinv2 byte error %v ", err)
						return nil, nil, nil, err
					}

					// we do not use burned coins since they will reduce the privacy level of the transaction.
					if !common.IsPublicKeyBurningAddress(coinDB.GetPublicKey().ToBytesS()) {
						break
					}
					attempts++
				}
				if attempts == privacy.MaxPrivacyAttempts {
					return nil, nil, nil, fmt.Errorf("cannot form decoys")
				}

				row[j] = coinDB.GetPublicKey()
				sumInputs.Add(sumInputs, coinDB.GetCommitment())
			}
		}
		row = append(row, sumInputs)
		if i == pi {
			commitmentToZero = sumInputs
		}
		ring[i] = row
		indexes[i] = rowIndexes
	}
	return mlsag.NewRing(ring), indexes, commitmentToZero, nil
}

func getMLSAGSigFromTxSigAndKeyImages(txSig []byte, keyImages []*privacy.Point) (*mlsag.Sig, error) {
	mlsagSig, err := new(mlsag.Sig).FromBytes(txSig)
	if err != nil {
		utils.Logger.Log.Errorf("Has error when converting byte to mlsag signature, err: %v", err)
		return nil, err
	}

	return mlsag.NewMlsagSig(mlsagSig.GetC(), keyImages, mlsagSig.GetR())
}

func (tx *Tx) verifySig(transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isNewTransaction bool) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("input transaction must be a signed one"))
	}
	var err error

	// Reform Ring
	sumOutputsWithFee := tx_generic.CalculateSumOutputsWithFee(tx.Proof.GetOutputCoins(), tx.Fee)
	ring, err := getRingFromSigPubKeyAndLastColumnCommitmentV2(tx.GetValidationEnv(), sumOutputsWithFee, transactionStateDB)
	if err != nil {
		utils.Logger.Log.Errorf("Error when querying database to construct mlsag ring: %v ", err)
		return false, err
	}

	// Reform MLSAG Signature
	inputCoins := tx.Proof.GetInputCoins()
	keyImages := make([]*privacy.Point, len(inputCoins)+1)
	for i := 0; i < len(inputCoins); i++ {
		if inputCoins[i].GetKeyImage() == nil {
			utils.Logger.Log.Errorf("Error when reconstructing mlsagSignature: missing keyImage")
			return false, err
		}
		keyImages[i] = inputCoins[i].GetKeyImage()
	}
	// The last column is gone, so just fill in any value
	keyImages[len(inputCoins)] = privacy.RandomPoint()
	mlsagSignature, err := getMLSAGSigFromTxSigAndKeyImages(tx.Sig, keyImages)
	if err != nil {
		return false, err
	}

	return mlsag.Verify(mlsagSignature, ring, tx.Hash()[:])
}

// Verify is the sub-function for ValidateTransaction.
// It is called in the verification flow of most transactions, excluding some special TX types.
// It takes in boolParams to reflect some big differences across code versions; and db pointer, shard ID & token ID to get coins from the chain database.
func (tx *Tx) Verify(boolParams map[string]bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	var err error
	var valid bool
	if tokenID, err = tx_generic.ParseTokenID(tokenID); err != nil {
		return false, err
	}
	proofAsV2, ok := tx.GetProof().(*privacy.ProofV2)
	if !ok {
		utils.Logger.Log.Errorf("Error in tx %s : ver2 transaction cannot have proofs of any other version - %v", tx.Hash().String(), err)
		return false, utils.NewTransactionErr(utils.UnexpectedError, err)
	}

	isNewTransaction, ok := boolParams["isNewTransaction"]
	if !ok {
		isNewTransaction = false
	}

	isConfAsset, err := proofAsV2.IsConfidentialAsset()
	if err != nil {
		utils.Logger.Log.Errorf("Error in tx %s : proof is invalid due to inconsistent asset tags - %v", tx.Hash().String(), err)
		return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
	}
	if isConfAsset {
		utils.Logger.Log.Infof("Verifying transaction with assetTag")
		valid, err = tx.verifySigCA(transactionStateDB, shardID, tokenID, isNewTransaction)
	} else {
		utils.Logger.Log.Infof("Verifying transaction without assetTag")
		valid, err = tx.verifySig(transactionStateDB, shardID, tokenID, isNewTransaction)
	}
	if !valid {
		utils.Logger.Log.Infof("Fail with CA = %v and tokenID = %s", isConfAsset, tokenID.String())
		if err != nil {
			utils.Logger.Log.Errorf("Error verifying signature ver2 with tx hash %s: %+v", tx.Hash().String(), err)
			return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
		}
		utils.Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String())
		return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String()))
	}

	boolParams["hasConfidentialAsset"] = isConfAsset

	if valid, err := tx.Proof.Verify(boolParams, tx.SigPubKey, tx.Fee, shardID, tokenID, nil); !valid {
		if err != nil {
			utils.Logger.Log.Error(err)
		}
		utils.Logger.Log.Error("FAILED VERIFICATION PAYMENT PROOF VER 2")
		err1, ok := err.(*privacy.PrivacyError)
		if ok {
			// parse error detail
			if err1.Code == privacy.ErrCodeMessage[errhandler.VerifyOneOutOfManyProofFailedErr].Code {
				if isNewTransaction {
					return false, utils.NewTransactionErr(utils.VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
				}
				// for old txs which be get from sync block or validate new block
				if tx.LockTime <= utils.ValidateTimeForOneoutOfManyProof {
					// only verify by signOnMessage on block because of issue #504(that mean we should pass old tx, which happen before this issue)
					return true, nil
				}
				return false, utils.NewTransactionErr(utils.VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
			}
		}
		return false, utils.NewTransactionErr(utils.TxProofVerifyFailError, err, tx.Hash().String())
	}
	utils.Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil
}

// VerifyMinerCreatedTxBeforeGettingInBlock checks that a transaction was created by a miner
func (tx Tx) VerifyMinerCreatedTxBeforeGettingInBlock(mintdata *metadata.MintData, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	return tx_generic.VerifyTxCreatedByMiner(&tx, mintdata, shardID, bcr, accumulatedValues, retriever, viewRetriever)
}

// ========== SALARY FUNCTIONS: INIT AND VALIDATE  ==========

func (tx Tx) IsSalaryTx() bool {
	if tx.GetType() != common.TxRewardType && tx.GetType() != common.TxReturnStakingType {
		return false
	}

	proof := tx.GetProof()
	if proof == nil {
		return false
	}
	if len(proof.GetInputCoins()) != 0 {
		return false
	}
	if len(proof.GetOutputCoins()) != 1 {
		return false
	}

	return true
}

// InitTxSalary is used to create a "mint" transaction of version 2. The minting rule is covered inside the metadata.
func (tx *Tx) InitTxSalary(otaCoin *privacy.CoinV2, privateKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata) error {
	tokenID := &common.Hash{}
	if err := tokenID.SetBytes(common.PRVCoinID[:]); err != nil {
		return utils.NewTransactionErr(utils.TokenIDInvalidError, err, tokenID.String())
	}
	found, status, err := statedb.HasOnetimeAddress(stateDB, *tokenID, otaCoin.GetPublicKey().ToBytesS())
	if err != nil {
		errStr := fmt.Sprintf("Checking onetimeaddress existence in database get error %v", err)
		return fmt.Errorf(errStr)
	}
	if found {
		switch status {
		case statedb.OTA_STATUS_STORED:
			utils.Logger.Log.Error("InitTxSalary got error: found onetimeaddress stored in database")
			return fmt.Errorf("cannot initTxSalary, onetimeaddress already exists in database")
		case statedb.OTA_STATUS_OCCUPIED:
			utils.Logger.Log.Warnf("Continue minting OTA %x since status is %d", otaCoin.GetPublicKey().ToBytesS(), status)
		default:
			return fmt.Errorf("invalid onetimeaddress status in database")
		}
	}

	tx.Version = utils.TxVersion2Number
	tx.Type = common.TxRewardType
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	tempOutputCoin := []privacy.Coin{otaCoin}
	proof := new(privacy.ProofV2)
	proof.Init()
	// no type-cast error since elements of tempOutputCoin are of correct type (*CoinV2)
	_ = proof.SetOutputCoins(tempOutputCoin)
	tx.Proof = proof

	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	tx.PubKeyLastByteSender = common.GetShardIDFromLastByte(publicKeyBytes[len(publicKeyBytes)-1])

	// signOnMessage Tx using ver1 schnorr
	tx.SetPrivateKey(*privateKey)
	tx.SetMetadata(metaData)

	if tx.Sig, tx.SigPubKey, err = tx_generic.SignNoPrivacy(privateKey, tx.Hash()[:]); err != nil {
		return utils.NewTransactionErr(utils.SignTxError, err)
	}
	return nil
}

// ValidateTxSalary checks the following conditions for salary transactions (s, rs):
//	- the signature is valid
//	- the number of output coins is 1
//	- all fields of the output coins are valid
//	- the commitment has been calculated correctly
//  - the ota has not existed
func (tx *Tx) ValidateTxSalary(db *statedb.StateDB) (bool, error) {
	// verify signature
	if valid, err := tx_generic.VerifySigNoPrivacy(tx.Sig, tx.SigPubKey, tx.Hash()[:]); !valid {
		if err != nil {
			utils.Logger.Log.Debugf("Error verifying signature of tx: %+v", err)
			return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
		}
		return false, nil
	}

	// check whether output coin's input exists in input list or not
	tokenID := &common.Hash{}
	if err := tokenID.SetBytes(common.PRVCoinID[:]); err != nil {
		return false, utils.NewTransactionErr(utils.TokenIDInvalidError, err, tokenID.String())
	}

	// Check commitment
	outputCoins := tx.Proof.GetOutputCoins()
	if len(outputCoins) != 1 {
		return false, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("length outputCoins of proof is not 1"))
	}
	outCoin, ok := outputCoins[0].(*privacy.CoinV2)
	if !ok {
		return false, utils.NewTransactionErr(utils.CommitOutputCoinError, fmt.Errorf("outCoin must be of version 2, got %v", outputCoins[0].GetVersion()))
	}
	cmpCommitment := privacy.PedCom.CommitAtIndex(outCoin.GetAmount(), outCoin.GetRandomness(), privacy.PedersenValueIndex)
	if !privacy.IsPointEqual(cmpCommitment, outCoin.GetCommitment()) {
		return false, utils.NewTransactionErr(utils.CommitOutputCoinError, fmt.Errorf("output coin's commitment isn't calculated correctly"))
	}

	return true, nil
}

// nolint:revive // skip linter since this function modifies a value receiver
// Hash returns the hash of this transaction.
// All non-signature fields are marshalled into JSON before hashing
func (tx Tx) Hash() *common.Hash {
	// leave out signature & its public key when hashing tx
	tx.Sig = []byte{}
	tx.SigPubKey = []byte{}
	inBytes, err := json.Marshal(tx)
	if err != nil {
		return nil
	}
	hash := common.HashH(inBytes)
	// after this returns, tx is restored since the receiver is not a pointer
	return &hash
}

// HashWithoutMetadataSig returns the hash of this transaction, but it leaves out the metadata's own signature field. It is used to verify that metadata signature.
func (tx Tx) HashWithoutMetadataSig() *common.Hash {
	md := tx.GetMetadata()
	mdHash := md.HashWithoutSig()
	tx.SetMetadata(nil)
	txHash := tx.Hash()
	if mdHash == nil || txHash == nil {
		return nil
	}
	// tx.SetMetadata(md)
	inBytes := append(mdHash[:], txHash[:]...)
	hash := common.HashH(inBytes)
	return &hash
}

// ========== VALIDATE FUNCTIONS ============

func (tx Tx) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	if tx.Proof == nil {
		return false, fmt.Errorf("tx Privacy Ver 2 must have proof")
	}

	if check, err := tx_generic.ValidateSanity(&tx, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight); !check || err != nil {
		utils.Logger.Log.Errorf("Cannot check sanity of version, size, proof, type and info: err %v", err)
		return false, err
	}

	if check, err := tx_generic.MdValidateSanity(&tx, chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight); !check || err != nil {
		utils.Logger.Log.Errorf("Cannot check sanity of metadata: err %v", err)
		return false, err
	}

	return true, nil
}

// ========== SHARED FUNCTIONS ============

func (tx Tx) GetTxMintData() (bool, privacy.Coin, *common.Hash, error) {
	return tx_generic.GetTxMintData(&tx, &common.PRVCoinID)
}

func (tx Tx) GetTxBurnData() (bool, privacy.Coin, *common.Hash, error) {
	return tx_generic.GetTxBurnData(&tx)
}

func (tx Tx) GetTxFullBurnData() (bool, privacy.Coin, privacy.Coin, *common.Hash, error) {
	isBurn, burnedCoin, burnedToken, err := tx.GetTxBurnData()
	return isBurn, burnedCoin, nil, burnedToken, err
}

func (tx *Tx) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	err := tx_generic.MdValidateWithBlockChain(tx, chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)
	if err != nil {
		return err
	}
	return tx.TxBase.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
}

func (tx Tx) ValidateTransaction(boolParams map[string]bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, []privacy.Proof, error) {
	jsb, _ := json.Marshal(tx)
	utils.Logger.Log.Infof("Begin verifying TX %s", string(jsb))
	switch tx.GetType() {
	case common.TxRewardType:
		valid, err := tx.ValidateTxSalary(transactionStateDB)
		return valid, nil, err
	case common.TxReturnStakingType:
		valid, err := tx.ValidateTxSalary(transactionStateDB)
		return valid, nil, err
	case common.TxConversionType:
		valid, err := validateConversionVer1ToVer2(&tx, transactionStateDB, shardID, tokenID)
		return valid, nil, err
	default:
		valid, err := tx.Verify(boolParams, transactionStateDB, bridgeStateDB, shardID, tokenID)
		resultProofs := []privacy.Proof{}
		isBatch, ok := boolParams["isBatch"]
		if !ok {
			isBatch = false
		}
		if isBatch {
			if tx.GetProof() != nil {
				resultProofs = append(resultProofs, tx.GetProof())
			}
		}
		return valid, resultProofs, err
	}
	// return validateTransaction(&tx, hasPrivacy, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}

func (tx Tx) ValidateTxByItself(boolParams map[string]bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	valid, _, err := tx.ValidateTransaction(boolParams, transactionStateDB, bridgeStateDB, shardID, prvCoinID)
	if !valid {
		return false, err
	}

	hasPrivacy, ok := boolParams["hasPrivacy"]
	if !ok {
		hasPrivacy = false
	}
	valid, err = tx_generic.MdValidate(&tx, hasPrivacy, transactionStateDB, bridgeStateDB, shardID)
	if !valid {
		return false, err
	}
	return true, nil
}

// GetTxActualSize returns the size of this TX.
// It is the length of its JSON form.
func (tx Tx) GetTxActualSize() uint64 {
	jsb, err := json.Marshal(tx)
	if err != nil {
		return 0
	}
	return uint64(math.Ceil(float64(len(jsb)) / 1024))
}

func (tx Tx) ListOTAHashH() []common.Hash {
	result := make([]common.Hash, 0)
	if tx.Proof != nil {
		for _, outputCoin := range tx.Proof.GetOutputCoins() {
			// Discard coins sent to the burning address
			if common.IsPublicKeyBurningAddress(outputCoin.GetPublicKey().ToBytesS()) {
				continue
			}
			hash := common.HashH(outputCoin.GetPublicKey().ToBytesS())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

func (tx Tx) validateDuplicateOTAsWithCurrentMempool(poolOTAHashH map[common.Hash][]common.Hash) error {
	if tx.Proof == nil {
		return nil
	}
	outCoinHash := make(map[common.Hash][32]byte)
	declaredOTAHash := make(map[common.Hash][32]byte)
	for _, outputCoin := range tx.Proof.GetOutputCoins() {
		// Skip coins sent to the burning address
		if common.IsPublicKeyBurningAddress(outputCoin.GetPublicKey().ToBytesS()) {
			continue
		}
		hash := common.HashH(outputCoin.GetPublicKey().ToBytesS())
		outCoinHash[hash] = outputCoin.GetPublicKey().ToBytes()
	}
	decls := tx_generic.GetOTADeclarationsFromTx(tx)
	for _, otaDeclaration := range decls {
		otaPublicKey := otaDeclaration.PublicKey[:]
		hash := common.HashH(otaPublicKey)
		declaredOTAHash[hash] = otaDeclaration.PublicKey
	}

	for key, listOTAs := range poolOTAHashH {
		for _, otaHash := range listOTAs {
			if pk, exists := declaredOTAHash[otaHash]; exists {
				return fmt.Errorf("duplicate OTA %x with current mempool for TX %v", pk, tx.Hash().String())
			}
			if pk, exists := outCoinHash[otaHash]; exists {
				declKey := common.Hash{}
				if key == declKey && tx.IsSalaryTx() {
					// minting over requested OTA
					continue
				}
				return fmt.Errorf("duplicate OTA %x with current mempool for TX %v", pk, tx.Hash().String())
			}
		}
	}
	return nil
}

func (tx Tx) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	if tx.Proof == nil {
		return nil
	}

	// Check if any OTA has been produced in current mempool
	poolSNDOutputsHashH := mr.GetOTAHashH()
	err := tx.validateDuplicateOTAsWithCurrentMempool(poolSNDOutputsHashH)
	if err != nil {
		return err
	}

	// Check if any serial number has been used in current mempool
	temp := make(map[common.Hash]interface{})
	for _, desc := range tx.Proof.GetInputCoins() {
		hash := common.HashH(desc.GetKeyImage().ToBytesS())
		temp[hash] = nil
	}
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	for _, listSerialNumbers := range poolSerialNumbersHashH {
		for _, serialNumberHash := range listSerialNumbers {
			if _, ok := temp[serialNumberHash]; ok {
				return fmt.Errorf("double spend in mempool")
			}
		}
	}
	return nil
}
