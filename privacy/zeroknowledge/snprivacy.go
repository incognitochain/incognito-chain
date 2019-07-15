package zkp

import (
	"errors"
	"github.com/incognitochain/incognito-chain/privacy"
	"math/big"
)

type SNPrivacyStatement struct {
	sn       *privacy.EllipticPoint // serial number
	comSK    *privacy.EllipticPoint // commitment to private key
	comInput *privacy.EllipticPoint // commitment to input of the pseudo-random function
}

type SNPrivacyWitness struct {
	stmt *SNPrivacyStatement // statement to be proved

	sk     *big.Int // private key
	rSK    *big.Int // blinding factor in the commitment to private key
	input  *big.Int // input of pseudo-random function
	rInput *big.Int // blinding factor in the commitment to input
}

type SNPrivacyProof struct {
	stmt *SNPrivacyStatement // statement to be proved

	tSK    *privacy.EllipticPoint // random commitment related to private key
	tInput *privacy.EllipticPoint // random commitment related to input
	tSN    *privacy.EllipticPoint // random commitment related to serial number

	zSK     *big.Int // first challenge-dependent information to open the commitment to private key
	zRSK    *big.Int // second challenge-dependent information to open the commitment to private key
	zInput  *big.Int // first challenge-dependent information to open the commitment to input
	zRInput *big.Int // second challenge-dependent information to open the commitment to input
}

// ValidateSanity validates sanity of proof
func (proof *SNPrivacyProof) ValidateSanity() bool {
	if !proof.stmt.sn.IsSafe() {
		return false
	}
	if !proof.stmt.comSK.IsSafe() {
		return false
	}
	if !proof.stmt.comInput.IsSafe() {
		return false
	}

	if !proof.tSK.IsSafe() {
		return false
	}
	if !proof.tInput.IsSafe() {
		return false
	}
	if !proof.tSN.IsSafe() {
		return false
	}

	if proof.zSK.BitLen() > 256 {
		return false
	}
	if proof.zRSK.BitLen() > 256 {
		return false
	}
	if proof.zInput.BitLen() > 256 {
		return false
	}
	return proof.zRInput.BitLen() <= 256
}

func (proof *SNPrivacyProof) isNil() bool {
	if proof.stmt.sn == nil {
		return true
	}
	if proof.stmt.comSK == nil {
		return true
	}
	if proof.stmt.comInput == nil {
		return true
	}
	if proof.tSK == nil {
		return true
	}
	if proof.tInput == nil {
		return true
	}
	if proof.tSN == nil {
		return true
	}
	if proof.zSK == nil {
		return true
	}
	if proof.zRSK == nil {
		return true
	}
	if proof.zInput == nil {
		return true
	}
	return proof.zRInput == nil
}

// Init inits Proof
func (proof *SNPrivacyProof) Init() *SNPrivacyProof {
	proof.stmt = new(SNPrivacyStatement)

	proof.tSK = new(privacy.EllipticPoint)
	proof.tInput = new(privacy.EllipticPoint)
	proof.tSN = new(privacy.EllipticPoint)

	proof.zSK = new(big.Int)
	proof.zRSK = new(big.Int)
	proof.zInput = new(big.Int)
	proof.zRInput = new(big.Int)

	return proof
}

// Set sets Statement
func (stmt *SNPrivacyStatement) Set(
	SN *privacy.EllipticPoint,
	comSK *privacy.EllipticPoint,
	comInput *privacy.EllipticPoint) {
	stmt.sn = SN
	stmt.comSK = comSK
	stmt.comInput = comInput
}

// Set sets Witness
func (wit *SNPrivacyWitness) Set(
	stmt *SNPrivacyStatement,
	SK *big.Int,
	rSK *big.Int,
	input *big.Int,
	rInput *big.Int) {

	wit.stmt = stmt
	wit.sk = SK
	wit.rSK = rSK
	wit.input = input
	wit.rInput = rInput
}

// Set sets Proof
func (proof *SNPrivacyProof) Set(
	stmt *SNPrivacyStatement,
	tSK *privacy.EllipticPoint,
	tInput *privacy.EllipticPoint,
	tSN *privacy.EllipticPoint,
	zSK *big.Int,
	zRSK *big.Int,
	zInput *big.Int,
	zRInput *big.Int) {
	proof.stmt = stmt
	proof.tSK = tSK
	proof.tInput = tInput
	proof.tSN = tSN

	proof.zSK = zSK
	proof.zRSK = zRSK
	proof.zInput = zInput
	proof.zRInput = zRInput
}

