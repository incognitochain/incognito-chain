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
	saleData := params.SaleData{
		SaleID:        make([]byte, 32),
		EndBlock:      1000,
		BuyingAsset:   common.BondTokenID[:],
		BuyingAmount:  uint64(1000),
		SellingAsset:  common.ConstantID[:],
		SellingAmount: uint64(2000),
	}
	genesisBlock.Header.DCBConstitution.DCBParams = params.DCBParams{
		ListSaleData:             []params.SaleData{saleData},
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
	boardPaymentAddress := []privacy.PaymentAddress{
		{
			Pk: []byte{3, 116, 183, 57, 32, 5, 157, 136, 217, 20, 89, 65, 18, 38, 23, 74, 7, 92, 219, 104, 208, 51, 28, 18, 72, 32, 190, 31, 120, 225, 206, 247, 71},
			Tk: []byte{3, 254, 61, 147, 129, 25, 38, 50, 162, 131, 221, 110, 0, 110, 91, 168, 163, 227, 34, 128, 246, 132, 168, 152, 225, 203, 180, 23, 155, 0, 117, 36, 48},
		},
		{
			Pk: []byte{3, 107, 44, 180, 170, 164, 107, 71, 126, 248, 38, 110, 212, 117, 79, 141, 188, 207, 244, 151, 226, 252, 47, 63, 69, 38, 11, 241, 199, 60, 85, 27, 74},
			Tk: []byte{3, 11, 201, 172, 23, 228, 134, 220, 28, 65, 222, 228, 156, 206, 142, 39, 23, 215, 237, 7, 61, 197, 246, 119, 251, 30, 105, 107, 131, 36, 156, 134, 76},
		},
		{
			Pk: []byte{3, 192, 159, 176, 226, 183, 190, 102, 43, 227, 172, 38, 53, 154, 235, 72, 106, 127, 1, 18, 213, 206, 25, 52, 72, 244, 29, 23, 130, 208, 138, 17, 170},
			Tk: []byte{2, 185, 191, 213, 246, 102, 18, 67, 247, 17, 25, 74, 169, 237, 67, 141, 165, 76, 249, 209, 183, 215, 253, 118, 118, 55, 24, 99, 5, 95, 71, 233, 174},
		},
		{
			Pk: []byte{3, 85, 237, 178, 30, 58, 190, 219, 126, 31, 9, 93, 40, 217, 109, 177, 70, 41, 64, 157, 2, 133, 2, 138, 23, 108, 228, 152, 234, 35, 101, 192, 173},
			Tk: []byte{3, 116, 125, 158, 22, 126, 79, 50, 46, 119, 52, 133, 6, 246, 156, 94, 138, 244, 107, 147, 25, 78, 231, 105, 162, 185, 245, 152, 196, 116, 86, 15, 30},
		},
	}
	genesisBlock.Header.DCBGovernor = DCBGovernor{
		GovernorInfo: GovernorInfo{
			boardIndex:          0,
			StartedBlock:        1,
			EndBlock:            1000, // = startedblock of decent governor
			BoardPaymentAddress: boardPaymentAddress,
			StartAmountToken:    0, //Sum of DCB token stack to all member of this board
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
	keyWallet, err := wallet.Base58CheckDeserialize(icoParams.InitialPaymentAddress)
	if err != nil {
		panic(err)
	}
	// Create genesis token tx for DCB
	dcbTokenTx := createSpecialTokenTx( // DCB
		common.Hash(common.DCBTokenID),
		"Decentralized central bank token",
		"DCB",
		icoParams.InitialDCBToken,
		keyWallet.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&dcbTokenTx)

	// Create genesis token tx for GOV
	govTokenTx := createSpecialTokenTx(
		common.Hash(common.GOVTokenID),
		"Government token",
		"GOV",
		icoParams.InitialGOVToken,
		keyWallet.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&govTokenTx)

	// Create genesis token tx for CMB
	cmbTokenTx := createSpecialTokenTx(
		common.Hash(common.CMBTokenID),
		"Commercial bank token",
		"CMB",
		icoParams.InitialCMBToken,
		keyWallet.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&cmbTokenTx)

	// Create genesis token tx for BOND test
	bondTokenTx := createSpecialTokenTx(
		common.Hash([common.HashSize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		"BondTest",
		"BONTest",
		icoParams.InitialBondToken,
		keyWallet.KeySet.PaymentAddress,
	)
	genesisBlock.AddTransaction(&bondTokenTx)

	// calculate merkle root tx for genesis block
	genesisBlock.Header.MerkleRoot = self.CalcMerkleRoot(genesisBlock.Transactions)

	// genesisBlock.Header.MerkleRootCommitments = self.calcCommitmentMerkleRoot(tx)
	// fmt.Printf("Anchor: %x\n", genesisBlock.Header.MerkleRootCommitments[:])

	return &genesisBlock
}
