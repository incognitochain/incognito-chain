package zkp

import (
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
)

// PKOneOfManyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type PKSNPrivacyWitness struct {
	// general info
	serialNumber *privacy.EllipticPoint
	comSK        *privacy.EllipticPoint
	comSND1      *privacy.EllipticPoint
	comSND2      *privacy.EllipticPoint

	sk    *big.Int
	rSK   *big.Int
	snd   *big.Int
	rSND1 *big.Int
	rSND2 *big.Int
}

// PKOneOfManyProof contains Proof's value
type PKSNPrivacyProof struct {
	// general info
	serialNumber *privacy.EllipticPoint
	comSK        *privacy.EllipticPoint
	comSND1      *privacy.EllipticPoint
	comSND2      *privacy.EllipticPoint

	tSK   *privacy.EllipticPoint
	tSND1 *privacy.EllipticPoint
	tSND2 *privacy.EllipticPoint
	tE    *privacy.EllipticPoint

	zSK    *big.Int
	zRSK *big.Int
	zSND   *big.Int
	zRSND1 *big.Int
	zRSND2 *big.Int
}

func (pro *PKSNPrivacyProof) IsNil() bool {
	if pro.serialNumber == nil {
		return true
	}
	if pro.comSK == nil {
		return true
	}
	if pro.comSND1 == nil {
		return true
	}

	if pro.comSND2 == nil {
		return true
	}
	if pro.tSK == nil {
		return true
	}
	if pro.tSND1 == nil {
		return true
	}
	if pro.tSND2 == nil {
		return true
	}
	if pro.tE == nil {
		return true
	}
	if pro.zSK == nil {
		return true
	}
	if pro.zRSK == nil {
		return true
	}
	if pro.zSND == nil {
		return true
	}
	if pro.zRSND1 == nil {
		return true
	}
	if pro.zRSND2 == nil {
		return true
	}
	return false
}

func (pro *PKSNPrivacyProof) Init() *PKSNPrivacyProof {
	pro.serialNumber = new(privacy.EllipticPoint)
	pro.comSK = new(privacy.EllipticPoint)
	pro.comSND1 = new(privacy.EllipticPoint)
	pro.comSND2 = new(privacy.EllipticPoint)

	pro.tSK = new(privacy.EllipticPoint)
	pro.tSND1 = new(privacy.EllipticPoint)
	pro.tSND2 = new(privacy.EllipticPoint)
	pro.tE = new(privacy.EllipticPoint)

	pro.zSK = new(big.Int)
	pro.zRSK = new(big.Int)
	pro.zSND = new(big.Int)
	pro.zRSND1 = new(big.Int)
	pro.zRSND2 = new(big.Int)

	return pro
}

// Set sets Witness
func (wit *PKSNPrivacyWitness) Set(
	serialNumber *privacy.EllipticPoint,
	comSK        *privacy.EllipticPoint,
	comSND1      *privacy.EllipticPoint,
	comSND2      *privacy.EllipticPoint,
	sk    *big.Int,
	rSK   *big.Int,
	snd   *big.Int,
	rSND1 *big.Int,
	rSND2 *big.Int) {

	if wit == nil {
		wit = new(PKSNPrivacyWitness)
	}

	wit.serialNumber = serialNumber
	wit.comSK = comSK
	wit.comSND1 = comSND1
	wit.comSND2 = comSND2

	wit.sk = sk
	wit.rSK = rSK
	wit.snd = snd
	wit.rSND1 = rSND1
	wit.rSND2 = rSND2
}

// Set sets Proof
func (pro *PKSNPrivacyProof) Set(
	serialNumber *privacy.EllipticPoint,
	comSK        *privacy.EllipticPoint,
	comSND1      *privacy.EllipticPoint,
	comSND2      *privacy.EllipticPoint,
	tSK   *privacy.EllipticPoint,
	tSND1 *privacy.EllipticPoint,
	tSND2 *privacy.EllipticPoint,
	tE    *privacy.EllipticPoint,
	zSK    *big.Int,
	zRSK    *big.Int,
	zSND   *big.Int,
	zRSND1 *big.Int,
	zRSND2 *big.Int ) {

	if pro == nil {
		pro = new(PKSNPrivacyProof)
	}

	pro.serialNumber = serialNumber
	pro.comSK = comSK
	pro.comSND1 = comSND1
	pro.comSND2 = comSND2

	pro.tSK = tSK
	pro.tSND1 = tSND1
	pro.tSND2 = tSND2
	pro.tE = tE

	pro.zSK = zSK
	pro.zRSK = zRSK
	pro.zSND = zSND
	pro.zRSND1 = zRSND1
	pro.zRSND2 = zRSND2
}

func (pro *PKSNPrivacyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if pro.IsNil() {
		return []byte{}
	}

	var bytes []byte
	bytes = append(bytes, pro.serialNumber.Compress()...)
	bytes = append(bytes, pro.comSK.Compress()...)
	bytes = append(bytes, pro.comSND1.Compress()...)
	bytes = append(bytes, pro.comSND2.Compress()...)

	bytes = append(bytes, pro.tSK.Compress()...)
	bytes = append(bytes, pro.tSND1.Compress()...)
	bytes = append(bytes, pro.tSND2.Compress()...)
	bytes = append(bytes, pro.tE.Compress()...)

	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zSK, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zRSK, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zSND, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zRSND1, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zRSND2, privacy.BigIntSize)...)

	return bytes
}

