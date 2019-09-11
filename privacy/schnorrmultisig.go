package privacy

import (
	"crypto/rand"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"golang.org/x/crypto/sha3"
)

func internalHash(data []byte) []byte {
	hashMachine := sha3.NewLegacyKeccak256()
	hashMachine.Write(data)
	return hashMachine.Sum(nil)
}

// MultiSigScheme ...
type MultiSigScheme struct {
	keyset    *MultiSigKeyset
	signature *SchnMultiSig
}

func (multiSigScheme *MultiSigScheme) GetKeyset() *MultiSigKeyset {
	return multiSigScheme.keyset
}

// MultiSigKeyset contains keyset for EC Schnorr MultiSig Scheme
type MultiSigKeyset struct {
	privateKey *PrivateKey
	publicKey  *PublicKey
}

// SchnMultiSig is struct of EC Schnorr Signature which is combinable
type SchnMultiSig struct {
	r *EllipticPoint
	s *big.Int
}

// SetBytes - Constructing multiSig from byte array
func (multiSig *SchnMultiSig) SetBytes(sigByte []byte) error {
	if len(sigByte) < CompressedEllipticPointSize+common.BigIntSize {
		return NewPrivacyErr(InvalidLengthMultiSigErr, nil)
	}
	multiSig.r = new(EllipticPoint)
	err := multiSig.r.Decompress(sigByte[0:CompressedEllipticPointSize])
	if err != nil {
		return err
	}
	multiSig.s = big.NewInt(0)
	multiSig.s.SetBytes(sigByte[CompressedEllipticPointSize : CompressedEllipticPointSize+common.BigIntSize])
	return nil
}

// Set - Constructing multiSig
func (multiSig *SchnMultiSig) Set(R *EllipticPoint, S *big.Int) {
	multiSig.r = R
	multiSig.s = S
}

// Set - Constructing MultiSigKeyset
func (multiSigKeyset *MultiSigKeyset) Set(priKey *PrivateKey, pubKey *PublicKey) {
	multiSigKeyset.privateKey = priKey
	multiSigKeyset.publicKey = pubKey
}

