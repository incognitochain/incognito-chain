package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/wallet"
)

type Tx struct {
	// Basic data
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`

	Fee  uint64 `json:"Fee"` // Fee applies: always consant
	Info []byte // 512 bytes

	// Sign and Privacy proof
	SigPubKey []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig       []byte `json:"Sig, omitempty"`       //
	Proof     *zkp.PaymentProof

	PubKeyLastByteSender byte
	// Metadata
	Metadata metadata.Metadata

	sigPrivKey       []byte       // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes
	cachedHash       *common.Hash // cached hash data of tx
	cachedActualSize *uint64      // cached actualsize data for tx
}

func (tx *Tx) UnmarshalJSON(data []byte) error {
	type Alias Tx
	temp := &struct {
		Metadata interface{}
		*Alias
	}{
		Alias: (*Alias)(tx),
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	meta, parseErr := metadata.ParseMetadata(temp.Metadata)
	if parseErr != nil {
		Logger.log.Error(parseErr)
		return parseErr
	}
	tx.SetMetadata(meta)

	return nil
}

// Init - init value for tx from inputcoin(old output coin from old tx)
// create new outputcoin and build privacy proof
// if not want to create a privacy tx proof, set hashPrivacy = false
// database is used like an interface which use to query info from db in building tx
func (tx *Tx) Init(
	senderSK *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	hasPrivacy bool,
	db database.DatabaseInterface,
	tokenID *common.Hash, // default is nil -> use for prv coin
	metaData metadata.Metadata,
) *TransactionError {

	Logger.log.Debugf("CREATING TX........\n")
	tx.Version = txVersion
	var err error

	if len(inputCoins) > 255 {
		return NewTransactionErr(UnexpectedErr, errors.New("Input coins in tx are very large:"+strconv.Itoa(len(inputCoins))))
	}

	if len(paymentInfo) > 254 {
		return NewTransactionErr(UnexpectedErr, errors.New("Input coins in tx are very large:"+strconv.Itoa(len(paymentInfo))))
	}

	if tokenID == nil {
		tokenID = &common.Hash{}
		tokenID.SetBytes(common.PRVCoinID[:])
	}

	// Calculate execution time
	start := time.Now()

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	// create sender's key set from sender's spending key
	senderFullKey := incognitokey.KeySet{}
	err = senderFullKey.ImportFromPrivateKey(senderSK)
	if err != nil {
		Logger.log.Error(err)
		return NewTransactionErr(UnexpectedErr, err)
	}
	// get public key last byte of sender
	pkLastByteSender := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]

	// init info of tx
	tx.Info = []byte{}

	// set metadata
	tx.Metadata = metaData

	// set tx type
	tx.Type = common.TxNormalType
	Logger.log.Debugf("len(inputCoins), fee, hasPrivacy: %d, %d, %v\n", len(inputCoins), fee, hasPrivacy)

	if len(inputCoins) == 0 && fee == 0 && !hasPrivacy {
		Logger.log.Infof("len(inputCoins) == 0 && fee == 0 && !hasPrivacy\n")
		tx.Fee = fee
		tx.sigPrivKey = *senderSK
		tx.PubKeyLastByteSender = pkLastByteSender
		err := tx.signTx()
		if err != nil {
			Logger.log.Error(err)
			return NewTransactionErr(UnexpectedErr, err)
		}
		return nil
	}

	shardID := common.GetShardIDFromLastByte(pkLastByteSender)
	var commitmentIndexs []uint64   // array index random of commitments in db
	var myCommitmentIndexs []uint64 // index in array index random of commitment in db

	if hasPrivacy {
		commitmentIndexs, myCommitmentIndexs, _ = RandomCommitmentsProcess(inputCoins, privacy.CMRingSize, db, shardID, tokenID)

		// Check number of list of random commitments, list of random commitment indices
		if len(commitmentIndexs) != len(inputCoins)*privacy.CMRingSize {
			return NewTransactionErr(RandomCommitmentErr, nil)
		}

		if len(myCommitmentIndexs) != len(inputCoins) {
			return NewTransactionErr(RandomCommitmentErr, errors.New("number of list my commitment indices must be equal to number of input coins"))
		}
	}

	// Calculate execution time for creating payment proof
	startPrivacy := time.Now()

	// Calculate sum of all output coins' value
	sumOutputValue := uint64(0)
	for _, p := range paymentInfo {
		sumOutputValue += p.Amount
	}

	// Calculate sum of all input coins' value
	sumInputValue := uint64(0)
	for _, coin := range inputCoins {
		sumInputValue += coin.CoinDetails.Value
	}
	Logger.log.Debugf("sumInputValue: %d\n", sumInputValue)

	// Calculate over balance, it will be returned to sender
	overBalance := int(sumInputValue - sumOutputValue - fee)

	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return NewTransactionErr(WrongInput, errors.New("input value less than output value"))
	}

	// if overBalance > 0, create a new payment info with pk is sender's pk and amount is overBalance
	if overBalance > 0 {
		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = uint64(overBalance)
		changePaymentInfo.PaymentAddress = senderFullKey.PaymentAddress
		paymentInfo = append(paymentInfo, changePaymentInfo)
	}

	// create new output coins
	outputCoins := make([]*privacy.OutputCoin, len(paymentInfo))

	// create SNDs for output coins
	ok := true
	sndOuts := make([]*big.Int, 0)
	for ok {
		var sndOut *big.Int
		for i := 0; i < len(paymentInfo); i++ {
			sndOut = privacy.RandScalar()
			for {

				ok1, err := CheckSNDerivatorExistence(tokenID, sndOut, shardID, db)
				if err != nil {
					Logger.log.Error(err)
				}
				// if sndOut existed, then re-random it
				if ok1 {
					sndOut = privacy.RandScalar()
				} else {
					break
				}
			}
			sndOuts = append(sndOuts, sndOut)
		}

		// if sndOuts has two elements that have same value, then re-generates it
		ok = common.CheckDuplicateBigIntArray(sndOuts)
		if ok {
			sndOuts = make([]*big.Int, 0)
		}
	}

	// create new output coins with info: Pk, value, last byte of pk, snd
	for i, pInfo := range paymentInfo {
		outputCoins[i] = new(privacy.OutputCoin)
		outputCoins[i].CoinDetails = new(privacy.Coin)
		outputCoins[i].CoinDetails.Value = pInfo.Amount
		outputCoins[i].CoinDetails.PublicKey = new(privacy.EllipticPoint)
		outputCoins[i].CoinDetails.PublicKey.Decompress(pInfo.PaymentAddress.Pk)
		outputCoins[i].CoinDetails.SNDerivator = sndOuts[i]
	}

	// assign fee tx
	tx.Fee = fee

	// create zero knowledge proof of payment
	tx.Proof = &zkp.PaymentProof{}

	// get list of commitments for proving one-out-of-many from commitmentIndexs
	commitmentProving := make([]*privacy.EllipticPoint, len(commitmentIndexs))
	for i, cmIndex := range commitmentIndexs {
		commitmentProving[i] = new(privacy.EllipticPoint)
		temp, err := db.GetCommitmentByIndex(*tokenID, cmIndex, shardID)
		if err != nil {
			return NewTransactionErr(UnexpectedErr, err)
		}
		err = commitmentProving[i].Decompress(temp)
		if err != nil {
			return NewTransactionErr(UnexpectedErr, err)
		}
	}

	// prepare witness for proving
	witness := new(zkp.PaymentWitness)
	err = witness.Init(hasPrivacy, new(big.Int).SetBytes(*senderSK), inputCoins, outputCoins, pkLastByteSender, commitmentProving, commitmentIndexs, myCommitmentIndexs, fee)
	if err.(*privacy.PrivacyError) != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}

	tx.Proof, err = witness.Prove(hasPrivacy)
	if err.(*privacy.PrivacyError) != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}

	Logger.log.Debugf("DONE PROVING........\n")

	// set private key for signing tx
	if hasPrivacy {
		tx.sigPrivKey = make([]byte, 64)
		randSK := witness.RandSK
		tx.sigPrivKey = append(*senderSK, randSK.Bytes()...)

		// encrypt coin details (Randomness)
		// hide information of output coins except coin commitments, public key, snDerivators
		for i := 0; i < len(tx.Proof.OutputCoins); i++ {
			err = tx.Proof.OutputCoins[i].Encrypt(paymentInfo[i].PaymentAddress.Tk)
			if err.(*privacy.PrivacyError) != nil {
				return NewTransactionErr(UnexpectedErr, err)
			}
			tx.Proof.OutputCoins[i].CoinDetails.SerialNumber = nil
			tx.Proof.OutputCoins[i].CoinDetails.Value = 0
			tx.Proof.OutputCoins[i].CoinDetails.Randomness = nil
		}

		// hide information of input coins except serial number of input coins
		for i := 0; i < len(tx.Proof.InputCoins); i++ {
			tx.Proof.InputCoins[i].CoinDetails.CoinCommitment = nil
			tx.Proof.InputCoins[i].CoinDetails.Value = 0
			tx.Proof.InputCoins[i].CoinDetails.SNDerivator = nil
			tx.Proof.InputCoins[i].CoinDetails.PublicKey = nil
			tx.Proof.InputCoins[i].CoinDetails.Randomness = nil
		}

	} else {
		tx.sigPrivKey = []byte{}
		randSK := big.NewInt(0)
		tx.sigPrivKey = append(*senderSK, randSK.Bytes()...)
	}

	// sign tx
	tx.PubKeyLastByteSender = pkLastByteSender
	err = tx.signTx()
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}

	elapsedPrivacy := time.Since(startPrivacy)
	elapsed := time.Since(start)
	Logger.log.Debugf("Creating payment proof time %s", elapsedPrivacy)
	Logger.log.Infof("Successfully Creating normal tx %+v in %s time", *tx.Hash(), elapsed)
	return nil
}

// signTx - signs tx
func (tx *Tx) signTx() error {
	//Check input transaction
	if tx.Sig != nil {
		return errors.New("input transaction must be an unsigned one")
	}

	/****** using Schnorr signature *******/
	// sign with sigPrivKey
	// prepare private key for Schnorr
	sk := new(big.Int).SetBytes(tx.sigPrivKey[:privacy.BigIntSize])
	r := new(big.Int).SetBytes(tx.sigPrivKey[privacy.BigIntSize:])
	sigKey := new(privacy.SchnPrivKey)
	sigKey.Set(sk, r)

	// save public key for verification signature tx
	tx.SigPubKey = sigKey.PubKey.PK.Compress()

	// signing
	if Logger.log != nil {
		Logger.log.Debugf(tx.Hash().String())
	}
	signature, err := sigKey.Sign(tx.Hash()[:])
	if err != nil {
		return err
	}

	// convert signature to byte array
	tx.Sig = signature.Bytes()

	return nil
}

// verifySigTx - verify signature on tx
func (tx *Tx) verifySigTx() (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, errors.New("input transaction must be an signed one")
	}

	var err error
	res := false

	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verKey := new(privacy.SchnPubKey)
	verKey.PK = new(privacy.EllipticPoint)
	err = verKey.PK.Decompress(tx.SigPubKey)
	if err != nil {
		return false, err
	}

	verKey.G = new(privacy.EllipticPoint)
	verKey.G.Set(privacy.PedCom.G[privacy.SK].X, privacy.PedCom.G[privacy.SK].Y)

	verKey.H = new(privacy.EllipticPoint)
	verKey.H.Set(privacy.PedCom.G[privacy.RAND].X, privacy.PedCom.G[privacy.RAND].Y)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	signature.SetBytes(tx.Sig)

	// verify signature
	// Logger.log.Infof(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash()[:])
	// Logger.log.Infof(" VERIFY SIGNATURE ----------- TX Proof bytes before verifing the signature: %v\n", tx.Proof.Bytes())
	// Logger.log.Infof(" VERIFY SIGNATURE ----------- TX meta: %v\n", tx.Metadata)
	res = verKey.Verify(signature, tx.Hash()[:])

	return res, nil
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
func (tx *Tx) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	//hasPrivacy = false
	Logger.log.Debugf("VALIDATING TX........\n")
	// start := time.Now()
	// Verify tx signature
	if tx.GetType() == common.TxRewardType {
		return tx.ValidateTxSalary(db)
	}
	if tx.GetType() == common.TxReturnStakingType {
		return tx.ValidateTxReturnStaking(db), nil
	}
	var valid bool
	var err error

	valid, err = tx.verifySigTx()
	if !valid {
		if err != nil {
			Logger.log.Errorf("Error verifying signature of tx: %+v \n", err)
		}
		//Logger.log.Error("FAILED VERIFICATION SIGNATURE")
		return false, nil
	}

	if tx.Proof != nil {
		if tokenID == nil {
			tokenID = &common.Hash{}
			tokenID.SetBytes(common.PRVCoinID[:])
		}

		sndOutputs := make([]*big.Int, len(tx.Proof.OutputCoins))
		for i := 0; i < len(tx.Proof.OutputCoins); i++ {
			sndOutputs[i] = tx.Proof.OutputCoins[i].CoinDetails.SNDerivator
		}

		if common.CheckDuplicateBigIntArray(sndOutputs) {
			Logger.log.Errorf("Duplicate output coins' snd\n")
			return false, errors.New("Duplicate output coins' snd\n")
		}

		for i := 0; i < len(tx.Proof.OutputCoins); i++ {
			// Check output coins' SND is not exists in SND list (Database)
			if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.OutputCoins[i].CoinDetails.SNDerivator, shardID, db); ok || err != nil {
				Logger.log.Errorf("snd existed: %d\n", i)
				return false, errors.New(fmt.Sprintf("snd existed: %d\n", i))
			}
		}

		if !hasPrivacy {
			// Check input coins' cm is exists in cm list (Database)
			for i := 0; i < len(tx.Proof.InputCoins); i++ {
				ok, err := tx.CheckCMExistence(tx.Proof.InputCoins[i].CoinDetails.CoinCommitment.Compress(), db, shardID, tokenID)
				if !ok || err != nil {
					return false, err
				}
			}
		}

		// Verify the payment proof
		valid = tx.Proof.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, db, shardID, tokenID)
		if !valid {
			Logger.log.Error("FAILED VERIFICATION PAYMENT PROOF")
			return false, errors.New("FAILED VERIFICATION PAYMENT PROOF")
		} else {
			//Logger.log.Infof("SUCCESSED VERIFICATION PAYMENT PROOF ")
		}
	}
	//@UNCOMMENT: metrics time
	//elapsed := time.Since(start)
	//Logger.log.Infof("Validation normal tx %+v in %s time \n", *tx.Hash(), elapsed)

	return true, nil
}

func (tx Tx) String() string {
	record := strconv.Itoa(int(tx.Version))
	//fmt.
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		tmp := base58.Base58Check{}.Encode(tx.Proof.Bytes()[:], 0x00)
		record += tmp
		// fmt.Printf("Proof check base 58: %v\n",tmp)
	}
	if tx.Metadata != nil {
		metadataHash := tx.Metadata.Hash()
		//Logger.log.Infof("\n\n\n\n test metadata after hashing: %v\n", metadataHash.GetBytes())
		metadata := metadataHash.String()
		record += metadata
	}

	//TODO: To be uncomment
	// record += string(tx.Info)
	return record
}

func (tx *Tx) Hash() *common.Hash {
	if tx.cachedHash != nil {
		return tx.cachedHash
	}
	bytes := []byte(tx.String())
	hash := common.HashH(bytes)
	tx.cachedHash = &hash
	return &hash
}

func (tx *Tx) GetSenderAddrLastByte() byte {
	return tx.PubKeyLastByteSender
}

func (tx *Tx) GetTxFee() uint64 {
	return tx.Fee
}

func (tx *Tx) GetTxFeeToken() uint64 {
	return uint64(0)
}

// GetTxActualSize computes the actual size of a given transaction in kilobyte
func (tx *Tx) GetTxActualSize() uint64 {
	if tx.cachedActualSize != nil {
		return *tx.cachedActualSize
	}
	sizeTx := uint64(1)                // int8
	sizeTx += uint64(len(tx.Type) + 1) // string
	sizeTx += uint64(8)                // int64
	sizeTx += uint64(8)

	sigPubKey := uint64(len(tx.SigPubKey))
	sizeTx += sigPubKey
	sig := uint64(len(tx.Sig))
	sizeTx += sig
	if tx.Proof != nil {
		proof := uint64(len(tx.Proof.Bytes()))
		sizeTx += proof
	}

	sizeTx += uint64(1)
	info := uint64(len(tx.Info))
	sizeTx += info

	meta := tx.Metadata
	if meta != nil {
		metaSize := meta.CalculateSize()
		sizeTx += metaSize
	}
	result := uint64(math.Ceil(float64(sizeTx) / 1024))
	tx.cachedActualSize = &result
	return *tx.cachedActualSize
}

// GetType returns the type of the transaction
func (tx *Tx) GetType() string {
	return tx.Type
}

func (tx *Tx) ListSerialNumbersHashH() []common.Hash {
	result := []common.Hash{}
	if tx.Proof != nil {
		for _, d := range tx.Proof.InputCoins {
			hash := common.HashH(d.CoinDetails.SerialNumber.Compress())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i,j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

// CheckCMExistence returns true if cm exists in cm list
func (tx Tx) CheckCMExistence(cm []byte, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	ok, err := db.HasCommitment(*tokenID, cm, shardID)
	return ok, err
}

func (tx *Tx) CheckTxVersion(maxTxVersion int8) bool {
	return !(tx.Version > maxTxVersion)
}

func (tx *Tx) CheckTransactionFee(minFeePerKbTx uint64) bool {
	if tx.IsSalaryTx() {
		return true
	}
	if tx.Metadata != nil {
		return tx.Metadata.CheckTransactionFee(tx, minFeePerKbTx)
	}
	fullFee := minFeePerKbTx * tx.GetTxActualSize()
	return tx.Fee >= fullFee
}

func (tx *Tx) IsSalaryTx() bool {
	// Check normal tx(not an action tx)
	if tx.GetType() != common.TxRewardType {
		return false
	}
	// Check serialNumber in every Descs
	if len(tx.Proof.InputCoins) == 0 {
		return true
	}
	return false
}

func (tx *Tx) GetSender() []byte {
	if tx.Proof == nil || len(tx.Proof.InputCoins) == 0 {
		return nil
	}
	return tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
}

func (tx *Tx) GetReceivers() ([][]byte, []uint64) {
	pubkeys := [][]byte{}
	amounts := []uint64{}
	if tx.Proof != nil && len(tx.Proof.OutputCoins) > 0 {
		for _, coin := range tx.Proof.OutputCoins {
			added := false
			coinPubKey := coin.CoinDetails.PublicKey.Compress()
			for i, key := range pubkeys {
				if bytes.Equal(coinPubKey, key) {
					added = true
					amounts[i] += coin.CoinDetails.Value
					break
				}
			}
			if !added {
				pubkeys = append(pubkeys, coinPubKey)
				amounts = append(amounts, coin.CoinDetails.Value)
			}
		}
	}
	return pubkeys, amounts
}

func (tx *Tx) GetUniqueReceiver() (bool, []byte, uint64) {
	sender := []byte{} // Empty byte slice for coinbase tx
	if tx.Proof != nil && len(tx.Proof.InputCoins) > 0 && !tx.IsPrivacy() {
		sender = tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	}
	pubkeys, amounts := tx.GetReceivers()
	pubkey := []byte{}
	amount := uint64(0)
	count := 0
	for i, pk := range pubkeys {
		if !bytes.Equal(pk, sender) {
			pubkey = pk
			amount = amounts[i]
			count += 1
		}
	}
	return count == 1, pubkey, amount
}

func (tx *Tx) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	unique, pk, amount := tx.GetUniqueReceiver()
	return unique, pk, amount, &common.PRVCoinID
}

func (tx *Tx) GetTokenReceivers() ([][]byte, []uint64) {
	return nil, nil
}

func (tx *Tx) GetTokenUniqueReceiver() (bool, []byte, uint64) {
	return false, nil, 0
}

func (tx *Tx) validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH map[common.Hash][]common.Hash) error {
	if tx.Proof == nil {
		return nil
	}
	temp := make(map[common.Hash]interface{})
	for _, desc := range tx.Proof.InputCoins {
		hash := common.HashH(desc.CoinDetails.SerialNumber.Compress())
		temp[hash] = nil
	}

	for _, listSerialNumbers := range poolSerialNumbersHashH {
		for _, serialNumberHash := range listSerialNumbers {
			if _, ok := temp[serialNumberHash]; ok {
				return errors.New("double spend")
			}
		}
	}
	return nil
}

func (tx *Tx) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	return tx.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
}

// ValidateDoubleSpend - check double spend for any transaction type
func (tx *Tx) ValidateConstDoubleSpendWithBlockchain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) error {

	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	for i := 0; tx.Proof != nil && i < len(tx.Proof.InputCoins); i++ {
		serialNumber := tx.Proof.InputCoins[i].CoinDetails.SerialNumber.Compress()
		ok, err := db.HasSerialNumber(*prvCoinID, serialNumber, shardID)
		if ok || err != nil {
			return errors.New("double spend")
		}
	}
	return nil
}

func (tx *Tx) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) error {
	if tx.GetType() == common.TxRewardType || tx.GetType() == common.TxReturnStakingType {
		return nil
	}
	if tx.Metadata != nil {
		fmt.Printf("[db] validate metadata with blockchain: %d %h\n", tx.GetMetadataType(), tx.Hash())
		isContinued, err := tx.Metadata.ValidateTxWithBlockChain(tx, bcr, shardID, db)
		if err != nil {
			return err
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateConstDoubleSpendWithBlockchain(bcr, shardID, db)
}

func (tx Tx) validateNormalTxSanityData() (bool, error) {
	//check version
	if tx.Version > txVersion {
		return false, errors.New(fmt.Sprintf("tx version is %d. Wrong version tx. Only support for version >= %d", tx.Version, txVersion))
	}
	// check LockTime before now
	if int64(tx.LockTime) > time.Now().Unix() {
		return false, errors.New("wrong tx locktime")
	}

	// check tx size
	if tx.GetTxActualSize() > common.MaxTxSize {
		return false, errors.New("tx size is too large")
	}

	// check sanity of Proof
	validateSanityOfProof, err := tx.validateSanityDataOfProof()
	if err != nil || !validateSanityOfProof {
		return false, err
	}

	if len(tx.SigPubKey) != privacy.SigPubKeySize {
		return false, errors.New("wrong tx Sig PK")
	}
	// check Type is normal or salary tx
	switch tx.Type {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, errors.New("wrong tx type")
	}

	//if txN.Type != common.TxNormalType && txN.Type != common.TxRewardType && txN.Type != common.TxCustomTokenType && txN.Type != common.TxCustomTokenPrivacyType { // only 1 byte
	//	return false, errors.New("wrong tx type")
	//}

	// check info field
	if len(tx.Info) > 512 {
		return false, errors.New("wrong tx info length")
	}

	return true, nil
}

func (txN Tx) validateSanityDataOfProof() (bool, error) {
	if txN.Proof != nil {

		if len(txN.Proof.InputCoins) > 255 {
			return false, errors.New("Input coins in tx are very large:" + strconv.Itoa(len(txN.Proof.InputCoins)))
		}

		if len(txN.Proof.OutputCoins) > 255 {
			return false, errors.New("Output coins in tx are very large:" + strconv.Itoa(len(txN.Proof.OutputCoins)))
		}

		isPrivacy := true
		// check Privacy or not

		if txN.Proof.AggregatedRangeProof == nil || len(txN.Proof.OneOfManyProof) == 0 || len(txN.Proof.SerialNumberProof) == 0 {
			isPrivacy = false
		}

		if isPrivacy {
			if !txN.Proof.AggregatedRangeProof.ValidateSanity() {
				return false, errors.New("validate sanity Aggregated range proof failed")
			}

			for i := 0; i < len(txN.Proof.OneOfManyProof); i++ {
				if !txN.Proof.OneOfManyProof[i].ValidateSanity() {
					return false, errors.New("validate sanity One out of many proof failed")
				}
			}
			for i := 0; i < len(txN.Proof.SerialNumberProof); i++ {
				if !txN.Proof.SerialNumberProof[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number proof failed")
				}
			}

			// check input coins with privacy
			for i := 0; i < len(txN.Proof.InputCoins); i++ {
				if !txN.Proof.InputCoins[i].CoinDetails.SerialNumber.IsSafe() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
			}
			// check output coins with privacy
			for i := 0; i < len(txN.Proof.OutputCoins); i++ {
				if !txN.Proof.OutputCoins[i].CoinDetails.PublicKey.IsSafe() {
					return false, errors.New("validate sanity Public key of output coin failed")
				}
				if !txN.Proof.OutputCoins[i].CoinDetails.CoinCommitment.IsSafe() {
					return false, errors.New("validate sanity Coin commitment of output coin failed")
				}
				if len(txN.Proof.OutputCoins[i].CoinDetails.SNDerivator.Bytes()) > privacy.BigIntSize {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
			// check ComInputSK
			if !txN.Proof.ComInputSK.IsSafe() {
				return false, errors.New("validate sanity ComInputSK of proof failed")
			}
			// check ComInputValue
			for i := 0; i < len(txN.Proof.ComInputValue); i++ {
				if !txN.Proof.ComInputValue[i].IsSafe() {
					return false, errors.New("validate sanity ComInputValue of proof failed")
				}
			}
			//check ComInputSND
			for i := 0; i < len(txN.Proof.ComInputSND); i++ {
				if !txN.Proof.ComInputSND[i].IsSafe() {
					return false, errors.New("validate sanity ComInputSND of proof failed")
				}
			}
			//check ComInputShardID
			if !txN.Proof.ComInputShardID.IsSafe() {
				return false, errors.New("validate sanity ComInputShardID of proof failed")
			}

			// check ComOutputShardID
			for i := 0; i < len(txN.Proof.ComOutputShardID); i++ {
				if !txN.Proof.ComOutputShardID[i].IsSafe() {
					return false, errors.New("validate sanity ComOutputShardID of proof failed")
				}
			}
			//check ComOutputSND
			for i := 0; i < len(txN.Proof.ComOutputSND); i++ {
				if !txN.Proof.ComOutputSND[i].IsSafe() {
					return false, errors.New("validate sanity ComOutputSND of proof failed")
				}
			}
			//check ComOutputValue
			for i := 0; i < len(txN.Proof.ComOutputValue); i++ {
				if !txN.Proof.ComOutputValue[i].IsSafe() {
					return false, errors.New("validate sanity ComOutputValue of proof failed")
				}
			}
			if len(txN.Proof.CommitmentIndices) != len(txN.Proof.InputCoins)*privacy.CMRingSize {
				return false, errors.New("validate sanity CommitmentIndices of proof failed")

			}
		}

		if !isPrivacy {
			for i := 0; i < len(txN.Proof.SNNoPrivacyProof); i++ {
				if !txN.Proof.SNNoPrivacyProof[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number no privacy proof failed")
				}
			}
			// check input coins without privacy
			for i := 0; i < len(txN.Proof.InputCoins); i++ {
				if !txN.Proof.InputCoins[i].CoinDetails.CoinCommitment.IsSafe() {
					return false, errors.New("validate sanity CoinCommitment of input coin failed")
				}
				if !txN.Proof.InputCoins[i].CoinDetails.PublicKey.IsSafe() {
					return false, errors.New("validate sanity PublicKey of input coin failed")
				}
				if !txN.Proof.InputCoins[i].CoinDetails.SerialNumber.IsSafe() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
				if len(txN.Proof.InputCoins[i].CoinDetails.Randomness.Bytes()) > privacy.BigIntSize {
					return false, errors.New("validate sanity Randomness of input coin failed")
				}
				if len(txN.Proof.InputCoins[i].CoinDetails.SNDerivator.Bytes()) > privacy.BigIntSize {
					return false, errors.New("validate sanity SNDerivator of input coin failed")
				}

			}

			// check output coins without privacy
			for i := 0; i < len(txN.Proof.OutputCoins); i++ {
				if !txN.Proof.OutputCoins[i].CoinDetails.CoinCommitment.IsSafe() {
					return false, errors.New("validate sanity CoinCommitment of output coin failed")
				}
				if !txN.Proof.OutputCoins[i].CoinDetails.PublicKey.IsSafe() {
					return false, errors.New("validate sanity PublicKey of output coin failed")
				}
				if len(txN.Proof.OutputCoins[i].CoinDetails.Randomness.Bytes()) > privacy.BigIntSize {
					return false, errors.New("validate sanity Randomness of output coin failed")
				}
				if len(txN.Proof.OutputCoins[i].CoinDetails.SNDerivator.Bytes()) > privacy.BigIntSize {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
		}
	}
	return true, nil
}

func (tx *Tx) ValidateSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
	Logger.log.Debugf("\n\n\n START Validating sanity data of metadata %+v\n\n\n", tx.Metadata)
	if tx.Metadata != nil {
		Logger.log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := tx.Metadata.ValidateSanityData(bcr, tx)
		Logger.log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	Logger.log.Debugf("\n\n\n END sanity data of metadata%+v\n\n\n")
	return tx.validateNormalTxSanityData()
}

func (tx *Tx) ValidateTxByItself(
	hasPrivacy bool,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) (bool, error) {
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	ok, err := tx.ValidateTransaction(hasPrivacy, db, shardID, prvCoinID)
	if !ok {
		return false, err
	}
	if tx.Metadata != nil {
		if hasPrivacy {
			return false, errors.New("Metadata can not exist in not privacy tx")
		}
		validateMetadata := tx.Metadata.ValidateMetadataByItself()
		if validateMetadata {
			return validateMetadata, nil
		} else {
			return validateMetadata, errors.New("Validate Metadata fail")
		}
	}
	return true, nil
}

// GetMetadataType returns the type of underlying metadata if is existed
func (tx *Tx) GetMetadataType() int {
	if tx.Metadata != nil {
		return tx.Metadata.GetType()
	}
	return metadata.InvalidMeta
}

// GetMetadata returns metadata of tx is existed
func (tx *Tx) GetMetadata() metadata.Metadata {
	return tx.Metadata
}

// SetMetadata sets metadata to tx
func (tx *Tx) SetMetadata(meta metadata.Metadata) {
	tx.Metadata = meta
}

// GetMetadata returns metadata of tx is existed
func (tx *Tx) GetInfo() []byte {
	return tx.Info
}

func (tx *Tx) GetLockTime() int64 {
	return tx.LockTime
}

func (tx *Tx) GetSigPubKey() []byte {
	return tx.SigPubKey
}

func (tx *Tx) GetProof() *zkp.PaymentProof {
	return tx.Proof
}

func (tx *Tx) IsPrivacy() bool {
	if tx.Proof == nil || tx.Proof.OneOfManyProof == nil || len(tx.Proof.OneOfManyProof) == 0 {
		return false
	}
	return true
}

func (tx *Tx) ValidateType() bool {
	return tx.Type == common.TxNormalType || tx.Type == common.TxRewardType || tx.Type == common.TxReturnStakingType
}

func (tx *Tx) IsCoinsBurning() bool {
	if tx.Proof == nil || len(tx.Proof.OutputCoins) == 0 {
		return false
	}
	senderPKBytes := []byte{}
	if len(tx.Proof.InputCoins) > 0 {
		senderPKBytes = tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	}
	keyWalletBurningAccount, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	keysetBurningAccount := keyWalletBurningAccount.KeySet
	paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress
	for _, outCoin := range tx.Proof.OutputCoins {
		outPKBytes := outCoin.CoinDetails.PublicKey.Compress()
		if !bytes.Equal(senderPKBytes, outPKBytes) && !bytes.Equal(outPKBytes, paymentAddressBurningAccount.Pk[:]) {
			return false
		}
	}
	return true
}

func (tx *Tx) CalculateTxValue() uint64 {
	if tx.Proof == nil {
		return 0
	}
	if tx.Proof.OutputCoins == nil || len(tx.Proof.OutputCoins) == 0 {
		return 0
	}
	if tx.Proof.InputCoins == nil || len(tx.Proof.InputCoins) == 0 { // coinbase tx
		txValue := uint64(0)
		for _, outCoin := range tx.Proof.OutputCoins {
			txValue += outCoin.CoinDetails.Value
		}
		return txValue
	}

	senderPKBytes := tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	txValue := uint64(0)
	for _, outCoin := range tx.Proof.OutputCoins {
		outPKBytes := outCoin.CoinDetails.PublicKey.Compress()
		if bytes.Equal(senderPKBytes, outPKBytes) {
			continue
		}
		txValue += outCoin.CoinDetails.Value
	}
	return txValue
}

func NewEmptyTx(minerPrivateKey *privacy.PrivateKey, db database.DatabaseInterface, meta metadata.Metadata) metadata.Transaction {
	tx := Tx{}
	keyWalletBurningAdd, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	tx.InitTxSalary(0,
		&keyWalletBurningAdd.KeySet.PaymentAddress,
		minerPrivateKey,
		db,
		meta,
	)
	return &tx
}

// InitTxSalary
// Blockchain use this tx to pay a reward(salary) to miner of chain
// #1 - salary:
// #2 - receiverAddr:
// #3 - privKey:
// #4 - snDerivators:
func (tx *Tx) InitTxSalary(
	salary uint64,
	receiverAddr *privacy.PaymentAddress,
	privKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	metaData metadata.Metadata,
) error {
	tx.Version = txVersion
	tx.Type = common.TxRewardType

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	var err error
	// create new output coins with info: Pk, value, input, randomness, last byte pk, coin commitment
	tx.Proof = new(zkp.PaymentProof)
	tx.Proof.OutputCoins = make([]*privacy.OutputCoin, 1)
	tx.Proof.OutputCoins[0] = new(privacy.OutputCoin)
	//tx.Proof.OutputCoins[0].CoinDetailsEncrypted = new(privacy.CoinDetailsEncrypted).Init()
	tx.Proof.OutputCoins[0].CoinDetails = new(privacy.Coin)
	tx.Proof.OutputCoins[0].CoinDetails.Value = salary
	tx.Proof.OutputCoins[0].CoinDetails.PublicKey = new(privacy.EllipticPoint)
	err = tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Decompress(receiverAddr.Pk)
	if err != nil {
		return err
	}
	tx.Proof.OutputCoins[0].CoinDetails.Randomness = privacy.RandScalar()

	sndOut := privacy.RandScalar()
	for {
		lastByte := receiverAddr.Pk[len(receiverAddr.Pk)-1]
		shardIDSender := common.GetShardIDFromLastByte(lastByte)
		tokenID := &common.Hash{}
		tokenID.SetBytes(common.PRVCoinID[:])
		ok, err := CheckSNDerivatorExistence(tokenID, sndOut, shardIDSender, db)
		if err != nil {
			return err
		}
		if ok {
			sndOut = privacy.RandScalar()
		} else {
			break
		}
	}

	tx.Proof.OutputCoins[0].CoinDetails.SNDerivator = sndOut

	// create coin commitment
	tx.Proof.OutputCoins[0].CoinDetails.CommitAll()
	// get last byte
	tx.PubKeyLastByteSender = receiverAddr.Pk[len(receiverAddr.Pk)-1]

	// sign Tx
	tx.SigPubKey = receiverAddr.Pk
	tx.sigPrivKey = *privKey
	tx.SetMetadata(metaData)
	err = tx.signTx()
	if err != nil {
		return err
	}

	return nil
}

func (tx Tx) ValidateTxReturnStaking(db database.DatabaseInterface,
) bool {
	return true
}

func (tx Tx) ValidateTxSalary(
	db database.DatabaseInterface,
) (bool, error) {
	// verify signature
	valid, err := tx.verifySigTx()
	if !valid {
		if err != nil {
			Logger.log.Infof("Error verifying signature of tx: %+v", err)
		}
		return false, err
	}

	// check whether output coin's input exists in input list or not
	lastByte := tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()[len(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress())-1]
	shardIDSender := common.GetShardIDFromLastByte(lastByte)
	tokenID := &common.Hash{}
	tokenID.SetBytes(common.PRVCoinID[:])
	if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.OutputCoins[0].CoinDetails.SNDerivator, shardIDSender, db); ok || err != nil {
		return false, err
	}

	// check output coin's coin commitment is calculated correctly
	cmTmp := tx.Proof.OutputCoins[0].CoinDetails.PublicKey
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(tx.Proof.OutputCoins[0].CoinDetails.Value))))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator))

	shardID := common.GetShardIDFromLastByte(tx.Proof.OutputCoins[0].CoinDetails.GetPubKeyLastByte())
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{shardID})))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(tx.Proof.OutputCoins[0].CoinDetails.Randomness))
	ok := cmTmp.IsEqual(tx.Proof.OutputCoins[0].CoinDetails.CoinCommitment)
	if !ok {
		return ok, errors.New("check output coin's coin commitment isn't calculated correctly")
	}
	return ok, nil
}

func (tx Tx) GetMetadataFromVinsTx(bcr metadata.BlockchainRetriever) (metadata.Metadata, error) {
	// implement this func if needed
	return nil, nil
}

func (tx Tx) GetTokenID() *common.Hash {
	return &common.PRVCoinID
}

func (tx *Tx) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []metadata.Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	bcr metadata.BlockchainRetriever,
) (bool, error) {
	if tx.IsPrivacy() {
		return true, nil
	}

	meta := tx.Metadata
	if tx.Proof != nil && len(tx.Proof.InputCoins) == 0 && len(tx.Proof.OutputCoins) > 0 { // coinbase tx
		if meta == nil {
			return false, nil
		}
		// TODO: uncomment below as we have fully validation for all tx/meta types in order to check strictly miner created tx
		if !meta.IsMinerCreatedMetaType() {
			return false, nil
		}
	}
	if meta != nil {
		return meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, tx, bcr)
	}
	return true, nil
}
