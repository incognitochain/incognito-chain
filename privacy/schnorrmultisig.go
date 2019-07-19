package privacy

import (
	"errors"
	"math/big"

	"golang.org/x/crypto/sha3"
)

func internalHash(data []byte) []byte {
	hashMachine := sha3.NewLegacyKeccak256()
	hashMachine.Write(data)
	return hashMachine.Sum(nil)
}

// MultiSigScheme ...
type MultiSigScheme struct {
	Keyset    *MultiSigKeyset
	Signature *SchnMultiSig
}

// MultiSigKeyset contains keyset for EC Schnorr MultiSig Scheme
type MultiSigKeyset struct {
	priKey *PrivateKey
	pubKey *PublicKey
}

// SchnMultiSig is struct of EC Schnorr Signature which is combinable
type SchnMultiSig struct {
	R *EllipticPoint
	S *big.Int
}

// SetBytes - Constructing multiSig from byte array
func (multiSig *SchnMultiSig) SetBytes(sigByte []byte) error {
	if len(sigByte) < CompressedPointSize+BigIntSize {
		return errors.New("Invalid sig length")
	}
	multiSig.R = new(EllipticPoint)
	err := multiSig.R.Decompress(sigByte[0:CompressedPointSize])
	if err != nil {
		return err
	}
	multiSig.S = big.NewInt(0)
	multiSig.S.SetBytes(sigByte[CompressedPointSize : CompressedPointSize+BigIntSize])
	return nil
}

// Set - Constructing multiSig
func (multiSig *SchnMultiSig) Set(R *EllipticPoint, S *big.Int) {
	multiSig.R = R
	multiSig.S = S
}

// Set - Constructing MultiSigKeyset
func (multiSigKeyset *MultiSigKeyset) Set(priKey *PrivateKey, pubKey *PublicKey) {
	multiSigKeyset.priKey = priKey
	multiSigKeyset.pubKey = pubKey
}

// Bytes - Converting SchnorrMultiSig to byte array
func (multiSig *SchnMultiSig) Bytes() []byte {
	if !Curve.IsOnCurve(multiSig.R.X, multiSig.R.Y) {
		panic("Throw Error from Byte() method")
	}
	res := multiSig.R.Compress()
	if multiSig.S == nil {
		panic("Throw Error from Byte() method")
	}
	temp := multiSig.S.Bytes()
	diff := BigIntSize - len(temp)
	for j := 0; j < diff; j++ {
		temp = append([]byte{0}, temp...)
	}
	res = append(res, temp...)
	return res
}

func (multisigScheme *MultiSigScheme) Init() {
	multisigScheme.Keyset = new(MultiSigKeyset)
	multisigScheme.Keyset.priKey = new(PrivateKey)
	multisigScheme.Keyset.pubKey = new(PublicKey)
	multisigScheme.Signature = new(SchnMultiSig)
	multisigScheme.Signature.R = new(EllipticPoint)
	multisigScheme.Signature.R.X = big.NewInt(0)
	multisigScheme.Signature.R.Y = big.NewInt(0)
	multisigScheme.Signature.S = big.NewInt(0)
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
	//r = R0+R1+R2+R3+...+Rn
	R := new(EllipticPoint)
	R.X = big.NewInt(0)
	R.Y = big.NewInt(0)
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
	selfPK.Decompress(*multiSigKeyset.pubKey)
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
	if len(res.Bytes()) != (BigIntSize + CompressedPointSize) {
		panic("can not sign multi sig")
	}

	return res
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
func (multiSig SchnMultiSig) VerifyMultiSig(data []byte, listCommonPK []*PublicKey, listCombinePK []*PublicKey, RCombine *EllipticPoint) bool {
	if len(multiSig.Bytes()) != (BigIntSize + CompressedPointSize) {
		panic("Wrong length")
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
	GSPoint.X, GSPoint.Y = big.NewInt(0), big.NewInt(0)
	GSPoint.X.Set(Curve.Params().Gx)
	GSPoint.Y.Set(Curve.Params().Gy)
	// fmt.Println("GSPoint: \n", GSPoint)
	// fmt.Println("multisig.S: \n", multiSig.S)
	GSPoint = GSPoint.ScalarMult(multiSig.S)

	//RXCPoint is r.X^C
	RXCPoint := X.ScalarMult(C)
	RXCPoint = RXCPoint.Add(multiSig.R)
	return GSPoint.IsEqual(RXCPoint)
}

//GenerateRandomFromSeed abc
func (multisigScheme *MultiSigScheme) GenerateRandomFromSeed(i *big.Int) (*EllipticPoint, *big.Int) {
	r := i
	GPoint := new(EllipticPoint)
	GPoint.X, GPoint.Y = big.NewInt(0), big.NewInt(0)
	GPoint.X.Set(Curve.Params().Gx)
	GPoint.Y.Set(Curve.Params().Gy)
	R := GPoint.ScalarMult(r)
	return R, r
}

func (multisigScheme *MultiSigScheme) GenerateRandom() (*EllipticPoint, *big.Int) {
	r := RandScalar()
	GPoint := new(EllipticPoint)
	GPoint.X, GPoint.Y = big.NewInt(0), big.NewInt(0)
	GPoint.X.Set(Curve.Params().Gx)
	GPoint.Y.Set(Curve.Params().Gy)
	R := GPoint.ScalarMult(r)
	return R, r
}

func generateCommonParams(listCommonPK []*PublicKey, listCombinePK []*PublicKey, R *EllipticPoint, data []byte) (*EllipticPoint, *big.Int, *EllipticPoint) {
	aggPubkey := new(EllipticPoint)
	aggPubkey.X = big.NewInt(0)
	aggPubkey.Y = big.NewInt(0)

	for i := 0; i < len(listCommonPK); i++ {
		temp := new(EllipticPoint)
		temp.Decompress(*listCommonPK[i])
		aggPubkey = aggPubkey.Add(temp)
	}

	X := new(EllipticPoint)
	X.X = big.NewInt(0)
	X.Y = big.NewInt(0)

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
		X.X.Set(big.NewInt(0))
		X.Y.Set(big.NewInt(0))
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