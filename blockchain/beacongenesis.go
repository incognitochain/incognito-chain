package blockchain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata/frombeaconins"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
)

func CreateBeaconGenesisBlock(
	version int,
	genesisParams GenesisParams,
) *BeaconBlock {
	inst := [][]string{}
	// build validator beacon
	// test generate public key in utility/generateKeys
	beaconAssingInstruction := []string{StakeAction}
	beaconAssingInstruction = append(beaconAssingInstruction, strings.Join(genesisParams.PreSelectBeaconNodeSerializedPubkey[:], ","))
	beaconAssingInstruction = append(beaconAssingInstruction, "beacon")

	shardAssingInstruction := []string{StakeAction}
	shardAssingInstruction = append(shardAssingInstruction, strings.Join(genesisParams.PreSelectShardNodeSerializedPubkey[:], ","))
	shardAssingInstruction = append(shardAssingInstruction, "shard")

	inst = append(inst, beaconAssingInstruction)
	inst = append(inst, shardAssingInstruction)

	// init network param
	inst = append(inst, []string{InitAction, salaryPerTx, fmt.Sprintf("%v", genesisParams.SalaryPerTx)})
	inst = append(inst, []string{InitAction, basicSalary, fmt.Sprintf("%v", genesisParams.BasicSalary)})
	inst = append(inst, []string{InitAction, salaryFund, strconv.Itoa(int(genesisParams.InitFundSalary))})
	inst = append(inst, []string{InitAction, feePerTxKb, fmt.Sprintf("%v", genesisParams.FeePerTxKb)})

	inst = append(inst, []string{SetAction, "randomnumber", strconv.Itoa(int(0))})

	// init stability params
	stabilityInsts := createStabilityGenesisInsts()
	inst = append(inst, stabilityInsts...)

	body := BeaconBody{ShardState: nil, Instructions: inst}
	header := BeaconHeader{
		Timestamp:           time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:              1,
		Version:             1,
		Round:               1,
		Epoch:               1,
		PrevBlockHash:       common.Hash{},
		ValidatorsRoot:      common.Hash{},
		BeaconCandidateRoot: common.Hash{},
		ShardCandidateRoot:  common.Hash{},
		ShardValidatorsRoot: common.Hash{},
		ShardStateHash:      common.Hash{},
		InstructionHash:     common.Hash{},
	}

	block := &BeaconBlock{
		Body:   body,
		Header: header,
	}

	return block
}

// createStabilityGenesisInsts generates instructions to initialize stability params for genesis block of beacon chain
func createStabilityGenesisInsts() [][]string {
	govInsts := createGOVGenesisInsts()
	dcbInsts := createDCBGenesisInsts()
	insts := [][]string{}
	insts = append(insts, govInsts...)
	insts = append(insts, dcbInsts...)
	return insts
}

func createGOVGenesisInsts() [][]string {
	return [][]string{createGOVGenesisBoardInst(), createGOVGenesisParamInst()}
}

func createGOVGenesisBoardInst() []string {
	boardAddress := []privacy.PaymentAddress{
		// Payment4: 1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba
		//privacy.PaymentAddress{
		//Pk: []byte{3, 36, 133, 3, 185, 44, 62, 112, 196, 239, 49, 190, 100, 172, 50, 147, 196, 154, 105, 211, 203, 57, 242, 110, 34, 126, 100, 226, 74, 148, 128, 167, 0},
		//Tk: []byte{2, 134, 3, 114, 89, 60, 134, 3, 185, 245, 176, 187, 244, 145, 250, 149, 67, 98, 68, 106, 69, 200, 228, 209, 3, 26, 231, 15, 36, 251, 211, 186, 159},
		//},
	}
	govBoardInst := &frombeaconins.AcceptGOVBoardIns{
		BoardPaymentAddress: boardAddress,
		StartAmountToken:    0,
	}
	govInst, _ := govBoardInst.GetStringFormat()
	return govInst
}

func createGOVGenesisParamInst() []string {
	// Bond
	sellingBonds := &component.SellingBonds{
		BondName:       "Bond 1000 blocks",
		BondSymbol:     "BND1000",
		TotalIssue:     1000,
		BondsToSell:    1000,
		BondPrice:      100, // 1 constant
		Maturity:       3,
		BuyBackPrice:   120, // 1.2 constant
		StartSellingAt: 0,
		SellingWithin:  100000,
	}
	sellingGOVTokens := &component.SellingGOVTokens{
		TotalIssue:      1000,
		GOVTokensToSell: 1000,
		GOVTokenPrice:   500, // 5 constant
		StartSellingAt:  0,
		SellingWithin:   10000,
	}

	govParams := component.GOVParams{
		SalaryPerTx:      uint64(0),
		BasicSalary:      uint64(0),
		FeePerKbTx:       uint64(0),
		SellingBonds:     sellingBonds,
		SellingGOVTokens: sellingGOVTokens,
		RefundInfo:       nil,
		OracleNetwork:    nil,
	}

	// First proposal created by GOV, reward back to itself
	keyWalletGOVAccount, _ := wallet.Base58CheckDeserialize(common.GOVAddress)
	govAddress := keyWalletGOVAccount.KeySet.PaymentAddress
	govUpdateInst := &frombeaconins.UpdateGOVConstitutionIns{
		SubmitProposalInfo: component.SubmitProposalInfo{
			ExecuteDuration:   0,
			Explanation:       "Genesis GOV proposal",
			PaymentAddress:    govAddress,
			ConstitutionIndex: 0,
		},
		GOVParams: govParams,
		Voters:    []privacy.PaymentAddress{},
	}
	govInst, _ := govUpdateInst.GetStringFormat()
	return govInst
}

