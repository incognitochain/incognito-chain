package blockchain

import (
	"time"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

func createSpecialTokenTx(
	tokenID common.Hash,
	tokenName string,
	tokenSymbol string,
	amount uint64,
	initialAddress privacy.PaymentAddress,
) transaction.TxCustomToken {
	//log.Printf("Init token %s: %s \n", tokenSymbol, tokenID.String())
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

func CreateShardGenesisBlock(
	version int,
	icoParams GenesisParams,
) *ShardBlock {

	// keyWallet, err := wallet.Base58CheckDeserialize(icoParams.InitialPaymentAddress)
	// if err != nil {
	// 	panic(err)
	// }

	body := ShardBody{}
	header := ShardHeader{
		Timestamp:       time.Date(2018, 8, 1, 0, 0, 0, 0, time.UTC).Unix(),
		Height:          1,
		Version:         1,
		PrevBlockHash:   common.Hash{},
		BeaconHeight:    1,
		Epoch:           1,
		Round:           1,
		ProducerAddress: privacy.PaymentAddress{},
	}

	for _, tx := range icoParams.InitialConstant {
		testSalaryTX := transaction.Tx{}
		testSalaryTX.UnmarshalJSON([]byte(tx))
		body.Transactions = append(body.Transactions, &testSalaryTX)
	}

	block := &ShardBlock{
		Body:   body,
		Header: header,
	}

	return block
}
