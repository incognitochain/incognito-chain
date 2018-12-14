package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
)

type Tx struct {
	// Basic data
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant

	// Sign and Privacy proof
	SigPubKey []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig       []byte `json:"Sig, omitempty"`       // 64 bytes
	Proof     *zkp.PaymentProof

	// Metadata
	Metadata metadata.Metadata

	sigPrivKey []byte // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes
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
) error {

	if len(inputCoins) == 0 && fee == 0 && !hasPrivacy {
		fmt.Printf("CREATE TX CUSTOM TOKEN\n")
		tx.Fee = fee
		tx.sigPrivKey = *senderSK

		err := tx.SignTx(hasPrivacy)
		if err != nil {
			return err
		}
		return nil
	}

	tx.Type = common.TxNormalType
	chainID := byte(14)
	var commitmentIndexs []uint64   // array index random of commitments in db
	var myCommitmentIndexs []uint64 // index in array index random of commitment in db

	commitmentIndexs, myCommitmentIndexs = RandomCommitmentsProcess(inputCoins, 8, db, chainID)

	// Print list of all input coins
	fmt.Printf("List of all input coins before building tx:\n")
	for _, coin := range inputCoins {
		fmt.Printf("%+v\n", coin)
	}

	// Check number of list of random commitments, list of random commitment indices
	if len(commitmentIndexs) != len(inputCoins)*privacy.CMRingSize {
		return fmt.Errorf("Number of list commitments indices must be corresponding with number of input coins")
	}

	if len(myCommitmentIndexs) != len(inputCoins) {
		return fmt.Errorf("Number of list my commitment indices must be equal to number of input coins")
	}

	// Calculate sum of all output coins' value
	var sumOutputValue uint64
	sumOutputValue = 0
	for _, p := range paymentInfo {
		sumOutputValue += p.Amount
		fmt.Printf("[CreateTx] paymentInfo.H: %+v, paymentInfo.PaymentAddress: %x\n", p.Amount, p.PaymentAddress.Pk)
	}

	// Calculate sum of all input coins' value
	var sumInputValue uint64
	sumInputValue = 0
	for _, coin := range inputCoins {
		sumInputValue += coin.CoinDetails.Value
	}

	// Calculate over balance, it will be returned to sender
	overBalance := sumInputValue - sumOutputValue - fee

	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return fmt.Errorf("Input value less than output value")
	}

	// create sender's key set from sender's spending key
	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKey(senderSK)

	// if overBalance > 0, create a new payment info with pk is sender's pk and amount is overBalance
	if overBalance > 0 {
		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = overBalance
		changePaymentInfo.PaymentAddress = senderFullKey.PaymentAddress
		paymentInfo = append(paymentInfo, changePaymentInfo)
	}

	// calculate serial number from SND and spending key
	for _, inputCoin := range inputCoins {
		inputCoin.CoinDetails.SerialNumber = privacy.Eval(new(big.Int).SetBytes(*senderSK), inputCoin.CoinDetails.SNDerivator)
	}

	// create new output coins
	outputCoins := make([]*privacy.OutputCoin, len(paymentInfo))

	// create SNDs for output coins
	ok := true
	sndOuts := make([]*big.Int, 0)
	for ok {
		sndOut := new(big.Int)
		for i := 0; i < len(paymentInfo); i++ {
			sndOut = privacy.RandInt()
			for true {
				ok1, err := CheckSNDerivatorExistence(sndOut, chainID, db)
				if err != nil {
					fmt.Println(err)
				}
				if ok1 {
					sndOut = privacy.RandInt()
				} else {
					break
				}
			}
			sndOuts = append(sndOuts, sndOut)
		}

		ok = common.CheckDuplicateBigInt(sndOuts)
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

	// get public key last byte of sender
	pkLastByteSender := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]
	tx.Proof = &zkp.PaymentProof{}
	tx.Proof.PubKeyLastByteSender = pkLastByteSender

	// get public key last byte of receivers
	pkLastByteReceivers := make([]byte, len(paymentInfo))
	for i, payInfo := range paymentInfo {
		pkLastByteReceivers[i] = payInfo.PaymentAddress.Pk[len(payInfo.PaymentAddress.Pk)-1]
	}

	// create zero knowledge proof of payment

	// get list of commitments for proving one-out-of-many from commitmentIndexs
	commitmentProving := make([]*privacy.EllipticPoint, len(commitmentIndexs))
	for i, cmIndex := range commitmentIndexs {
		commitmentProving[i] = new(privacy.EllipticPoint)
		temp, _ := db.GetCommitmentByIndex(cmIndex, chainID)
		commitmentProving[i], _ = privacy.DecompressKey(temp)
	}

	// prepare witness for proving
	witness := new(zkp.PaymentWitness)
	err := witness.Build(hasPrivacy, new(big.Int).SetBytes(*senderSK), inputCoins, outputCoins, pkLastByteSender, pkLastByteReceivers, commitmentProving, commitmentIndexs, myCommitmentIndexs, fee)
	if err != nil {
		return err
	}
	tx.Proof, err = witness.Prove(hasPrivacy)
	if err != nil {
		return err
	}

	// set private key for signing tx
	if hasPrivacy {
		tx.sigPrivKey = make([]byte, 64)
		openings := witness.ComInputOpeningsWitness[0].Openings
		lastOpening := openings[len(openings)-1]
		tx.sigPrivKey = append(*senderSK, lastOpening.Bytes()...)

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
		tx.sigPrivKey = *senderSK
	}

	// sign tx
	err = tx.SignTx(hasPrivacy)

	return err
}

