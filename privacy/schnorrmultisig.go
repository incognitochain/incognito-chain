package privacy

import (
	"math/big"
	"sync"

	"github.com/ninjadotorg/constant/common"
)

var isTesting bool

//#if isTesting
var pubkeyTest []*PublicKey
var RTest []*EllipticPoint
var mutex sync.Mutex
var counter int
var wg sync.WaitGroup
var wgchild sync.WaitGroup
var Numbs int

//#endif

// MultiSigSchemeInterface define all of function for create EC Schnorr signature which could be combine by adding with another EC Schnorr Signature
type MultiSigSchemeInterface interface {
	SignMultiSig(data []byte, listPK []*PublicKey, listR []*EllipticPoint, r *big.Int) *SchnMultiSig
	VerifyMultiSig(data []byte, listPK []*PublicKey, pubKey *PublicKey, RCombine *EllipticPoint) bool
	CombineMultiSig(listSignatures []*SchnMultiSig) *SchnMultiSig
}

// MultiSigScheme ...
type MultiSigScheme struct {
	Keyset    *MultiSigKeyset
	Signature *SchnMultiSig
}

// MultiSigKeyset contains keyset for EC Schnorr MultiSig Scheme
type MultiSigKeyset struct {
	priKey *SpendingKey
	pubKey *PublicKey
}

// SchnMultiSig is struct of EC Schnorr Signature which is combinable
type SchnMultiSig struct {
	R *EllipticPoint
	S *big.Int
}

// SetBytes - Constructing multiSig from byte array
func (multiSig *SchnMultiSig) SetBytes(sigByte []byte) {
	multiSig.R.Decompress(sigByte[0:CompressedPointSize])
	multiSig.S = big.NewInt(0)
	multiSig.S.SetBytes(sigByte[CompressedPointSize : CompressedPointSize+BigIntSize])
}

// Set - Constructing multiSig
func (multiSig *SchnMultiSig) Set(R *EllipticPoint, S *big.Int) {
	multiSig.R = R
	multiSig.S = S
}

// Set - Constructing MultiSigKeyset
func (multiSigKeyset *MultiSigKeyset) Set(priKey *SpendingKey, pubKey *PublicKey) {
	multiSigKeyset.priKey = priKey
	multiSigKeyset.pubKey = pubKey
}

