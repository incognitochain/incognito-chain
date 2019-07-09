package zkp

import (
	"errors"
	"github.com/incognitochain/incognito-chain/privacy"
	"math/big"
)

type SNNoPrivacyStatement struct {
	output *privacy.EllipticPoint
	vKey   *privacy.EllipticPoint
	input  *big.Int
}

// SNNoPrivacyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type SNNoPrivacyWitness struct {
	stmt SNNoPrivacyStatement
	seed *big.Int
}

// SNNoPrivacyProof contains Proof's value
type SNNoPrivacyProof struct {
	// general info
	stmt SNNoPrivacyStatement

	tSeed   *privacy.EllipticPoint
	tOutput *privacy.EllipticPoint

	zSeed *big.Int
}

func (proof *SNNoPrivacyProof) ValidateSanity() bool {
	if !proof.stmt.output.IsSafe() {
		return false
	}
	if !proof.stmt.vKey.IsSafe() {
		return false
	}
	if proof.stmt.input.BitLen() > 256 {
		return false
	}

	if !proof.tSeed.IsSafe() {
		return false
	}
	if !proof.tOutput.IsSafe() {
		return false
	}
	return proof.zSeed.BitLen() <= 256
}

func (pro *SNNoPrivacyProof) isNil() bool {
	if pro.stmt.output == nil {
		return true
	}
	if pro.stmt.vKey == nil {
		return true
	}
	if pro.stmt.input == nil {
		return true
	}
	if pro.tSeed == nil {
		return true
	}
	if pro.tOutput == nil {
		return true
	}
	if pro.zSeed == nil {
		return true
	}
	return false
}

func (pro *SNNoPrivacyProof) Init() *SNNoPrivacyProof {
	pro.stmt.output = new(privacy.EllipticPoint)
	pro.stmt.vKey = new(privacy.EllipticPoint)
	pro.stmt.input = new(big.Int)

	pro.tSeed = new(privacy.EllipticPoint)
	pro.tOutput = new(privacy.EllipticPoint)

	pro.zSeed = new(big.Int)

	return pro
}

// Set sets Witness
func (wit *SNNoPrivacyWitness) Set(
	output *privacy.EllipticPoint,
	vKey *privacy.EllipticPoint,
	input *big.Int,
	seed *big.Int) {

	if wit == nil {
		wit = new(SNNoPrivacyWitness)
	}

	wit.stmt.output = output
	wit.stmt.vKey = vKey
	wit.stmt.input = input

	wit.seed = seed
}

// Set sets Proof
func (pro *SNNoPrivacyProof) Set(
	output *privacy.EllipticPoint,
	vKey *privacy.EllipticPoint,
	input *big.Int,
	tSeed *privacy.EllipticPoint,
	tOutput *privacy.EllipticPoint,
	zSeed *big.Int) {

	if pro == nil {
		pro = new(SNNoPrivacyProof)
	}

	pro.stmt.output = output
	pro.stmt.vKey = vKey
	pro.stmt.input = input

	pro.tSeed = tSeed
	pro.tOutput = tOutput

	pro.zSeed = zSeed
}

func (pro *SNNoPrivacyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if pro.isNil() {
		return []byte{}
	}

	var bytes []byte
	bytes = append(bytes, pro.stmt.output.Compress()...)
	bytes = append(bytes, pro.stmt.vKey.Compress()...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.stmt.input, privacy.BigIntSize)...)

	bytes = append(bytes, pro.tSeed.Compress()...)
	bytes = append(bytes, pro.tOutput.Compress()...)

	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zSeed, privacy.BigIntSize)...)

	return bytes
}

func (pro *SNNoPrivacyProof) SetBytes(bytes []byte) error {
	if len(bytes) == 0 {
		return errors.New("Bytes array is empty")
	}

	offset := 0
	var err error

	pro.stmt.output = new(privacy.EllipticPoint)
	err = pro.stmt.output.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.stmt.vKey = new(privacy.EllipticPoint)
	err = pro.stmt.vKey.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.stmt.input.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	pro.tSeed = new(privacy.EllipticPoint)

	err = pro.tSeed.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tOutput = new(privacy.EllipticPoint)
	err = pro.tOutput.Decompress(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.zSeed.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}

	return nil
}

func (wit *SNNoPrivacyWitness) Prove(mess []byte) (*SNNoPrivacyProof, error) {
	// randomness
	eSK := privacy.RandScalar()

	// calculate tSeed = g_SK^eSK
	tSK := privacy.PedCom.G[privacy.SK].ScalarMult(eSK)

	// calculate tOutput = sn^eSK
	tE := wit.stmt.output.ScalarMult(eSK)

	x := big.NewInt(0)
	if mess == nil {
		// calculate x = hash(tSeed || tInput || tSND2 || tOutput)
		x.Set(generateChallenge([][]byte{tSK.Compress(), tE.Compress()}))
	} else {
		x.SetBytes(mess)
	}

	// Calculate zSeed = SK * x + eSK
	zSK := new(big.Int).Mul(wit.seed, x)
	zSK.Add(zSK, eSK)
	zSK.Mod(zSK, privacy.Curve.Params().N)

	proof := new(SNNoPrivacyProof).Init()
	proof.Set(wit.stmt.output, wit.stmt.vKey, wit.stmt.input, tSK, tE, zSK)
	return proof, nil
}

func (pro *SNNoPrivacyProof) Verify(mess []byte) bool {
	// re-calculate x = hash(tSeed || tOutput)
	x := big.NewInt(0)
	if mess == nil {
		// calculate x = hash(tSeed || tInput || tSND2 || tOutput)
		x.Set(generateChallenge([][]byte{pro.tSeed.Compress(), pro.tOutput.Compress()}))
	} else {
		x.SetBytes(mess)
	}

	// Check gSK^zSeed = vKey^x * tSeed
	leftPoint1 := privacy.PedCom.G[privacy.SK].ScalarMult(pro.zSeed)

	rightPoint1 := pro.stmt.vKey.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(pro.tSeed)

	if !leftPoint1.IsEqual(rightPoint1) {
		privacy.Logger.Log.Errorf("Failed verify serial number no privacy 1")
		return false
	}

	// Check sn^(zSeed + x*input) = gSK^x * tOutput
	leftPoint2 := pro.stmt.output.ScalarMult(new(big.Int).Add(pro.zSeed, new(big.Int).Mul(x, pro.stmt.input)))

	rightPoint2 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint2 = rightPoint2.Add(pro.tOutput)

	if !leftPoint2.IsEqual(rightPoint2) {
		privacy.Logger.Log.Errorf("Failed verify serial number no privacy 1")
		return false
	}

	return true
}
