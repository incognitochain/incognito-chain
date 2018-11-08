package main

import (
	"fmt"
	// "crypto/rand"
	// "encoding/json"
	// "math/big"
	// "unsafe"


	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/privacy/proto/zksnark"
	// "github.com/ninjadotorg/constant/transaction"
)

type ProveResult struct {
	Proof  *zksnark.PHGRProof
	Nf, Cm [][]byte
	Reward uint64
}

func runProve() (*ProveResult, error) {
	// ask := client.RandSpendingKey()
	ask := client.SpendingKey{127, 9, 42, 195, 53, 40, 231, 23, 127, 206, 167, 170, 20, 82, 217, 40, 248, 110, 181, 16, 253, 131, 117, 145, 0, 30, 57, 18, 84, 57, 189, 8}
	outApk := client.GenSpendingAddress(ask)
	skenc := client.GenReceivingKey(ask)
	ekey := client.GenTransmissionKey(skenc)
	outNote1 := &client.Note{Value: 1, Apk: outApk}
	outNote2 := &client.Note{Value: 0, Apk: outApk}
	outputs := []*client.JSOutput{
		&client.JSOutput{EncKey: ekey, OutputNote: outNote1},
		&client.JSOutput{EncKey: ekey, OutputNote: outNote2},
	}

	const merkleTreeDepth = 29
	hash1 := [32]byte{3}
	mhash1 := [merkleTreeDepth]client.MerkleHash{}
	for i := 0; i < merkleTreeDepth; i++ {
		mh := client.MerkleHash{}
		mh = make([]byte, len(hash1))
		copy(mh[:], hash1[:])
		mhash1[i] = mh
	}
	index1 := [merkleTreeDepth]bool{true}
	mpath1 := client.MerklePath{AuthPath: mhash1[:], Index: index1[:]}
	hash2 := [32]byte{4}
	mhash2 := [merkleTreeDepth]client.MerkleHash{}
	for i := 0; i < merkleTreeDepth; i++ {
		mh := client.MerkleHash{}
		mh = make([]byte, len(hash2))
		copy(mh[:], hash2[:])
		mhash2[i] = mh
	}
	index2 := [merkleTreeDepth]bool{true}
	mpath2 := client.MerklePath{AuthPath: mhash2[:], Index: index2[:]}

	// inpApk := client.SpendingAddress{6}
	inpApk := outApk
	rho1 := [32]byte{7}
	rho2 := [32]byte{8}
	r1 := [32]byte{11}
	r2 := [32]byte{12}
	inpNote1 := client.Note{Value: 0, Apk: inpApk, Rho: rho1[:], R: r1[:]} // Value, Apk, Rho and R should be gotten from note's memo
	inpNote2 := client.Note{Value: 0, Apk: inpApk, Rho: rho2[:], R: r2[:]}
	input1 := client.JSInput{WitnessPath: &mpath1, Key: &ask, InputNote: &inpNote1}
	input2 := client.JSInput{WitnessPath: &mpath2, Key: &ask, InputNote: &inpNote2}
	inputs := []*client.JSInput{&input1, &input2}

	pubKey := [32]byte{9}
	rt := [32]byte{10}
	var reward uint64 = 1
	proof, _, _, _, err := client.Prove(inputs, outputs, pubKey[:], rt[:], reward, nil, nil, nil)
	// var proof *zksnark.PHGRProof
	// var err error
	// fmt.Println(pubKey, rt)
	if err != nil {
		return nil, err
	}

	var nf, cm [][]byte
	for _, input := range inputs {
		nf = append(nf, input.InputNote.Nf)
	}
	for _, output := range outputs {
		cm = append(cm, output.OutputNote.Cm)
	}
	result := &ProveResult{
		Proof:  proof,
		Nf:     nf,
		Cm:     cm,
		Reward: reward,
	}
	return result, nil
}

func runVerify(proveResult *ProveResult) {
	var nf [][]byte
	var cm [][]byte
	var proof *zksnark.PHGRProof
	var reward uint64

	if proveResult == nil {
		gA := [33]byte{0, 7}
		gAPrime := [33]byte{0, 8}
		gB := [65]byte{0, 9}
		gBPrime := [33]byte{0, 10}
		gC := [33]byte{0, 11}
		gCPrime := [33]byte{0, 12}
		gH := [33]byte{0, 13}
		gK := [33]byte{0, 14}

		proof = &zksnark.PHGRProof{
			G_A: gA[:], G_APrime: gAPrime[:],
			G_B: gB[:], G_BPrime: gBPrime[:],
			G_C: gC[:], G_CPrime: gCPrime[:],
			G_H: gH[:], G_K: gK[:],
		}

		nf1 := [32]byte{1}
		nf2 := [32]byte{2}
		nf = append(nf, nf1[:])
		nf = append(nf, nf2[:])
		cm1 := [32]byte{3}
		cm2 := [32]byte{4}
		cm = append(cm, cm1[:])
		cm = append(cm, cm2[:])
	} else {
		proof = proveResult.Proof
		nf = proveResult.Nf
		cm = proveResult.Cm
		reward = proveResult.Reward
	}

	rt := [32]byte{5}
	hSig := [32]byte{6}
	valid, err := client.Verify(proof, &nf, &cm, rt[:], hSig[:], reward)
	if err != nil {
		panic(err)
	}
	fmt.Println("Verify valid:", valid)
}

func runProveThenVerify() {
	if proveResult, err := runProve(); err == nil {
		runVerify(proveResult)
	} else {
		panic(err)
	}
}

func main() {
	// runProve()
	// runVerify()
	//runProveThenVerify()
	// client.TestEncrypt()

	// Test transaction signing
	// privKey, _ := client.GenerateKey(rand.Reader)
	// tx := new(transaction.Tx)
	// tx.Version = 1
	// tx.Type = "Normal"
	// tx.LockTime = 123
	// tx.Fee = 1234
	// tx.SetTxID(tx.Hash())

	// fmt.Printf("Hash tx: %s\n", tx.GetTxID())
	// signed_tx, err := transaction.SignTx(tx, privKey)
	// if err != nil {
	// 	fmt.Printf("Error: %s", err)
	// }

	// pub := transaction.PubKeyToByteArray(&privKey.PaymentAddress)
	// signed_tx.JSPubKey = pub
	// fmt.Printf("Pub key: %v\n",signed_tx.JSPubKey)
	// fmt.Printf("Size of pub key: %d\n", len(signed_tx.JSPubKey))

	// fmt.Printf("Signature: %v\n", signed_tx.JSSig)
	// fmt.Printf("Size of signature: %d\n", len(signed_tx.JSSig))

	// fmt.Printf("Hash tx: %s\n", signed_tx.GetTxID())
	// res, _ := transaction.VerifySign(signed_tx)

	// fmt.Println(res)
}
