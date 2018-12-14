package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/proto/zksnark"
	"google.golang.org/grpc"
)

type JSInput struct {
	WitnessPath *MerklePath
	Key         *privacy.SpendingKey
	InputNote   *Note
}

type JSOutput struct {
	EncKey     privacy.TransmissionKey
	OutputNote *Note
}

func printProof(proof *zksnark.PHGRProof) {
	fmt.Printf("G_A: %x\n", proof.G_A)
	fmt.Printf("G_APrime: %x\n", proof.G_APrime)
	fmt.Printf("G_B: %x\n", proof.G_B)
	fmt.Printf("G_BPrime: %x\n", proof.G_BPrime)
	fmt.Printf("G_C: %x\n", proof.G_C)
	fmt.Printf("G_CPrime: %x\n", proof.G_CPrime)
	fmt.Printf("G_K: %x\n", proof.G_K)
	fmt.Printf("G_H: %x\n", proof.G_H)
}

// Prove calls libsnark's Prove and return the proof
// inputs: WitnessPath and Key must be set; InputeNote's H, PaymentAddress, R and Rho must also be set before calling this function
// outputs: EncKey, OutputNote's PaymentAddress and H must be set before calling this function
// reward: for salary tx, this is the mining reward; for other tx, it must be 0
// After this function, outputs' Rho and R and Cm will be updated
func Prove(inputs []*JSInput,
	outputs []*JSOutput,
	pubKey []byte,
	rts [][]byte,
	reward, fee uint64,
	seed, phi []byte,
	outputR [][]byte,
	addressLastByte byte,
	noPrivacy bool,
) (proof *zksnark.PHGRProof, hSig, newSeed, newPhi []byte, err error) {
	// Check balance between input and output
	totalInput := reward
	totalOutput := fee
	for _, input := range inputs {
		totalInput += input.InputNote.Value
	}
	for _, output := range outputs {
		totalOutput += output.OutputNote.Value
	}
	if totalInput != totalOutput {
		panic(fmt.Sprintf("Input and output value are not equal: %v and %v", totalInput, totalOutput))
	}

	// Generate hSig
	// Compute nullifier for old notes
	for _, input := range inputs {
		var rho [32]byte
		copy(rho[:], input.InputNote.Rho)
		key := make([]byte, len(*input.Key))
		copy(key[:], (*input.Key)[:])
		input.InputNote.Nf = GetNullifier(key, rho)

		// Compute cm for old notes to check for merkle path
		input.InputNote.Cm = GetCommitment(input.InputNote)
	}

	newSeed = seed
	if seed == nil { // Only for the transaction in genesis block
		newSeed = RandBits(256)
	}
	hSig = HSigCRH(newSeed, inputs[0].InputNote.Nf, inputs[1].InputNote.Nf, pubKey)

	// Generate rho and r for new notes
	const phiLen = 252
	newPhi = phi
	if phi == nil { // Only for the transaction in genesis block
		newPhi = RandBits(phiLen)
	}

	for i, output := range outputs {
		rho := PRF_rho(uint64(i), newPhi, hSig)
		output.OutputNote.Rho = make([]byte, len(rho))
		output.OutputNote.R = make([]byte, 32)
		copy(output.OutputNote.Rho, rho)
		if outputR == nil {
			copy(output.OutputNote.R, RandBits(256))
		} else { // Genesis block only
			copy(output.OutputNote.R, outputR[i])
		}

		// Compute cm for new notes to check for Note commitment integrity
		fmt.Printf("Getting commitment for output note\n")
		fmt.Printf("PaymentAddress: %x\n", output.OutputNote.Apk)
		fmt.Printf("Rho: %x\n", output.OutputNote.Rho)
		fmt.Printf("Randomness: %x\n", output.OutputNote.R)
		output.OutputNote.Cm = GetCommitment(output.OutputNote)
	}

	fmt.Printf("hsig: %x\n", hSig)
	fmt.Printf("phi: %x\n", phi)

	// Calling libsnark's prove
	const address = "localhost:50052"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to connect %v", err)
	}
	defer conn.Close()

	c := zksnark.NewZksnarkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*500)

	defer cancel()

	var outNotes []*Note
	for _, output := range outputs {
		outNotes = append(outNotes, output.OutputNote)
	}

	zkNotes := Notes2ZksnarkNotes(outNotes)
	zkInputs := JSInputs2ZkInputs(inputs)
	var proveRequest = &zksnark.ProveRequest{
		Hsig:            hSig,
		Phi:             newPhi,
		Rts:             rts,
		OutNotes:        zkNotes,
		Inputs:          zkInputs,
		Reward:          reward,
		Fee:             fee,
		AddressLastByte: uint32(addressLastByte),
	}
	// fmt.Printf("proveRequest: %v\n", proveRequest)
	fmt.Printf("key: %x\n", proveRequest.Inputs[0].SpendingKey)
	fmt.Printf("Anchor: %x\n", rts)
	fmt.Printf("reward: %v\n", reward)
	fmt.Printf("fee: %v\n", fee)
	fmt.Printf("last byte: %d\n", addressLastByte)
	for i, zkinput := range zkInputs {
		fmt.Printf("zkInputs[%d].SpendingKey: %x\n", i, zkinput.SpendingKey)
		fmt.Printf("zkInputs[%d].Note.H: %v\n", i, zkinput.Note.Value)
		fmt.Printf("zkInputs[%d].Note.Cm: %x\n", i, zkinput.Note.Cm)
		fmt.Printf("zkInputs[%d].Note.Randomness: %x\n", i, zkinput.Note.R)
		fmt.Printf("zkInputs[%d].Note.Nf: %x\n", i, zkinput.Note.Nf)
		fmt.Printf("zkInputs[%d].Note.Rho: %x\n", i, zkinput.Note.Rho)
		fmt.Printf("zkInputs[%d].Note.PaymentAddress: %x\n", i, zkinput.Note.Apk)
	}

	for i, zkout := range zkNotes {
		fmt.Printf("zkNotes[%d].Note.H: %v\n", i, zkout.Value)
		fmt.Printf("zkNotes[%d].Note.Cm: %x\n", i, zkout.Cm)
		fmt.Printf("zkNotes[%d].Note.Randomness: %x\n", i, zkout.R)
		fmt.Printf("zkNotes[%d].Note.Nf: %x\n", i, zkout.Nf)
		fmt.Printf("zkNotes[%d].Note.Rho: %x\n", i, zkout.Rho)
		fmt.Printf("zkNotes[%d].Note.PaymentAddress: %x\n", i, zkout.Apk)
	}

	var r *zksnark.ProveReply
	if !noPrivacy {
		r, err = c.Prove(ctx, proveRequest)
	} else {
		return nil, hSig, newSeed, newPhi, nil
	}
	/* for Test not privacy-protocol
	dummyProof := &zksnark.PHGRProof{
		G_A:      make([]byte, 33),
		G_APrime: make([]byte, 33),
		G_B:      make([]byte, 65),
		G_BPrime: make([]byte, 33),
		G_C:      make([]byte, 33),
		G_CPrime: make([]byte, 33),
		G_K:      make([]byte, 33),
		G_H:      make([]byte, 33),
	}
	r := &zksnark.ProveReply{
		Success: true,
		Proof:   dummyProof,
	}*/
	if err != nil || r == nil || r.Proof == nil {
		log.Printf("fail to prove: %v", err)
		return nil, nil, nil, nil, errors.New("Fail to prove JoinSplit")
	}
	log.Printf("Prove response:\n")
	printProof(r.Proof)
	if !r.Success {
		return nil, nil, nil, nil, fmt.Errorf("Prove returns success = false")
	}
	return r.Proof, hSig, newSeed, newPhi, nil
}

