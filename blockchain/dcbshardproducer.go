package blockchain

import (
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
)

func (blockchain *BlockChain) GetAmountPerAccount(tokenID *common.Hash) (uint64, []privacy.PaymentAddress, []uint64, error) {
	tokenHoldersMap, err := blockchain.config.DataBase.GetCustomTokenPaymentAddressesBalanceUnreward(tokenID)
	if err != nil {
		return 0, nil, nil, err
	}

	// Get total token supply
	totalTokenAmount := uint64(0)
	for _, value := range tokenHoldersMap {
		totalTokenAmount += value
	}

	// Get amount per account (only count unrewarded utxo)
	tokenHolders := []privacy.PaymentAddress{}
	amounts := []uint64{}
	for holder, amount := range tokenHoldersMap {
		paymentAddressInBytes, _, _ := base58.Base58Check{}.Decode(holder)
		keySet := cashec.KeySet{}
		keySet.PaymentAddress = privacy.PaymentAddress{}
		keySet.PaymentAddress.SetBytes(paymentAddressInBytes)
		//vouts, err := blockchain.GetUnspentTxCustomTokenVout(keySet, tokenID)
		//if err != nil {
		//	return 0, nil, nil, err
		//}
		//amount := uint64(0)
		//for _, vout := range vouts {
		//	amount += vout.Value
		//}

		if amount > 0 {
			tokenHolders = append(tokenHolders, keySet.PaymentAddress)
			amounts = append(amounts, amount)
		}
	}
	return totalTokenAmount, tokenHolders, amounts, nil
}

func (blockgen *BlkTmplGenerator) buildInstitutionDividendPaymentTxs(forDCB bool, producerPrivateKey *privacy.SpendingKey) ([]*transaction.Tx, error) {
	// Get latest dividend proposal id and amount
	id, cstToPayout := blockgen.chain.BestState.Beacon.GetLatestDividendProposal(forDCB)
	if id == 0 || cstToPayout == 0 {
		return nil, nil // No Dividend proposal found
	}

	// Check in shard state if DividendSubmit tx has been included in chain
	receivers, amounts, hasValue, err := blockgen.chain.config.DataBase.GetDividendReceiversForID(id, forDCB)
	if err != nil {
		return nil, err
	}
	if !hasValue {
		// fmt.Printf("[db] waiting for dividend submit tx\n")
		return nil, nil // Waiting for Dividend submit tx to be included in block
	}

	if len(receivers) == 0 || len(amounts) == 0 {
		// fmt.Printf("[db] paid to all receivers\n")
		return nil, nil // Paid to all receivers for the latest dividend proposal
	}

	// Get dividend info
	tokenID := &common.DCBTokenID
	if !forDCB {
		tokenID = &common.GOVTokenID
	}
	totalTokenOnAllShards, cstToPayout, aggregated := blockgen.chain.BestState.Beacon.GetDividendAggregatedInfo(id, tokenID)
	if !aggregated {
		// fmt.Printf("[db] waiting for aggregation\n")
		return nil, nil // Waiting for beacon to aggregate dividend infos
	}

	// Make dividend payments to token holders
	paymentAddresses := []*privacy.PaymentAddress{}
	payoutAmounts := []uint64{}
	for i, amount := range amounts {
		if i > metadata.MaxDivTxsPerBlock {
			break
		}
		receiver := &privacy.PaymentAddress{
			Pk: receivers[i].Pk,
			Tk: receivers[i].Tk,
		}

		receiverCstAmount := amount * cstToPayout / totalTokenOnAllShards
		paymentAddresses = append(paymentAddresses, receiver)
		payoutAmounts = append(payoutAmounts, receiverCstAmount)
		// fmt.Printf("[db] div rec, amount: %x %d\n", receiver.Pk[:], receiverCstAmount)
	}
	// fmt.Printf("[db] paymentAddresses: %v\n", paymentAddresses)

	txs, err := transaction.BuildDividendTxs(
		id,
		tokenID,
		paymentAddresses,
		payoutAmounts,
		producerPrivateKey,
		blockgen.chain.GetDatabase(),
	)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("[db] built divPays: %d\n", len(txs))
	return txs, nil
}

