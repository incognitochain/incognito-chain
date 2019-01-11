package zkp

import (
	"fmt"
	"github.com/ninjadotorg/constant/privacy"
	"math/big"
	"time"
)

// OneOutOfManyWitness is a protocol for Zero-knowledge Proof of Knowledge of one out of many commitments containing 0
// include Witness: CommitedValue, r []byte
type PKSNPrivacyWitness struct {
	// general info
	output   *privacy.EllipticPoint
	comSeed  *privacy.EllipticPoint
	comInput *privacy.EllipticPoint

	seed   *big.Int
	rSeed  *big.Int
	input  *big.Int
	rInput *big.Int
}

// OneOutOfManyProof contains Proof's value
type PKSNPrivacyProof struct {
	// general info
	output   *privacy.EllipticPoint
	comSeed  *privacy.EllipticPoint
	comInput *privacy.EllipticPoint

	tSeed   *privacy.EllipticPoint
	tInput  *privacy.EllipticPoint
	tOutput *privacy.EllipticPoint

	zSeed   *big.Int
	zRSeed  *big.Int
	zInput  *big.Int
	zRInput *big.Int
}

func (pro *PKSNPrivacyProof) isNil() bool {
	if pro.output == nil {
		return true
	}
	if pro.comSeed == nil {
		return true
	}
	if pro.comInput == nil {
		return true
	}

	if pro.tSeed == nil {
		return true
	}
	if pro.tInput == nil {
		return true
	}
	if pro.tOutput == nil {
		return true
	}
	if pro.zSeed == nil {
		return true
	}
	if pro.zRSeed == nil {
		return true
	}
	if pro.zInput == nil {
		return true
	}
	if pro.zRInput == nil {
		return true
	}
	return false
}

func (pro *PKSNPrivacyProof) Init() *PKSNPrivacyProof {
	pro.output = new(privacy.EllipticPoint)
	pro.comSeed = new(privacy.EllipticPoint)
	pro.comInput = new(privacy.EllipticPoint)

	pro.tSeed = new(privacy.EllipticPoint)
	pro.tInput = new(privacy.EllipticPoint)
	pro.tOutput = new(privacy.EllipticPoint)

	pro.zSeed = new(big.Int)
	pro.zRSeed = new(big.Int)
	pro.zInput = new(big.Int)
	pro.zRInput = new(big.Int)

	return pro
}

// Set sets Witness
func (wit *PKSNPrivacyWitness) Set(
	output *privacy.EllipticPoint,
	comSeed *privacy.EllipticPoint,
	comInput *privacy.EllipticPoint,
	seed *big.Int,
	rSeed *big.Int,
	input *big.Int,
	rInput *big.Int) {

	if wit == nil {
		wit = new(PKSNPrivacyWitness)
	}

	wit.output = output
	wit.comSeed = comSeed
	wit.comInput = comInput

	wit.seed = seed
	wit.rSeed = rSeed
	wit.input = input
	wit.rInput = rInput
}

// Set sets Proof
func (pro *PKSNPrivacyProof) Set(
	output *privacy.EllipticPoint,
	comSeed *privacy.EllipticPoint,
	comInput *privacy.EllipticPoint,
	tSeed *privacy.EllipticPoint,
	tInput *privacy.EllipticPoint,
	tOutput *privacy.EllipticPoint,
	zSeed *big.Int,
	zRSeed *big.Int,
	zInput *big.Int,
	zRInput *big.Int ) {

	if pro == nil {
		pro = new(PKSNPrivacyProof)
	}

	pro.output = output
	pro.comSeed = comSeed
	pro.comInput = comInput

	pro.tSeed = tSeed
	pro.tInput = tInput
	pro.tOutput = tOutput

	pro.zSeed = zSeed
	pro.zRSeed = zRSeed
	pro.zInput = zInput
	pro.zRInput = zRInput
}