// Bytes - Converting SchnorrMultiSig to byte array
func (multiSig SchnMultiSig) Bytes() ([]byte, error) {
	if !Curve.IsOnCurve(multiSig.r.x, multiSig.r.y) {
		return nil, NewPrivacyErr(InvalidMultiSigErr, nil)
	}
	res := multiSig.r.Compress()
	if multiSig.s == nil {
		return nil, NewPrivacyErr(InvalidMultiSigErr, nil)
	}
	temp := multiSig.s.Bytes()
	diff := common.BigIntSize - len(temp)
	for j := 0; j < diff; j++ {
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	return res, nil
}

func (multisigScheme *MultiSigScheme) Init() {
	multisigScheme.keyset = new(MultiSigKeyset)
	multisigScheme.keyset.privateKey = new(PrivateKey)
	multisigScheme.keyset.publicKey = new(PublicKey)
	multisigScheme.signature = new(SchnMultiSig)
	multisigScheme.signature.r = new(EllipticPoint)
	multisigScheme.signature.r.x = big.NewInt(0)
	multisigScheme.signature.r.y = big.NewInt(0)
	multisigScheme.signature.s = big.NewInt(0)
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
func (multiSigKeyset MultiSigKeyset) SignMultiSig(data []byte, listPK []*PublicKey, listR []*EllipticPoint, r *big.Int) (*SchnMultiSig, error) {
	//r = R0+R1+R2+R3+...+Rn
	R := new(EllipticPoint)
	R.x = big.NewInt(0)
	R.y = big.NewInt(0)
	for i := 0; i < len(listR); i++ {
		R = R.Add(listR[i])
	}

	//Calculate common component:
	//	aggKey = PK0+PK1+PK2+...+PKn
	//	X = (PK0*a0) + (PK1*a1) + ... + (PKn*an)
	//	C = Hash(X||r||data)
	aggKey, C, _ := generateCommonParams(listPK, listPK, R, data)
	//recalculate a0
	selfPK := new(EllipticPoint)
	selfPK.Decompress(*multiSigKeyset.publicKey)
	temp := aggKey.Add(selfPK)
	a := internalHash(temp.Compress())
	aInt := big.NewInt(0)
	aInt.SetBytes(a)
	aInt.Mod(aInt, Curve.Params().N)

	//sig = r + C*a0*privKey
	sig := big.NewInt(0)
	sig.Set(aInt)
	sig.Mul(sig, C)
	sig.Mod(sig, Curve.Params().N)
	sig.Mul(sig, new(big.Int).SetBytes(*multiSigKeyset.privateKey))
	sig.Mod(sig, Curve.Params().N)
	sig.Add(sig, r)
	sig.Mod(sig, Curve.Params().N)

	selfR := new(EllipticPoint)
	selfR.x, selfR.y = big.NewInt(0), big.NewInt(0)
	selfR.x.Set(Curve.Params().Gx)
	selfR.y.Set(Curve.Params().Gy)
	selfR = selfR.ScalarMult(r)
	res := new(SchnMultiSig)
	res.Set(selfR, sig)
	sigInBytes, err := res.Bytes()
	if err != nil {
		Logger.Log.Error("Convert multisig to bytes array error when signing", err)
		return nil, NewPrivacyErr(ConvertMultiSigToBytesErr, err)
	}

	if len(sigInBytes) != (common.BigIntSize + CompressedEllipticPointSize) {
		return nil, NewPrivacyErr(SignMultiSigErr, nil)
	}

	return res, nil
}

// VerifyMultiSig ...
/*
	function: Verify signature
	component:
		#1: data need to be signed
		#2: List of public key join phase 1 (create RCombine)
		#3: List of public key of signer who create multi signature
		#4: r combine in phase 1
	return: true or false
*/
func (multiSig SchnMultiSig) VerifyMultiSig(data []byte, listCommonPK []*PublicKey, listCombinePK []*PublicKey, RCombine *EllipticPoint) (bool, error) {
	multiSigInByte, err := multiSig.Bytes()
	if err != nil {
		Logger.Log.Error("Convert multisig to bytes array error when verifying", err)
		return false, NewPrivacyErr(ConvertMultiSigToBytesErr, err)
	}

	if len(multiSigInByte) != (common.BigIntSize + CompressedEllipticPointSize) {
		return false, NewPrivacyErr(InvalidLengthMultiSigErr, nil)
	}
	//Calculate common params:
	//	aggKey = PK0+PK1+PK2+...+PKn, PK0 is selfPK
	//	X = (PK0*a0) + (PK1*a1) + ... + (PKn*an)
	//	C = Hash(X||r||data)
	//for verify signature of a Signer, which wasn't combined, |listPK| = 1 and contain publickey of the Signer
	var C *big.Int
	var X *EllipticPoint
	// if RCombine == nil {
	_, C, X = generateCommonParams(listCommonPK, listCombinePK, RCombine, data)
	// fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	// fmt.Println(C.Text(16))
	// fmt.Println(X.X.Text(16))
	// fmt.Println(X.Y.Text(16))
	// fmt.Println("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	//GSPoint = G*S
	GSPoint := new(EllipticPoint)
	GSPoint.x, GSPoint.y = big.NewInt(0), big.NewInt(0)
	GSPoint.x.Set(Curve.Params().Gx)
	GSPoint.y.Set(Curve.Params().Gy)
	// fmt.Println("GSPoint: \n", GSPoint)
	// fmt.Println("multisig.S: \n", multiSig.S)
	GSPoint = GSPoint.ScalarMult(multiSig.s)

	//RXCPoint is r.X^C
	RXCPoint := X.ScalarMult(C)
	RXCPoint = RXCPoint.Add(multiSig.r)
	return GSPoint.IsEqual(RXCPoint), nil
}

//GenerateRandomFromSeed abc
func (multisigScheme MultiSigScheme) GenerateRandomFromSeed(i *big.Int) (*EllipticPoint, *big.Int) {
	r := i
	GPoint := new(EllipticPoint)
	GPoint.x, GPoint.y = big.NewInt(0), big.NewInt(0)
	GPoint.x.Set(Curve.Params().Gx)
	GPoint.y.Set(Curve.Params().Gy)
	R := GPoint.ScalarMult(r)
	return R, r
}

func (multisigScheme MultiSigScheme) GenerateRandom() (*EllipticPoint, *big.Int) {
	var r = rand.Reader
	randomNumer := RandScalar(r)
	GPoint := new(EllipticPoint)
	GPoint.x, GPoint.y = big.NewInt(0), big.NewInt(0)
	GPoint.x.Set(Curve.Params().Gx)
	GPoint.y.Set(Curve.Params().Gy)
	R := GPoint.ScalarMult(randomNumer)
	return R, randomNumer
}

/*
	function: aggregate signature
	param: list of signature
	return: agg sign
*/
// CombineMultiSig Combining all EC Schnorr MultiSig in given list
func (multisigScheme MultiSigScheme) CombineMultiSig(listSignatures []*SchnMultiSig) *SchnMultiSig {
	res := new(SchnMultiSig)
	res.r = new(EllipticPoint)
	res.r.x, res.r.y = big.NewInt(0), big.NewInt(0)
	res.s = big.NewInt(0)

	for i := 0; i < len(listSignatures); i++ {
		res.r = res.r.Add(listSignatures[i].r)
		res.s.Add(res.s, listSignatures[i].s)
		res.s.Mod(res.s, Curve.Params().N)
	}

	return res
}

func generateCommonParams(listCommonPK []*PublicKey, listCombinePK []*PublicKey, R *EllipticPoint, data []byte) (*EllipticPoint, *big.Int, *EllipticPoint) {
	aggPubkey := new(EllipticPoint)
	aggPubkey.x = big.NewInt(0)
	aggPubkey.y = big.NewInt(0)

	for i := 0; i < len(listCommonPK); i++ {
		temp := new(EllipticPoint)
		temp.Decompress(*listCommonPK[i])
		aggPubkey = aggPubkey.Add(temp)
	}

	X := new(EllipticPoint)
	X.x = big.NewInt(0)
	X.y = big.NewInt(0)

	for i := 0; i < len(listCommonPK); i++ {
		temp := new(EllipticPoint)
		temp.Decompress(*listCommonPK[i])
		temp1 := aggPubkey.Add(temp)
		a := internalHash(temp1.Compress())
		aInt := big.NewInt(0)
		aInt.SetBytes(a)
		aInt.Mod(aInt, Curve.Params().N)
		X = X.Add(temp.ScalarMult(aInt))
	}

	Cbyte := X.Compress()
	Cbyte = append(Cbyte, R.Compress()...)
	Cbyte = append(Cbyte, data...)
	Cbyte = internalHash(Cbyte)
	C := big.NewInt(0)
	C.SetBytes(Cbyte)
	C.Mod(C, Curve.Params().N)

	if len(listCommonPK) > len(listCombinePK) {
		X.x.Set(big.NewInt(0))
		X.y.Set(big.NewInt(0))
		for i := 0; i < len(listCombinePK); i++ {
			temp := new(EllipticPoint)
			temp.Decompress(*listCombinePK[i])
			temp1 := aggPubkey.Add(temp)
			a := internalHash(temp1.Compress())
			aInt := big.NewInt(0)
			aInt.SetBytes(a)
			X = X.Add(temp.ScalarMult(aInt))
		}
	}
	return aggPubkey, C, X
}
