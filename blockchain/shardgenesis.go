package blockchain

import (
	"log"
	"time"

	"github.com/ninjadotorg/constant/wallet"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

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

//TODO Write function to create Shard block of shard chain here
func CreateShardGenesisBlock(
	version int,
	shardNodes []string,
	icoParams IcoParams,
) *ShardBlock {

	log.Printf("Ico payment address:", icoParams.InitialPaymentAddress)
	keyWallet, err := wallet.Base58CheckDeserialize(icoParams.InitialPaymentAddress)
	if err != nil {
		panic(err)
	}

	body := ShardBody{}
	header := ShardHeader{
		Timestamp:     time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:        1,
		Version:       1,
		PrevBlockHash: common.Hash{},
		BeaconHeight:  1,

		//TODO:
		SalaryFund: icoParams.InitFundSalary,
	}

	block := &ShardBlock{
		Body:   body,
		Header: header,
	}

	// Create genesis token tx for DCB
	dcbTokenTx := createSpecialTokenTx( // DCB
		common.Hash(common.DCBTokenID),
		"Decentralized central bank token",
		"DCB",
		icoParams.InitialDCBToken,
		keyWallet.KeySet.PaymentAddress,
	)
	block.Body.Transactions = append(block.Body.Transactions, &dcbTokenTx)

	// Create genesis token tx for GOV
	govTokenTx := createSpecialTokenTx(
		common.Hash(common.GOVTokenID),
		"Government token",
		"GOV",
		icoParams.InitialGOVToken,
		keyWallet.KeySet.PaymentAddress,
	)
	block.Body.Transactions = append(block.Body.Transactions, &govTokenTx)

	// Create genesis token tx for CMB
	cmbTokenTx := createSpecialTokenTx(
		common.Hash(common.CMBTokenID),
		"Commercial bank token",
		"CMB",
		icoParams.InitialCMBToken,
		keyWallet.KeySet.PaymentAddress,
	)
	block.Body.Transactions = append(block.Body.Transactions, &cmbTokenTx)

	// Create genesis token tx for BOND test
	bondTokenTx := createSpecialTokenTx(
		common.Hash([common.HashSize]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		"BondTest",
		"BONTest",
		icoParams.InitialBondToken,
		keyWallet.KeySet.PaymentAddress,
	)
	block.Body.Transactions = append(block.Body.Transactions, &bondTokenTx)

	// Create genesis vote token tx for DCB
	// voteDCBTokenTx := createSpecialTokenTx(
	// 	common.Hash(common.VoteDCBTokenID),
	// 	"Bond",
	// 	"BON",
	// 	icoParams.InitialVoteDCBToken,
	// 	initialKeySet.PaymentAddress,
	// )
	// block.Body.Transactions = append(block.Body.Transactions, &voteDCBTokenTx)

	// voteGOVTokenTx := createSpecialTokenTx(
	// 	common.Hash(common.VoteGOVTokenID),
	// 	"Bond",
	// 	"BON",
	// 	icoParams.InitialVoteGOVToken,
	// 	initialKeySet.PaymentAddress,
	// )
	// block.Body.Transactions = append(block.Body.Transactions, &voteGOVTokenTx)

	return block
}