// Bytes - Converting SchnorrMultiSig to byte array
func (multiSig *SchnMultiSig) Bytes() []byte {
	res := multiSig.R.Compress()
	temp := multiSig.S.Bytes()
	for j := 0; j < BigIntSize-len(temp); j++ {
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	return res
}

func (multisigScheme *MultiSigScheme) Init() {
	multisigScheme.Keyset = new(MultiSigKeyset)
	multisigScheme.Keyset.priKey = new(SpendingKey)
	multisigScheme.Keyset.pubKey = new(PublicKey)
	multisigScheme.Signature = new(SchnMultiSig)
	multisigScheme.Signature.R = new(EllipticPoint)
	multisigScheme.Signature.S = new(big.Int)
}

/*
	function: sign
	param
	#1 data need to be signed
	#2 notice verifymultisignature
	#3 g^r
	#4 r: random number of signer
*/
// SignMultiSig ...
func (multiSigKeyset *MultiSigKeyset) SignMultiSig(data []byte, listPK []*PublicKey, listR []*EllipticPoint, r *big.Int) *SchnMultiSig {
	//R = R0+R1+R2+R3+...+Rn
	R := new(EllipticPoint)
	R.X = big.NewInt(0)
	R.Y = big.NewInt(0)
	for i := 0; i < len(listR); i++ {
		R = R.Add(listR[i])
	}

	//Calculate common params:
	//	aggKey = PK0+PK1+PK2+...+PKn
	//	X = (PK0*a0) + (PK1*a1) + ... + (PKn*an)
	//	C = Hash(X||R||data)
	aggKey, C, _ := generateCommonParams(nil, listPK, R, data)
	//recalculate a0
	selfPK := new(EllipticPoint)
	selfPK.Decompress(*multiSigKeyset.pubKey)
	temp := aggKey.Add(selfPK)
	a := common.DoubleHashB(temp.Compress())
	aInt := big.NewInt(0)
	aInt.SetBytes(a)
	aInt.Mod(aInt, Curve.Params().N)

	//sig = r + C*a0*privKey
	sig := big.NewInt(0)
	sig.Set(aInt)
	sig.Mul(sig, C)
	sig.Mod(sig, Curve.Params().N)
	sig.Mul(sig, new(big.Int).SetBytes(*multiSigKeyset.priKey))
	sig.Mod(sig, Curve.Params().N)
	sig.Add(sig, r)
	sig.Mod(sig, Curve.Params().N)

	selfR := new(EllipticPoint)
	selfR.X, selfR.Y = big.NewInt(0), big.NewInt(0)
	selfR.X.Set(Curve.Params().Gx)
	selfR.Y.Set(Curve.Params().Gy)
	selfR = selfR.ScalarMult(r)
	res := new(SchnMultiSig)
	res.Set(selfR, sig)

	return res
}

// VerifyMultiSig ...
/*
	function: Verify signature
	params:
		#1: data need to be signed
		#2: List of public key join (either sign or un-sign) -> alway not null
		#3: public key of signer
			if null: verify of many party
			if not null: verify of this publickey
		#4: R combine from params #2

	return: true or false (notice param 3)
*/
func (multiSig *SchnMultiSig) VerifyMultiSig(data []byte, listPK []*PublicKey, pubKey *PublicKey, RCombine *EllipticPoint) bool {
	//Calculate common params:
	//	aggKey = PK0+PK1+PK2+...+PKn, PK0 is selfPK
	//	X = (PK0*a0) + (PK1*a1) + ... + (PKn*an)
	//	C = Hash(X||R||data)
	//for verify signature of a Signer, which wasn't combined, |listPK| = 1 and contain publickey of the Signer
	var C *big.Int
	var X *EllipticPoint
	if RCombine == nil {
		_, C, X = generateCommonParams(nil, listPK, multiSig.R, data)
	} else {
		_, C, X = generateCommonParams(pubKey, listPK, RCombine, data)
	}

	//GSPoint = G*S
	GSPoint := new(EllipticPoint)
	GSPoint.X, GSPoint.Y = big.NewInt(0), big.NewInt(0)
	GSPoint.X.Set(Curve.Params().Gx)
	GSPoint.Y.Set(Curve.Params().Gy)
	GSPoint = GSPoint.ScalarMult(multiSig.S)

	//RXCPoint is R.X^C
	RXCPoint := X.ScalarMult(C)
	RXCPoint = RXCPoint.Add(multiSig.R)
	return GSPoint.IsEqual(RXCPoint)
}

func (multisigScheme *MultiSigScheme) GenerateRandom() (*EllipticPoint, *big.Int) {
	r := RandInt()
	GPoint := new(EllipticPoint)
	GPoint.X, GPoint.Y = big.NewInt(0), big.NewInt(0)
	GPoint.X.Set(Curve.Params().Gx)
	GPoint.Y.Set(Curve.Params().Gy)
	R := GPoint.ScalarMult(r)
	return R, r
}

func generateCommonParams(pubKey *PublicKey, listPubkey []*PublicKey, R *EllipticPoint, data []byte) (*EllipticPoint, *big.Int, *EllipticPoint) {
	aggPubkey := new(EllipticPoint)
	aggPubkey.X = big.NewInt(0)
	aggPubkey.Y = big.NewInt(0)

	for i := 0; i < len(listPubkey); i++ {
		temp := new(EllipticPoint)
		temp.Decompress(*listPubkey[i])
		aggPubkey = aggPubkey.Add(temp)
	}

	X := new(EllipticPoint)
	X.X = big.NewInt(0)
	X.Y = big.NewInt(0)

	for i := 0; i < len(listPubkey); i++ {
		temp := new(EllipticPoint)
		temp.Decompress(*listPubkey[i])
		temp1 := aggPubkey.Add(temp)
		a := common.DoubleHashB(temp1.Compress())
		aInt := big.NewInt(0)
		aInt.SetBytes(a)
		aInt.Mod(aInt, Curve.Params().N)
		X = X.Add(temp.ScalarMult(aInt))
	}

	Cbyte := X.Compress()
	Cbyte = append(Cbyte, R.Compress()...)
	Cbyte = append(Cbyte, data...)
	C := big.NewInt(0)
	C.SetBytes(Cbyte)
	C.Mod(C, Curve.Params().N)

	if pubKey != nil {
		temp := new(EllipticPoint)
		temp.Decompress(*pubKey)
		temp1 := aggPubkey.Add(temp)
		a := common.DoubleHashB(temp1.Compress())
		aInt := big.NewInt(0)
		aInt.SetBytes(a)
		X = temp.ScalarMult(aInt)
	}
	return aggPubkey, C, X
}

/*
	function: aggregate signature
	param: list of signature
	return: agg sign
*/
// CombineMultiSig Combining all EC Schnorr MultiSig in given list
func (multisigScheme *MultiSigScheme) CombineMultiSig(listSignatures []*SchnMultiSig) *SchnMultiSig {
	res := new(SchnMultiSig)
	res.R = new(EllipticPoint)
	res.R.X, res.R.Y = big.NewInt(0), big.NewInt(0)
	res.S = big.NewInt(0)

	for i := 0; i < len(listSignatures); i++ {
		res.R = res.R.Add(listSignatures[i].R)
		res.S.Add(res.S, listSignatures[i].S)
		res.S.Mod(res.S, Curve.Params().N)
	}

	return res
}

// Functions for testing
// -------------------------------------------------------------------------------------------------

func broadcastR(R *EllipticPoint) {
	if isTesting {
		mutex.Lock()
		RTest[counter] = R
		counter++
		mutex.Unlock()
	}
	//todo
}

// TestMultiSig EC Schnorr MultiSig Scheme
// func TestMultiSig() {
// 	isTesting = true
// 	Numbs = 20
// 	counter = 0
// 	listSigners := make([]*MultiSigKeyset, Numbs)
// 	pubkeyTest = make([]*PublicKey, Numbs)
// 	RTest = make([]*EllipticPoint, Numbs)
// 	// REachSigner := make([]*EllipticPoint, Numbs)
// 	Sig := make([]*SchnMultiSig, Numbs)
// 	R := new(EllipticPoint)
// 	R.X = big.NewInt(0)
// 	R.Y = big.NewInt(0)
// 	for i := 0; i < Numbs; i++ {
// 		listSigners[i] = new(MultiSigKeyset)
// 		listSigners[i].priKey = new(SpendingKey)
// 		*listSigners[i].priKey = GenerateSpendingKey(RandInt().Bytes())
// 		pubkeyTest[i] = new(PublicKey)
// 		listSigners[i].pubKey = new(PublicKey)
// 		*pubkeyTest[i] = GeneratePublicKey(*listSigners[i].priKey)
// 		listSigners[i].pubKey = pubkeyTest[i]
// 	}
// 	for i := 0; i < Numbs; i++ {
// 		wg.Add(1)
// 		go func(j int) {
// 			defer wg.Done()
// 			Ri, ri := generateRandom()
// 			broadcastR(Ri)
// 			time.Sleep(500 * time.Millisecond)
// 			for counter < Numbs {
// 			}
// 			Sig[j] = listSigners[j].SignMultiSig([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, pubkeyTest, RTest, ri)
// 		}(i)
// 	}
// 	wg.Wait()
// 	aggSig := CombineMultiSig(Sig)
// 	for i := 0; i < Numbs; i++ {
// 		R = R.Add(RTest[i])
// 		fmt.Printf("\n**********************************************************************************************************************************************************************************")
// 		fmt.Printf("\n* Signature of signer %v\n", i)
// 		fmt.Printf("*\tR  [%v]: %v\n", i, Sig[i].R)
// 		fmt.Printf("*\tSig[%v]: %v\n", i, Sig[i].S)
// 		fmt.Printf("* Verifing... ")
// 		fmt.Printf("Signature %v is %v\n", i, Sig[i].VerifyMultiSig([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, pubkeyTest, pubkeyTest[i], aggSig.R))
// 		fmt.Println("**********************************************************************************************************************************************************************************")
// 	}

// 	fmt.Println("\tAggregate:")
// 	fmt.Printf("\t\tAggSignature: %v\n", aggSig.S)
// 	fmt.Printf("\t\tAggR        : %v\n", aggSig.R)
// 	fmt.Printf("\tVerify result: %v\n", aggSig.VerifyMultiSig([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, pubkeyTest, nil, nil))
// }

// -------------------------------------------------------------------------------------------------