func (blockgen *BlkTmplGenerator) buildDividendPaymentTxs(producerPrivateKey *privacy.SpendingKey, shardID byte) ([]metadata.Transaction, error) {
	// Build dividend payments for DCB
	forDCB := true
	dcbDividendPaymentTxs, err := blockgen.buildInstitutionDividendPaymentTxs(forDCB, producerPrivateKey)
	if err != nil {
		return nil, err
	}

	// Build dividend payments for GOV
	forDCB = false
	govDividendPaymentTxs, err := blockgen.buildInstitutionDividendPaymentTxs(forDCB, producerPrivateKey)
	if err != nil {
		return nil, err
	}

	txs := []metadata.Transaction{}
	for _, tx := range dcbDividendPaymentTxs {
		txs = append(txs, tx)
	}
	for _, tx := range govDividendPaymentTxs {
		txs = append(txs, tx)
	}

	return txs, nil
}

func (blockgen *BlkTmplGenerator) buildInstitutionDividendSubmitInst(forDCB bool, shardID byte) ([][]string, error) {
	// Get latest dividend proposal id and amount
	id, cstToPayout := blockgen.chain.BestState.Beacon.GetLatestDividendProposal(forDCB)
	if id == 0 || cstToPayout == 0 {
		// fmt.Printf("[db] no div proposal found: %t\n", forDCB)
		return nil, nil // No Dividend proposal found
	}

	// Check in shard state if DividendSubmit tx has been created
	_, _, hasValue, err := blockgen.chain.config.DataBase.GetDividendReceiversForID(id, forDCB)
	if err != nil {
		// fmt.Printf("[db] buildDivSub err: %v\n", err)
		return nil, err
	}
	if hasValue {
		// fmt.Printf("[db] divsub created: %d %t\n", id, forDCB)
		return nil, nil // Already created DividendSubmit tx in previous blocks
	}

	tokenID := &common.DCBTokenID
	if !forDCB {
		tokenID = &common.GOVTokenID
	}
	totalTokenAmount, _, _, err := blockgen.chain.GetAmountPerAccount(tokenID)
	// fmt.Printf("[db] buildDivSubmit: %t, %d, %d, %d\n", forDCB, id, totalTokenAmount, shardID)
	if err != nil {
		return nil, err
	}

	// Create instruction
	return metadata.BuildDividendSubmitInst(tokenID, id, totalTokenAmount, shardID)
}

func (blockgen *BlkTmplGenerator) buildDividendSubmitInsts(producerPrivateKey *privacy.SpendingKey, shardID byte) ([][]string, error) {
	// Dividend proposals for DCB
	submitInsts := [][]string{}
	forDCB := true
	dcbInst, err := blockgen.buildInstitutionDividendSubmitInst(forDCB, shardID)
	if err != nil {
		// fmt.Printf("[db] error building dividend submit tx for dcb: %v\n", err)
		return nil, err
	} else if len(dcbInst) > 0 {
		// fmt.Printf("[db] added divsub inst: %v\n", dcbInst)
		submitInsts = append(submitInsts, dcbInst...)
	}

	// For GOV
	forDCB = false
	govInst, err := blockgen.buildInstitutionDividendSubmitInst(forDCB, shardID)
	if err != nil {
		// fmt.Printf("[db] error building dividend submit tx for dcb: %v\n", err)
		return nil, err
	} else if len(govInst) > 0 {
		submitInsts = append(submitInsts, govInst...)
	}

	return submitInsts, nil
}

func buildBuySellConfirmInst(inst []string, shardID byte) []string {
	fmt.Printf("[db] build buy sell confirm inst\n")
	return []string{
		strconv.Itoa(component.ConfirmBuySellRequestMeta),
		strconv.Itoa(int(shardID)),
		inst[2],
		inst[3],
		inst[4],
	}
}

func buildBuyBackConfirmInst(inst []string, shardID byte) []string {
	fmt.Printf("[db] build buy back confirm inst\n")
	return []string{
		strconv.Itoa(component.ConfirmBuyBackRequestMeta),
		strconv.Itoa(int(shardID)),
		inst[2],
		inst[3],
		inst[4],
	}
}

func (blockgen *BlkTmplGenerator) buildTradeBondConfirmInsts(beaconBlocks []*BeaconBlock, shardID byte) ([][]string, error) {
	insts := [][]string{}
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
	dcbShardID := common.GetShardIDFromLastByte(dcbPk[len(dcbPk)-1])

	if shardID == dcbShardID {
		for _, beaconBlock := range beaconBlocks {
			for _, l := range beaconBlock.Body.Instructions {
				if len(l) <= 2 {
					continue
				}
				switch l[0] {
				case strconv.Itoa(metadata.BuyFromGOVRequestMeta):
					insts = append(insts, buildBuySellConfirmInst(l, shardID))

				case strconv.Itoa(metadata.BuyBackRequestMeta):
					insts = append(insts, buildBuyBackConfirmInst(l, shardID))
				}
			}
		}
	}
	return insts, nil
}