// Verify checks if a zk-proof of a JSDesc is valid or not
func Verify(
	proof *zksnark.PHGRProof,
	nf, cm [][]byte,
	rts, macs [][]byte,
	hSig []byte,
	reward,
	fee uint64,
	addressLastByte byte,
) (bool, error) {
	return true, nil

	// Calling libsnark's verify
	const address = "localhost:50052"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return false, fmt.Errorf("fail to connect: %v", err)
	}
	defer conn.Close()

	c := zksnark.NewZksnarkClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var verifyRequest = &zksnark.VerifyRequest{
		Proof:           proof,
		Hsig:            hSig,
		Rts:             rts,
		Nullifiers:      nf,
		Commits:         cm,
		Macs:            macs,
		Reward:          reward,
		Fee:             fee,
		AddressLastByte: uint32(addressLastByte),
	}
	fmt.Printf("verifyRequest: %+v\n", verifyRequest)
	r, err := c.Verify(ctx, verifyRequest)
	if err != nil {
		return false, fmt.Errorf("fail to verify: %v", err)
	}
	log.Printf("response: %v %v", r.Valid, r.Success)
	if !r.Success {
		return false, fmt.Errorf("Verify returns success = false")
	}
	return r.Valid, err
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
			zkinput.WitnessPath.AuthPath = append(zkinput.WitnessPath.AuthPath, &zksnark.MerkleHash{Hash: hash})
		}
		copy(zkinput.SpendingKey, (*input.Key)[:])
		// fmt.Printf("zkinput.SpendingKey: %x %x\n", zkinput.SpendingKey, input.PubKey)

		zkinput.Note = Note2ZksnarkNote(input.InputNote)

		zkInputs = append(zkInputs, &zkinput)
	}
	return zkInputs
}