func (pro *PKSNPrivacyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if pro.isNil() {
		return []byte{}
	}

	var bytes []byte
	bytes = append(bytes, pro.output.Compress()...)
	bytes = append(bytes, pro.comSeed.Compress()...)
	bytes = append(bytes, pro.comInput.Compress()...)

	bytes = append(bytes, pro.tSeed.Compress()...)
	bytes = append(bytes, pro.tInput.Compress()...)
	bytes = append(bytes, pro.tOutput.Compress()...)

	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zSeed, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zRSeed, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zInput, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pro.zRInput, privacy.BigIntSize)...)

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

	pro.output, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.comSeed, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.comInput, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tSeed, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tInput, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.tOutput, err = privacy.DecompressKey(bytes[offset: offset + privacy.CompressedPointSize])
	if err != nil{
		return err
	}
	offset += privacy.CompressedPointSize

	pro.zSeed.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize

	pro.zRSeed.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize

	pro.zInput.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize

	pro.zRInput.SetBytes(bytes[offset: offset + privacy.BigIntSize])
	if err != nil{
		return err
	}
	offset += privacy.BigIntSize
	return nil
}

func (wit *PKSNPrivacyWitness) Prove() (*PKSNPrivacyProof, error){
	start := time.Now()

	// randomness
	eSK := privacy.RandInt()
	eSND := privacy.RandInt()
	dSK := privacy.RandInt()
	dSND := privacy.RandInt()

	// calculate tSeed = g_SK^eSK * h^dSK
	tSeed := privacy.PedCom.CommitAtIndex(eSK, dSK, privacy.SK)

	// calculate tSND = g_SND^eSND * h^dSND
	tInput := privacy.PedCom.CommitAtIndex(eSND, dSND, privacy.SND)

	// calculate tSND = g_SK^eSND * h^dSND2
	tOutput := wit.output.ScalarMult(new(big.Int).Add(eSK, eSND))

	// calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	x := generateChallengeFromPoint([]*privacy.EllipticPoint{tSeed, tInput, tOutput})

	// Calculate zSeed = SK * x + eSK
	zSeed := new(big.Int).Mul(wit.seed, x)
	zSeed.Add(zSeed, eSK)
	zSeed.Mod(zSeed, privacy.Curve.Params().N)

	// Calculate zRSeed = rSeed * x + dSK
	zRSeed := new(big.Int).Mul(wit.rSeed, x)
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

	proof := new(PKSNPrivacyProof).Init()
	proof.Set(wit.output, wit.comSeed, wit.comInput, tSeed, tInput, tOutput, zSeed, zRSeed, zInput, zRInput)
	end := time.Since(start)
	fmt.Printf("Serial number proving time: %v\n", end)
	return proof, nil
}

func (pro *PKSNPrivacyProof) Verify() bool{
	start := time.Now()
	// re-calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	x := generateChallengeFromPoint([]*privacy.EllipticPoint{pro.tSeed, pro.tInput, pro.tOutput})

	// Check gSND^zInput * h^zRInput = input^x * tInput
	leftPoint1 := privacy.PedCom.CommitAtIndex(pro.zInput, pro.zRInput, privacy.SND)

	rightPoint1 := pro.comInput.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(pro.tInput)

	if !leftPoint1.IsEqual(rightPoint1){
		return false
	}

	// Check gSK^zSeed * h^zRSeed = vKey^x * tSeed
	leftPoint3 := privacy.PedCom.CommitAtIndex(pro.zSeed, pro.zRSeed, privacy.SK)

	rightPoint3 := pro.comSeed.ScalarMult(x)
	rightPoint3 = rightPoint3.Add(pro.tSeed)

	if !leftPoint3.IsEqual(rightPoint3){
		return false
	}

	// Check SN^(zSeed + zInput) = gSK^x * tOutput
	leftPoint4 := pro.output.ScalarMult(new(big.Int).Add(pro.zSeed, pro.zInput))

	rightPoint4 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint4 = rightPoint4.Add(pro.tOutput)

	if !leftPoint4.IsEqual(rightPoint4){
		return false
	}

	end := time.Since(start)
	fmt.Printf("Serial number verification time: %v\n", end)

	return true
}
