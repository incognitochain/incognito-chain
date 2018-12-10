package transaction

import (
	"fmt"
	"math/big"
	"strconv"
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
	"math"
	rand2 "math/rand"
)

type Tx struct {
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant

	SigPubKey []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig       []byte `json:"Sig, omitempty"`       // 64 bytes
	Proof     *zkp.PaymentProof

	txId       *common.Hash
	sigPrivKey []byte // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes

	// this one is a hash id of requested tx
	// and is used inside response txs
	// so that we can determine pair of req/res txs
	// for example, BuySellRequestTx/BuySellResponseTx
	//RequestedTxID *common.Hash

	// temp variable to validate tx
	//snDerivators []*big.Int

	Metadata interface{}
}

// randomCommitmentsProcess - process list commitments and useable tx to create
// a list commitment random which be used to create a proof for new tx
func randomCommitmentsProcess(commitments [][]byte, useableTx []*Tx, randNum int) (commitmentIndexs []uint64, myCommitmentIndexs []uint64) {
	commitmentIndexs = []uint64{}
	myCommitmentIndexs = []uint64{}
	if randNum == 0 {
		randNum = 7
	}
	listCommitmentsInUsableTx := [][]byte{}
	mapIndexCommitmentsInUsableTx := make(map[string]uint64)
	for _, tx := range useableTx {
		for _, out := range tx.Proof.OutputCoins {
			listCommitmentsInUsableTx = append(listCommitmentsInUsableTx, out.CoinDetails.CoinCommitment.Compress())
			index, _ := common.SliceBytesExists(commitments, out.CoinDetails.CoinCommitment.Compress())
			mapIndexCommitmentsInUsableTx[string(out.CoinDetails.CoinCommitment.Compress())] = uint64(index)
		}
	}
	cpRandNum := (len(listCommitmentsInUsableTx) * randNum) - len(listCommitmentsInUsableTx)
	for i := 0; i < cpRandNum; i++ {
		for true {
			index := rand2.Int63n(int64(len(commitments)))
			choosenCommitment := commitments[index]
			if k, err := common.SliceBytesExists(listCommitmentsInUsableTx, choosenCommitment); k != -1 && err != nil {
				commitmentIndexs = append(commitmentIndexs, uint64(index))
			} else {
				continue
			}
		}
	}
	for _, temp := range listCommitmentsInUsableTx {
		key := string(temp)
		index := mapIndexCommitmentsInUsableTx[key]
		i := rand2.Int63n(int64(len(commitmentIndexs)))
		commitmentIndexs = append(commitmentIndexs[:i], append([]uint64{index}, commitmentIndexs[i:]...)...)
		myCommitmentIndexs = append(myCommitmentIndexs, uint64(i))
	}
	return nil, nil
}

func getInputCoins(usableTx []*Tx) []*privacy.InputCoin {
	var inputCoins []*privacy.InputCoin
	inCoin := new(privacy.InputCoin)

	for _, tx := range usableTx {
		for _, coin := range tx.Proof.OutputCoins {
			inCoin.CoinDetails = coin.CoinDetails
			inputCoins = append(inputCoins, inCoin)
		}
	}
	return inputCoins
}