// SignTx - signs tx
func (tx *Tx) SignTx(hasPrivacy bool) error {
	//Check input transaction
	if tx.Sig != nil {
		return fmt.Errorf("input transaction must be an unsigned one")
	}

	if hasPrivacy {
		/****** using Schnorr *******/
		// sign with sigPrivKey
		// prepare private key for Schnorr
		sigKey := new(privacy.SchnPrivKey)
		sigKey.SK = new(big.Int).SetBytes(tx.sigPrivKey[:32])
		sigKey.R = new(big.Int).SetBytes(tx.sigPrivKey[32:])

		// save public key for verification signature tx
		sigKey.PubKey = new(privacy.SchnPubKey)
		sigKey.PubKey.G = new(privacy.EllipticPoint)
		sigKey.PubKey.G.X, sigKey.PubKey.G.Y = privacy.Curve.Params().Gx, privacy.Curve.Params().Gy

		sigKey.PubKey.H = new(privacy.EllipticPoint)
		sigKey.PubKey.H.X, sigKey.PubKey.H.Y = privacy.PedCom.G[privacy.RAND].X, privacy.PedCom.G[privacy.RAND].Y

		sigKey.PubKey.PK = &privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
		tmp := new(privacy.EllipticPoint)
		tmp.X, tmp.Y = privacy.Curve.ScalarMult(sigKey.PubKey.G.X, sigKey.PubKey.G.Y, sigKey.SK.Bytes())
		sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y = privacy.Curve.Add(sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y, tmp.X, tmp.Y)
		tmp.X, tmp.Y = privacy.Curve.ScalarMult(sigKey.PubKey.H.X, sigKey.PubKey.H.Y, sigKey.R.Bytes())
		sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y = privacy.Curve.Add(sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y, tmp.X, tmp.Y)
		tx.SigPubKey = sigKey.PubKey.PK.Compress()

		// signing
		signature, err := sigKey.Sign(tx.Hash()[:])
		if err != nil {
			return err
		}

		// convert signature to byte array
		tx.Sig = signature.ToBytes()

	} else {
		/***** using ECDSA ****/
		// sign with sigPrivKey
		// prepare private key for ECDSA
		sigKey := new(ecdsa.PrivateKey)
		sigKey.PublicKey.Curve = privacy.Curve
		sigKey.D = new(big.Int).SetBytes(tx.sigPrivKey)
		sigKey.PublicKey.X, sigKey.PublicKey.Y = privacy.Curve.ScalarBaseMult(tx.sigPrivKey)

		// save public key for verification signature tx
		verKey := new(privacy.EllipticPoint)
		verKey.X, verKey.Y = sigKey.PublicKey.X, sigKey.PublicKey.Y
		tx.SigPubKey = verKey.Compress()

		// signing
		r, s, err := ecdsa.Sign(rand.Reader, sigKey, tx.Hash()[:])
		if err != nil {
			return err
		}

		// convert signature to byte array
		tx.Sig = common.ECDSASigToByteArray(r, s)
	}

	return nil
}

