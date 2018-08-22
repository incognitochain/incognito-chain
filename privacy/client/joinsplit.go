package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/thaibaoautonomous/btcd/privacy/proto/zksnark"
	"google.golang.org/grpc"
)

type JSInput struct {
	WitnessPath MerklePath
	Key         SpendingKey
	InputNote   Note
}

type JSOutput struct {
	EncKey     TransmissionKey
	OutputNote Note
}

// Prove calls libsnark's Prove and return the proof
// inputs: WitnessPath and Key must be set; InputeNote's Value, Apk and Rho must also be set before calling this function
// outputs: EncKey, OutputNote's Apk and Value must be set before calling this function
func Prove(inputs []*JSInput, outputs []*JSOutput, pubKey []byte, rt []byte) {
	// TODO: think how to implement vpub
	// TODO: check for inputs (witness root & cm)

	if len(inputs) != 2 || len(outputs) != 2 {
		panic("Number of inputs/outputs to Prove is incorrect")
	}

	// Check balance between input and output
	var totalInput, totalOutput uint64
	for _, input := range inputs {
		totalInput += input.InputNote.Value
	}
	for _, output := range outputs {
		totalOutput += output.OutputNote.Value
	}
	if totalInput != totalOutput {
		panic("Input and output value are not equal")
	}

	// Generate hSig
	// Compute nullifier for old notes
	for _, input := range inputs {
		var rho [32]byte
		copy(rho[:], input.InputNote.Rho)
		input.InputNote.Nf = GetNullifier(input.Key, rho)
	}
	seed := RandBits(256)
	hSig := HSigCRH(seed, inputs[0].InputNote.Nf, inputs[1].InputNote.Nf, pubKey)

	// Generate rho and r for new notes
	const phiLen = 252
	phi := RandBits(phiLen)
	for i, output := range outputs {
		rho := PRF_rho(uint64(i), phi, hSig)
		copy(output.OutputNote.Rho[:], rho)
		copy(output.OutputNote.R[:], RandBits(256))

		// Compute cm for new notes to check for Note commitment integrity
		output.OutputNote.Cm = GetCommitment(output.OutputNote)
	}

	// TODO: encrypt note's data
	// TODO: malleability

	// Calling libsnark's prove
	const address = "localhost:50052"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to connect %v", err)
	}
	defer conn.Close()

	c := zksnark.NewZksnarkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var outNotes []*Note
	for _, output := range outputs {
		outNotes = append(outNotes, &output.OutputNote)
	}

	zkNotes := Notes2ZksnarkNotes(outNotes)
	zkInputs := JSInputs2ZkInputs(inputs)
	var proveRequest = &zksnark.ProveRequest{Hsig: hSig, Phi: phi, Rt: rt, OutNotes: zkNotes, Inputs: zkInputs}
	fmt.Printf("%v\n", proveRequest)
	r, err := c.Prove(ctx, proveRequest)
	if err != nil {
		log.Fatalf("fail to prove: %v", err)
	}
	log.Printf("response: %v", r.Dummy)
}

func Note2ZksnarkNote(note *Note) *zksnark.Note {
	var zknote = zksnark.Note{Value: note.Value}
	copy(zknote.Cm, note.Cm)
	copy(zknote.R, note.R)
	copy(zknote.Nf, note.Nf) // Might be 0 for output notes
	copy(zknote.Rho, note.Rho)
	copy(zknote.Apk, note.Apk[:])
	return &zknote
}

func Notes2ZksnarkNotes(notes []*Note) []*zksnark.Note {
	var zkNotes []*zksnark.Note
	for _, note := range notes {
		zkNotes = append(zkNotes, Note2ZksnarkNote(note))
	}
	return zkNotes
}

func JSInputs2ZkInputs(inputs []*JSInput) []*zksnark.JSInput {
	var zkInputs []*zksnark.JSInput
	for _, input := range inputs {
		var zkinput zksnark.JSInput
		for _, hash := range input.WitnessPath.AuthPath {
			zkinput.WitnessPath.AuthPath = append(zkinput.WitnessPath.AuthPath, &zksnark.MerkleHash{Hash: hash})
		}
		copy(zkinput.WitnessPath.Index, input.WitnessPath.Index)
		copy(zkinput.SpendingKey, input.Key[:])

		zkinput.Note = Note2ZksnarkNote(&input.InputNote)

		zkInputs = append(zkInputs, &zkinput)
	}
	return zkInputs
}
