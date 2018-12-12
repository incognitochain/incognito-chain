package blockchain

import (
	"log"
	"time"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

type GenesisBlockGenerator struct {
}

func (self GenesisBlockGenerator) CalcMerkleRoot(txns []metadata.Transaction) common.Hash {
	if len(txns) == 0 {
		return common.Hash{}
	}

	utilTxns := make([]metadata.Transaction, 0, len(txns))
	for _, tx := range txns {
		utilTxns = append(utilTxns, tx)
	}
	merkles := Merkle{}.BuildMerkleTreeStore(utilTxns)
	return *merkles[len(merkles)-1]
}

/*func createGenesisInputNote(spendingKey *privacy.SpendingKey, idx uint) *client.Note {
	addr := privacy.GeneratePaymentAddress(*spendingKey)
	rho := [32]byte{byte(idx)}
	r := [32]byte{byte(idx)}
	note := &client.Note{
		H: 0,
		Apk:   addr.Pk,
		Rho:   rho[:],
		R:     r[:],
	}
	return note
}*/

/*func createGenesisJSInput(idx uint) *client.JSInput {
	spendingKey := &privacy.SpendingKey{} // SpendingKey for input of genesis transaction is 0x0
	input := new(client.JSInput)
	input.InputNote = createGenesisInputNote(spendingKey, idx)
	input.PubKey = spendingKey
	input.WitnessPath = (&client.MerklePath{}).CreateDummyPath()
	return input
}*/

/*
Use to get hardcode for genesis block
*/

/*func (self GenesisBlockGenerator) createGenesisTx(initialCoin uint64, initialAddress string) (*transaction.TxNormal, error) {
	// Create deterministic inputs (note, receiver's address and rho)
	var inputs []*client.JSInput
	inputs = append(inputs, createGenesisJSInput(0))
	inputs = append(inputs, createGenesisJSInput(1))

	// Create new notes: first one is a salary UTXO, second one has 0 value
	key, err := wallet.Base58CheckDeserialize(initialAddress)
	if err != nil {
		return nil, err
	}
	outNote := &client.Note{H: initialCoin, Apk: key.KeySet.PaymentAddress.Pk}
	placeHolderOutputNote := &client.Note{H: 0, Apk: key.KeySet.PaymentAddress.Pk}

	fmt.Printf("EncKey: %x\n", key.KeySet.PaymentAddress.Tk)

	// Create deterministic outputs
	outputs := []*client.JSOutput{
		&client.JSOutput{EncKey: key.KeySet.PaymentAddress.Tk, OutputNote: outNote},
		&client.JSOutput{EncKey: key.KeySet.PaymentAddress.Tk, OutputNote: placeHolderOutputNote},
	}

	// Wrap ephemeral private key
	var ephemeralPrivKey client.EphemeralPrivKey
	copy(ephemeralPrivKey[:], GENESIS_BLOCK_EPHEMERAL_PRIVKEY[:])

	// Since input notes of genesis tx have 0 value, rt can be anything
	rts := [][]byte{make([]byte, 32), make([]byte, 32)}
	tx, err := transaction.GenerateProofForGenesisTx(
		inputs,
		outputs,
		rts,
		initialCoin,
		GENESIS_BLOCK_SEED[:],
		GENESIS_BLOCK_PHI[:],
		GENESIS_BLOCK_OUTPUT_R,
		ephemeralPrivKey,
		//common.AssetTypeCoin,
	)
	return tx, err
}*/

/*func (self GenesisBlockGenerator) getGenesisTx(genesisBlockReward uint64) (*transaction.TxNormal, error) {
	// Convert zk-proof from hex string to byte array
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

	// Convert nullifiers from hex string to byte array
	nf1, err := hex.DecodeString(GENESIS_BLOCK_NULLIFIERS[0])
	if err != nil {
		return nil, err
	}
	nf2, err := hex.DecodeString(GENESIS_BLOCK_NULLIFIERS[1])
	if err != nil {
		return nil, err
	}
	nullfiers := [][]byte{nf1, nf2}

	// Convert commitments from hex string to byte array
	cm1, err := hex.DecodeString(GENESIS_BLOCK_COMMITMENTS[0])
	if err != nil {
		return nil, err
	}
	cm2, err := hex.DecodeString(GENESIS_BLOCK_COMMITMENTS[1])
	if err != nil {
		return nil, err
	}
	commitments := [][]byte{cm1, cm2}

	// Convert encrypted data from hex string to byte array
	encData1, err := hex.DecodeString(GENESIS_BLOCK_ENCRYPTED_DATA[0])
	if err != nil {
		return nil, err
	}
	encData2, err := hex.DecodeString(GENESIS_BLOCK_ENCRYPTED_DATA[1])
	if err != nil {
		return nil, err
	}
	encryptedData := [][]byte{encData1, encData2}

	// Convert ephemeral public key from hex string to byte array
	ephemeralPubKey, err := hex.DecodeString(GENESIS_BLOCK_EPHEMERAL_PUBKEY)
	if err != nil {
		return nil, err
	}

	// Convert vmacs from hex string to byte array
	vmacs1, err := hex.DecodeString(GENESIS_BLOCK_VMACS[0])
	if err != nil {
		return nil, err
	}
	vmacs2, err := hex.DecodeString(GENESIS_BLOCK_VMACS[1])
	if err != nil {
		return nil, err
	}
	vmacs := [][]byte{vmacs1, vmacs2}

	anchors := [][]byte{make([]byte, 32), make([]byte, 32)}
	copy(anchors[0], GENESIS_BLOCK_ANCHORS[0][:])
	copy(anchors[1], GENESIS_BLOCK_ANCHORS[1][:])
	desc := []*transaction.JoinSplitDesc{&transaction.JoinSplitDesc{
		Anchor:          anchors,
		Nullifiers:      nullfiers,
		Commitments:     commitments,
		Proof:           proof,
		EncryptedData:   encryptedData,
		EphemeralPubKey: ephemeralPubKey,
		HSigSeed:        GENESIS_BLOCK_SEED[:],
		Reward:          genesisBlockReward,
		Vmacs:           vmacs,
	}}

	jsPubKey, err := hex.DecodeString(GENESIS_BLOCK_JSPUBKEY)
	if err != nil {
		return nil, err
	}

	//tempKeySet, _ := wallet.Base58CheckDeserialize(GENESIS_BLOCK_PAYMENT_ADDR)
	//lastByte := tempKeySet.KeySet.PaymentAddress.PaymentAddress[len(tempKeySet.KeySet.PaymentAddress.PaymentAddress)-1]

	tx := &transaction.TxNormal{
		Version:  transaction.TxVersion,
		Type:     common.TxNormalType,
		LockTime: 0,
		Fee:      0,
		Descs:    desc,
		JSPubKey: jsPubKey,
		JSSig:    nil,
		//AddressLastByte: lastByte,
	}
	return tx, nil
}*/

/*func (self GenesisBlockGenerator) calcCommitmentMerkleRoot(tx *transaction.Tx) common.Hash {
	tree := new(client.IncMerkleTree)
	for _, desc := range tx.Descs {
		for _, cm := range desc.Commitments {
			tree.AddNewNode(cm[:])
		}
	}

	rt := tree.GetRoot(common.IncMerkleTreeHeight)
	hash := common.Hash{}
	copy(hash[:], rt[:])
	return hash
}*/

/*func (self GenesisBlockGenerator) CreateGenesisBlock(
	time time.Time,
	nonce int,
	difficulty uint32,
	version int,
	genesisReward uint64,
) *Block {
	genesisBlock := Block{}
	// update default genesis block
	genesisBlock.Header.Timestamp = time.Unix()
	//genesisBlock.Header.PrevBlockHash = (&common.Hash{}).String()
	genesisBlock.Header.Nonce = nonce
	genesisBlock.Header.Difficulty = difficulty
	genesisBlock.Header.Version = version

	tx, err := self.getGenesisTx()
	//tx, err := self.createGenesisTx(genesisReward)

	if err != nil {
		Logger.log.Error(err)
		return nil
	}

	genesisBlock.Header.MerkleRootCommitments = self.calcCommitmentMerkleRoot(tx)
	fmt.Printf("Anchor: %x\n", genesisBlock.Header.MerkleRootCommitments[:])

	genesisBlock.Transactions = append(genesisBlock.Transactions, tx)
	genesisBlock.Header.MerkleRoot = self.CalcMerkleRoot(genesisBlock.Transactions)
	return &genesisBlock
}*/

// CreateSpecialTokenTx - create special token such as GOV, BANK, VOTE
func createSpecialTokenTx(
	tokenID common.Hash,
	tokenName string,
	tokenSymbol string,
	amount uint64,
	initialAddress privacy.PaymentAddress,
) transaction.TxCustomToken {
	log.Printf("Init token %s: %s \n", tokenSymbol, tokenID.String())
	paymentAddr := initialAddress
	vout := transaction.TxTokenVout{
		Value:          amount,
		PaymentAddress: paymentAddr,
	}
	vout.SetIndex(0)
	txTokenData := transaction.TxTokenData{
		PropertyID:     tokenID,
		PropertyName:   tokenName,
		PropertySymbol: tokenSymbol,
		Type:           transaction.CustomTokenInit,
		Amount:         amount,
		Vins:           []transaction.TxTokenVin{},
		Vouts:          []transaction.TxTokenVout{vout},
	}
	result := transaction.TxCustomToken{
		TxTokenData: txTokenData,
	}
	result.Type = common.TxCustomTokenType
	return result
}

func (self GenesisBlockGenerator) CreateGenesisBlockPoSParallel(
	version int,
	preSelectValidators []string,
	icoParams IcoParams,
	salaryPerTx uint64,
	basicSalary uint64,
) *Block {
	//init the loc
	loc, _ := time.LoadLocation("America/New_York")
	time := time.Date(2018, 8, 1, 0, 0, 0, 0, loc)
	genesisBlock := Block{
		Transactions: []metadata.Transaction{},
	}
	genesisBlock.Header = BlockHeader{}

	// update default genesis block
	genesisBlock.Header.Timestamp = time.Unix()
	genesisBlock.Header.Version = version
	genesisBlock.Header.Committee = make([]string, len(preSelectValidators))

	// Gov param
	genesisBlock.Header.GOVConstitution.GOVParams = params.GOVParams{
		SalaryPerTx:  salaryPerTx,
		BasicSalary:  basicSalary,
		SellingBonds: &params.SellingBonds{},
		RefundInfo:   &params.RefundInfo{},
	}
	// Decentralize central bank params
	loanParams := []params.LoanParams{
		params.LoanParams{
			InterestRate:     0,
			Maturity:         7776000, // 3 months in seconds
			LiquidationStart: 15000,   // 150%
		},
	}
	genesisBlock.Header.DCBConstitution.DCBParams = params.DCBParams{
		LoanParams: loanParams,
	}

	// Commercial bank params
	genesisBlock.Header.CBParams = CBParams{}
	copy(genesisBlock.Header.Committee, preSelectValidators)

	genesisBlock.Header.Height = 1
	genesisBlock.Header.SalaryFund = icoParams.InitFundSalary

	// Get Ico payment address
	key, err := wallet.Base58CheckDeserialize(icoParams.InitialPaymentAddress)
	if err != nil {
		panic(err)
	}
	// Create genesis token tx for DCB
	dcbTokenTx := createSpecialTokenTx( // DCB
		common.Hash(common.DCBTokenID),
		"Decentralized central bank token",
		"DCB",
		icoParams.InitialDCBToken,
		key.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&dcbTokenTx)

	// Create genesis token tx for GOV
	govTokenTx := createSpecialTokenTx(
		common.Hash(common.GOVTokenID),
		"Government token",
		"GOV",
		icoParams.InitialGOVToken,
		key.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&govTokenTx)

	// Create genesis token tx for CMB
	cmbTokenTx := createSpecialTokenTx(
		common.Hash(common.CMBTokenID),
		"Commercial bank token",
		"CMB",
		icoParams.InitialCMBToken,
		key.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&cmbTokenTx)

	// Create genesis token tx for BOND test
	bondTokenTx := createSpecialTokenTx(
		common.Hash([common.HashSize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		"BondTest",
		"BONTest",
		icoParams.InitialBondToken,
		key.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&bondTokenTx)

	// Create genesis vote token tx for DCB
	VoteDCBTokenTx := createSpecialTokenTx(
		common.Hash(common.VoteDCBTokenID),
		"Bond",
		"BON",
		icoParams.InitialVoteDCBToken,
		key.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&VoteDCBTokenTx)

	// Create genesis vote token tx for GOV
	VoteGOVTokenTx := createSpecialTokenTx(
		common.Hash(common.VoteGOVTokenID),
		"Bond",
		"BON",
		icoParams.InitialVoteGOVToken,
		key.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&VoteGOVTokenTx)

	// calculate merkle root tx for genesis block
	genesisBlock.Header.MerkleRoot = self.CalcMerkleRoot(genesisBlock.Transactions)

	// genesisBlock.Header.MerkleRootCommitments = self.calcCommitmentMerkleRoot(tx)
	// fmt.Printf("Anchor: %x\n", genesisBlock.Header.MerkleRootCommitments[:])

	return &genesisBlock
}
