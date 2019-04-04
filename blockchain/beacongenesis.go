package blockchain

import (
	"encoding/json"
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
	// inst = append(inst, []string{InitAction, salaryFund, strconv.Itoa(int(genesisParams.InitFundSalary))})
	inst = append(inst, []string{SetAction, "randomnumber", strconv.Itoa(int(0))})

	// init stability params
	stabilityInsts := createStabilityGenesisInsts(genesisParams)
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
func createStabilityGenesisInsts(genesisParams GenesisParams) [][]string {
	govInsts := createGOVGenesisInsts(genesisParams)
	dcbInsts := createDCBGenesisInsts()
	insts := [][]string{}
	insts = append(insts, govInsts...)
	insts = append(insts, dcbInsts...)
	return insts
}

func createGOVGenesisInsts(genesisParams GenesisParams) [][]string {

	return [][]string{
		createGOVGenesisBoardInst(),
		createGOVGenesisParamInst(genesisParams),
		createGOVGenesisOracleInst(),
		createGOVGenesisSalaryFund(genesisParams),
	}
}

func createGOVGenesisSalaryFund(genesisParams GenesisParams) []string {
	return []string{InitAction, salaryFund, strconv.Itoa(int(genesisParams.InitFundSalary))}
}

func createGOVGenesisOracleInst() []string {
	initialPrices := component.Oracle{
		DCBToken: 1000,   // $10
		GOVToken: 2000,   // $20
		Constant: 100,    // $1
		ETH:      15000,  // $150
		BTC:      400000, // $4000
	}
	content, err := json.Marshal(initialPrices)
	if err != nil {
		fmt.Println("Error to marshal oracleInitialPrices: ", err)
		return []string{}
	}
	return []string{
		InitAction,
		oracleInitialPrices,
		string(content),
	}
}

func createGOVGenesisBoardInst() []string {
	govMemberAddr := privacy.PaymentAddress{
		Pk: []byte{3, 159, 2, 42, 22, 163, 195, 221, 129, 31, 217, 133, 149, 16, 68, 108, 42, 192, 58, 95, 39, 204, 63, 68, 203, 132, 221, 48, 181, 131, 40, 189, 0},
		Tk: []byte{2, 58, 116, 58, 73, 55, 129, 154, 193, 197, 40, 130, 50, 242, 99, 84, 59, 31, 107, 85, 68, 234, 250, 118, 66, 188, 15, 139, 89, 254, 12, 38, 211},
	}
	boardAddress := []privacy.PaymentAddress{govMemberAddr}
	govBoardInst := &frombeaconins.AcceptGOVBoardIns{
		BoardPaymentAddress: boardAddress,
		StartAmountToken:    0,
	}
	govInst, _ := govBoardInst.GetStringFormat()
	return govInst
}

func createGOVGenesisParamInst(genesisParams GenesisParams) []string {
	// Bond
	// sellingBonds := &component.SellingBonds{
	// 	BondName:       "Bond 1000 blocks",
	// 	BondSymbol:     "BND1000",
	// 	TotalIssue:     1000,
	// 	BondsToSell:    1000,
	// 	BondPrice:      100, // 1 constant
	// 	Maturity:       3,
	// 	BuyBackPrice:   120, // 1.2 constant
	// 	StartSellingAt: 0,
	// 	SellingWithin:  100000,
	// }
	// sellingGOVTokens := &component.SellingGOVTokens{
	// 	TotalIssue:      1000,
	// 	GOVTokensToSell: 1000,
	// 	GOVTokenPrice:   500, // 5 constant
	// 	StartSellingAt:  0,
	// 	SellingWithin:   10000,
	// }

	oracleNetwork := &component.OracleNetwork{
		OraclePubKeys:          []string{"039f022a16a3c3dd811fd9859510446c2ac03a5f27cc3f44cb84dd30b58328bd00"},
		UpdateFrequency:        10,
		OracleRewardMultiplier: 1,
		AcceptableErrorMargin:  5,
	}

	govParams := component.GOVParams{
		SalaryPerTx:      uint64(genesisParams.SalaryPerTx),
		BasicSalary:      uint64(genesisParams.BasicSalary),
		FeePerKbTx:       uint64(genesisParams.FeePerTxKb),
		SellingBonds:     nil,
		SellingGOVTokens: nil,
		RefundInfo:       nil,
		OracleNetwork:    oracleNetwork,
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
		// Payment4: 112t8rqJHgJp2TPpNpLNx34aWHB5VH5Pys3hVjjhhf9tctVeCNmX2zQLBqzHau6LpUbSV52kXtG2hRZsuYWkXWF5kw2v24RJq791fWmQxVqy
		privacy.PaymentAddress{
			Pk: []byte{3, 159, 2, 42, 22, 163, 195, 221, 129, 31, 217, 133, 149, 16, 68, 108, 42, 192, 58, 95, 39, 204, 63, 68, 203, 132, 221, 48, 181, 131, 40, 189, 0},
			Tk: []byte{2, 58, 116, 58, 73, 55, 129, 154, 193, 197, 40, 130, 50, 242, 99, 84, 59, 31, 107, 85, 68, 234, 250, 118, 66, 188, 15, 139, 89, 254, 12, 38, 211},
		},
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
	bondID, _ := common.NewHashFromStr("a1bdba2624828899959bd3704df90859539623d89ba6767d0000000000000000")
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
