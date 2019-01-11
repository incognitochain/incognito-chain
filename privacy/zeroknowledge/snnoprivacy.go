package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
	"time"
)

// SNNoPrivacyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type SNNoPrivacyWitness struct {
	// general info
	output *privacy.EllipticPoint
	vKey   *privacy.EllipticPoint
	input  *big.Int

	seed *big.Int
}

// SNNoPrivacyProof contains Proof's value
type SNNoPrivacyProof struct {
	// general info
	output *privacy.EllipticPoint
	vKey   *privacy.EllipticPoint
	input  *big.Int

	tSeed   *privacy.EllipticPoint
	tOutput *privacy.EllipticPoint

	zSeed *big.Int
}

func (pro *SNNoPrivacyProof) isNil() bool {
	if pro.output == nil {
		return true
	}
	if pro.vKey == nil {
		return true
	}
	if pro.input == nil {
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
	pro.output = new(privacy.EllipticPoint)
	pro.vKey = new(privacy.EllipticPoint)
	pro.input = new(big.Int)

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

	wit.output = output
	wit.vKey = vKey
	wit.input = input

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

	pro.output = output
	pro.vKey = vKey
	pro.input = input

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
	bytes = append(bytes, pro.output.Compress()...)
	bytes = append(bytes, pro.vKey.Compress()...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.input, privacy.BigIntSize)...)

	bytes = append(bytes, pro.tSeed.Compress()...)
	bytes = append(bytes, pro.tOutput.Compress()...)

	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zSeed, privacy.BigIntSize)...)

	return bytes
}

func (pro *SNNoPrivacyProof) SetBytes(bytes []byte) error {
	if pro == nil {
		pro = pro.Init()
	}

	if len(bytes) == 0 {
		return nil
	}

	offset := 0
	var err error

	pro.output, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.vKey, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.input.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	pro.tSeed, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tOutput, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
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

func (wit *SNNoPrivacyWitness) Prove() (*SNNoPrivacyProof, error) {
	start := time.Now()
	// randomness
	eSK := privacy.RandInt()

	// calculate tSeed = g_SK^eSK
	tSK := privacy.PedCom.G[privacy.SK].ScalarMult(eSK)

	// calculate tOutput = SN^eSK
	tE := wit.output.ScalarMult(eSK)

	// calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	x := generateChallengeFromPoint([]*privacy.EllipticPoint{tSK, tE})

	// Calculate zSeed = SK * x + eSK
	zSK := new(big.Int).Mul(wit.seed, x)
	zSK.Add(zSK, eSK)
	zSK.Mod(zSK, privacy.Curve.Params().N)

	proof := new(SNNoPrivacyProof).Init()
	proof.Set(wit.output, wit.vKey, wit.input, tSK, tE, zSK)
	end := time.Since(start)
	fmt.Printf("Serial number no privacy proving time: %v\n", end)
	return proof, nil
}

func (pro *SNNoPrivacyProof) Verify() bool {
	start := time.Now()
	// re-calculate x = hash(tSeed || tOutput)
	x := generateChallengeFromPoint([]*privacy.EllipticPoint{pro.tSeed, pro.tOutput})

	// Check gSK^zSeed = vKey^x * tSeed
	leftPoint1 := privacy.PedCom.G[privacy.SK].ScalarMult(pro.zSeed)

	rightPoint1 := pro.vKey.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(pro.tSeed)

	if !leftPoint1.IsEqual(rightPoint1) {
		return false
	}

	// Check SN^(zSeed + x*input) = gSK^x * tOutput
	leftPoint2 := pro.output.ScalarMult(new(big.Int).Add(pro.zSeed, new(big.Int).Mul(x, pro.input)))

	rightPoint4 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint4 = rightPoint4.Add(pro.tOutput)

	if !leftPoint2.IsEqual(rightPoint4) {
		return false
	}
	end := time.Since(start)
	fmt.Printf("Serial number no privacy verification time: %v\n", end)


	return true
}
