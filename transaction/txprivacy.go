package transaction

import (
	"bytes"
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
	"github.com/ninjadotorg/constant/wallet"
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

	PubKeyLastByteSender byte

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
	tokenID *common.Hash, // default is nill -> use for constant coin
) *TrasactionError {
	if tokenID == nil {
		tokenID = &common.Hash{}
		tokenID.SetBytes(common.ConstantID[:])
	}

	// create sender's key set from sender's spending key
	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKey(senderSK)
	// get public key last byte of sender
	pkLastByteSender := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]

	if len(inputCoins) == 0 && fee == 0 && !hasPrivacy {
		Logger.log.Infof("CREATE TX CUSTOM TOKEN\n")
		tx.Fee = fee
		tx.sigPrivKey = *senderSK
		tx.PubKeyLastByteSender = pkLastByteSender

		err := tx.SignTx(hasPrivacy)
		if err != nil {
			return NewTransactionErr(UnexpectedErr, err)
		}
		return nil
	}

	tx.Type = common.TxNormalType
	chainID, _ := common.GetTxSenderChain(pkLastByteSender)
	var commitmentIndexs []uint64   // array index random of commitments in db
	var myCommitmentIndexs []uint64 // index in array index random of commitment in db

	commitmentIndexs, myCommitmentIndexs = RandomCommitmentsProcess(inputCoins, 8, db, chainID, tokenID)

	// Print list of all input coins
	//Logger.log.Infof("List of all input coins before building tx:\n")
	//for _, coin := range inputCoins {
	//	Logger.log.Infof("%+v\n", coin)
	//}

	// Check number of list of random commitments, list of random commitment indices
	if len(commitmentIndexs) != len(inputCoins)*privacy.CMRingSize {
		return NewTransactionErr(UnexpectedErr, errors.New("Number of list commitments indices must be corresponding with number of input coins"))
	}

	if len(myCommitmentIndexs) != len(inputCoins) {
		return NewTransactionErr(UnexpectedErr, errors.New("Number of list my commitment indices must be equal to number of input coins"))
	}

	// Calculate sum of all output coins' value
	var sumOutputValue uint64
	sumOutputValue = 0
	for _, p := range paymentInfo {
		sumOutputValue += p.Amount
		//Logger.log.Infof("[CreateTx] paymentInfo.Value: %+v, paymentInfo.PaymentAddress: %x\n", p.Amount, p.PaymentAddress.Pk)
	}

	// Calculate sum of all input coins' value
	var sumInputValue uint64
	sumInputValue = 0
	for _, coin := range inputCoins {
		sumInputValue += coin.CoinDetails.Value
	}

	// Calculate over balance, it will be returned to sender
	overBalance := sumInputValue - sumOutputValue - fee

	valueMax := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(64)), nil)
	valueMax = valueMax.Sub(valueMax, big.NewInt(1))
	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 || overBalance > valueMax.Uint64() {
		return NewTransactionErr(UnexpectedErr, errors.New("Input value less than output value"))
	}

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
				ok1, err := CheckSNDerivatorExistence(tokenID, sndOut, chainID, db)
				if err != nil {
					Logger.log.Error(err)
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

	tx.Proof = &zkp.PaymentProof{}

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
		temp, _ := db.GetCommitmentByIndex(tokenID, cmIndex, chainID)
		commitmentProving[i], _ = privacy.DecompressKey(temp)
	}

	// check
	com := make([]*privacy.EllipticPoint, len(inputCoins))
	for i := 0; i < len(inputCoins); i++ {
		com[i] = new(privacy.EllipticPoint)
		com[i].X, com[i].Y = big.NewInt(0), big.NewInt(0)
		com[i].X.Set(inputCoins[i].CoinDetails.PublicKey.X)
		com[i].Y.Set(inputCoins[i].CoinDetails.PublicKey.Y)

		tmp := new(privacy.EllipticPoint)
		tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
		tmp.X.Set(privacy.PedCom.G[privacy.VALUE].X)
		tmp.Y.Set(privacy.PedCom.G[privacy.VALUE].Y)
		tmp = tmp.ScalarMul(new(big.Int).SetUint64(inputCoins[i].CoinDetails.Value))
		com[i] = com[i].Add(tmp)

		tmp = new(privacy.EllipticPoint)
		tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
		tmp.X.Set(privacy.PedCom.G[privacy.SND].X)
		tmp.Y.Set(privacy.PedCom.G[privacy.SND].Y)
		tmp = tmp.ScalarMul(inputCoins[i].CoinDetails.SNDerivator)
		com[i] = com[i].Add(tmp)

		tmp = new(privacy.EllipticPoint)
		tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
		tmp.X.Set(privacy.PedCom.G[privacy.SHARDID].X)
		tmp.Y.Set(privacy.PedCom.G[privacy.SHARDID].Y)
		tmp = tmp.ScalarMul(new(big.Int).SetBytes([]byte{inputCoins[i].CoinDetails.GetPubKeyLastByte()}))
		com[i] = com[i].Add(tmp)

		tmp = new(privacy.EllipticPoint)
		tmp.X, tmp.Y = big.NewInt(0), big.NewInt(0)
		tmp.X.Set(privacy.PedCom.G[privacy.RAND].X)
		tmp.Y.Set(privacy.PedCom.G[privacy.RAND].Y)
		tmp = tmp.ScalarMul(inputCoins[i].CoinDetails.Randomness)
		com[i] = com[i].Add(tmp)
		inputCoins[i].CoinDetails.CommitAll()
		if !com[i].IsEqual(commitmentProving[myCommitmentIndexs[i]]) {
			Logger.log.Infof("WRONG 1")
		} else {
			Logger.log.Infof("Right")
		}

		if !inputCoins[i].CoinDetails.CoinCommitment.IsEqual(commitmentProving[myCommitmentIndexs[i]]) {
			Logger.log.Infof("WRONG 2")
		} else {
			Logger.log.Infof("Right")
		}
		if !inputCoins[i].CoinDetails.CoinCommitment.IsEqual(com[i]) {
			Logger.log.Infof("WRONG 3")
		} else {
			Logger.log.Infof("Right")
		}

		//openingWitnessInputCoin := new(zkp.PKComOpeningsWitness)
		//openingWitnessInputCoin.Set(inputCoins[i].CoinDetails.CoinCommitment,
		//	[]*big.Int{, new(big.Int).SetUint64(inputCoins[i].CoinDetails.Value), inputCoins[i].CoinDetails.SNDerivator, big.NewInt(int64(wit.pkLastByteSender)), randInputSum[i]},
		//	[]byte{privacy.SK, privacy.VALUE, privacy.SND, privacy.SHARDID, privacy.RAND})
		//
		//openingProofHien, _ := openingWitnessHien.Prove()
		//Logger.log.Infof(openingProofHien.Verify())

		//inputCoins[i].CoinDetails.CommitAll()
		//if !com[i].IsEqual(inputCoins[i].CoinDetails.CoinCommitment){
		//	Logger.log.Infof("WRONG")
		//}

	}

	// prepare witness for proving
	witness := new(zkp.PaymentWitness)
	err := witness.Build(hasPrivacy, new(big.Int).SetBytes(*senderSK), inputCoins, outputCoins, pkLastByteSender, pkLastByteReceivers, commitmentProving, commitmentIndexs, myCommitmentIndexs, fee)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	tx.Proof, err = witness.Prove(hasPrivacy)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}

	// set private key for signing tx
	if hasPrivacy {
		tx.sigPrivKey = make([]byte, 64)
		randSK := witness.RandSK
		tx.sigPrivKey = append(*senderSK, randSK.Bytes()...)
		//gSK := privacy.PedCom.G[privacy.SK].ScalarMul(new(big.Int).SetBytes(*senderSK))
		//gRandSK := privacy.PedCom.G[privacy.RAND].ScalarMul(randSK)
		//pubKeyPoint := gSK.Add(gRandSK)
		//tx.SigPubKey = pubKeyPoint.Compress()

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
		//tx.SigPubKey = senderFullKey.PaymentAddress.Pk
	}

	// sign tx
	tx.PubKeyLastByteSender = pkLastByteSender
	err = tx.SignTx(hasPrivacy)

	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	return nil
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
		sigKey.PubKey.G.X, sigKey.PubKey.G.Y = privacy.PedCom.G[privacy.SK].X, privacy.PedCom.G[privacy.SK].Y

		sigKey.PubKey.H = new(privacy.EllipticPoint)
		sigKey.PubKey.H.X, sigKey.PubKey.H.Y = privacy.PedCom.G[privacy.RAND].X, privacy.PedCom.G[privacy.RAND].Y

		sigKey.PubKey.PK = &privacy.EllipticPoint{big.NewInt(0), big.NewInt(0)}
		//Logger.log.Infof(sigKey)
		tmp := new(privacy.EllipticPoint)
		tmp.X, tmp.Y = privacy.Curve.ScalarMult(sigKey.PubKey.G.X, sigKey.PubKey.G.Y, sigKey.SK.Bytes())
		sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y = privacy.Curve.Add(sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y, tmp.X, tmp.Y)

		tmp.X, tmp.Y = privacy.Curve.ScalarMult(sigKey.PubKey.H.X, sigKey.PubKey.H.Y, sigKey.R.Bytes())
		sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y = privacy.Curve.Add(sigKey.PubKey.PK.X, sigKey.PubKey.PK.Y, tmp.X, tmp.Y)
		//Logger.log.Infof("SIGN ------ PUBLICKEY: %+v\n", sigKey.PubKey.PK)
		tx.SigPubKey = sigKey.PubKey.PK.Compress()
		//Logger.log.Infof("SIGN ------ PUBLICKEY BYTE: %+v\n", tx.SigPubKey)

		// signing
		//Logger.log.Infof("SIGN ------ HASH TX: %+v\n", tx.Hash().String())
		//Logger.log.Infof(" SIGN SIGNATURE ----------- HASH: %v\n", tx.Hash().String())
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

	if tx.Proof != nil {
		//Logger.log.Infof("VERIFY SIGNATURE ------------- TX.PROOF: %v\n", tx.Proof.Bytes())
	}

	var err error
	res := false

	if hasPrivacy {
		/****** verify Schnorr signature *****/
		// prepare Public key for verification
		verKey := new(privacy.SchnPubKey)
		//Logger.log.Infof("VERIFY ------ PUBLICKEY BYTE: %+v\n", tx.SigPubKey)
		verKey.PK, err = privacy.DecompressKey(tx.SigPubKey)
		if err != nil {
			return false, err
		}
		//Logger.log.Infof("VERIFY ------ PUBLICKEY: %+v\n", verKey.PK)

		verKey.G = new(privacy.EllipticPoint)
		verKey.G.X, verKey.G.Y = privacy.PedCom.G[privacy.SK].X, privacy.PedCom.G[privacy.SK].Y

		verKey.H = new(privacy.EllipticPoint)
		verKey.H.X, verKey.H.Y = privacy.PedCom.G[privacy.RAND].X, privacy.PedCom.G[privacy.RAND].Y
		//Logger.log.Infof(verKey)
		// convert signature from byte array to SchnorrSign
		signature := new(privacy.SchnSignature)
		signature.FromBytes(tx.Sig)

		// verify signature
		//Logger.log.Infof(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash().String())
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
// - Check double spendingComInputOpeningsWitnessval
func (tx *Tx) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, chainId byte, tokenID *common.Hash) bool {
	// Verify tx signature
	if tx.GetType() == common.TxSalaryType {
		return tx.ValidateTxSalary(db)
	}
	var valid bool
	var err error
	valid, err = tx.VerifySigTx(hasPrivacy)
	if valid == false {
		if err != nil {
			Logger.log.Infof("[PRIVACY LOG] - Error verifying signature of tx: %+v", err)
		}
		Logger.log.Infof("[PRIVACY LOG] - FAILED VERIFICATION SIGNATURE")
		return false
	}

	if tx.Proof != nil {
		tokenID := &common.Hash{}
		tokenID.SetBytes(common.ConstantID[:])
		for i := 0; i < len(tx.Proof.OutputCoins); i++ {
			// Check output coins' SND is not exists in SND list (Database)
			if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.OutputCoins[i].CoinDetails.SNDerivator, chainId, db); ok || err != nil {
				return false
			}
		}

		if !hasPrivacy {
			// Check input coins' cm is exists in cm list (Database)
			for i := 0; i < len(tx.Proof.InputCoins); i++ {
				ok, err := tx.CheckCMExistence(tx.Proof.InputCoins[i].CoinDetails.CoinCommitment.Compress(), db, chainId, tokenID)
				if !ok || err != nil {
					return false
				}
			}
		}

		// Verify the payment proof
		valid = tx.Proof.Verify(hasPrivacy, tx.SigPubKey, db, chainId, tokenID)
		if valid == false {
			Logger.log.Infof("[PRIVACY LOG] - FAILED VERIFICATION PAYMENT PROOF")
			return false
		}
	}

	return true
}