func (tx *Tx) CreateTx(
	senderSK *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	usableTx []*Tx,
	fee uint64,
	commitmentsDB [][]byte,
	hasPrivacy bool,
) (error) {

	var commitmentIndexs []uint64   // array index random of commitments in db
	var myCommitmentIndexs []uint64 // index in array index random of commitment in db

	commitmentIndexs, myCommitmentIndexs = randomCommitmentsProcess(commitmentsDB, usableTx, 7)

	inputCoins := getInputCoins(usableTx)
	//Get input coins from usableTX

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

	// tx.proof.Input
	//tx.Proof = new(zkp.PaymentProof)
	//tx.Proof.InputCoins = inputCoins

	// create sender's key set from sender's spending key
	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKeyByte((*senderSK)[:])

	// if overBalance > 0, create a new payment info with pk is sender's pk and amount is overBalance
	if overBalance > 0 {
		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = overBalance
		changePaymentInfo.PaymentAddress = senderFullKey.PaymentAddress
		paymentInfo = append(paymentInfo, changePaymentInfo)
	}

	// create new output coins
	outputCoins := make([]*privacy.OutputCoin, len(paymentInfo))

	var sndOuts []*big.Int
	sndOut := new(big.Int)
	ok := true

	// create SNDs for output coins
	for ok {
		sndOuts := make([]*big.Int, len(paymentInfo))

		for i := 0; i < len(paymentInfo); i++ {
			sndOut = privacy.RandInt()
			for common.CheckSNDExistence(sndOut) {
				sndOut = privacy.RandInt()
			}
			sndOuts = append(sndOuts, sndOut)
		}

		ok = common.CheckDuplicateBigInt(sndOuts)
	}

	// create new output coins with info: Pk, value, last byte of pk, snd
	for i, pInfo := range paymentInfo {
		outputCoins[i] = new(privacy.OutputCoin)
		outputCoins[i].CoinDetails.Value = pInfo.Amount
		outputCoins[i].CoinDetails.PublicKey, _ = privacy.DecompressKey(pInfo.PaymentAddress.Pk)
		outputCoins[i].CoinDetails.PubKeyLastByte = pInfo.PaymentAddress.Pk[len(pInfo.PaymentAddress.Pk)-1]
		outputCoins[i].CoinDetails.SNDerivator = sndOuts[i]
	}

	// assign fee tx
	tx.Fee = fee

	// get public key last byte of sender
	pkLastByteSender := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]
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
		commitmentProving[i], _ = privacy.DecompressKey(commitmentsDB[cmIndex])
	}

	// prepare witness for proving
	witness := new(zkp.PaymentWitness)
	witness.Build(hasPrivacy, new(big.Int).SetBytes(*senderSK), inputCoins, outputCoins, pkLastByteSender, pkLastByteReceivers, commitmentProving, commitmentIndexs, myCommitmentIndexs)
	tx.Proof, _ = witness.Prove(false)

	// set private key for signing tx
	if hasPrivacy {
		tx.sigPrivKey = make([]byte, 64)
		tx.sigPrivKey = append(*senderSK, witness.ComInputOpeningsWitness[0].Openings[privacy.RAND].Bytes()...)

		// encrypt coin details (Randomness)
		// hide information of output coins except coin commitments, last byte of public key, snDerivators
		for i := 0; i < len(tx.Proof.OutputCoins); i++ {
			tx.Proof.OutputCoins[i].Encrypt(paymentInfo[i].PaymentAddress.Tk)
			tx.Proof.OutputCoins[i].CoinDetails.SerialNumber = nil
			tx.Proof.OutputCoins[i].CoinDetails.Value = 0
			tx.Proof.OutputCoins[i].CoinDetails.PublicKey = nil
			tx.Proof.OutputCoins[i].CoinDetails.Randomness = nil
			tx.Proof.OutputCoins[i].CoinDetails.PubKeyLastByte = tx.Proof.OutputCoins[i].CoinDetails.PublicKey.Compress()[len(tx.Proof.OutputCoins[i].CoinDetails.PublicKey.Compress())-1]
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
	tx.Hash()
	tx.SignTx(hasPrivacy)

	return nil
}

// SignTx signs tx
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
		tx.Sig = ECDSASigToByteArray(r, s)
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

		// convert signature from byte array to ECDSASign
		r, s := FromByteArrayToECDSASig(tx.Sig)

		// verify signature
		res = ecdsa.Verify(verKey, tx.Hash()[:], r, s)
	}

	return res, nil
}

// ECDSASigToByteArray converts signature to byte array
func ECDSASigToByteArray(r, s *big.Int) (sig []byte) {
	sig = append(sig, r.Bytes()...)
	sig = append(sig, s.Bytes()...)
	return
}

// FromByteArrayToECDSASig converts a byte array to signature
func FromByteArrayToECDSASig(sig []byte) (r, s *big.Int) {
	r = new(big.Int).SetBytes(sig[0:32])
	s = new(big.Int).SetBytes(sig[32:64])
	return
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
// Note: This method doesn't check for double spending
func (tx *Tx) ValidateTransaction(hasPrivacy bool) bool {
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

	// Verify the payment proof
	valid = tx.Proof.Verify(false, tx.SigPubKey)
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
		sizeDescs = uint64(common.Max(1, (len(usableTx) + len(payments) - 3))) * EstimateJSDescSize()
	} else {
		sizeDescs = uint64(common.Max(1, (len(usableTx) - 3))) * EstimateJSDescSize()
	}
	var sizejSPubKey uint64 = 64 // [64]byte
	var sizejSSig uint64 = 64    // [64]byte
	estimateTxSizeInByte := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeDescs + sizejSPubKey + sizejSSig
	return uint64(math.Ceil(float64(estimateTxSizeInByte) / 1024))
}

// todo: thunderbird
// CheckSND return true if snd exists in snDerivators list
func CheckSNDExistence(snd *big.Int) bool {
	//todo: query from db to get snDerivators
	return false
}
