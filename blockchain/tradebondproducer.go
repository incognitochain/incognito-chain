package blockchain

import (
	"fmt"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

func (bc *BlockChain) CalcTradeData(inst string) (*component.TradeData, error) {
	tradeID, reqAmount, err := metadata.ParseTradeActivationActionValue(inst)
	if err != nil {
		return nil, fmt.Errorf("failed parsing ta action: %v", err)
	}

	// Okay to use unsynced BeaconBestState since we can skip activating trade if it is deleted
	bondID, buy, activated, amount, err := bc.GetLatestTradeActivation(tradeID)
	if err != nil {
		return nil, fmt.Errorf("failed getting latest trade: %v", err)
	}

	return &component.TradeData{
		TradeID:   tradeID,
		BondID:    bondID,
		Buy:       buy,
		Activated: activated,
		Amount:    amount,
		ReqAmount: reqAmount,
	}, nil
}

func (bc *BlockChain) GetSellBondPrice(bondID *common.Hash) uint64 {
	buyPrice := bc.BestState.Beacon.StabilityInfo.Oracle.Bonds[bondID.String()]
	if buyPrice == 0 {
		buyPrice = bc.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams.SellingBonds.BondPrice
	}
	return buyPrice
}

func (blockgen *BlkTmplGenerator) buildTradeActivationTx(
	inst string,
	unspentTokens map[string]([]transaction.TxTokenVout),
	producerPrivateKey *privacy.PrivateKey,
	tradeActivated map[string]bool,
) ([]metadata.Transaction, error) {
	data, err := blockgen.chain.CalcTradeData(inst)
	if err != nil {
		fmt.Printf("[db] calcTradeData err: %+v\n", err)
		return nil, nil // Skip activation
	}

	// Ignore activation request if params are unsynced
	activatedInBlock := tradeActivated[string(data.TradeID)]
	if data.Activated || data.ReqAmount > data.Amount || activatedInBlock {
		fmt.Printf("[db] skip building buy sell tx: %t %t %d %d\n", data.Activated, activatedInBlock, data.ReqAmount, data.Amount)
		return nil, nil
	}

	fmt.Printf("[db] trade act tx data: %h %t %d\n", data.BondID, data.Buy, data.ReqAmount)
	txs := []metadata.Transaction{}
	if data.Buy {
		txs, err = blockgen.buildTradeBuySellRequestTx(data.TradeID, data.BondID, data.ReqAmount, producerPrivateKey)
	} else {
		txs, err = blockgen.buildTradeBuyBackRequestTx(data.TradeID, data.BondID, data.ReqAmount, unspentTokens, producerPrivateKey)
	}

	if err != nil {
		return nil, err
	}

	tradeActivated[string(data.TradeID)] = true
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
	buyPrice := blockgen.chain.GetSellBondPrice(bondID)

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
		fmt.Printf("[db] build buysell request err: %v\n", err)
		// Skip building tx buyback/buysell if error (retry later)
		return nil, nil
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
	fmt.Printf("[db] building buyback request tx: %d %h\n", amount, bondID)
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
	if err != nil {
		fmt.Printf("[db] build buyback request err: %v\n", err)
		// Skip building tx buyback/buysell if error (retry later)
		return nil, nil
	}

	// Update list of token available for next request
	if usedID >= 0 {
		unspentTokens[bondID.String()] = unspentTokens[bondID.String()][usedID:]
	}

	fmt.Printf("[db] done built buyback request tx: %v\n", usedID)
	return []metadata.Transaction{txToken}, nil
}
