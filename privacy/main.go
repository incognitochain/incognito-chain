package main

import (
	"fmt"

	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

func main() {
	zkinput := zksnark.JSInput{SpendingKey: make([]byte, 32)}
	var k [32]byte
	b := [32]byte{1, 2, 3}
	copy(zkinput.SpendingKey, b[:])
	fmt.Println(zkinput, k, b)

	outApk := client.SpendingAddress{1}
	ekey := client.TransmissionKey{2}
	outNote1 := client.Note{Value: 1000, Apk: outApk}
	outNote2 := client.Note{Value: 2000, Apk: outApk}
	outputs := []*client.JSOutput{
		&client.JSOutput{EncKey: ekey, OutputNote: outNote1},
		&client.JSOutput{EncKey: ekey, OutputNote: outNote2}}

	hash1 := [29]byte{3}
	mhash1 := [29]*client.MerkleHash{}
	for i := 0; i < 29; i++ {
		mh := client.MerkleHash{}
		copy(mh[:], hash1[:])
		mhash1[i] = &mh
	}
	index1 := [29]bool{true}
	mpath1 := client.MerklePath{AuthPath: mhash1[:], Index: index1[:]}
	hash2 := [32]byte{4}
	mhash2 := [29]*client.MerkleHash{}
	for i := 0; i < 29; i++ {
		mh := client.MerkleHash{}
		copy(mh[:], hash2[:])
		mhash2[i] = &mh
	}
	index2 := [32]bool{true}
	mpath2 := client.MerklePath{AuthPath: mhash2[:], Index: index2[:]}

	skey := client.SpendingKey{5}
	inpApk := client.SpendingAddress{6}
	rho1 := [32]byte{7}
	rho2 := [32]byte{8}
	inpNote1 := client.Note{Value: 400, Apk: inpApk, Rho: rho1[:]}
	inpNote2 := client.Note{Value: 2600, Apk: inpApk, Rho: rho2[:]}
	input1 := client.JSInput{WitnessPath: &mpath1, Key: &skey, InputNote: &inpNote1}
	input2 := client.JSInput{WitnessPath: &mpath2, Key: &skey, InputNote: &inpNote2}
	inputs := []*client.JSInput{&input1, &input2}

	pubKey := [32]byte{9}
	rt := [32]byte{10}
	client.Prove(inputs, outputs, pubKey[:], rt[:])
}
