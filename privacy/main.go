package main

import (
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

func runProve() {
	outApk := client.SpendingAddress{1}
	ekey := client.TransmissionKey{2}
	outNote1 := &client.Note{Value: 1000, Apk: outApk}
	outNote2 := &client.Note{Value: 2000, Apk: outApk}
	outputs := []*client.JSOutput{
		&client.JSOutput{EncKey: ekey, OutputNote: outNote1},
		&client.JSOutput{EncKey: ekey, OutputNote: outNote2}}

	const merkleTreeDepth = 29
	hash1 := [32]byte{3}
	mhash1 := [merkleTreeDepth]*client.MerkleHash{}
	for i := 0; i < merkleTreeDepth; i++ {
		mh := client.MerkleHash{}
		mh = make([]byte, len(hash1))
		copy(mh[:], hash1[:])
		mhash1[i] = &mh
	}
	index1 := [merkleTreeDepth]bool{true}
	mpath1 := client.MerklePath{AuthPath: mhash1[:], Index: index1[:]}
	hash2 := [32]byte{4}
	mhash2 := [merkleTreeDepth]*client.MerkleHash{}
	for i := 0; i < merkleTreeDepth; i++ {
		mh := client.MerkleHash{}
		mh = make([]byte, len(hash2))
		copy(mh[:], hash2[:])
		mhash2[i] = &mh
	}
	index2 := [merkleTreeDepth]bool{true}
	mpath2 := client.MerklePath{AuthPath: mhash2[:], Index: index2[:]}

	skey := client.SpendingKey{5}
	inpApk := client.SpendingAddress{6}
	rho1 := [32]byte{7}
	rho2 := [32]byte{8}
	r1 := [32]byte{11}
	r2 := [32]byte{12}
	inpNote1 := client.Note{Value: 400, Apk: inpApk, Rho: rho1[:], R: r1[:]} // Value, Apk, Rho and R should be gotten from note's memo
	inpNote2 := client.Note{Value: 2600, Apk: inpApk, Rho: rho2[:], R: r2[:]}
	input1 := client.JSInput{WitnessPath: &mpath1, Key: &skey, InputNote: &inpNote1}
	input2 := client.JSInput{WitnessPath: &mpath2, Key: &skey, InputNote: &inpNote2}
	inputs := []*client.JSInput{&input1, &input2}

	pubKey := [32]byte{9}
	rt := [32]byte{10}
	client.Prove(inputs, outputs, pubKey[:], rt[:])
}

func runVerify() {
	gA := [33]byte{0, 7}
	gAPrime := [33]byte{0, 8}
	gB := [65]byte{0, 9}
	gBPrime := [33]byte{0, 10}
	gC := [33]byte{0, 11}
	gCPrime := [33]byte{0, 12}
	gH := [33]byte{0, 13}
	gK := [33]byte{0, 14}

	proof := &zksnark.PHGRProof{
		G_A: gA[:], G_APrime: gAPrime[:],
		G_B: gB[:], G_BPrime: gBPrime[:],
		G_C: gC[:], G_CPrime: gCPrime[:],
		G_H: gH[:], G_K: gK[:]}

	nf1 := [32]byte{1}
	nf2 := [32]byte{2}
	var nf [][]byte
	nf = append(nf, nf1[:])
	nf = append(nf, nf2[:])

	cm1 := [32]byte{3}
	cm2 := [32]byte{4}
	var cm [][]byte
	cm = append(cm, cm1[:])
	cm = append(cm, cm2[:])
	rt := [32]byte{5}
	hSig := [32]byte{6}
	client.Verify(proof, &nf, &cm, rt[:], hSig[:])
}

func main() {
	// runProve()
	runVerify()
}
