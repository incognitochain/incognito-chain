package blockchain

import (
	"bytes"
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

type tradeData struct {
	tradeID   []byte
	bondID    *common.Hash
	buy       bool
	activated bool
	amount    uint64
	reqAmount uint64
}

func (td *tradeData) Compare(td2 *tradeData) bool {
	return bytes.Equal(td.tradeID, td2.tradeID) &&
		td.bondID.IsEqual(td2.bondID) &&
		td.buy == td2.buy &&
		td.activated == td2.activated &&
		td.amount == td2.amount &&
		td.reqAmount == td2.reqAmount
}

func (bc *BlockChain) calcTradeData(inst string) (*tradeData, error) {
	tradeID, reqAmount, err := metadata.ParseTradeActivationActionValue(inst)
	if err != nil {
		fmt.Printf("[db] err parsing ta action: %v\n", err)
		return nil, err
	}

	bondID, buy, activated, amount, err := bc.GetLatestTradeActivation(tradeID)
	if err != nil {
		fmt.Printf("[db] err getting latest trade: %v\n", err)
		return nil, err
	}

	return &tradeData{
		tradeID:   tradeID,
		bondID:    bondID,
		buy:       buy,
		activated: activated,
		amount:    amount,
		reqAmount: reqAmount,
	}, nil
}

func (blockgen *BlkTmplGenerator) buildTradeActivationTx(
	inst string,
	unspentTokens map[string]([]transaction.TxTokenVout),
	producerPrivateKey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {
	fmt.Printf("[db] building trade act tx\n")
	data, err := blockgen.chain.calcTradeData(inst)
	if err != nil {
		return nil, err
	}

	// Ignore activation request if params are unsynced
	if data.activated || data.reqAmount > data.amount {
		fmt.Printf("[db] skip building buy sell tx: %t %d %d\n", data.activated, data.reqAmount, data.amount)
		return nil, nil
	}

	fmt.Printf("[db] trade act tx data: %s %t %d\n", data.bondID.String(), data.buy, data.reqAmount)
	txs := []metadata.Transaction{}
	if data.buy {
		txs, err = blockgen.buildTradeBuySellRequestTx(data.tradeID, data.bondID, data.reqAmount, producerPrivateKey)
	} else {
		txs, err = blockgen.buildTradeBuyBackRequestTx(data.tradeID, data.bondID, data.reqAmount, unspentTokens, producerPrivateKey)
	}

	if err != nil {
		fmt.Printf("[db] 3\n")
		return nil, err
	}

	fmt.Printf("[db] done built trade act tx\n")
	return txs, nil
}

func (blockgen *BlkTmplGenerator) buildTradeBuySellRequestTx(
	tradeID []byte,
	bondID *common.Hash,
	amount uint64,
	producerPrivateKey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	keyWalletBurnAccount, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	buyPrice := blockgen.chain.BestState.Beacon.StabilityInfo.Oracle.Bonds[bondID.String()]
	if buyPrice == 0 {
		buyPrice = blockgen.chain.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams.SellingBonds.BondPrice
	}

	buySellMeta := &metadata.BuySellRequest{
		PaymentAddress: keyWalletDCBAccount.KeySet.PaymentAddress,
		TokenID:        *bondID,
		Amount:         amount,
		BuyPrice:       buyPrice,
		TradeID:        tradeID,
		MetadataBase:   metadata.MetadataBase{Type: metadata.BuyFromGOVRequestMeta},
	}
	cstAmount := amount * buyPrice
	txs, err := transaction.BuildCoinbaseTxs(
		[]*privacy.PaymentAddress{&keyWalletBurnAccount.KeySet.PaymentAddress},
		[]uint64{cstAmount},
		producerPrivateKey,
		blockgen.chain.GetDatabase(),
		[]metadata.Metadata{buySellMeta},
	)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[db] built buy sell req: %d\n", cstAmount)
	return []metadata.Transaction{txs[0]}, nil
}

func (blockgen *BlkTmplGenerator) buildTradeBuyBackRequestTx(
	tradeID []byte,
	bondID *common.Hash,
	amount uint64,
	unspentTokens map[string]([]transaction.TxTokenVout),
	producerPrivateKey *privacy.PrivateKey,
) ([]metadata.Transaction, error) {
	fmt.Printf("[db] building buyback request tx: %d\n", amount)
	// TODO(@0xbunyip): not enough bonds to send ==> update activated status to retry later
	// Build metadata to send to GOV
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	buyBackMeta := &metadata.BuyBackRequest{
		PaymentAddress: keyWalletDCBAccount.KeySet.PaymentAddress,
		Amount:         amount,
		TradeID:        tradeID,
		MetadataBase:   metadata.MetadataBase{Type: metadata.BuyBackRequestMeta},
	}

	// Save list of UTXO to prevent double spending in current block
	if _, ok := unspentTokens[bondID.String()]; !ok {
		unspentTxTokenOuts, err := blockgen.chain.GetUnspentTxCustomTokenVout(keyWalletDCBAccount.KeySet, bondID)
		if err == nil {
			unspentTokens[bondID.String()] = unspentTxTokenOuts
		} else {
			unspentTokens[bondID.String()] = []transaction.TxTokenVout{}
		}
	}

	// Build tx
	keyWalletBurnAccount, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	txToken, usedID, err := transferTxToken(
		amount,
		unspentTokens[bondID.String()],
		*bondID,
		keyWalletBurnAccount.KeySet.PaymentAddress,
		buyBackMeta,
	)
	// TODO(@0xbunyip): skip building tx buyback/buysell if error (retry later)
	if err != nil {
		fmt.Printf("[db] build buyback request err: %v\n", err)
		return nil, err
	}

	// Update list of token available for next request
	if usedID >= 0 {
		unspentTokens[bondID.String()] = unspentTokens[bondID.String()][usedID:]
	}

	fmt.Printf("[db] done built buyback request tx: %v\n", usedID)
	return []metadata.Transaction{txToken}, nil
}
