package zkp

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ninjadotorg/constant/privacy-protocol"
)

const(
	VMAX = 18446744073709551615		// 2^64 - 1
)

// PKMaxValue is a protocol for Zero-knowledge Proof of Knowledge of max value is 2^64-1
// include witnesses: commitedValue, r []byte
type PKMaxValueProtocol struct{
	Witnesses [][]byte
	Proof PKMaxValueProof
}

// PKOneOfManyProof contains proof's value
type PKMaxValueProof struct {
	commitments [][]byte
	proofZeroOneCommitments []*PKComZeroOneProof
	PKComZeroOneProtocol *PKComZeroOneProtocol
}

// SetWitness sets witnesses
//func (pro *PKMaxValueProtocol) SetWitness(witnesses [][]byte) {
//	pro.witnesses = make([][]byte, len(witnesses))
//	for i := 0; i < len(witnesses); i++ {
//		pro.witnesses[i] = make([]byte, len(witnesses[i]))
//		copy(pro.witnesses[i], witnesses[i])
//	}
//}

func (pro * PKMaxValueProtocol) Prove(value []byte) (*PKMaxValueProof, error) {
	proof := new(PKMaxValueProof)

	threshold := [4]int{8, 16, 32, 64}

	//max := big.NewInt(0)
	//max.Exp(big.NewInt(2), big.NewInt(64), nil)
	//max.Sub(max, big.NewInt(1))
	//
	//fmt.Printf("max: %v\n", max)

	//convert value from byte array to int64
	vInt, n := binary.Varint(value)
	if uint64(vInt) > VMAX{
		return nil, fmt.Errorf("value must be less than 2^64-1")
	}
	fmt.Printf("vInt is: %v, n is: %v\n", vInt, n)

	vIntTmp := vInt
	//vBinary represent value in binary
	var vBinary []byte

	// find index of bit-length threshold in which i lies
	// nBitsValue: bit-length of value
	nBitsValue := 0
	// indexThreshold: index of bit-length threshold in which i lies
	indexThreshold := 0
	for vIntTmp > 0 {
		vBinary = append(vBinary, byte(vIntTmp % 2))
		nBitsValue += 1
		if nBitsValue > threshold[indexThreshold] {
			indexThreshold++
		}
		vIntTmp = vIntTmp / 2
	}

	// Set witnesses and commit
	privacy.Pcm.InitCommitment()
	witnesses := make([][][]byte, threshold[indexThreshold])
	proof.commitments = make([][]byte, threshold[indexThreshold])
	proof.proofZeroOneCommitments = make([]*PKComZeroOneProof, threshold[indexThreshold])
	proof.PKComZeroOneProtocol = new(PKComZeroOneProtocol)

	for i := 0 ; i < nBitsValue; i++{
		witnesses[i] = [][]byte{big.NewInt(int64(vBinary[i])).Bytes(), privacy.RandBytes(32)}
		proof.commitments[i] = make([]byte, 34)
		fmt.Printf("witness[%v][0] : %v\n", i, witnesses[i][0])
		fmt.Printf("witness[%v][1] : %v\n", i, witnesses[i][1])
		proof.commitments[i] = privacy.Pcm.CommitSpecValue(witnesses[i][0], witnesses[i][1], privacy.VALUE)

		var witness PKComZeroOneWitness
		witness.commitment = proof.commitments[i]
		witness.commitedValue = witnesses[i][0]
		witness.rand = witnesses[i][1]
		witness.index = privacy.VALUE
		proof.PKComZeroOneProtocol.SetWitness(witness)

		proof.proofZeroOneCommitments[i] = new(PKComZeroOneProof)
		proof.proofZeroOneCommitments[i], _ = proof.PKComZeroOneProtocol.Prove()
		fmt.Printf("Proof %v: %+v\n", i, proof.proofZeroOneCommitments[i])
	}

	for j := nBitsValue ; j < threshold[indexThreshold] ; j++{
		witnesses[j] = [][]byte{big.NewInt(0).Bytes(), privacy.RandBytes(32)}
		proof.commitments[j] = make([]byte, 34)
		fmt.Printf("witness[%v][0] : %v\n", j, witnesses[j][0])
		fmt.Printf("witness[%v][1] : %v\n", j, witnesses[j][1])
		proof.commitments[j] = privacy.Pcm.CommitSpecValue(witnesses[j][0], witnesses[j][1], privacy.VALUE)

		var witness PKComZeroOneWitness
		witness.commitment = proof.commitments[j]
		witness.commitedValue = witnesses[j][0]
		witness.rand = witnesses[j][1]
		witness.index = privacy.VALUE
		proof.PKComZeroOneProtocol.SetWitness(witness)

		proof.proofZeroOneCommitments[j] = new(PKComZeroOneProof)
		proof.proofZeroOneCommitments[j], _ = proof.PKComZeroOneProtocol.Prove()
		fmt.Printf("Proof %v: %+v\n", j, proof.proofZeroOneCommitments[j])
	}

	return proof, nil
}

func (pro * PKMaxValueProtocol) SetProof(proof PKMaxValueProof){
	pro.Proof = proof
}

func (pro * PKMaxValueProtocol) Verify(proof PKMaxValueProof) bool {
	if len(proof.commitments) > 64 || len(proof.proofZeroOneCommitments) > 64{
		return false
	}

	if len(proof.commitments) != len(proof.proofZeroOneCommitments) {
		return false
	}

	//pk := new(PKComZeroOneProtocol)

	for i := 0; i < len(proof.commitments); i++{
		res := proof.PKComZeroOneProtocol.Verify(proof.proofZeroOneCommitments[i], proof.commitments[i], privacy.VALUE)
		if !res {
			fmt.Printf("verify fail at %v\n", i)
			return false
		}
	}

	return true
}


func TestPKMaxValue(){
	pk := new(PKMaxValueProtocol)

	res := true

	for res == true {
		var value int64 = 123

		// convert int64 to []byte
		vBytes := make([]byte, binary.MaxVarintLen64)
		n := binary.PutVarint(vBytes, value)
		vBytes = vBytes[:n]

		proof, err := pk.Prove(vBytes)

		pk.SetProof(*proof)

		if err != nil{
			fmt.Println(err)
		}
		fmt.Printf("Proof.commitments: %v\n", proof.commitments)
		fmt.Printf("Proof.proofZeroOne: %v\n", proof.proofZeroOneCommitments)

		res = pk.Verify(*proof)

		fmt.Println(res)
	}
}
