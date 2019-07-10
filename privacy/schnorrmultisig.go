package privacy

import (
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
)

// MultiSigScheme represents to Schnorr multi signature
// include key set (signing key and verification key) and signature
type MultiSigScheme struct {
	Keyset    *MultiSigKeyset
	Signature *SchnMultiSig
}

// MultiSigKeyset represents a key set for Schnorr MultiSig Scheme
// key set contains priKey and pubKey
type MultiSigKeyset struct {
	priKey *PrivateKey
	pubKey *PublicKey
}

// SchnMultiSig represents a EC Schnorr Signature which is combined
type SchnMultiSig struct {
	R *EllipticPoint
	S *big.Int
}

// SetBytes (SchnMultiSig) reverts multi sig from bytes array
func (multiSig *SchnMultiSig) SetBytes(sigByte []byte) error {
	if len(sigByte) < SchnMultiSigSize {
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

// Set (SchnMultiSig) sets values to multi sig
func (multiSig *SchnMultiSig) Set(R *EllipticPoint, S *big.Int) {
	multiSig.R = R
	multiSig.S = S
}

// Set (MultiSigKeyset) sets values to multi key set
func (multiSigKeyset *MultiSigKeyset) Set(priKey *PrivateKey, pubKey *PublicKey) {
	multiSigKeyset.priKey = priKey
	multiSigKeyset.pubKey = pubKey
}

// Bytes (SchnMultiSig) converts SchnorrMultiSig to bytes array
func (multiSig *SchnMultiSig) Bytes() []byte {
	if !Curve.IsOnCurve(multiSig.R.X, multiSig.R.Y) || multiSig.S == nil{
		return []byte{}
	}

	res := append(multiSig.R.Compress(), AddPaddingBigInt(multiSig.S, BigIntSize)...)
	return res
}

// Init (MultiSigScheme) initializes multi sig scheme
func (multisigScheme *MultiSigScheme) Init() {
	multisigScheme.Keyset = new(MultiSigKeyset)
	multisigScheme.Keyset.priKey = new(PrivateKey)
	multisigScheme.Keyset.pubKey = new(PublicKey)

	multisigScheme.Signature = new(SchnMultiSig)
	multisigScheme.Signature.R = new(EllipticPoint).Zero()
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
// SignMultiSig receives a data in bytes array,
// list of public keys, list of randomness and a secret random
// it returns a Schnorr multi signature
func (multiSigKeyset *MultiSigKeyset) SignMultiSig(data []byte, listPK []*PublicKey, listR []*EllipticPoint, r *big.Int) (*SchnMultiSig, error) {
	//r = R0+R1+R2+R3+...+Rn
	R := new(EllipticPoint).Zero()
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
	err := selfPK.Decompress(*multiSigKeyset.pubKey)
	if err != nil{
		return nil, err
	}

	temp := aggKey.Add(selfPK)
	a := common.HashB(temp.Compress())
	aInt := new(big.Int).SetBytes(a)
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
	selfR.Set(Curve.Params().Gx, Curve.Params().Gy)
	selfR = selfR.ScalarMult(r)

	res := new(SchnMultiSig)
	res.Set(selfR, sig)
	if len(res.Bytes()) != (SchnMultiSigSize) {
		return nil, errors.New("can not sign multi sig")
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

// VerifyMultiSig receives a data in bytes array
// list of common public keys, list of combined public key
// and combined randomness
// It verifies the multi sig is valid or not
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
	_, C, X = generateCommonParams(listCommonPK, listCombinePK, RCombine, data)

	//GSPoint = G*S
	GSPoint := new(EllipticPoint)
	GSPoint.Set(Curve.Params().Gx, Curve.Params().Gy)
	GSPoint = GSPoint.ScalarMult(multiSig.S)

	//RXCPoint is r.X^C
	RXCPoint := X.ScalarMult(C)
	RXCPoint = RXCPoint.Add(multiSig.R)
	return GSPoint.IsEqual(RXCPoint)
}

// GenerateRandom generates a secret randomness r
// ands returns G^r in which G is base generator of Curve
func (multisigScheme *MultiSigScheme) GenerateRandom() (*EllipticPoint, *big.Int) {
	r := RandScalar()
	GPoint := new(EllipticPoint)
	GPoint.Set(Curve.Params().Gx, Curve.Params().Gy)
	R := GPoint.ScalarMult(r)
	return R, r
}

func generateCommonParams(listCommonPK []*PublicKey, listCombinePK []*PublicKey, R *EllipticPoint, data []byte) (*EllipticPoint, *big.Int, *EllipticPoint) {
	aggPubkey := new(EllipticPoint).Zero()

	mapCommonPKPoints := make(map[*PublicKey]*EllipticPoint, len(listCommonPK))
	for i := 0; i < len(listCommonPK); i++ {
		tmp := new(EllipticPoint)
		tmp.Decompress(*listCommonPK[i])
		mapCommonPKPoints[listCommonPK[i]] = tmp
		aggPubkey = aggPubkey.Add(tmp)
	}

	X := new(EllipticPoint).Zero()
	mapCommonPKaInt := make(map[*PublicKey]*big.Int, len(listCommonPK))
	for i := 0; i < len(listCommonPK); i++ {
		temp := aggPubkey.Add(mapCommonPKPoints[listCommonPK[i]])

		a := common.HashB(temp.Compress())
		aInt := new(big.Int).SetBytes(a)
		aInt.Mod(aInt, Curve.Params().N)
		mapCommonPKaInt[listCommonPK[i]] = new(big.Int).Set(aInt)
		X = X.Add(mapCommonPKPoints[listCommonPK[i]].ScalarMult(aInt))
	}

	Cbytes := append(X.Compress(), R.Compress()...)
	Cbytes = append(Cbytes, data...)
	Cbytes = common.HashB(Cbytes)
	C := new(big.Int).SetBytes(Cbytes)
	C.Mod(C, Curve.Params().N)

	if len(listCommonPK) > len(listCombinePK) {
		X.Set(big.NewInt(0), big.NewInt(0))

		for i := 0; i < len(listCombinePK); i++ {
			publicKeyPoint, _:= mapCommonPKPoints[listCombinePK[i]]
			aInt, _ := mapCommonPKaInt[listCombinePK[i]]

			X = X.Add(publicKeyPoint.ScalarMult(aInt))
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
	res.R = new(EllipticPoint).Zero()
	res.S = big.NewInt(0)

	for i := 0; i < len(listSignatures); i++ {
		res.R = res.R.Add(listSignatures[i].R)
		res.S.Add(res.S, listSignatures[i].S)
		res.S.Mod(res.S, Curve.Params().N)
	}

	return res
}