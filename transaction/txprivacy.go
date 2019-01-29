package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	zkp "github.com/ninjadotorg/constant/privacy/zeroknowledge"
	"github.com/ninjadotorg/constant/wallet"
)

type sigWitness struct {
	// hasPrivacy bool
	SNPrivacyWitness   []*zkp.PKSNPrivacyWitness
	SNNoPrivacyWitness []*zkp.SNNoPrivacyWitness
}

type Tx struct {
	// Basic data
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`

	Fee  uint64 `json:"Fee"` // Fee applies: always consant
	Info []byte

	// Sign and Privacy proof
	SigPubKey []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig       []byte `json:"Sig, omitempty"`       //
	Proof     *zkp.PaymentProof

	PubKeyLastByteSender byte

	// Metadata
	Metadata metadata.Metadata

	sigPrivKey *sigWitness // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes
}

// func (sigWit * sigWitness) get

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
	senderSK *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	hasPrivacy bool,
	db database.DatabaseInterface,
	tokenID *common.Hash, // default is nil -> use for constant coin
	metaData metadata.Metadata,
) *TransactionError {
	//hasPrivacy = false
	tx.Version = TxVersion
	var err error
	if tokenID == nil {
		tokenID = &common.Hash{}
		tokenID.SetBytes(common.ConstantID[:])
	}

	// Calculate execution time
	start := time.Now()

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	// check number of outputs
	//if len(paymentInfo) > privacy.ValueMax{
	//	return NewTransactionErr(WrongInput, errors.New("Number of outputs is exceed max value"))
	//}

	// create sender's key set from sender's spending key
	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKey(senderSK)
	// get public key last byte of sender
	pkLastByteSender := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]
	// init info of tx
	pubKeyData := &privacy.EllipticPoint{}
	pubKeyData.Decompress(senderFullKey.PaymentAddress.Pk)
	tx.Info, err = privacy.ElGamalEncrypt(senderFullKey.PaymentAddress.Tk[:], pubKeyData)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	// set metadata
	tx.Metadata = metaData
	// set tx type
	tx.Type = common.TxNormalType
	Logger.log.Infof("len(inputCoins), fee, hasPrivacy: %d, %d, %v\n", len(inputCoins), fee, hasPrivacy)
	tx.sigPrivKey = new(sigWitness)
	tx.sigPrivKey.SNNoPrivacyWitness = nil
	tx.sigPrivKey.SNPrivacyWitness = nil
	if len(inputCoins) == 0 && fee == 0 && !hasPrivacy {
		Logger.log.Infof("CREATE TX CUSTOM TOKEN\n")
		tx.Fee = fee
		tx.PubKeyLastByteSender = pkLastByteSender

		err := tx.signTx(hasPrivacy)
		if err != nil {
			return NewTransactionErr(UnexpectedErr, err)
		}
		return nil
	}

	chainID, _ := common.GetTxSenderChain(pkLastByteSender)
	var commitmentIndexs []uint64   // array index random of commitments in db
	var myCommitmentIndexs []uint64 // index in array index random of commitment in db

	if hasPrivacy {
		commitmentIndexs, myCommitmentIndexs = RandomCommitmentsProcess(inputCoins, privacy.CMRingSize, db, chainID, tokenID)

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
	Logger.log.Infof("sumInputValue: %d\n", sumInputValue)

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
			sndOut = privacy.RandInt()
			for {

				ok1, err := CheckSNDerivatorExistence(tokenID, sndOut, chainID, db)
				if err != nil {
					Logger.log.Error(err)
				}
				// if sndOut existed, then re-random it
				if ok1 {
					sndOut = privacy.RandInt()
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
		outputCoins[i].CoinDetails.PublicKey, _ = privacy.DecompressKey(pInfo.PaymentAddress.Pk)
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
		temp, _ := db.GetCommitmentByIndex(tokenID, cmIndex, chainID)
		commitmentProving[i], _ = privacy.DecompressKey(temp)
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

	// set private key for signing tx
	if hasPrivacy {
		tx.sigPrivKey.SNPrivacyWitness = witness.SerialNumberWitness

		// encrypt coin details (Randomness)
		// hide information of output coins except coin commitments, public key, snDerivators
		for i := 0; i < len(tx.Proof.OutputCoins); i++ {
			tx.Proof.OutputCoins[i].Encrypt(paymentInfo[i].PaymentAddress.Tk)
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
		tx.sigPrivKey.SNNoPrivacyWitness = witness.SNNoPrivacyWitness
	}

	// sign tx
	tx.PubKeyLastByteSender = pkLastByteSender
	err = tx.signTx(hasPrivacy)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}

	elapsedPrivacy := time.Since(startPrivacy)
	elapsed := time.Since(start)
	Logger.log.Infof("Creating payment proof time %s", elapsedPrivacy)
	Logger.log.Infof("Creating normal tx time %s", elapsed)
	return nil
}

// signTx - signs tx
func (tx *Tx) signTx(hasPrivacy bool) error {
	//Check input transaction
	if tx.Sig != nil {
		return errors.New("input transaction must be an unsigned one")
	}

	/****** using Schnorr signature *******/
	// sign with sigPrivKey
	// // prepare private key for Schnorr
	// sk := new(big.Int).SetBytes(tx.sigPrivKey[:privacy.BigIntSize])
	// r := new(big.Int).SetBytes(tx.sigPrivKey[privacy.BigIntSize:])
	// sigKey := new(privacy.SchnPrivKey)
	// sigKey.Set(sk, r)

	// // save public key for verification signature tx
	// tx.SigPubKey = sigKey.PubKey.PK.Compress()

	// signing
	Logger.log.Infof(tx.Hash().String())
	if hasPrivacy {
		SerialSignature := make([]byte, privacy.SNPrivacyProofSize*len(tx.Proof.InputCoins))
		for i := 0; i < len(tx.Proof.InputCoins); i++ {
			proof, snerr := tx.sigPrivKey.SNPrivacyWitness[i].Prove(tx.Hash()[:])
			if snerr != nil {
				return snerr
			}
			copy(SerialSignature[i*privacy.SNPrivacyProofSize:(i+1)*privacy.SNPrivacyProofSize], proof.Bytes())
		}
		tx.Sig = SerialSignature
	} else {
		SerialSignature := make([]byte, privacy.SNNoPrivacyProofSize*len(tx.Proof.InputCoins))
		for i := 0; i < len(tx.Proof.InputCoins); i++ {
			proof, snerr := tx.sigPrivKey.SNNoPrivacyWitness[i].Prove(tx.Hash()[:])
			if snerr != nil {
				return snerr
			}
			copy(SerialSignature[i*privacy.SNPrivacyProofSize:(i+1)*privacy.SNPrivacyProofSize], proof.Bytes())
		}
		tx.Sig = SerialSignature
	}
	// signature, err := sigKey.Sign(tx.Hash()[:])
	// if err != nil {
	// 	return err
	// }

	// // convert signature to byte array
	// tx.Sig = signature.Bytes()

	return nil
}

// verifySigTx - verify signature on tx
func (tx *Tx) verifySigTx(hasPrivacy bool) (bool, error) {
	// check input transaction
	if tx.Sig == nil { //|| tx.SigPubKey == nil {
		return false, errors.New("input transaction must be an signed one")
	}

	// var err error
	res := true

	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	// verKey := new(privacy.SchnPubKey)
	// verKey.PK, err = privacy.DecompressKey(tx.SigPubKey)
	// if err != nil {
	// 	return false, err
	// }

	// verKey.G = new(privacy.EllipticPoint)
	// verKey.G.Set(privacy.PedCom.G[privacy.sk].X, privacy.PedCom.G[privacy.sk].Y)

	// verKey.H = new(privacy.EllipticPoint)
	// verKey.H.Set(privacy.PedCom.G[privacy.RAND].X, privacy.PedCom.G[privacy.RAND].Y)

	// // convert signature from byte array to SchnorrSign
	// signature := new(privacy.SchnSignature)
	// signature.SetBytes(tx.Sig)

	// verify signature
	// Logger.log.Infof(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash().String())
	// res = verKey.Verify(signature, tx.Hash()[:])
	//fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", tx.Sig)
	if hasPrivacy {
		SNPrivacyProof := make([]*zkp.PKSNPrivacyProof, len(tx.Proof.InputCoins))
		for i := 0; i < len(tx.Proof.InputCoins); i++ {
			SNPrivacyProof[i] = new(zkp.PKSNPrivacyProof).Init()
			SNPrivacyProof[i].SetBytes(tx.Sig[i*privacy.SNPrivacyProofSize : (i+1)*privacy.SNPrivacyProofSize])
			if !SNPrivacyProof[i].Verify(tx.Hash()[:]) {
				//fmt.Println("Falseeeeeeeeeeeeeeeeeeeeeeeee")
				return false, nil
			}
		}
	} else {
		SNPrivacyProof := make([]*zkp.SNNoPrivacyProof, len(tx.Proof.InputCoins))
		for i := 0; i < len(tx.Proof.InputCoins); i++ {
			SNPrivacyProof[i] = new(zkp.SNNoPrivacyProof).Init()
			SNPrivacyProof[i].SetBytes(tx.Sig[i*privacy.SNNoPrivacyProofSize : (i+1)*privacy.SNNoPrivacyProofSize])
			if !SNPrivacyProof[i].Verify(tx.Hash()[:]) {
				//fmt.Println("Falseeeeeeeeeeeeeeeeeeeeeeeee")
				return false, nil
			}
		}
	}
	// SNPrivacyProof.SetBytes(tx.Sig)
	// res = SNPrivacyProof.Verify(tx.Hash()[:])
	return res, nil
}

func (tx *Tx) verifyMultiSigsTx(db database.DatabaseInterface) (bool, error) {
	meta := tx.GetMetadata()
	if meta == nil {
		return false, nil
	}
	if meta.GetType() != metadata.MultiSigsSpendingMeta {
		return false, nil
	}
	return meta.VerifyMultiSigs(tx, db)
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
// - Check double spendingComInputOpeningsWitnessval
func (tx *Tx) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, chainId byte, tokenID *common.Hash) bool {
	//hasPrivacy = falsenite
	Logger.log.Infof("[db] Validating Transaction tx\n")
	start := time.Now()
	// Verify tx signature
	Logger.log.Infof("tx.GetType(): %v\n", tx.GetType())
	if tx.GetType() == common.TxSalaryType {
		return tx.ValidateTxSalary(db)
	}

	var valid bool
	var err error

	valid, err = tx.verifySigTx(hasPrivacy)
	if !valid {
		if err != nil {
			Logger.log.Infof("[PRIVACY LOG] - Error verifying signature of tx: %+v", err)
		}
		Logger.log.Infof("[PRIVACY LOG] - FAILED VERIFICATION SIGNATURE")
		return false
	}

	senderPK := tx.GetSigPubKey()
	multisigsRegBytes, getMSRErr := db.GetMultiSigsRegistration(senderPK)
	if getMSRErr != nil {
		Logger.log.Infof("getMSRErr: %v\n", getMSRErr)
		return false
	}

	// found, spending on multisigs address
	// Multi signatures
	if len(multisigsRegBytes) > 0 {
		valid, err = tx.verifyMultiSigsTx(db)
		if err != nil {
			Logger.log.Infof("%+v", err)
		}
		if !valid {
			return false
		}
	}

	Logger.log.Infof("[db]tx.Proof: %+v\n", tx.Proof)
	if tx.Proof != nil {
		if tokenID == nil {
			tokenID = &common.Hash{}
			tokenID.SetBytes(common.ConstantID[:])
		}
		for i := 0; i < len(tx.Proof.OutputCoins); i++ {
			// Check output coins' input is not exists in input list (Database)
			if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.OutputCoins[i].CoinDetails.SNDerivator, chainId, db); ok || err != nil {
				Logger.log.Infof("snd existed: %d\n", i)
				return false
			}
		}
		if !hasPrivacy {
			// Check input coins' cm is exists in cm list (Database)
			for i := 0; i < len(tx.Proof.InputCoins); i++ {
				ok, err := tx.CheckCMExistence(tx.Proof.InputCoins[i].CoinDetails.CoinCommitment.Compress(), db, chainId, tokenID)
				if !ok || err != nil {
					Logger.log.Infof("[db]cm existed: %d\n", i)
					return false
				}
			}
		}

		// Verify the payment proof
		valid = tx.Proof.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, db, chainId, tokenID)
		Logger.log.Infof("proof valid: %v\n", valid)
		if !valid {
			Logger.log.Infof("[PRIVACY LOG] - FAILED VERIFICATION PAYMENT PROOF")
			return false
		}
	}
	elapsed := time.Since(start)
	Logger.log.Infof("Validation normal tx time %s", elapsed)

	return true
}

func (tx Tx) String() string {
	record := strconv.Itoa(int(tx.Version))
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		record += string(tx.Proof.Bytes()[:])
	}
	if tx.Metadata != nil {
		metadata := tx.Metadata.Hash().String()
		record += metadata
	}
	return record
}

func (tx *Tx) Hash() *common.Hash {
	hash := common.DoubleHashH([]byte(tx.String()))
	return &hash
}

func (tx *Tx) GetSenderAddrLastByte() byte {
	return tx.PubKeyLastByteSender
}

func (tx *Tx) GetTxFee() uint64 {
	return tx.Fee
}

// GetTxActualSize computes the actual size of a given transaction in kilobyte
func (tx *Tx) GetTxActualSize() uint64 {
	sizeTx := uint64(1)                // int8
	sizeTx += uint64(len(tx.Type) + 1) // string
	sizeTx += uint64(8)                // int64
	sizeTx += uint64(8)

	sizeTx += uint64(len(tx.SigPubKey))
	sizeTx += uint64(len(tx.Sig))
	if tx.Proof != nil {
		sizeTx += uint64(len(tx.Proof.Bytes()))
	}

	sizeTx += uint64(1)
	sizeTx += uint64(len(tx.Info))

	if tx.Metadata != nil {
		// TODO 0xjackpolope
		_ = 1
	}

	return uint64(math.Ceil(float64(sizeTx) / 1024))
}

// GetType returns the type of the transaction
func (tx *Tx) GetType() string {
	return tx.Type
}

func (tx *Tx) ListNullifiers() [][]byte {
	result := [][]byte{}
	if tx.Proof != nil {
		for _, d := range tx.Proof.InputCoins {
			result = append(result, d.CoinDetails.SerialNumber.Compress())
		}
	}
	return result
}

// CheckCMExistence returns true if cm exists in cm list
func (tx Tx) CheckCMExistence(cm []byte, db database.DatabaseInterface, chainID byte, tokenID *common.Hash) (bool, error) {
	ok, err := db.HasCommitment(tokenID, cm, chainID)
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
	return !(tx.Fee < fullFee)
}

func (tx *Tx) IsSalaryTx() bool {
	// Check normal tx(not an action tx)
	if tx.GetType() != common.TxSalaryType {
		return false
	}
	// Check nullifiers in every Descs
	if len(tx.Proof.InputCoins) == 0 {
		return true
	}
	return false
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
	sender := tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
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

func (tx *Tx) validateDoubleSpendTxWithCurrentMempool(poolNullifiers map[common.Hash][][]byte) error {
	if tx.Proof == nil {
		return nil
	}
	for _, temp1 := range poolNullifiers {
		for _, desc := range tx.Proof.InputCoins {
			if ok, err := common.SliceBytesExists(temp1, desc.CoinDetails.SerialNumber.Compress()); ok > -1 || err != nil {
				return errors.New("double spend")
			}
		}
	}
	return nil
}

func (tx *Tx) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	if tx.Type == common.TxSalaryType {
		return errors.New("can not receive a salary tx from other node, this is a violation")
	}
	poolNullifiers := mr.GetSerialNumbers()
	return tx.validateDoubleSpendTxWithCurrentMempool(poolNullifiers)
}

// ValidateDoubleSpend - check double spend for any transaction type
func (tx *Tx) ValidateConstDoubleSpendWithBlockchain(
	bcr metadata.BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) error {

	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	for i := 0; tx.Proof != nil && i < len(tx.Proof.InputCoins); i++ {
		serialNumber := tx.Proof.InputCoins[i].CoinDetails.SerialNumber.Compress()
		ok, err := db.HasSerialNumber(constantTokenID, serialNumber, chainID)
		if ok || err != nil {
			return errors.New("double spend")
		}
	}
	return nil
}

func (tx *Tx) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) error {
	if tx.GetType() == common.TxSalaryType {
		return nil
	}
	if tx.Metadata != nil {
		isContinued, err := tx.Metadata.ValidateTxWithBlockChain(tx, bcr, chainID, db)
		if err != nil {
			return err
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateConstDoubleSpendWithBlockchain(bcr, chainID, db)
}

func (tx *Tx) validateNormalTxSanityData() (bool, error) {
	txN := tx
	//check version
	if txN.Version > TxVersion {
		return false, errors.New("wrong tx version")
	}
	// check LockTime before now
	if int64(txN.LockTime) > time.Now().Unix() {
		return false, errors.New("wrong tx locktime")
	}
	// check Type is normal or salary tx
	/*if len(txN.Type) != 1 || (txN.Type != common.TxNormalType && txN.Type != common.TxSalaryType) { // only 1 byte
		return false, errors.New("Wrong tx type")
	}*/

	return true, nil
}

func (tx *Tx) ValidateSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
	Logger.log.Info("Validating sanity data!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", tx.Metadata)
	if tx.Metadata != nil {
		isContinued, ok, err := tx.Metadata.ValidateSanityData(bcr, tx)
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return tx.validateNormalTxSanityData()
}

func (tx *Tx) ValidateTxByItself(
	hasPrivacy bool,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	chainID byte,
) bool {
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	ok := tx.ValidateTransaction(hasPrivacy, db, chainID, constantTokenID)
	Logger.log.Infof("[db]ok validatetxbyitself: %v\n", ok)
	if !ok {
		return false
	}
	if tx.Metadata != nil {
		if hasPrivacy {
			Logger.log.Infof("[db]validatetxbyitself metadata: Transaction with metadata should not enable privacy feature.")
			return false
		}
		Logger.log.Infof("[db]validatetxbyitself metadata: %v\n", tx.Metadata.ValidateMetadataByItself())
		return tx.Metadata.ValidateMetadataByItself()
	}
	return true
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
	return tx.Type == common.TxNormalType || tx.Type == common.TxSalaryType
}

func (tx *Tx) IsCoinsBurning() bool {
	if tx.Proof == nil || len(tx.Proof.InputCoins) == 0 || len(tx.Proof.OutputCoins) == 0 {
		return false
	}
	senderPKBytes := tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
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
	if tx.Proof == nil || len(tx.Proof.InputCoins) == 0 || len(tx.Proof.OutputCoins) == 0 {
		return 0
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

func (tx *Tx) GetSenderAddress() *privacy.PaymentAddress {
	meta := tx.GetMetadata()
	if meta == nil {
		return nil
	}
	if meta.GetType() != metadata.WithSenderAddressMeta {
		return nil
	}
	withSenderAddrMeta, ok := meta.(*metadata.WithSenderAddress)
	if !ok {
		return nil
	}
	return &withSenderAddrMeta.SenderAddress
}

func NewEmptyTx(minerPrivateKey *privacy.SpendingKey, db database.DatabaseInterface, meta metadata.Metadata) metadata.Transaction {
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
	privKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	metaData metadata.Metadata,
) error {
	tx.Version = TxVersion
	tx.Type = common.TxSalaryType

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().UnixNano()
	}

	var err error
	// create new output coins with info: Pk, value, input, randomness, last byte pk, coin commitment
	tx.Proof = new(zkp.PaymentProof)
	tx.Proof.OutputCoins = make([]*privacy.OutputCoin, 1)
	tx.Proof.OutputCoins[0] = new(privacy.OutputCoin)
	//tx.Proof.OutputCoins[0].CoinDetailsEncrypted = new(privacy.CoinDetailsEncrypted).Init()
	tx.Proof.OutputCoins[0].CoinDetails = new(privacy.Coin)
	tx.Proof.OutputCoins[0].CoinDetails.Value = salary
	tx.Proof.OutputCoins[0].CoinDetails.PublicKey, err = privacy.DecompressKey(receiverAddr.Pk)
	if err != nil {
		return err
	}
	tx.Proof.OutputCoins[0].CoinDetails.Randomness = privacy.RandInt()

	sndOut := privacy.RandInt()
	for {
		lastByte := receiverAddr.Pk[len(receiverAddr.Pk)-1]
		chainIdSender, err := common.GetTxSenderChain(lastByte)
		if err != nil {
			return err
		}
		tokenID := &common.Hash{}
		tokenID.SetBytes(common.ConstantID[:])
		ok, err := CheckSNDerivatorExistence(tokenID, sndOut, chainIdSender, db)
		if err != nil {
			return err
		}
		if ok {
			sndOut = privacy.RandInt()
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
	// tx.sigPrivKey = *privKey
	tx.SetMetadata(metaData)
	err = tx.signTx(false)
	if err != nil {
		return err
	}

	return nil
}

func (tx Tx) ValidateTxSalary(
	db database.DatabaseInterface,
) bool {
	// verify signature
	valid, err := tx.verifySigTx(false)
	if !valid {
		if err != nil {
			Logger.log.Infof("Error verifying signature of tx: %+v", err)
		}
		return false
	}

	// check whether output coin's input exists in input list or not
	lastByte := tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()[len(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress())-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	if err != nil {
		return false
	}
	tokenID := &common.Hash{}
	tokenID.SetBytes(common.ConstantID[:])
	if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.OutputCoins[0].CoinDetails.SNDerivator, chainIdSender, db); ok || err != nil {
		return false
	}

	// check output coin's coin commitment is calculated correctly
	cmTmp := tx.Proof.OutputCoins[0].CoinDetails.PublicKey
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMult(big.NewInt(int64(tx.Proof.OutputCoins[0].CoinDetails.Value))))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMult(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMult(new(big.Int).SetBytes([]byte{tx.Proof.OutputCoins[0].CoinDetails.GetPubKeyLastByte()})))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMult(tx.Proof.OutputCoins[0].CoinDetails.Randomness))
	return cmTmp.IsEqual(tx.Proof.OutputCoins[0].CoinDetails.CoinCommitment)
}