func (tx *Tx) VerifySigTx(hasPrivacy bool) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, fmt.Errorf("input transaction must be an signed one!")
	}

	var err error
	res := false

	if hasPrivacy {
		/****** verify Schnorr signature *****/
		// prepare Public key for verification
		verKey := new(privacy.SchnPubKey)
		verKey.PK, err = privacy.DecompressKey(tx.SigPubKey)
		if err != nil {
			return false, err
		}
		verKey.G = new(privacy.EllipticPoint)
		verKey.G.X, verKey.G.Y = privacy.Curve.Params().Gx, privacy.Curve.Params().Gy

		verKey.H = new(privacy.EllipticPoint)
		verKey.H.X, verKey.H.Y = privacy.PedCom.G[privacy.RAND].X, privacy.PedCom.G[privacy.RAND].Y

		// convert signature from byte array to SchnorrSign
		signature := new(privacy.SchnSignature)
		signature.FromBytes(tx.Sig)

		// verify signature
		res = verKey.Verify(signature, tx.Hash()[:])

	} else {
		/****** verify ECDSA signature *****/
		// prepare Public key for verification
		verKey := new(ecdsa.PublicKey)
		point := new(privacy.EllipticPoint)
		point, _ = privacy.DecompressKey(tx.SigPubKey)
		verKey.X, verKey.Y = point.X, point.Y
		verKey.Curve = privacy.Curve

		// convert signature from byte array to ECDSASign
		r, s := common.FromByteArrayToECDSASig(tx.Sig)

		// verify signature
		res = ecdsa.Verify(verKey, tx.Hash()[:], r, s)
	}

	return res, nil
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
// - Check double spending
func (tx *Tx) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, chainId byte) bool {
	// Verify tx signature
	var valid bool
	var err error
	valid, err = tx.VerifySigTx(hasPrivacy)
	if valid == false {
		if err != nil {
			fmt.Printf("Error verifying signature of tx: %+v", err)
		}
		return false
	}

	for i := 0; i < len(tx.Proof.OutputCoins); i++ {
		// Check output coins' SND is not exists in SND list (Database)
		if ok, err := CheckSNDerivatorExistence(tx.Proof.OutputCoins[i].CoinDetails.SNDerivator, chainId, db); ok || err != nil {
			return false
		}
	}

	if !hasPrivacy {
		// Check input coins' cm is exists in cm list (Database)
		for i := 0; i < len(tx.Proof.InputCoins); i++ {
			ok, err := tx.CheckCMExistence(tx.Proof.InputCoins[i].CoinDetails.CoinCommitment.Compress(), db, chainId)
			if !ok || err != nil {
				return false
			}
		}
	}

	// Verify the payment proof
	valid = tx.Proof.Verify(hasPrivacy, tx.SigPubKey, db, chainId)
	if valid == false {
		fmt.Printf("Error verifying the payment proof")
		return false
	}

	return true
}

func (tx *Tx) Hash() *common.Hash {
	record := strconv.Itoa(int(tx.Version))
	record += tx.Type
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		record += string(tx.Proof.Bytes()[:])
	}
	if tx.Metadata != nil {
		record += string(tx.Metadata.Hash()[:])
	}
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *Tx) GetSenderAddrLastByte() byte {
	return tx.Proof.PubKeyLastByteSender
}

func (tx *Tx) GetTxFee() uint64 {
	return tx.Fee
}

// GetTxVirtualSize computes the virtual size of a given transaction
func (tx *Tx) GetTxVirtualSize() uint64 {
	// TODO 0xkraken
	return 0
}

// GetType returns the type of the transaction
func (tx *Tx) GetType() string {
	return tx.Type
}

func (tx *Tx) ListNullifiers() [][]byte {
	result := [][]byte{}
	for _, d := range tx.Proof.InputCoins {
		result = append(result, d.CoinDetails.SerialNumber.Compress())
	}
	return result
}

// EstimateTxSize returns the estimated size of the tx in kilobyte
func EstimateTxSize(usableTx []*Tx, payments []*privacy.PaymentInfo) uint64 {
	var sizeVersion uint64 = 1  // int8
	var sizeType uint64 = 8     // string
	var sizeLockTime uint64 = 8 // int64
	var sizeFee uint64 = 8      // uint64
	var sizeDescs uint64        // uint64
	if payments != nil {
		sizeDescs = uint64(common.Max(1, (len(usableTx)+len(payments)-3))) * EstimateJSDescSize()
	} else {
		sizeDescs = uint64(common.Max(1, (len(usableTx)-3))) * EstimateJSDescSize()
	}
	var sizejSPubKey uint64 = 64 // [64]byte
	var sizejSSig uint64 = 64    // [64]byte
	estimateTxSizeInByte := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeDescs + sizejSPubKey + sizejSSig
	return uint64(math.Ceil(float64(estimateTxSizeInByte) / 1024))
}

// CheckCMExistence returns true if cm exists in cm list
func (tx Tx) CheckCMExistence(cm []byte, db database.DatabaseInterface, chainID byte) (bool, error) {
	ok, err := db.HasCommitment(cm, chainID)
	return ok, err
}

// /////////////// merging ----------------------
// func (tx *Tx) SetTxID(txId *common.Hash) {
// 	tx.txId = txId
// }

