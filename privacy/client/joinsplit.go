package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
	"google.golang.org/grpc"
)

type JSInput struct {
	WitnessPath *MerklePath
	Key         *SpendingKey
	InputNote   *Note
}

type JSOutput struct {
	EncKey     TransmissionKey
	OutputNote *Note
}

// Prove calls libsnark's Prove and return the proof
// inputs: WitnessPath and Key must be set; InputeNote's Value, Apk, R and Rho must also be set before calling this function
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
		input.InputNote.Nf = GetNullifier(*input.Key, rho)

		// Compute cm for old notes to check for merkle path
		input.InputNote.Cm = GetCommitment(input.InputNote)
	}

	// TODO:(0xbunyip): remove debug here
	// seed := RandBits(256)
	// seed := []byte{155, 31, 215, 9, 16, 242, 239, 233, 201, 109, 141, 58, 24, 239, 210, 117, 155, 17, 23, 188, 70, 125, 245, 85, 154, 42, 212, 0, 164, 221, 80, 94}
	// hSig := HSigCRH(seed, inputs[0].InputNote.Nf, inputs[1].InputNote.Nf, pubKey)
	hSig := []byte{155, 31, 215, 9, 16, 242, 239, 233, 201, 109, 141, 58, 24, 239, 210, 117, 155, 17, 23, 188, 70, 125, 245, 85, 154, 42, 212, 0, 164, 221, 80, 94}

	// Generate rho and r for new notes
	const phiLen = 252
	// TODO:(0xbunyip): remove debug here
	// phi := RandBits(phiLen)
	phi := []byte{80, 163, 129, 14, 224, 14, 22, 199, 9, 222, 152, 68, 97, 249, 132, 138, 69, 64, 195, 13, 46, 200, 79, 248, 16, 161, 73, 187, 200, 122, 235, 6}

	for i, output := range outputs {
		rho := PRF_rho(uint64(i), phi, hSig)
		output.OutputNote.Rho = make([]byte, len(rho))
		output.OutputNote.R = make([]byte, 32)
		copy(output.OutputNote.Rho, rho)
		copy(output.OutputNote.R, RandBits(256))

		// Compute cm for new notes to check for Note commitment integrity
		output.OutputNote.Cm = GetCommitment(output.OutputNote)
	}

	fmt.Printf("hsig: %x\n", hSig)
	fmt.Printf("phi: %x\n", phi)
	fmt.Printf("rt: %x\n", rt)
	fmt.Printf("rho0: %x\n", outputs[0].OutputNote.Rho)
	fmt.Printf("rho1: %x\n", outputs[1].OutputNote.Rho)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var outNotes []*Note
	for _, output := range outputs {
		outNotes = append(outNotes, output.OutputNote)
	}

	zkNotes := Notes2ZksnarkNotes(outNotes)
	zkInputs := JSInputs2ZkInputs(inputs)
	fmt.Printf("zkInputs[0].Note.R: %x\n", zkInputs[0].Note.R)
	fmt.Printf("zkInputs[0].WitnessPath.AuthPath[0]: %x\n", zkInputs[0].WitnessPath.AuthPath[0].Hash)
	var proveRequest = &zksnark.ProveRequest{Hsig: hSig, Phi: phi, Rt: rt, OutNotes: zkNotes, Inputs: zkInputs}
	// fmt.Printf("proveRequest: %v\n", proveRequest)
	fmt.Printf("key: %x\n", proveRequest.Inputs[0].SpendingKey)
	r, err := c.Prove(ctx, proveRequest)
	if err != nil {
		log.Fatalf("fail to prove: %v", err)
	}
	log.Printf("response: %v", r.Proof)
}

func Verify(proof *zksnark.PHGRProof, nf, cm *[][]byte, rt, hSig []byte) bool {
	// Calling libsnark's verify
	const address = "localhost:50052"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to connect %v", err)
	}
	defer conn.Close()

	c := zksnark.NewZksnarkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var verifyRequest = &zksnark.VerifyRequest{Proof: proof, Hsig: hSig, Rt: rt, Nullifiers: *nf, Commits: *cm}
	fmt.Printf("verifyRequest: %v\n", verifyRequest)
	r, err := c.Verify(ctx, verifyRequest)
	if err != nil {
		log.Fatalf("fail to verify: %v", err)
	}
	log.Printf("response: %v", r.Valid)

	return true
}

func Note2ZksnarkNote(note *Note) *zksnark.Note {
	var zknote = zksnark.Note{
		Value: note.Value,
		Cm:    make([]byte, len(note.Cm)),
		R:     make([]byte, len(note.R)),
		Nf:    make([]byte, len(note.Nf)),
		Rho:   make([]byte, len(note.Rho)),
		Apk:   make([]byte, len(note.Apk)),
	}
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
		zkinput := zksnark.JSInput{SpendingKey: make([]byte, 32)}
		zkinput.WitnessPath = &zksnark.MerklePath{Index: make([]bool, len(input.WitnessPath.Index))}
		copy(zkinput.WitnessPath.Index, input.WitnessPath.Index)
		for _, hash := range input.WitnessPath.AuthPath {
			zkinput.WitnessPath.AuthPath = append(zkinput.WitnessPath.AuthPath, &zksnark.MerkleHash{Hash: *hash})
		}
		copy(zkinput.SpendingKey, input.Key[:])
		// fmt.Printf("zkinput.SpendingKey: %x %x\n", zkinput.SpendingKey, input.Key)

		zkinput.Note = Note2ZksnarkNote(input.InputNote)

		zkInputs = append(zkInputs, &zkinput)
	}
	return zkInputs
}
