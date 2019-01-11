package blockchain

import (
	"log"
	"time"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
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
		SalaryPerTx: salaryPerTx,
		BasicSalary: basicSalary,
		SellingBonds: &params.SellingBonds{
			BondName:       "Bonds have 2 blocks maturity",
			BondSymbol:     "BND2",
			TotalIssue:     10000,
			BondsToSell:    10000,
			BondPrice:      100,
			Maturity:       2,
			BuyBackPrice:   120,
			StartSellingAt: 1,
			SellingWithin:  500,
		},

		RefundInfo: &params.RefundInfo{},
		OracleNetwork: &params.OracleNetwork{
			OraclePubKeys:         [][]byte{},
			WrongTimesAllowed:     2,
			Quorum:                1,
			AcceptableErrorMargin: 200, // 2 USD
			UpdateFrequency:       10,
		},
	}
	// Decentralize central bank params
	loanParams := []params.LoanParams{
		params.LoanParams{
			InterestRate:     100,   // 1%
			Maturity:         1000,  // 1 month in blocks
			LiquidationStart: 15000, // 150%
		},
	}
	genesisBlock.Header.DCBConstitution.DCBParams = params.DCBParams{
		LoanParams:               loanParams,
		MinLoanResponseRequire:   1,
		MinCMBApprovalRequire:    1,
		LateWithdrawResponseFine: 1000,
		SaleDCBTokensByUSDData: &params.SaleDCBTokensByUSDData{
			Amount:   0,
			EndBlock: 0,
		},
	}

	// TODO(@0xjackalope): fill correct values
	genesisBlock.Header.DCBGovernor = DCBGovernor{
		GovernorInfo: GovernorInfo{
			boardIndex:   0,
			StartedBlock: 1,
			EndBlock:     1000, // = startedblock of decent governor
			BoardPubKeys: [][]byte{
				[]byte{3, 85, 237, 178, 30, 58, 190, 219, 126, 31, 9, 93, 40, 217, 109, 177, 70, 41, 64, 157, 2, 133, 2, 138, 23, 108, 228, 152, 234, 35, 101, 192, 173},
				//				[]byte{3, 116, 125, 158, 22, 126, 79, 50, 46, 119, 52, 133, 6, 246, 156, 94, 138, 244, 107, 147, 25, 78, 231, 105, 162, 185, 245, 152, 196, 116, 86, 15, 30},
			},
			StartAmountToken: 0, //Sum of DCB token stack to all member of this board
		},
	}

	// Commercial bank params
	genesisBlock.Header.CBParams = CBParams{}
	copy(genesisBlock.Header.Committee, preSelectValidators)

	genesisBlock.Header.Height = 1
	genesisBlock.Header.SalaryFund = icoParams.InitFundSalary
	genesisBlock.Header.Oracle = &params.Oracle{
		Bonds:    map[string]uint64{},
		DCBToken: 1,
		Constant: 1,
	}

	// Get Ico payment address
	log.Printf("Ico payment address:", icoParams.InitialPaymentAddress)
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

	// calculate merkle root tx for genesis block
	genesisBlock.Header.MerkleRoot = self.CalcMerkleRoot(genesisBlock.Transactions)

	// genesisBlock.Header.MerkleRootCommitments = self.calcCommitmentMerkleRoot(tx)
	// fmt.Printf("Anchor: %x\n", genesisBlock.Header.MerkleRootCommitments[:])

	return &genesisBlock
}
