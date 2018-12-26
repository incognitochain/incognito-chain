package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)

// PKOneOfManyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type PKSNNoPrivacyWitness struct {
	// general info
	serialNumber *privacy.EllipticPoint
	PK           *privacy.EllipticPoint
	SND          *big.Int

	sk    *big.Int
}

// PKOneOfManyProof contains Proof's value
type PKSNNoPrivacyProof struct {
	// general info
	serialNumber *privacy.EllipticPoint
	PK           *privacy.EllipticPoint
	SND          *big.Int

	tSK   *privacy.EllipticPoint
	tE    *privacy.EllipticPoint

	zSK    *big.Int
}

func (pro *PKSNNoPrivacyProof) IsNil() bool {
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

func (pro *PKSNNoPrivacyProof) Init() *PKSNNoPrivacyProof {
	pro.serialNumber = new(privacy.EllipticPoint)
	pro.PK = new(privacy.EllipticPoint)
	pro.SND = new(big.Int)

	pro.tSK = new(privacy.EllipticPoint)
	pro.tE = new(privacy.EllipticPoint)

	pro.zSK = new(big.Int)

	return pro
}

// Set sets Witness
func (wit *PKSNNoPrivacyWitness) Set(
	serialNumber *privacy.EllipticPoint,
	Pk *privacy.EllipticPoint,
	SND *big.Int,
	sk    *big.Int) {

	if wit == nil {
		wit = new(PKSNNoPrivacyWitness)
	}

	wit.serialNumber = serialNumber
	wit.PK = Pk
	wit.SND = SND

	wit.sk = sk
}

// Set sets Proof
func (pro *PKSNNoPrivacyProof) Set(
	serialNumber *privacy.EllipticPoint,
	PK *privacy.EllipticPoint,
	SND *big.Int,
	tSK   *privacy.EllipticPoint,
	tE    *privacy.EllipticPoint,
	zSK    *big.Int ) {

	if pro == nil {
		pro = new(PKSNNoPrivacyProof)
	}

	pro.serialNumber = serialNumber
	pro.PK = PK
	pro.SND = SND

	pro.tSK = tSK
	pro.tE = tE

	pro.zSK = zSK
}

func (pro *PKSNNoPrivacyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if pro.IsNil() {
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

func (pro *PKSNNoPrivacyProof) SetBytes(bytes []byte) error {
	if pro == nil {
		pro = pro.Init()
	}

	if len(bytes) == 0 {
		return nil
	}

	offset := 0
	var err error

	pro.serialNumber, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.PK, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.SND.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize

	pro.tSK, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tE, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.zSK.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}

	return nil
}

func (wit *PKSNNoPrivacyWitness) Prove() (*PKSNNoPrivacyProof, error){
	// randomness
	eSK := privacy.RandInt()

	// calculate tSK = g_SK^eSK
	tSK := privacy.PedCom.G[privacy.SK].ScalarMult(eSK)

	// calculate tE = SN^eSK
	tE := wit.serialNumber.ScalarMult(eSK)

	// calculate x = hash(tSK || tSND1 || tSND2 || tE)
	x := GenerateChallengeFromPoint([]*privacy.EllipticPoint{tSK, tE})

	// Calculate zSK = SK * x + eSK
	zSK := new(big.Int).Mul(wit.sk, x)
	zSK.Add(zSK, eSK)
	zSK.Mod(zSK, privacy.Curve.Params().N)

	proof := new(PKSNNoPrivacyProof).Init()
	proof.Set(wit.serialNumber, wit.PK, wit.SND, tSK, tE, zSK)
	return proof, nil
}

func (pro *PKSNNoPrivacyProof) Verify() bool{
	// re-calculate x = hash(tSK || tE)
	x := GenerateChallengeFromPoint([]*privacy.EllipticPoint{pro.tSK, pro.tE})

	// Check gSK^zSK = PK^x * tSK
	leftPoint1 := privacy.PedCom.G[privacy.SK].ScalarMult(pro.zSK)

	rightPoint1 := pro.PK.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(pro.tSK)

	if !leftPoint1.IsEqual(rightPoint1){
		return false
	}

	// Check SN^(zSK + x*SND) = gSK^x * tE
	leftPoint2 := pro.serialNumber.ScalarMult(new(big.Int).Add(pro.zSK, new(big.Int).Mul(x, pro.SND)))

	rightPoint4 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint4 = rightPoint4.Add(pro.tE)

	if !leftPoint2.IsEqual(rightPoint4){
		return false
	}

	return true
}
