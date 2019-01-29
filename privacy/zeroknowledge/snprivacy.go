package zkp

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ninjadotorg/constant/privacy"
)

type PKSNPrivacyStatement struct {
	SN       *privacy.EllipticPoint // serial number
	comSK    *privacy.EllipticPoint // commitment to private key
	comInput *privacy.EllipticPoint // commitment to input of the pseudo-random function
}

type PKSNPrivacyWitness struct {
	stmt   *PKSNPrivacyStatement // statement to be proved
	SK     *big.Int              // private key
	rSK    *big.Int              // blinding factor in the commitment to private key
	input  *big.Int              // input of pseudo-random function
	rInput *big.Int              // blinding factor in the commitment to input
}

type PKSNPrivacyProof struct {
	stmt   *PKSNPrivacyStatement  // statement to be proved
	tSK    *privacy.EllipticPoint // random commitment related to private key
	tInput *privacy.EllipticPoint // random commitment related to input
	tSN    *privacy.EllipticPoint // random commitment related to serial number

	zSK     *big.Int // first challenge-dependent information to open the commitment to private key
	zRSK    *big.Int // second challenge-dependent information to open the commitment to private key
	zInput  *big.Int // first challenge-dependent information to open the commitment to input
	zRInput *big.Int // second challenge-dependent information to open the commitment to input
}

func (pi *PKSNPrivacyProof) isNil() bool {
	if pi.stmt.SN == nil {
		return true
	}
	if pi.stmt.SN == nil {
		return true
	}
	if pi.stmt.comInput == nil {
		return true
	}

	if pi.tSK == nil {
		return true
	}
	if pi.tInput == nil {
		return true
	}
	if pi.tSN == nil {
		return true
	}
	if pi.zSK == nil {
		return true
	}
	if pi.zRSK == nil {
		return true
	}
	if pi.zInput == nil {
		return true
	}
	if pi.zRInput == nil {
		return true
	}
	return false
}

// Init inits Proof
func (pi *PKSNPrivacyProof) Init() *PKSNPrivacyProof {
	pi.stmt = new(PKSNPrivacyStatement)
	pi.tSK = new(privacy.EllipticPoint)
	pi.tInput = new(privacy.EllipticPoint)
	pi.tSN = new(privacy.EllipticPoint)

	pi.zSK = new(big.Int)
	pi.zRSK = new(big.Int)
	pi.zInput = new(big.Int)
	pi.zRInput = new(big.Int)

	return pi
}

// Set sets Statement
func (stmt *PKSNPrivacyStatement) Set(
	SN *privacy.EllipticPoint,
	comSK *privacy.EllipticPoint,
	comInput *privacy.EllipticPoint) {
	stmt.SN = SN
	stmt.comSK = comSK
	stmt.comInput = comInput
}

// Set sets Witness
func (wit *PKSNPrivacyWitness) Set(
	stmt *PKSNPrivacyStatement,
	SK *big.Int,
	rSK *big.Int,
	input *big.Int,
	rInput *big.Int) {

	if wit == nil {
		wit = new(PKSNPrivacyWitness)
	}

	wit.stmt = stmt
	wit.SK = SK
	wit.rSK = rSK
	wit.input = input
	wit.rInput = rInput
}

// Set sets Proof
func (pi *PKSNPrivacyProof) Set(
	stmt *PKSNPrivacyStatement,
	tSK *privacy.EllipticPoint,
	tInput *privacy.EllipticPoint,
	tSN *privacy.EllipticPoint,
	zSK *big.Int,
	zRSK *big.Int,
	zInput *big.Int,
	zRInput *big.Int) {

	if pi == nil {
		pi = new(PKSNPrivacyProof)
	}

	pi.stmt = stmt
	pi.tSK = tSK
	pi.tInput = tInput
	pi.tSN = tSN

	pi.zSK = zSK
	pi.zRSK = zRSK
	pi.zInput = zInput
	pi.zRInput = zRInput
}

