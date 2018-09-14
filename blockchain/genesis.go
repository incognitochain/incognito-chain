package blockchain

import (
	"encoding/hex"
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/privacy/proto/zksnark"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/wallet"
)

type GenesisBlockGenerator struct {
}

func (self GenesisBlockGenerator) CalcMerkleRoot(txns []transaction.Transaction) common.Hash {
	if len(txns) == 0 {
		return common.Hash{}
	}

	utilTxns := make([]transaction.Transaction, 0, len(txns))
	for _, tx := range txns {
		utilTxns = append(utilTxns, tx)
	}
	merkles := Merkle{}.BuildMerkleTreeStore(utilTxns)
	return *merkles[len(merkles)-1]
}

func createGenesisInputNote(spendingKey *client.SpendingKey, idx uint) *client.Note {
	addr := client.GenSpendingAddress(*spendingKey)
	rho := [32]byte{byte(idx)}
	r := [32]byte{byte(idx)}
	note := &client.Note{
		Value: 0,
		Apk:   addr,
		Rho:   rho[:],
		R:     r[:],
	}
	return note
}

func createGenesisJSInput(idx uint) *client.JSInput {
	spendingKey := &client.SpendingKey{} // SpendingKey for input of genesis transaction is 0x0
	input := new(client.JSInput)
	input.InputNote = createGenesisInputNote(spendingKey, idx)
	input.Key = spendingKey
	input.WitnessPath = (&client.MerklePath{}).CreateDummyPath()
	return input
}

/**
Use to get hardcode for genesis block
*/
func (self GenesisBlockGenerator) createGenesisTx(coinReward uint64) (*transaction.Tx, error) {
	// Create deterministic inputs (note, receiver's address and rho)
	var inputs []*client.JSInput
	inputs = append(inputs, createGenesisJSInput(0))
	inputs = append(inputs, createGenesisJSInput(1))

	// Create new notes: first one is a coinbase UTXO, second one has 0 value
	key, err := wallet.Base58CheckDeserialize(GENESIS_BLOCK_PAYMENT_ADDR)
	if err != nil {
		panic(err)
	}
	outNote := &client.Note{Value: coinReward, Apk: key.KeyPair.PublicKey.Apk}
	placeHolderOutputNote := &client.Note{Value: 0, Apk: key.KeyPair.PublicKey.Apk}

	// Create deterministic outputs
	outputs := []*client.JSOutput{
		&client.JSOutput{EncKey: key.KeyPair.PublicKey.Pkenc, OutputNote: outNote},
		&client.JSOutput{EncKey: key.KeyPair.PublicKey.Pkenc, OutputNote: placeHolderOutputNote},
	}

	// Wrap ephemeral private key
	var ephemeralPrivKey client.EphemeralPrivKey
	copy(ephemeralPrivKey[:], GENESIS_BLOCK_EPHEMERAL_PRIVKEY[:])

	// Since input notes of genesis tx have 0 value, rt can be anything
	rt := make([]byte, 32)
	tx, err := transaction.GenerateProofForGenesisTx(
		inputs,
		outputs,
		rt,
		coinReward,
		GENESIS_BLOCK_SEED[:],
		GENESIS_BLOCK_PHI[:],
		GENESIS_BLOCK_OUTPUT_R,
		ephemeralPrivKey,
	)
	return tx, err
}

func (self GenesisBlockGenerator) getGenesisTx() (*transaction.Tx, error) {
	gA, _ := hex.DecodeString(GENESIS_BLOCK_G_A)
	gAPrime, _ := hex.DecodeString(GENESIS_BLOCK_G_APrime)
	gB, _ := hex.DecodeString(GENESIS_BLOCK_G_B)
	gBPrime, _ := hex.DecodeString(GENESIS_BLOCK_G_BPrime)
	gC, _ := hex.DecodeString(GENESIS_BLOCK_G_C)
	gCPrime, _ := hex.DecodeString(GENESIS_BLOCK_G_CPrime)
	gK, _ := hex.DecodeString(GENESIS_BLOCK_G_K)
	gH, _ := hex.DecodeString(GENESIS_BLOCK_G_H)
	proof := &zksnark.PHGRProof{
		G_A:      gA,
		G_APrime: gAPrime,
		G_B:      gB,
		G_BPrime: gBPrime,
		G_C:      gC,
		G_CPrime: gCPrime,
		G_K:      gK,
		G_H:      gH,
	}

	nf1, err := hex.DecodeString(GENESIS_BLOCK_NULLIFIERS[0])
	if err != nil {
		panic(err)
	}
	nf2, err := hex.DecodeString(GENESIS_BLOCK_NULLIFIERS[1])
	if err != nil {
		panic(err)
	}
	nullfiers := [][]byte{nf1, nf2}

	cm1, err := hex.DecodeString(GENESIS_BLOCK_COMMITMENTS[0])
	if err != nil {
		panic(err)
	}
	cm2, err := hex.DecodeString(GENESIS_BLOCK_COMMITMENTS[1])
	if err != nil {
		panic(err)
	}
	commitments := [][]byte{cm1, cm2}

	encData1, err := hex.DecodeString(GENESIS_BLOCK_ENCRYPTED_DATA[0])
	if err != nil {
		panic(err)
	}
	encData2, err := hex.DecodeString(GENESIS_BLOCK_ENCRYPTED_DATA[1])
	if err != nil {
		panic(err)
	}
	encryptedData := [][]byte{encData1, encData2}
	ephemeralPubKey, err := hex.DecodeString(GENESIS_BLOCK_EPHEMERAL_PUBKEY)
	if err != nil {
		panic(err)
	}

	desc := []*transaction.JoinSplitDesc{&transaction.JoinSplitDesc{
		Anchor:          GENESIS_BLOCK_ANCHOR[:],
		Nullifiers:      nullfiers,
		Commitments:     commitments,
		Proof:           proof,
		EncryptedData:   encryptedData,
		EphemeralPubKey: ephemeralPubKey,
		HSigSeed:        GENESIS_BLOCK_SEED[:],
		Type:            common.TxOutCoinType,
		Reward:          GENESIS_BLOCK_REWARD,
	}}

	jsPubKey, err := hex.DecodeString(GENESIS_BLOCK_JSPUBKEY)
	if err != nil {
		panic(err)
	}
	tx := &transaction.Tx{
		Version:  transaction.TxVersion,
		Type:     common.TxNormalType,
		Descs:    desc,
		JSPubKey: jsPubKey,
		JSSig:    nil,
	}
	return tx, nil
}

func (self GenesisBlockGenerator) CreateGenesisBlock(
	time time.Time,
	nonce int,
	difficulty uint32,
	version int,
	genesisReward uint64,
) *Block {
	genesisBlock := Block{}
	// update default genesis block
	genesisBlock.Header.Timestamp = time
	//genesisBlock.Header.PrevBlockHash = (&common.Hash{}).String()
	genesisBlock.Header.Nonce = nonce
	genesisBlock.Header.Difficulty = difficulty
	genesisBlock.Header.Version = version

	tx, err := self.getGenesisTx()
	// tx, err := self.createGenesisTx(genesisReward)

	if err != nil {
		panic(err)
	}

	genesisBlock.Header.MerkleRoot = self.CalcMerkleRoot(genesisBlock.Transactions)
	genesisBlock.Transactions = append(genesisBlock.Transactions, tx)
	return &genesisBlock
}
