package client

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

// JoinSplitDesc stores the UTXO of a transaction
// TODO(@0xbunyip): add randomSeed, MACs and epk
type JoinSplitDesc struct {
	Anchor        []byte             `json:"Anchor"`
	Nullifiers    [][]byte           `json:"Nullifiers"`
	Commitments   [][]byte           `json:"Commitments"`
	Proof         *zksnark.PHGRProof `json:"Proof"`
	EncryptedData []byte             `json:"EncryptedData"`
	Type          string             `json:"Type"`
	Reward        uint64             `json:"Reward"` // For coinbase tx
}

// Tx represents a coin-transfer-transaction stored in a block
type Tx struct {
	Version  int    `json:"Version"`
	Type     string `json:"Type"` // NORMAL / ACTION_PARAMS
	LockTime int    `json:"LockTime"`
	Fee      uint64 `json:"Fee"`

	Descs    []*JoinSplitDesc `json:"Descs"`
	JSPubKey []byte           `json:"JSPubKey"` // 32 bytes
	JSSig    []byte           `json:"JSSig"`    // 64 bytes
}

func (desc *JoinSplitDesc) toString() string {
	s := string(desc.Anchor)
	for _, nf := range desc.Nullifiers {
		s += string(nf)
	}
	for _, cm := range desc.Commitments {
		s += string(cm)
	}
	s += desc.Proof.String()
	s += string(desc.EncryptedData)
	return s
}

// Hash returns the hash of all fields of the transaction
func (tx *Tx) Hash() *common.Hash {
	record := strconv.Itoa(tx.Version)
	record += tx.Type
	record += strconv.Itoa(tx.LockTime)
	record += strconv.Itoa(len(tx.Descs))
	for _, desc := range tx.Descs {
		record += desc.toString()
	}
	record += string(tx.JSPubKey)
	record += string(tx.JSSig)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// ValidateTransaction returns true if transaction is valid:
// - All data fields are well formed
// - JSDescriptions are valid (zk-snark proof satisfied)
// - Signature matches the signing public key
// Note: This method doesn't check for double spending
func (tx *Tx) ValidateTransaction() bool {
	for _, desc := range tx.Descs {
		if desc.Reward != 0 {
			return false // Coinbase tx shouldn't be broadcasted across the network
		}
	}

	// TODO(@0xbunyip): implement
	return true
}

// GetType returns the type of the transaction
func (tx *Tx) GetType() string {
	return tx.Type
}

func collectUnspentNotes(ask *SpendingKey, valueWanted uint64) ([]*Note, error) {
	return make([]*Note, 2), nil
}

// CreateTx creates transaction with appropriate proof for a private payment
// value: total value of the coins to transfer
// rt: root of the commitment merkle tree at current block (the latest block of the node creating this tx)
func CreateTx(senderKey *SpendingKey, receiverAddr *PaymentAddress, value uint64, rt []byte) (*Tx, error) {
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

	// Generate proof and sign tx
	var reward uint64 // Zero reward for non-coinbase transaction
	tx, err := generateProofAndSign(inputs, outputs, rt, reward)
	return tx, err
}

func createDummyNote(randomKey *SpendingKey) *Note {
	// TODO(@0xbunyip): create dummy note according to 4.7.1
	return nil
}

func createRandomJSInput() *JSInput {
	randomKey := RandSpendingKey()
	input := new(JSInput)
	input.InputNote = createDummyNote(&randomKey)
	input.Key = &randomKey
	input.WitnessPath = new(MerklePath) // TODO(@0xbunyip): create dummy path if necessary
	return input
}

func generateProofAndSign(inputs []*JSInput, outputs []*JSOutput, rt []byte, reward uint64) (*Tx, error) {
	// Generate JoinSplit key pair and sign the tx to prevent tx malleability
	keyBytes := []byte{} // TODO(0xbunyip): randomize seed?
	keyPair, err := (&cashec.KeyPair{}).GenerateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	proof, err := Prove(inputs, outputs, keyPair.PublicKey, rt, reward)
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
		Reward:      reward,
	}}

	tx := &Tx{
		Version:  1,
		Type:     common.TxNormalType,
		Descs:    desc,
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

func createCoinbaseTx(
	params *blockchain.Params,
	receiverAddr *PaymentAddress,
	rewardMap map[string]uint64,
	rt []byte,
) (*Tx, error) {
	// Create Proof for the joinsplit op
	inputs := make([]*JSInput, 2)
	inputs[0] = createRandomJSInput()
	inputs[1] = createRandomJSInput()

	// Get reward
	// TODO(@0xbunyip): implement bonds reward
	var reward uint64
	for rewardType, rewardValue := range rewardMap {
		if rewardValue <= 0 {
			continue
		}
		if rewardType == "coins" {
			reward = rewardValue
		}
	}

	// Create new notes: first one is coinbase UTXO, second one has 0 value
	outNote := &Note{Value: reward, Apk: receiverAddr.Apk}
	placeHolderOutputNote := &Note{Value: 0, Apk: receiverAddr.Apk}

	outputs := make([]*JSOutput, 2)
	outputs[0].EncKey = receiverAddr.Pkenc
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = receiverAddr.Pkenc
	outputs[1].OutputNote = placeHolderOutputNote

	// Shuffle output notes randomly (if necessary)

	// Generate proof and sign tx
	tx, err := generateProofAndSign(inputs, outputs, rt, reward)
	return tx, err
}
