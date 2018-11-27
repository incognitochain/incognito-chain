package zkp
//
//import (
//	"encoding/binary"
//	"fmt"
//	"math/big"
//
//	"github.com/ninjadotorg/constant/privacy-protocol"
//)
//
//const (
//	//VMAX is maximum value what user can spend in one transaction
//	VMAX = 18446744073709551615 // 2^64 - 1
//)
//
//// PKMaxValue is a protocol for Zero-knowledge Proof of Knowledge of max value is 2^64-1
//// include Witness: commitedValue, r []byte
//type PKMaxValueProtocol struct {
//	Witness PKMaxValueWitness
//	Proof   PKMaxValueProof
//}
//
//// PKOneOfManyProof contains Proof's value
//type PKMaxValueProof struct {
//	commitments             []privacy.EllipticPoint
//	proofZeroOneCommitments []*PKComZeroOneProof
//	PKComZeroOneProtocol    []*PKComZeroOneProtocol
//}
//
//type PKMaxValueWitness struct {
//	value []byte
//}
//
////SetWitness sets Witness
//func (pro *PKMaxValueProtocol) SetWitness(witness PKMaxValueWitness) {
//	pro.Witness = witness
//}
//
//func (pro *PKMaxValueProtocol) Prove() (*PKMaxValueProof, error) {
//	proof := new(PKMaxValueProof)
//
//	threshold := [4]int{8, 16, 32, 64}
//
//	//max := big.NewInt(0)
//	//max.Exp(big.NewInt(2), big.NewInt(64), nil)
//	//max.Sub(max, big.NewInt(1))
//	//
//	//fmt.Printf("max: %v\n", max)
//
//	//convert value from byte array to int64
//	vInt, n := binary.Varint(pro.Witness.value)
//	if uint64(vInt) > VMAX {
//		return nil, fmt.Errorf("value must be less than 2^64-1")
//	}
//	fmt.Printf("vInt is: %v, n is: %v\n", vInt, n)
//
//	vIntTmp := vInt
//	//vBinary represent value in binary
//	var vBinary []byte
//
//	// find index of bit-length threshold in which i lies
//	// nBitsValue: bit-length of value
//	nBitsValue := 0
//	// indexThreshold: index of bit-length threshold in which i lies
//	indexThreshold := 0
//	for vIntTmp > 0 {
//		vBinary = append(vBinary, byte(vIntTmp%2))
//		nBitsValue++
//		if nBitsValue > threshold[indexThreshold] {
//			indexThreshold++
//		}
//		vIntTmp = vIntTmp / 2
//	}
//
//	// Set Witness and commit
//	privacy.PedCom.InitCommitment()
//	witnesses := make([][][]byte, threshold[indexThreshold])
//	proof.commitments = make([][]byte, threshold[indexThreshold])
//	proof.proofZeroOneCommitments = make([]*PKComZeroOneProof, threshold[indexThreshold])
//	proof.PKComZeroOneProtocol = make([]*PKComZeroOneProtocol, threshold[indexThreshold])
//
//	for i := 0; i < nBitsValue; i++ {
//		witnesses[i] = [][]byte{big.NewInt(int64(vBinary[i])).Bytes(), privacy.RandBytes(32)}
//		proof.commitments[i] = make([]byte, 34)
//		fmt.Printf("witness[%v][0] : %v\n", i, witnesses[i][0])
//		fmt.Printf("witness[%v][1] : %v\n", i, witnesses[i][1])
//		proof.commitments[i] = privacy.PedCom.CommitAtIndex(witnesses[i][0], witnesses[i][1], privacy.VALUE)
//
//		var witness PKComZeroOneWitness
//		witness.commitment = proof.commitments[i]
//		witness.commitedValue = witnesses[i][0]
//		witness.rand = witnesses[i][1]
//		witness.index = privacy.VALUE
//		proof.PKComZeroOneProtocol[i] = new(PKComZeroOneProtocol)
//		proof.PKComZeroOneProtocol[i].SetWitness(witness)
//
//		proof.proofZeroOneCommitments[i] = new(PKComZeroOneProof)
//		proof.proofZeroOneCommitments[i], _ = proof.PKComZeroOneProtocol[i].Prove()
//		fmt.Printf("Proof %v: %+v\n", i, proof.proofZeroOneCommitments[i])
//	}
//
//	for j := nBitsValue; j < threshold[indexThreshold]; j++ {
//		witnesses[j] = [][]byte{big.NewInt(0).Bytes(), privacy.RandBytes(32)}
//		proof.commitments[j] = make([]byte, 34)
//		fmt.Printf("witness[%v][0] : %v\n", j, witnesses[j][0])
//		fmt.Printf("witness[%v][1] : %v\n", j, witnesses[j][1])
//		proof.commitments[j] = privacy.PedCom.CommitAtIndex(witnesses[j][0], witnesses[j][1], privacy.VALUE)
//
//		var witness PKComZeroOneWitness
//		witness.commitment = proof.commitments[j]
//		witness.commitedValue = witnesses[j][0]
//		witness.rand = witnesses[j][1]
//		witness.index = privacy.VALUE
//		proof.PKComZeroOneProtocol[j] = new(PKComZeroOneProtocol)
//		proof.PKComZeroOneProtocol[j].SetWitness(witness)
//
//		proof.proofZeroOneCommitments[j] = new(PKComZeroOneProof)
//		proof.proofZeroOneCommitments[j], _ = proof.PKComZeroOneProtocol[j].Prove()
//		fmt.Printf("Proof %v: %+v\n", j, proof.proofZeroOneCommitments[j])
//	}
//
//	return proof, nil
//}
//
//func (pro *PKMaxValueProtocol) SetProof(proof PKMaxValueProof) {
//	pro.Proof = proof
//	for i := 0; i < len(pro.Proof.commitments); i++ {
//		pro.Proof.PKComZeroOneProtocol[i].SetProof(*pro.Proof.proofZeroOneCommitments[i])
//	}
//}
//
//func (pro *PKMaxValueProtocol) Verify() bool {
//	if len(pro.Proof.commitments) > 64 || len(pro.Proof.proofZeroOneCommitments) > 64 {
//		return false
//	}
//
//	if len(pro.Proof.commitments) != len(pro.Proof.proofZeroOneCommitments) {
//		return false
//	}
//
//	for i := 0; i < len(pro.Proof.commitments); i++ {
//		//pro.Proof.PKComZeroOneProtocol[i].SetProof(*pro.Proof.proofZeroOneCommitments[i])
//		res := pro.Proof.PKComZeroOneProtocol[i].Verify()
//		if !res {
//			fmt.Printf("verify fail at %v\n", i)
//			return false
//		}
//	}
//	return true
//}
//
//func TestPKMaxValue() {
//	pk := new(PKMaxValueProtocol)
//
//	res := true
//
//	for res == true {
//		var value int64 = 123
//
//		// convert int64 to []byte
//		vBytes := make([]byte, binary.MaxVarintLen64)
//		n := binary.PutVarint(vBytes, value)
//		vBytes = vBytes[:n]
//		fmt.Printf("VBytes: %v\n", vBytes)
//
//		var witness PKMaxValueWitness
//		witness.value = vBytes
//		pk.SetWitness(witness)
//
//		proof, err := pk.Prove()
//		fmt.Printf("Proof: %+v\n", proof)
//
//		pk.SetProof(*proof)
//
//		if err != nil {
//			fmt.Println(err)
//		}
//		//fmt.Printf("Proof.commitments: %v\n", proof.commitments)
//		//fmt.Printf("Proof.proofZeroOne: %v\n", proof.proofZeroOneCommitments)
//
//		res = pk.Verify()
//
//		fmt.Println(res)
//	}
//}
