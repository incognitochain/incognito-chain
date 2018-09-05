package client

import (
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

// TODO(@0xbunyip): add randomSeed, MACs and epk
type JoinSplitDesc struct {
	Anchor        []byte
	Nullifiers    [][]byte
	Commitments   [][]byte
	Proof         *zksnark.PHGRProof
	EncryptedData []byte
}

type Tx struct {
	Version  int    `json:"Version"`
	Type     string `json:"Type"` // NORMAL / ACTION_PARAMS
	LockTime int    `json:"LockTime"`
	NumDescs int
	Desc     []*JoinSplitDesc
	JSPubKey []byte // 32 bytes
	JSSig    []byte // 64 bytes
}

func collectUnspentNotes(ask *SpendingKey, valueWanted uint64) ([]*Note, error) {
	return make([]*Note, 2), nil
}

// Send creates transaction with appropriate proof for a private payment
// value: total value of the coins to transfer
// rt: root of the commitment merkle tree at current block (the latest block of the node creating this tx)
func Send(senderKey *SpendingKey, receiverAddr *PaymentAddress, value uint64, rt []byte) (*Tx, error) {
	inputNotes, err := collectUnspentNotes(senderKey, value)
	if err != nil {
		return nil, err
	}

	if len(inputNotes) == 0 {
		return nil, errors.New("Cannot find notes with sufficient fund")
	}

	// Create Proof for the joinsplit op
	inputs := make([]*JSInput, 2)
	inputs[0].InputNote = inputNotes[0]
	inputs[0].Key = senderKey
	inputs[0].WitnessPath = new(MerklePath) // TODO: get path

	if len(inputNotes) <= 1 {
		// Create dummy note
	} else if len(inputNotes) <= 2 {
		inputs[1].InputNote = inputNotes[1]
		inputs[1].Key = senderKey
		inputs[1].WitnessPath = new(MerklePath)
	} else {
		return nil, errors.New("More than 2 notes for input is not supported")
	}

	// Left side value
	var sumInputValue uint64
	for _, input := range inputs {
		sumInputValue += input.InputNote.Value
	}
	if sumInputValue < value {
		panic("Input value less than output value")
	}

	senderFullKey := senderKey.GenFullKey()

	// Create new notes: first one send `value` to receiverAddr, second one sends `change` back to sender
	outNote := &Note{Value: value, Apk: receiverAddr.Apk}
	changeNote := &Note{Value: sumInputValue - value, Apk: senderFullKey.Addr.Apk}

	outputs := make([]*JSOutput, 2)
	outputs[0].EncKey = receiverAddr.Pkenc
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = senderFullKey.Addr.Pkenc
	outputs[1].OutputNote = changeNote

	// Shuffle output notes randomly (if necessary)

	// Generate JoinSplit key pair and sign the tx to prevent tx malleability
	keyBytes := []byte{} // TODO(0xbunyip): randomize seed?
	keyPair, err := (&cashec.KeyPair{}).GenerateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	proof, err := Prove(inputs, outputs, keyPair.PublicKey, rt)
	if err != nil {
		return nil, err
	}
	nullifiers := [][]byte{inputs[0].InputNote.Nf, inputs[1].InputNote.Nf}
	commitments := [][]byte{outputs[0].OutputNote.Cm, outputs[1].OutputNote.Cm}

	// TODO: add encrypted data
	desc := []*JoinSplitDesc{&JoinSplitDesc{
		Proof:       proof,
		Anchor:      rt,
		Nullifiers:  nullifiers,
		Commitments: commitments,
	}}

	tx := &Tx{
		Version:  1,
		Type:     common.TxNormalType,
		NumDescs: 2,
		Desc:     desc,
		JSPubKey: keyPair.PublicKey,
		JSSig:    nil,
	}

	// Sign tx
	dataToBeSigned, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	jsSig, err := keyPair.Sign(dataToBeSigned)
	if err != nil {
		return nil, err
	}
	tx.JSSig = jsSig

	return tx, nil
}