func (proof *SNPrivacyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if proof.isNil() {
		return []byte{}
	}

	var bytes []byte
	bytes = append(bytes, proof.stmt.sn.Compress()...)
	bytes = append(bytes, proof.stmt.comSK.Compress()...)
	bytes = append(bytes, proof.stmt.comInput.Compress()...)

	bytes = append(bytes, proof.tSK.Compress()...)
	bytes = append(bytes, proof.tInput.Compress()...)
	bytes = append(bytes, proof.tSN.Compress()...)

	bytes = append(bytes, privacy.AddPaddingBigInt(proof.zSK, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(proof.zRSK, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(proof.zInput, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(proof.zRInput, privacy.BigIntSize)...)

	return bytes
}

func (proof *SNPrivacyProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("Bytes array is empty")
	}

	offset := 0
	var err error

	proof.stmt.sn = new(privacy.EllipticPoint)
	err = proof.stmt.sn.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.stmt.comSK = new(privacy.EllipticPoint)
	err = proof.stmt.comSK.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.stmt.comInput = new(privacy.EllipticPoint)
	err = proof.stmt.comInput.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.tSK = new(privacy.EllipticPoint)
	err = proof.tSK.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.tInput = new(privacy.EllipticPoint)
	err = proof.tInput.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.tSN = new(privacy.EllipticPoint)
	err = proof.tSN.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	proof.zSK.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	proof.zRSK.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	proof.zInput.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	proof.zRInput.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize
	return nil
}

func (wit *SNPrivacyWitness) Prove(mess []byte) (*SNPrivacyProof, error) {
	// randomness
	eSK := privacy.RandScalar()
	eSND := privacy.RandScalar()
	dSK := privacy.RandScalar()
	dSND := privacy.RandScalar()

	// calculate tSeed = g_SK^eSK * h^dSK
	tSeed := privacy.PedCom.CommitAtIndex(eSK, dSK, privacy.SK)

	// calculate tSND = g_SND^eSND * h^dSND
	tInput := privacy.PedCom.CommitAtIndex(eSND, dSND, privacy.SND)

	// calculate tSND = g_SK^eSND * h^dSND2
	tOutput := wit.stmt.sn.ScalarMult(new(big.Int).Add(eSK, eSND))

	// calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	var x *big.Int
	if mess == nil {
		x = generateChallenge([][]byte{tSeed.Compress(), tInput.Compress(), tOutput.Compress()})
	} else {
		x = big.NewInt(0).SetBytes(mess)
	}

	// Calculate zSeed = sk * x + eSK
	zSeed := new(big.Int).Mul(wit.sk, x)
	zSeed.Add(zSeed, eSK)
	zSeed.Mod(zSeed, privacy.Curve.Params().N)

	// Calculate zRSeed = rSK * x + dSK
	zRSeed := new(big.Int).Mul(wit.rSK, x)
	zRSeed.Add(zRSeed, dSK)
	zRSeed.Mod(zRSeed, privacy.Curve.Params().N)

	// Calculate zInput = input * x + eSND
	zInput := new(big.Int).Mul(wit.input, x)
	zInput.Add(zInput, eSND)
	zInput.Mod(zInput, privacy.Curve.Params().N)

	// Calculate zRInput = rInput * x + dSND
	zRInput := new(big.Int).Mul(wit.rInput, x)
	zRInput.Add(zRInput, dSND)
	zRInput.Mod(zRInput, privacy.Curve.Params().N)

	proof := new(SNPrivacyProof).Init()
	proof.Set(wit.stmt, tSeed, tInput, tOutput, zSeed, zRSeed, zInput, zRInput)
	return proof, nil
}

func (proof *SNPrivacyProof) Verify(mess []byte) bool {
	// re-calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	var x *big.Int
	if mess == nil {
		x = generateChallenge([][]byte{proof.tSK.Compress(), proof.tInput.Compress(), proof.tSN.Compress()})
	} else {
		x = big.NewInt(0).SetBytes(mess)
	}

	// Check gSND^zInput * h^zRInput = input^x * tInput
	leftPoint1 := privacy.PedCom.CommitAtIndex(proof.zInput, proof.zRInput, privacy.SND)

	rightPoint1 := proof.stmt.comInput.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(proof.tInput)

	if !leftPoint1.IsEqual(rightPoint1) {
		return false
	}

	// Check gSK^zSeed * h^zRSeed = vKey^x * tSeed
	leftPoint2 := privacy.PedCom.CommitAtIndex(proof.zSK, proof.zRSK, privacy.SK)

	rightPoint2 := proof.stmt.comSK.ScalarMult(x)
	rightPoint2 = rightPoint2.Add(proof.tSK)

	if !leftPoint2.IsEqual(rightPoint2) {
		return false
	}

	// Check sn^(zSeed + zInput) = gSK^x * tOutput
	leftPoint3 := proof.stmt.sn.ScalarMult(new(big.Int).Add(proof.zSK, proof.zInput))

	rightPoint3 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint3 = rightPoint3.Add(proof.tSN)

	return leftPoint3.IsEqual(rightPoint3)
}
