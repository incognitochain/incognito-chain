package main

import (
	"github.com/thaibaoautonomous/btcd/privacy/client"
)

func main() {
	outApk := client.SpendingAddress{1}
	ekey := client.TransmissionKey{2}
	outNote1 := client.Note{Value: 1000, Apk: outApk}
	outNote2 := client.Note{Value: 2000, Apk: outApk}
	outputs := []*client.JSOutput{
		&client.JSOutput{EncKey: ekey, OutputNote: outNote1},
		&client.JSOutput{EncKey: ekey, OutputNote: outNote2}}

	path1 := [32]byte{3}
	path2 := [32]byte{4}
	skey := client.SpendingKey{5}
	inpApk := client.SpendingAddress{6}
	rho1 := [32]byte{7}
	rho2 := [32]byte{8}
	inpNote1 := client.Note{Value: 400, Apk: inpApk, Rho: rho1[:]}
	inpNote2 := client.Note{Value: 2600, Apk: inpApk, Rho: rho2[:]}
	input1 := client.JSInput{WitnessPath: path1[:], Key: skey, InputNote: inpNote1}
	input2 := client.JSInput{WitnessPath: path2[:], Key: skey, InputNote: inpNote2}
	inputs := []*client.JSInput{&input1, &input2}

	pubKey := [32]byte{9}
	rt := [32]byte{10}
	client.Prove(inputs, outputs, pubKey[:], rt[:])
	// const address = "localhost:50052"
	// conn, err := grpc.Dial(address, grpc.WithInsecure())
	// if err != nil {
	// 	log.Fatalf("fail to connect %v", err)
	// }
	// defer conn.Close()

	// c := zksnark.NewZksnarkClient(conn)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()

	// hsig := [32]byte{65, 66}
	// phi := [32]byte{67, 68}
	// rt := [32]byte{69, 70}
	// var proveRequest = &zksnark.ProveRequest{Hsig: hsig[:], Phi: phi[:], Rt: rt[:], OutNotes: outNotes, Inputs: inputs}
	// fmt.Printf("%v\n", proveRequest)
	// r, err := c.Prove(ctx, proveRequest)
	// if err != nil {
	// 	log.Fatalf("fail to prove: %v", err)
	// }
	// log.Printf("response: %v", r.Dummy)
}