func createDCBGenesisInsts() [][]string {
	return [][]string{createDCBGenesisBoardInst(), createDCBGenesisParamsInst()}
}

func createDCBGenesisBoardInst() []string {
	// TODO(@0xbunyip): set correct board address
	boardAddress := []privacy.PaymentAddress{
		// Payment4: 1Uv3VB24eUszt5xqVfB87ninDu7H43gGxdjAUxs9j9JzisBJcJr7bAJpAhxBNvqe8KNjM5G9ieS1iC944YhPWKs3H2US2qSqTyyDNS4Ba
		//privacy.PaymentAddress{
		//Pk: []byte{3, 36, 133, 3, 185, 44, 62, 112, 196, 239, 49, 190, 100, 172, 50, 147, 196, 154, 105, 211, 203, 57, 242, 110, 34, 126, 100, 226, 74, 148, 128, 167, 0},
		//Tk: []byte{2, 134, 3, 114, 89, 60, 134, 3, 185, 245, 176, 187, 244, 145, 250, 149, 67, 98, 68, 106, 69, 200, 228, 209, 3, 26, 231, 15, 36, 251, 211, 186, 159},
		//},
	}

	dcbBoardInst := &frombeaconins.AcceptDCBBoardIns{
		BoardPaymentAddress: boardAddress,
		StartAmountToken:    0,
	}
	dcbInst, _ := dcbBoardInst.GetStringFormat()
	return dcbInst
}

func createDCBGenesisParamsInst() []string {
	// Crowdsale bonds
	bondID, _ := common.NewHashFromStr("4c420b974449ac188c155a7029706b8419a591ee398977d00000000000000000")
	buyBondSaleID := [32]byte{1}
	sellBondSaleID := [32]byte{2}
	saleData := []component.SaleData{
		component.SaleData{
			SaleID:           buyBondSaleID[:],
			EndBlock:         1000,
			BuyingAsset:      *bondID,
			BuyingAmount:     100, // 100 bonds
			DefaultBuyPrice:  100, // 100 cent per bond
			SellingAsset:     common.ConstantID,
			SellingAmount:    15000, // 150 CST in Nano
			DefaultSellPrice: 100,   // 100 cent per CST
		},
		component.SaleData{
			SaleID:           sellBondSaleID[:],
			EndBlock:         2000,
			BuyingAsset:      common.ConstantID,
			BuyingAmount:     25000, // 250 CST in Nano
			DefaultBuyPrice:  100,   // 100 cent per CST
			SellingAsset:     *bondID,
			SellingAmount:    200, // 200 bonds
			DefaultSellPrice: 100, // 100 cent per bond
		},
	}

	// Reserve
	raiseReserveData := map[common.Hash]*component.RaiseReserveData{
		common.ETHAssetID: &component.RaiseReserveData{
			EndBlock: 1000,
			Amount:   1000,
		},
		common.USDAssetID: &component.RaiseReserveData{
			EndBlock: 1000,
			Amount:   1000,
		},
	}
	spendReserveData := map[common.Hash]*component.SpendReserveData{
		common.ETHAssetID: &component.SpendReserveData{
			EndBlock:        1000,
			ReserveMinPrice: 1000,
			Amount:          10000000,
		},
	}

	// Dividend
	divAmounts := []uint64{0}

	// Collateralized loan
	loanParams := []component.LoanParams{
		component.LoanParams{
			InterestRate:     100,   // 1%
			Maturity:         1000,  // 1 month in blocks
			LiquidationStart: 15000, // 150%
		},
	}

	dcbParams := component.DCBParams{
		ListSaleData:             saleData,
		MinLoanResponseRequire:   1,
		MinCMBApprovalRequire:    1,
		LateWithdrawResponseFine: 0,
		RaiseReserveData:         raiseReserveData,
		SpendReserveData:         spendReserveData,
		DividendAmount:           divAmounts[0],
		ListLoanParams:           loanParams,
	}

	// First proposal created by DCB, reward back to itself
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbAddress := keyWalletDCBAccount.KeySet.PaymentAddress
	dcbUpdateInst := &frombeaconins.UpdateDCBConstitutionIns{
		SubmitProposalInfo: component.SubmitProposalInfo{
			ExecuteDuration:   0,
			Explanation:       "Genesis DCB proposal",
			PaymentAddress:    dcbAddress,
			ConstitutionIndex: 0,
		},
		DCBParams: dcbParams,
		Voters:    []privacy.PaymentAddress{},
	}
	dcbInst, _ := dcbUpdateInst.GetStringFormat()
	return dcbInst
}
