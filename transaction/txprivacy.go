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
)

type TxPrivacy struct {
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant

	SigPubKey []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig       []byte `json:"Sig, omitempty"`       // 64 bytes
	Proof     *zkp.PaymentProof

	PubKeyLastByte byte `json:"AddressLastByte"`

	TxId       *common.Hash
	sigPrivKey []byte // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes

	// this one is a hash id of requested tx
	// and is used inside response txs
	// so that we can determine pair of req/res txs
	// for example, BuySellRequestTx/BuySellResponseTx
	//RequestedTxID *common.Hash

	// all input of verify function
	// outputcoin []OutputCoin
}

func (tx *TxPrivacy) CreateTx(
	senderSK *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	hasPrivacy bool,
) (*TxPrivacy, error) {

	// Print list of all input coins
	fmt.Printf("List of all input coins before building tx:\n")
	for _, coin := range inputCoins {
		fmt.Printf("%+v\n", coin)
	}

	// Calculate sum of all output coins' value
	var sumOutputValue uint64
	for _, p := range paymentInfo {
		sumOutputValue += p.Amount
		fmt.Printf("[CreateTx] paymentInfo.H: %+v, paymentInfo.PaymentAddress: %x\n", p.Amount, p.PaymentAddress.Pk)
	}

	// Calculate sum of all input coins' value
	var sumInputValue uint64
	for _, coin := range inputCoins {
		sumInputValue += coin.CoinDetails.Value
	}

	// Calculate over balance, it will be returned to sender
	overBalance := sumInputValue - sumOutputValue - fee

	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return nil, fmt.Errorf("Input value less than output value")
	}

	// create sender's key set from sender's spending key
	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKeyByte((*senderSK)[:])

	// get public key last byte
	pkLastByte := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]
	tx.PubKeyLastByte = pkLastByte

	// create new output coins
	outputCoins := make([]*privacy.OutputCoin, len(paymentInfo))

	// create new output coins with info: Pk, value, SND
	for i, pInfo := range paymentInfo {
		outputCoins[i] = new(privacy.OutputCoin)
		outputCoins[i].CoinDetails.Value = pInfo.Amount
		outputCoins[i].CoinDetails.PublicKey, _ = privacy.DecompressKey(pInfo.PaymentAddress.Pk)
		outputCoins[i].CoinDetails.SNDerivator = privacy.RandInt()
	}

	// if overBalance > 0, create a output coin with pk is pk's sender and value is overBalance
	if overBalance > 0 {
		changeCoin := new(privacy.OutputCoin)
		changeCoin.CoinDetails.Value = overBalance
		changeCoin.CoinDetails.PublicKey, _ = privacy.DecompressKey(senderFullKey.PaymentAddress.Pk)
		changeCoin.CoinDetails.SNDerivator = privacy.RandInt()

		outputCoins = append(outputCoins, changeCoin)

		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = overBalance
		changePaymentInfo.PaymentAddress = senderFullKey.PaymentAddress
		paymentInfo = append(paymentInfo, changePaymentInfo)
	}

	// create zero knowledge proof of payment
	// prepare witness for proving
	witness := new(zkp.PaymentWitness)
	witness.Build(hasPrivacy, new(big.Int).SetBytes(*senderSK), inputCoins, outputCoins )
	tx.Proof = witness.Prove(false)

	// set private key for signing tx
	if hasPrivacy{
		tx.sigPrivKey = make([]byte, 64)
		tx.sigPrivKey = append(*senderSK, witness.ComOpeningsWitness[0].Openings[privacy.RAND].Bytes()...)
	} else{
		tx.sigPrivKey = *senderSK
	}

	// encrypt coin details (Randomness)
	for i := 0; i < len(outputCoins); i++ {
		outputCoins[i].Encrypt(paymentInfo[i].PaymentAddress.Tk)
	}

	// sign tx
	tx.Hash()
	tx.SignTx(hasPrivacy)

	return tx, nil
}

// SignTx signs tx
func (tx * TxPrivacy) SignTx(hasPrivacy bool) error {
	//Check input transaction
	if tx.Sig != nil {
		return fmt.Errorf("input transaction must be an unsigned one")
	}

	if hasPrivacy{
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
		signature, err := sigKey.Sign(tx.TxId[:])
		if err != nil {
			return err
		}

		// convert signature to byte array
		tx.Sig = signature.ToBytes()

	} else{
		/***** using ECDSA ****/
		// sign with sigPrivKey
		// prepare private key for ECDSA
		sigKey := new(ecdsa.PrivateKey)
		sigKey.PublicKey.Curve = privacy.Curve
		sigKey.D = new(big.Int).SetBytes(tx.sigPrivKey)
		sigKey.PublicKey.X, sigKey.PublicKey.Y = privacy.Curve.ScalarBaseMult(tx.sigPrivKey)

		// save public key for verification signature tx
		verKey:= new(privacy.EllipticPoint)
		verKey.X, verKey.Y = sigKey.PublicKey.X, sigKey.PublicKey.Y
		tx.SigPubKey = verKey.Compress()

		// signing
		r, s, err := ecdsa.Sign(rand.Reader, sigKey, tx.TxId[:])
		if err != nil {
			return err
		}

		// convert signature to byte array
		tx.Sig = ECDSASigToByteArray(r, s)
	}

	return nil
}

func (tx *TxPrivacy) VerifySigTx(hasPrivacy bool) (bool, error){
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, fmt.Errorf("input transaction must be an signed one!")
	}

	var err error
	res := false

	if hasPrivacy{
		/****** verify Schnorr signature *****/
		// prepare Public key for verification
		verKey := new(privacy.SchnPubKey)
		verKey.PK, err = privacy.DecompressKey(tx.SigPubKey)
		if err!= nil{
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
		res = verKey.Verify(signature, tx.TxId[:])


	} else{
		/****** verify ECDSA signature *****/
		// prepare Public key for verification
		verKey := new(ecdsa.PublicKey)
		point := new(privacy.EllipticPoint)
		point, _ = privacy.DecompressKey(tx.SigPubKey)
		verKey.X, verKey.Y = point.X, point.Y

		// convert signature from byte array to ECDSASign
		r, s  := FromByteArrayToECDSASig(tx.Sig)

		// verify signature
		res = ecdsa.Verify(verKey, tx.TxId[:], r, s)
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
func (tx *TxPrivacy) ValidateTx(hasPrivacy bool) bool {
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
	valid = tx.Proof.Verify(false)
	if valid == false{
		fmt.Printf("Error verifying the payment proof")
		return false
	}

	return true
}

func (tx *TxPrivacy) Hash() *common.Hash {
	record := strconv.Itoa(int(tx.Version))
	record += tx.Type
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	record += string(tx.Proof.Bytes()[:])
	record += string(tx.PubKeyLastByte)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