// func (tx *Tx) GetTxID() *common.Hash {
// 	return tx.txId
// }

func (tx *Tx) CheckTxVersion(maxTxVersion int8) bool {
	if tx.Version > maxTxVersion {
		return false
	}
	return true
}

func (tx *Tx) CheckTransactionFee(minFee uint64) bool {
	if tx.IsSalaryTx() {
		return true
	}
	if tx.Metadata != nil {
		return tx.Metadata.CheckTransactionFee(tx, minFee)
	}
	if tx.Fee < minFee {
		return false
	}
	return true
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
	return pubkeys, amounts

	// TODO: @bunyip - update logic here

	// for _, desc := range tx.Descs {
	// 	for _, note := range desc.Note {
	// 		added := false
	// 		for i, key := range pubkeys {
	// 			if bytes.Equal(note.Apk[:], key) {
	// 				added = true
	// 				amounts[i] += note.Value
	// 			}
	// 		}
	// 		if !added {
	// 			pubkeys = append(pubkeys, note.Apk[:])
	// 			amounts = append(amounts, note.Value)
	// 		}
	// 	}
	// }
	// return pubkeys, amounts
}

func (tx *Tx) validateDoubleSpendTxWithCurrentMempool(poolNullifiers map[common.Hash][][]byte) error {
	for _, temp1 := range poolNullifiers {
		for _, desc := range tx.Proof.InputCoins {
			if ok, err := common.SliceBytesExists(temp1, desc.CoinDetails.SerialNumber.Compress()); ok == -1 || err != nil {
				return errors.New("Double spend")
			}
		}
	}
	return nil
}

func (tx *Tx) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	if tx.Type == common.TxSalaryType {
		return errors.New("Can not receive a salary tx from other node, this is a violation")
	}
	poolNullifiers := mr.GetPoolNullifiers()
	return tx.validateDoubleSpendTxWithCurrentMempool(poolNullifiers)
}

// ValidateDoubleSpend - check double spend for any transaction type
func (tx *Tx) ValidateConstDoubleSpendWithBlockchain(
	bcr metadata.BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) error {
	// txHash := tx.Hash()
	// nullifierDb, err := bcr.GetNulltifiersList(chainID)
	// if err != nil {
	// 	return errors.New(fmt.Sprintf("Can not check double spend for tx"))
	// }

	// ins := tx.Proof.InputCoins
	// for _, in := range ins {
	// 	existed, err := common.SliceBytesExists(nullifierDb, in.CoinDetails.SerialNumber.Compress())
	// 	if err != nil {
	// 		return errors.New(fmt.Sprintf("Can not check double spend for tx"))
	// 	}
	// 	if existed != -1 {
	// 		return errors.New(fmt.Sprintf("Nullifiers of transaction %+v already existed", txHash.String()))
	// 	}
	// }
	// return nil

	for i := 0; i < len(tx.Proof.InputCoins); i++ {
		serialNumber := tx.Proof.InputCoins[i].CoinDetails.SerialNumber.Compress()
		ok, err := db.HasSerialNumber(serialNumber, chainID)
		if ok || err != nil {
			return errors.New("Double spend")
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
		return false, errors.New("Wrong tx version")
	}
	// check LockTime before now
	if int64(txN.LockTime) > time.Now().Unix() {
		return false, errors.New("Wrong tx locktime")
	}
	// check Type is normal or salary tx
	if len(txN.Type) != 1 || (txN.Type != common.TxNormalType && txN.Type != common.TxSalaryType) { // only 1 byte
		return false, errors.New("Wrong tx type")
	}

	return true, nil
}

func (tx *Tx) ValidateSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
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
	ok := tx.ValidateTransaction(hasPrivacy, db, chainID)
	if !ok {
		return false
	}
	if tx.Metadata != nil {
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

func (tx *Tx) CalculateTxValue() (*privacy.PaymentAddress, uint64) {
	// TODO: 0xankylosaurus - update here
	return nil, 0

	// initiatorPubKey := tx.JSPubKey
	// txValue := uint64(0)
	// var addr *privacy.PaymentAddress
	// for _, desc := range tx.Descs {
	// 	for _, note := range desc.Note {
	// 		if string(note.Apk[:]) == string(initiatorPubKey) {
	// 			continue
	// 		}
	// 		addr = &privacy.PaymentAddress{
	// 			Pk: note.Apk,
	// 		}
	// 		txValue += note.Value
	// 	}
	// }
	// return addr, txValue
}

func (tx *Tx) GetJSPubKey() []byte {
	return tx.SigPubKey
}

func (tx *Tx) IsPrivacy() bool {
	// TODO: update here
	return false
}