func (pro *PKSNPrivacyProof) SetBytes(bytes []byte) error {
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

	pro.comSK, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.comSND1, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.comSND2, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tSK, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tSND1, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tSND2, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
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
	offset += privacy.BigIntSize

	pro.zRSK.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize

	pro.zSND.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize

	pro.zRSND1.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize

	pro.zRSND2.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	return nil
}

func (wit *PKSNPrivacyWitness) Prove() (*PKSNPrivacyProof, error){
	// randomness
	eSK := privacy.RandInt()
	eSND := privacy.RandInt()
	dSK := privacy.RandInt()
	dSND1 := privacy.RandInt()
	dSND2 := privacy.RandInt()

	// calculate tSK = g_SK^eSK * h^dSK
	tSK := privacy.PedCom.CommitAtIndex(eSK, dSK, privacy.SK)

	// calculate tSND = g_SND^eSND * h^dSND1
	tSND1 := privacy.PedCom.CommitAtIndex(eSND, dSND1, privacy.SND)

	// calculate tSND = g_SK^eSND * h^dSND2
	tSND2 := privacy.PedCom.CommitAtIndex(eSND, dSND2, privacy.SK)

	// calculate tSND = g_SK^eSND * h^dSND2
	tE := wit.serialNumber.ScalarMult(new(big.Int).Add(eSK, eSND))

	// calculate x = hash(tSK || tSND1 || tSND2 || tE)
	x := GenerateChallengeFromPoint([]*privacy.EllipticPoint{tSK, tSND1, tSND2, tE})

	// Calculate zSK = SK * x + eSK
	zSK := new(big.Int).Mul(wit.sk, x)
	zSK.Add(zSK, eSK)
	zSK.Mod(zSK, privacy.Curve.Params().N)

	// Calculate zRSK = rSK * x + dSK
	zRSK := new(big.Int).Mul(wit.rSK, x)
	zRSK.Add(zRSK, dSK)
	zRSK.Mod(zRSK, privacy.Curve.Params().N)

	// Calculate zSND = SND * x + eSND
	zSND := new(big.Int).Mul(wit.snd, x)
	zSND.Add(zSND, eSND)
	zSND.Mod(zSND, privacy.Curve.Params().N)

	// Calculate zRSND1 = rSND1 * x + dSND1
	zRSND1 := new(big.Int).Mul(wit.rSND1, x)
	zRSND1.Add(zRSND1, dSND1)
	zRSND1.Mod(zRSND1, privacy.Curve.Params().N)

	// Calculate zRSND2 = rSND2 * x + dSND2
	zRSND2 := new(big.Int).Mul(wit.rSND2, x)
	zRSND2.Add(zRSND2, dSND2)
	zRSND2.Mod(zRSND2, privacy.Curve.Params().N)

	proof := new(PKSNPrivacyProof).Init()
	proof.Set(wit.serialNumber, wit.comSK, wit.comSND1, wit.comSND2, tSK, tSND1, tSND2, tE, zSK, zRSK, zSND, zRSND1, zRSND2)
	return proof, nil
}

func (pro *PKSNPrivacyProof) Verify() bool{
	// re-calculate x = hash(tSK || tSND1 || tSND2 || tE)
	x := GenerateChallengeFromPoint([]*privacy.EllipticPoint{pro.tSK, pro.tSND1, pro.tSND2, pro.tE})

	// Check gSND^zSND * h^zRSND1 = SND^x * tSND1
	leftPoint1 := privacy.PedCom.CommitAtIndex(pro.zSND, pro.zRSND1, privacy.SND)

	rightPoint1 := pro.comSND1.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(pro.tSND1)

	if !leftPoint1.IsEqual(rightPoint1){
		return false
	}

	// Check gSK^zSND * h^zRSND2 = comSND2^x * tSND2
	leftPoint2 := privacy.PedCom.CommitAtIndex(pro.zSND, pro.zRSND2, privacy.SK)

	rightPoint2 := pro.comSND2.ScalarMult(x)
	rightPoint2 = rightPoint2.Add(pro.tSND2)

	if !leftPoint2.IsEqual(rightPoint2){
		return false
	}

	// Check gSK^zSK * h^zRSK = PK^x * tSK
	leftPoint3 := privacy.PedCom.CommitAtIndex(pro.zSK, pro.zRSK, privacy.SK)

	rightPoint3 := pro.comSK.ScalarMult(x)
	rightPoint3 = rightPoint3.Add(pro.tSK)

	if !leftPoint3.IsEqual(rightPoint3){
		return false
	}

	// Check SN^(zSK + zSND) = gSK^x * tE
	leftPoint4 := pro.serialNumber.ScalarMult(new(big.Int).Add(pro.zSK, pro.zSND))

	rightPoint4 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint4 = rightPoint4.Add(pro.tE)

	if !leftPoint4.IsEqual(rightPoint4){
		return false
	}

	return true
}
