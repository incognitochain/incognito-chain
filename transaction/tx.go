package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"crypto/rand"
	// "crypto/sha256"
	// "math/big"

	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
)

// Tx represents a coin-transfer-transaction stored in a block
type Tx struct {
	Version  int    `json:"Version"`
	Type     string `json:"Type"` // n
	LockTime int    `json:"LockTime"`
	Fee      uint64 `json:"Fee"`

	Descs    []*JoinSplitDesc `json:"Descs"`
	JSPubKey []byte           `json:"JSPubKey,omitempty"` // 32 bytes
	JSSig    []byte           `json:"JSSig,omitempty"`    // 64 bytes

	txId *common.Hash
}

func (tx *Tx) SetTxId(txId *common.Hash) {
	tx.txId = txId
}

func (tx *Tx) GetTxId() (*common.Hash) {
	return tx.txId
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

func collectUnspentNotes(ask *client.SpendingKey, valueWanted uint64) ([]*client.Note, error) {
	return make([]*client.Note, 2), nil
}

// CreateTx creates transaction with appropriate proof for a private payment
// value: total value of the coins to transfer
// rt: root of the commitment merkle tree at current block (the latest block of the node creating this tx)
func CreateTx(
	senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rt *common.Hash,
	usableTx []*Tx,
	nullifiers [][]byte,
	commitments [][]byte,
) (*Tx, error) {
	receiverAddr := paymentInfo[0].PaymentAddress
	value := paymentInfo[0].Amount
	inputNotes, err := collectUnspentNotes(senderKey, value)
	if err != nil {
		return nil, err
	}

	if len(inputNotes) == 0 {
		return nil, errors.New("Cannot find notes with sufficient fund")
	}

	// Create Proof for the joinsplit op
	var inputsToBuildWitness []*client.JSInput
	inputs := make([]*client.JSInput, 2)
	inputs[0].InputNote = inputNotes[0]
	inputs[0].Key = senderKey
	inputs[0].WitnessPath = new(client.MerklePath)
	inputsToBuildWitness = append(inputsToBuildWitness, inputs[0])

	if len(inputNotes) <= 1 {
		inputs[1].InputNote = createDummyNote(senderKey)
		inputs[1].Key = senderKey
		inputs[1].WitnessPath = (&client.MerklePath{}).CreateDummyPath() // No need to build commitment merkle path for dummy note
	} else if len(inputNotes) <= 2 {
		inputs[1].InputNote = inputNotes[1]
		inputs[1].Key = senderKey
		inputs[1].WitnessPath = new(client.MerklePath)
		inputsToBuildWitness = append(inputsToBuildWitness, inputs[1])
	} else {
		return nil, errors.New("More than 2 notes for input is not supported")
	}

	// Get commitments of input notes and build witness path
	// TODO: calculate cm and check if it's in commitments list
	client.BuildWitnessPath(inputsToBuildWitness, commitments)

	// Left side value
	var sumInputValue uint64
	for _, input := range inputs {
		sumInputValue += input.InputNote.Value
	}
	if sumInputValue < value {
		panic("Input value less than output value")
	}

	senderFullKey := cashec.KeySet{}
	senderFullKey.ImportFromPrivateKeyByte(senderKey[:])

	// Create new notes: first one send `value` to receiverAddr, second one sends `change` back to sender
	outNote := &client.Note{Value: value, Apk: receiverAddr.Apk}
	changeNote := &client.Note{Value: sumInputValue - value, Apk: senderFullKey.PublicKey.Apk}

	outputs := make([]*client.JSOutput, 2)
	outputs[0].EncKey = receiverAddr.Pkenc
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = senderFullKey.PublicKey.Pkenc
	outputs[1].OutputNote = changeNote

	// Shuffle output notes randomly (if necessary)

	// Generate proof and sign tx
	var reward uint64 // Zero reward for non-coinbase transaction
	tx, err := GenerateProofAndSign(inputs, outputs, rt[:], reward)
	return tx, err
}

func createDummyNote(randomKey *client.SpendingKey) *client.Note {
	addr := client.GenSpendingAddress(*randomKey)
	var rho [32]byte
	copy(rho[:], client.RandBits(32*8))
	note := &client.Note{
		Value: 0,
		Apk:   addr,
		Rho:   rho[:],
		Nf:    client.GetNullifier(*randomKey, rho),
	}
	return note
}

// CreateRandomJSInput creates a dummy input with 0 value note that is sended to a random address
func CreateRandomJSInput() *client.JSInput {
	randomKey := client.RandSpendingKey()
	input := new(client.JSInput)
	input.InputNote = createDummyNote(&randomKey)
	input.Key = &randomKey
	input.WitnessPath = new(client.MerklePath) // TODO(@0xbunyip): create dummy path if necessary
	return input
}

func SignTx(tx *Tx, privKey *client.PrivateKey) (*Tx, error) {
	//Check input transaction
	if tx.JSSig != nil || tx.JSPubKey != nil {
		return nil, errors.New("Input transaction must be an unsigned one!")
	}

	// Hash transaction
	hash := tx.GetTxId()
	data := make([]byte, common.HashSize)
	copy(data, hash[:])
	// dataToBeSigned, err := json.Marshal(tx)
	// if err != nil {
	// 	return nil, err
	// }
	// hash := sha256.Sum256([]byte(dataToBeSigned))

	// Sign
	ecdsaSignature := *new(client.EcdsaSignature)
	ecdsaSignature.R, ecdsaSignature.S, _ = client.Sign(rand.Reader, privKey, data[:])
	// if err != nil {
	// 	return nil, err
	// }

	tx.JSSig, _ = json.Marshal(ecdsaSignature)
	// if err != nil {
	// 	return nil, err
	// }

	return tx, nil
}

func VerifySign(tx *Tx) (bool, error) {
	//Check input transaction
	if tx.JSSig == nil || tx.JSPubKey == nil {
		return false, errors.New("Input transaction must be an signed one!")
	}
	// UnParse Public key
	pubKey := new(client.PublicKey)
	err := json.Unmarshal(tx.JSPubKey, pubKey)
	if err != nil {
		return false, err
	}
	// fmt.Printf("Pub key: %+v\n", *pubKey)

	// UnParse ECDSA signature
	ecdsaSignature := new(client.EcdsaSignature)
	err = json.Unmarshal(tx.JSSig, ecdsaSignature)
	if err != nil {
		return false, err
	}
	// fmt.Printf("JSsig : %+v\n", jsSig)

	// Hash origin transaction
	hash := tx.GetTxId()
	data := make([]byte, common.HashSize)
	copy(data, hash[:])

	valid := client.VerifySign(pubKey, data[:], ecdsaSignature.R, ecdsaSignature.S)
	return valid, nil
}

func generateTx(
	inputs []*client.JSInput,
	outputs []*client.JSOutput,
	proof *zksnark.PHGRProof,
	rt []byte,
	reward uint64,
	hSig, seed, jsPubKey []byte,
	ephemeralPrivKey *client.EphemeralPrivKey,
) (*Tx, error) {
	nullifiers := [][]byte{inputs[0].InputNote.Nf, inputs[1].InputNote.Nf}
	commitments := [][]byte{outputs[0].OutputNote.Cm, outputs[1].OutputNote.Cm}
	notes := [2]client.Note{*outputs[0].OutputNote, *outputs[1].OutputNote}
	keys := [2]client.TransmissionKey{outputs[0].EncKey, outputs[1].EncKey}

	ephemeralPubKey := new(client.EphemeralPubKey)
	if ephemeralPrivKey == nil {
		ephemeralPrivKey = new(client.EphemeralPrivKey)
		*ephemeralPubKey, *ephemeralPrivKey = client.GenEphemeralKey()
	} else { // Genesis block only
		ephemeralPrivKey.GenPubKey()
		*ephemeralPubKey = ephemeralPrivKey.GenPubKey()
	}
	fmt.Printf("hSig: %x\n", hSig)
	fmt.Printf("jsPubKey: %x\n", jsPubKey)
	fmt.Printf("ephemeralPrivKey: %x\n", *ephemeralPrivKey)
	fmt.Printf("ephemeralPubKey: %x\n", *ephemeralPubKey)
	fmt.Printf("tranmissionKey[0]: %x\n", keys[0])
	fmt.Printf("tranmissionKey[1]: %x\n", keys[1])
	fmt.Printf("notes[0].Value: %v\n", notes[0].Value)
	fmt.Printf("notes[0].Rho: %x\n", notes[0].Rho)
	fmt.Printf("notes[0].R: %x\n", notes[0].R)
	fmt.Printf("notes[0].Memo: %v\n", notes[0].Memo)
	fmt.Printf("notes[1].Value: %v\n", notes[1].Value)
	fmt.Printf("notes[1].Rho: %x\n", notes[1].Rho)
	fmt.Printf("notes[1].R: %x\n", notes[1].R)
	fmt.Printf("notes[1].Memo: %v\n", notes[1].Memo)
	noteciphers := client.EncryptNote(notes, keys, *ephemeralPrivKey, *ephemeralPubKey, hSig)

	desc := []*JoinSplitDesc{&JoinSplitDesc{
		Anchor:          rt,
		Nullifiers:      nullifiers,
		Commitments:     commitments,
		Proof:           proof,
		EncryptedData:   noteciphers,
		EphemeralPubKey: ephemeralPubKey[:],
		HSigSeed:        seed,
		Type:            common.TxOutCoinType,
		Reward:          reward,
	}}

	fmt.Println("desc[0]:")
	fmt.Printf("Anchor: %x\n", desc[0].Anchor)
	fmt.Printf("Nullifiers: %x\n", desc[0].Nullifiers)
	fmt.Printf("Commitments: %x\n", desc[0].Commitments)
	fmt.Printf("Proof: %x\n", desc[0].Proof)
	fmt.Printf("EncryptedData: %x\n", desc[0].EncryptedData)
	fmt.Printf("EphemeralPubKey: %x\n", desc[0].EphemeralPubKey)
	fmt.Printf("HSigSeed: %x\n", desc[0].HSigSeed)
	fmt.Printf("Type: %x\n", desc[0].Type)
	fmt.Printf("Reward: %x\n", desc[0].Reward)

	// TODO(@0xbunyip): use Apk of PubKey temporarily, we should derive another scheme for signing tx later
	tx := &Tx{
		Version:  TxVersion,
		Type:     common.TxNormalType,
		Descs:    desc,
		JSPubKey: jsPubKey,
		JSSig:    nil,
	}
	return tx, nil
}

// GenerateProofAndSign creates zk-proof, build the transaction and sign it using a random generated key pair
func GenerateProofAndSign(inputs []*client.JSInput, outputs []*client.JSOutput, rt []byte, reward uint64) (*Tx, error) {
	// Generate JoinSplit key pair and sign the tx to prevent tx malleability
	keyBytes := []byte{} // TODO(0xbunyip): randomize seed?
	keyPair := (&cashec.KeySet{}).GenerateKey(keyBytes)

	var seed, phi *[]byte
	var outputR [][]byte
	proof, hSig, err := client.Prove(inputs, outputs, keyPair.PublicKey.Apk[:], rt, reward, seed, phi, outputR)
	if err != nil {
		return nil, err
	}

	var ephemeralPrivKey *client.EphemeralPrivKey

	//Generate signing key
	sigPrivKey, err := client.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	// Verification key
	sigPubKey, err := json.Marshal(sigPrivKey.PublicKey)
	if err != nil {
		return nil, err
	}

	tx, err := generateTx(inputs, outputs, proof, rt, reward, hSig, *seed, sigPubKey, ephemeralPrivKey)
	if err != nil {
		return nil, err
	}
	tx, err = SignTx(tx, sigPrivKey)
	if err != nil {
		return tx, err
	}
	//Calculate vmacs to prove this transaction is signed by this user

	//nullifiers := [][]byte{inputs[0].InputNote.Nf, inputs[1].InputNote.Nf}

	vmacs := make([][]byte, 2)
	for i, _ := range inputs {
		var ask []byte
		copy(ask[:], inputs[i].Key[:])
		vmacs[i] = client.PRF_pk(uint64(i), ask, hSig)
	}
	tx.Descs[0].Vmacs = vmacs

	return tx, nil
}

// GenerateProofForGenesisTx creates zk-proof and build the transaction (without signing) for genesis block
func GenerateProofForGenesisTx(
	inputs []*client.JSInput,
	outputs []*client.JSOutput,
	rt []byte,
	reward uint64,
	seed, phi []byte,
	outputR [][]byte,
	ephemeralPrivKey client.EphemeralPrivKey,
) (*Tx, error) {
	// Generate JoinSplit key pair and sign the tx to prevent tx malleability
	privateSignKey := [32]byte{1}
	keyPair := &cashec.KeySet{}
	keyPair.ImportFromPrivateKeyByte(privateSignKey[:])

	proof, hSig, err := client.Prove(inputs, outputs, keyPair.PublicKey.Apk[:], rt, reward, &seed, &phi, outputR)
	if err != nil {
		return nil, err
	}

	return generateTx(inputs, outputs, proof, rt, reward, hSig, seed, keyPair.PublicKey.Apk[:], &ephemeralPrivKey)
}
