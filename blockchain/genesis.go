package blockchain

import (
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
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
	note := &client.Note{
		Value: 0,
		Apk:   addr,
		Rho:   rho[:],
		Nf:    client.GetNullifier(*spendingKey, rho),
	}
	return note
}

func createGenesisJSInput(idx uint) *client.JSInput {
	spendingKey := &client.SpendingKey{} // SpendingKey for input of genesis transaction is 0x0
	input := new(client.JSInput)
	input.InputNote = createGenesisInputNote(spendingKey, idx)
	input.Key = spendingKey
	input.WitnessPath = new(client.MerklePath)
	return input
}

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

	// Since input notes of genesis tx have 0 value, rt can be anything
	rt := make([]byte, 32)
	// TODO: move seed and phi of genesis block to constants.go
	genesisBlockSeed := [32]byte{1}
	genesisBlockPhi := [32]byte{2}
	tx, err := transaction.GenerateProofAndSignForGenesisTx(inputs, outputs, rt, coinReward, genesisBlockSeed[:], genesisBlockPhi[:])
	return tx, err
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

	tx, err := self.createGenesisTx(genesisReward)

	if err != nil {
		panic(err)
	}

	genesisBlock.Header.MerkleRoot = self.CalcMerkleRoot(genesisBlock.Transactions)
	genesisBlock.Transactions = append(genesisBlock.Transactions, tx)
	return &genesisBlock
}