func (tx *Tx) Hash() *common.Hash {
	record := strconv.Itoa(int(tx.Version))
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
	// TODO 0xjackpolope
	if tx.Metadata != nil {
		//
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
	if tx.Version > maxTxVersion {
		return false
	}
	return true
}

func (tx *Tx) CheckTransactionFee(minFeePerKbTx uint64) bool {
	if tx.IsSalaryTx() {
		return true
	}
	if tx.Metadata != nil {
		return tx.Metadata.CheckTransactionFee(tx, minFeePerKbTx)
	}
	fullFee := minFeePerKbTx * tx.GetTxActualSize()
	if tx.Fee < fullFee {
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
	if tx.Proof == nil {
		return nil
	}
	for _, temp1 := range poolNullifiers {
		for _, desc := range tx.Proof.InputCoins {
			if ok, err := common.SliceBytesExists(temp1, desc.CoinDetails.SerialNumber.Compress()); ok > -1 || err != nil {
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
	/*if len(txN.Type) != 1 || (txN.Type != common.TxNormalType && txN.Type != common.TxSalaryType) { // only 1 byte
		return false, errors.New("Wrong tx type")
	}*/

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
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	ok := tx.ValidateTransaction(hasPrivacy, db, chainID, constantTokenID)
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

func (tx *Tx) GetJSPubKey() []byte {
	result := []byte{}
	if len(tx.Proof.InputCoins) > 0 {
		pubkey := tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
		result = make([]byte, len(pubkey))
		copy(result, pubkey)
	}
	return result
}

func (tx *Tx) IsPrivacy() bool {
	switch tx.GetType() {
	case common.TxSalaryType:
		return false
	default:
		return true
	}
}

func (tx *Tx) ValidateType() bool {
	return tx.Type == common.TxNormalType || tx.Type == common.TxSalaryType
}

func (tx *Tx) IsCoinsBurning() bool {
	if tx.Proof == nil || len(tx.Proof.InputCoins) == 0 || len(tx.Proof.OutputCoins) == 0 {
		return false
	}
	senderPKBytes := tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	buringAcc, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	for _, outCoin := range tx.Proof.OutputCoins {
		outPKBytes := outCoin.CoinDetails.PublicKey.Compress()
		if !bytes.Equal(senderPKBytes, outPKBytes) && !bytes.Equal(outPKBytes, buringAcc.KeySet.PaymentAddress.Pk[:]) {
			return false
		}
	}
	return true
}

func (tx *Tx) CalculateTxValue() (*privacy.PaymentAddress, uint64) {
	if tx.Proof == nil || len(tx.Proof.InputCoins) == 0 || len(tx.Proof.OutputCoins) == 0 {
		return nil, 0
	}
	senderPKBytes := tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	senderAddr := &privacy.PaymentAddress{
		Pk: senderPKBytes,
	}
	txValue := uint64(0)
	for _, outCoin := range tx.Proof.OutputCoins {
		outPKBytes := outCoin.CoinDetails.PublicKey.Compress()
		if bytes.Equal(senderPKBytes, outPKBytes) {
			continue
		}
		txValue += outCoin.CoinDetails.Value
	}
	return senderAddr, txValue
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
) (error) {
	tx.Type = common.TxSalaryType

	var err error
	// create new output coins with info: Pk, value, SND, randomness, last byte pk, coin commitment
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
	for true {
		lastByte := receiverAddr.Pk[len(receiverAddr.Pk)-1]
		chainIdSender, err := common.GetTxSenderChain(lastByte)

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
	tx.sigPrivKey = *privKey
	err = tx.SignTx(false)
	if err != nil {
		return err
	}

	if len(tx.Proof.InputCoins) > 0 {
		Logger.log.Info(11111)
	}
	return nil
}

func (tx Tx) ValidateTxSalary(
	db database.DatabaseInterface,
) bool {
	// verify signature
	valid, err := tx.VerifySigTx(false)
	if valid == false {
		if err != nil {
			Logger.log.Infof("Error verifying signature of tx: %+v", err)
		}
		return false
	}

	// check whether output coin's SND exists in SND list or not
	lastByte := tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress()[len(tx.Proof.OutputCoins[0].CoinDetails.PublicKey.Compress())-1]
	chainIdSender, err := common.GetTxSenderChain(lastByte)
	tokenID := &common.Hash{}
	tokenID.SetBytes(common.ConstantID[:])
	if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.OutputCoins[0].CoinDetails.SNDerivator, chainIdSender, db); ok || err != nil {
		return false
	}

	// check output coin's coin commitment is calculated correctly
	cmTmp := tx.Proof.OutputCoins[0].CoinDetails.PublicKey
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.VALUE].ScalarMul(big.NewInt(int64(tx.Proof.OutputCoins[0].CoinDetails.Value))))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.SNDerivator))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.SHARDID].ScalarMul(new(big.Int).SetBytes([]byte{tx.Proof.OutputCoins[0].CoinDetails.GetPubKeyLastByte()})))
	cmTmp = cmTmp.Add(privacy.PedCom.G[privacy.RAND].ScalarMul(tx.Proof.OutputCoins[0].CoinDetails.Randomness))
	if !cmTmp.IsEqual(tx.Proof.OutputCoins[0].CoinDetails.CoinCommitment) {
		return false
	}

	return true
}
