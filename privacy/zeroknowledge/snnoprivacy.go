package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)

// SNNoPrivacyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type SNNoPrivacyWitness struct {
	// general info
	serialNumber *privacy.EllipticPoint
	PK           *privacy.EllipticPoint
	SND          *big.Int

	sk *big.Int
}

// SNNoPrivacyProof contains Proof's value
type SNNoPrivacyProof struct {
	// general info
	serialNumber *privacy.EllipticPoint
	PK           *privacy.EllipticPoint
	SND          *big.Int

	tSK *privacy.EllipticPoint
	tE  *privacy.EllipticPoint

	zSK *big.Int
}

func (pro *SNNoPrivacyProof) isNil() bool {
	if pro.serialNumber == nil {
		return true
	}
	if pro.PK == nil {
		return true
	}
	if pro.SND == nil {
		return true
	}
	if pro.tSK == nil {
		return true
	}
	if pro.tE == nil {
		return true
	}
	if pro.zSK == nil {
		return true
	}
	return false
}

func (pro *SNNoPrivacyProof) Init() *SNNoPrivacyProof {
	pro.serialNumber = new(privacy.EllipticPoint)
	pro.PK = new(privacy.EllipticPoint)
	pro.SND = new(big.Int)

	pro.tSK = new(privacy.EllipticPoint)
	pro.tE = new(privacy.EllipticPoint)

	pro.zSK = new(big.Int)

	return pro
}

// Set sets Witness
func (wit *SNNoPrivacyWitness) Set(
	serialNumber *privacy.EllipticPoint,
	Pk *privacy.EllipticPoint,
	SND *big.Int,
	sk *big.Int) {

	if wit == nil {
		wit = new(SNNoPrivacyWitness)
	}

	wit.serialNumber = serialNumber
	wit.PK = Pk
	wit.SND = SND

	wit.sk = sk
}

// Set sets Proof
func (pro *SNNoPrivacyProof) Set(
	serialNumber *privacy.EllipticPoint,
	PK *privacy.EllipticPoint,
	SND *big.Int,
	tSK *privacy.EllipticPoint,
	tE *privacy.EllipticPoint,
	zSK *big.Int) {

	if pro == nil {
		pro = new(SNNoPrivacyProof)
	}

	pro.serialNumber = serialNumber
	pro.PK = PK
	pro.SND = SND

	pro.tSK = tSK
	pro.tE = tE

	pro.zSK = zSK
}

func (pro *SNNoPrivacyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if pro.isNil() {
		return []byte{}
	}

	var bytes []byte
	bytes = append(bytes, pro.serialNumber.Compress()...)
	bytes = append(bytes, pro.PK.Compress()...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.SND, privacy.BigIntSize)...)

	bytes = append(bytes, pro.tSK.Compress()...)
	bytes = append(bytes, pro.tE.Compress()...)

	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zSK, privacy.BigIntSize)...)

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

	pro.serialNumber, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.PK, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.SND.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	pro.tSK, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tE, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pro.zSK.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}

	return nil
}

func (wit *SNNoPrivacyWitness) Prove() (*SNNoPrivacyProof, error) {
	// randomness
	eSK := privacy.RandInt()

	// calculate tSeed = g_SK^eSK
	tSK := privacy.PedCom.G[privacy.SK].ScalarMult(eSK)

	// calculate tOutput = SN^eSK
	tE := wit.serialNumber.ScalarMult(eSK)

	// calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	x := generateChallengeFromPoint([]*privacy.EllipticPoint{tSK, tE})

	// Calculate zSeed = SK * x + eSK
	zSK := new(big.Int).Mul(wit.sk, x)
	zSK.Add(zSK, eSK)
	zSK.Mod(zSK, privacy.Curve.Params().N)

	proof := new(SNNoPrivacyProof).Init()
	proof.Set(wit.serialNumber, wit.PK, wit.SND, tSK, tE, zSK)
	return proof, nil
}

func (pro *SNNoPrivacyProof) Verify() bool {
	// re-calculate x = hash(tSeed || tOutput)
	x := generateChallengeFromPoint([]*privacy.EllipticPoint{pro.tSK, pro.tE})

	// Check gSK^zSeed = PK^x * tSeed
	leftPoint1 := privacy.PedCom.G[privacy.SK].ScalarMult(pro.zSK)

	rightPoint1 := pro.PK.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(pro.tSK)

	if !leftPoint1.IsEqual(rightPoint1) {
		return false
	}

	// Check SN^(zSeed + x*SND) = gSK^x * tOutput
	leftPoint2 := pro.serialNumber.ScalarMult(new(big.Int).Add(pro.zSK, new(big.Int).Mul(x, pro.SND)))

	rightPoint4 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint4 = rightPoint4.Add(pro.tE)

	if !leftPoint2.IsEqual(rightPoint4) {
		return false
	}

	return true
}