func (pi *PKSNPrivacyProof) Bytes() []byte {
	// if proof is nil, return an empty array
	if pi.isNil() {
		return []byte{}
	}

	var bytes []byte
	bytes = append(bytes, pi.stmt.SN.Compress()...)
	bytes = append(bytes, pi.stmt.comSK.Compress()...)
	bytes = append(bytes, pi.stmt.comInput.Compress()...)

	bytes = append(bytes, pi.tSK.Compress()...)
	bytes = append(bytes, pi.tInput.Compress()...)
	bytes = append(bytes, pi.tSN.Compress()...)

	bytes = append(bytes, privacy.AddPaddingBigInt(pi.zSK, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pi.zRSK, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pi.zInput, privacy.BigIntSize)...)
	bytes = append(bytes, privacy.AddPaddingBigInt(pi.zRInput, privacy.BigIntSize)...)

	return bytes
}

func (pi *PKSNPrivacyProof) SetBytes(bytes []byte) error {
	if pi == nil {
		pi = pi.Init()
	}

	if len(bytes) == 0 {
		return nil
	}

	offset := 0
	var err error

	pi.stmt.SN, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pi.stmt.comSK, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pi.stmt.comInput, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pi.tSK, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pi.tInput, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pi.tSN, err = privacy.DecompressKey(bytes[offset : offset+privacy.CompressedPointSize])
	if err != nil {
		return err
	}
	offset += privacy.CompressedPointSize

	pi.zSK.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	pi.zRSK.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	pi.zInput.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize

	pi.zRInput.SetBytes(bytes[offset : offset+privacy.BigIntSize])
	if err != nil {
		return err
	}
	offset += privacy.BigIntSize
	return nil
}

func (wit *PKSNPrivacyWitness) Prove(mess []byte) (*PKSNPrivacyProof, error) {
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
	tOutput := wit.stmt.SN.ScalarMult(new(big.Int).Add(eSK, eSND))

	// calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	x := new(big.Int)
	if mess == nil {
		x = generateChallengeFromPoint([]*privacy.EllipticPoint{tSeed, tInput, tOutput})
	} else {
		x = big.NewInt(0).SetBytes(mess)
	}

	// Calculate zSeed = SK * x + eSK
	zSeed := new(big.Int).Mul(wit.SK, x)
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

	proof := new(PKSNPrivacyProof).Init()
	proof.Set(wit.stmt, tSeed, tInput, tOutput, zSeed, zRSeed, zInput, zRInput)
	end := time.Since(start)
	fmt.Printf("Serial number proving time: %v\n", end)
	return proof, nil
}

func (pi *PKSNPrivacyProof) Verify(mess []byte) bool {
	start := time.Now()
	// re-calculate x = hash(tSeed || tInput || tSND2 || tOutput)
	x := new(big.Int)
	if mess == nil {
		x = generateChallengeFromPoint([]*privacy.EllipticPoint{pi.tSK, pi.tInput, pi.tSN})
	} else {
		x = big.NewInt(0).SetBytes(mess)
	}

	// Check gSND^zInput * h^zRInput = input^x * tInput
	leftPoint1 := privacy.PedCom.CommitAtIndex(pi.zInput, pi.zRInput, privacy.SND)

	rightPoint1 := pi.stmt.comInput.ScalarMult(x)
	rightPoint1 = rightPoint1.Add(pi.tInput)

	if !leftPoint1.IsEqual(rightPoint1) {
		return false
	}

	// Check gSK^zSeed * h^zRSeed = vKey^x * tSeed
	leftPoint3 := privacy.PedCom.CommitAtIndex(pi.zSK, pi.zRSK, privacy.SK)

	rightPoint3 := pi.stmt.comSK.ScalarMult(x)
	rightPoint3 = rightPoint3.Add(pi.tSK)

	if !leftPoint3.IsEqual(rightPoint3) {
		return false
	}

	// Check SN^(zSeed + zInput) = gSK^x * tOutput
	leftPoint4 := pi.stmt.SN.ScalarMult(new(big.Int).Add(pi.zSK, pi.zInput))

	rightPoint4 := privacy.PedCom.G[privacy.SK].ScalarMult(x)
	rightPoint4 = rightPoint4.Add(pi.tSN)

	if !leftPoint4.IsEqual(rightPoint4) {
		return false
	}

	end := time.Since(start)
	fmt.Printf("Serial number verification time: %v\n", end)

	return true
}
