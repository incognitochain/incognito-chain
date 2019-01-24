package privacy

import (
	"github.com/ninjadotorg/constant/common"
	"math/big"
)

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

// SignMultiSig ...
func (privKey *SpendingKey) SignMultiSig(data []byte, listPK []*PublicKey, pubKey *PublicKey, listR []*EllipticPoint, r *big.Int) *SchnMultiSig {
	//C1 = R0+R1+R2+R3+...+Rn
	R := new(EllipticPoint)
	R.X = big.NewInt(0)
	R.Y = big.NewInt(0)
	for i := 0; i < len(listR); i++ {
		R = R.Add(listR[i])
	}

	//Calculate common params:
	//	aggKey = PK0+PK1+PK2+...+PKn
	//	X = (PK0*a0) + (PK1*a1) + ... + (PKn*an)
	//	C2 = Hash(X||C1||data)
	aggKey, C, _ := generateCommonParams(nil, listPK, R, data)
	//recalculate a0
	selfPK := new(EllipticPoint)
	selfPK.Decompress(*pubKey)
	temp := aggKey.Add(selfPK)
	a := common.DoubleHashB(temp.Compress())
	aInt := big.NewInt(0)
	aInt.SetBytes(a)
	aInt.Mod(aInt, Curve.Params().N)

	//sig = r + C2*a0*privKey
	sig := big.NewInt(0)
	sig.Set(aInt)
	sig.Mul(sig, C)
	sig.Mod(sig, Curve.Params().N)
	sig.Mul(sig, new(big.Int).SetBytes(*privKey))
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
func (multiSig *SchnMultiSig) VerifyMultiSig(data []byte, listPK []*PublicKey, pubKey *PublicKey, RCombine *EllipticPoint) bool {
	//Calculate common params:
	//	aggKey = PK0+PK1+PK2+...+PKn, PK0 is selfPK
	//	X = (PK0*a0) + (PK1*a1) + ... + (PKn*an)
	//	C2 = Hash(X||C1||data)
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

	//RXCPoint is C1.X^C2
	RXCPoint := X.ScalarMult(C)
	RXCPoint = RXCPoint.Add(multiSig.R)
	return GSPoint.IsEqual(RXCPoint)
}

/*func generateRandom() (*EllipticPoint, *big.Int) {
	r := RandInt()
	GPoint := new(EllipticPoint)
	GPoint.X, GPoint.Y = big.NewInt(0), big.NewInt(0)
	GPoint.X.Set(Curve.Params().Gx)
	GPoint.Y.Set(Curve.Params().Gy)
	R := GPoint.ScalarMult(r)
	return R, r
}*/

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

// CombineMultiSig Combining all EC Schnorr MultiSig in given list
func CombineMultiSig(listSignatures []*SchnMultiSig) *SchnMultiSig {
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
