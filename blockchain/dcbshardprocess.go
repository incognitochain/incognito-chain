package blockchain

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

type dividendPair struct {
	DividendID uint64
	ForDCB     bool
}

func (bc *BlockChain) ProcessLoanForBlock(block *ShardBlock) error {
	for _, tx := range block.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.LoanUnlockMeta:
			{
				// Confirm that loan is withdrawed
				tx := tx.(*transaction.Tx)
				meta := tx.GetMetadata().(*metadata.LoanUnlock)
				err := bc.config.DataBase.StoreLoanWithdrawed(meta.LoanID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (bc *BlockChain) processDividendPayment(receiversToRemove map[dividendPair][][]byte) error {
	for pair, receivers := range receiversToRemove {
		// fmt.Printf("[db] pair, rec: %+v %+v\n", pair, receivers)
		// Get list of token holders left
		paymentAddresses, amounts, _, _ := bc.config.DataBase.GetDividendReceiversForID(pair.DividendID, pair.ForDCB)

		// Update list of token holders left
		addrNotPaid := []privacy.PaymentAddress{}
		amountsNotPaid := []uint64{}
		for i, addr := range paymentAddresses {
			remove := false
			for _, pubkey := range receivers {
				if bytes.Equal(pubkey, addr.Pk[:]) {
					// fmt.Printf("[db] remove divRec %x, %d\n", pubkey, amounts[i])
					remove = true
					break
				}
			}
			if !remove {
				addrNotPaid = append(addrNotPaid, addr)
				amountsNotPaid = append(amountsNotPaid, amounts[i])
			}
		}
		err := bc.config.DataBase.StoreDividendReceiversForID(pair.DividendID, pair.ForDCB, addrNotPaid, amountsNotPaid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockChain) ProcessDividendForBlock(block *ShardBlock) error {
	receiversToRemove := map[dividendPair][][]byte{}
	for _, tx := range block.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.DividendPaymentMeta:
			_, receiver, _ := tx.GetUniqueReceiver()
			meta := tx.GetMetadata().(*metadata.DividendPayment)
			forDCB := meta.TokenID.IsEqual(&common.DCBTokenID)
			pair := dividendPair{
				DividendID: meta.DividendID,
				ForDCB:     forDCB,
			}
			receiversToRemove[pair] = append(receiversToRemove[pair], receiver)
			// fmt.Printf("[db] receiversToRemove: %x %x\n", receiversToRemove, receiver)
		}
	}
	if len(receiversToRemove) > 0 {
		return bc.processDividendPayment(receiversToRemove)
	}
	return nil
}

func (bc *BlockChain) StoreMetadataInstructions(inst []string, shardID byte) error {
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
	}
	switch inst[0] {
	case strconv.Itoa(metadata.IssuingRequestMeta):
		return bc.storeIssuingResponseInstruction(inst, shardID)
	case strconv.Itoa(metadata.ContractingRequestMeta):
		return bc.storeContractingResponseInstruction(inst, shardID)
	}
	return nil
}

func (bc *BlockChain) storeIssuingResponseInstruction(inst []string, shardID byte) error {
	// fmt.Printf("[db] store meta inst: %+v\n", inst)
	if strconv.Itoa(int(shardID)) != inst[1] {
		return nil
	}

	issuingInfo := &IssuingInfo{}
	err := json.Unmarshal([]byte(inst[3]), issuingInfo)
	if err != nil {
		return err
	}

	instType := inst[2]
	return bc.config.DataBase.StoreIssuingInfo(issuingInfo.RequestedTxID, issuingInfo.Amount, instType)
}

func (bc *BlockChain) storeContractingResponseInstruction(inst []string, shardID byte) error {
	// fmt.Printf("[db] store meta inst: %+v\n", inst)
	if strconv.Itoa(int(shardID)) != inst[1] {
		return nil
	}

	contractingInfo := &ContractingInfo{}
	err := json.Unmarshal([]byte(inst[3]), contractingInfo)
	if err != nil {
		return err
	}

	instType := inst[2]
	return bc.config.DataBase.StoreContractingInfo(contractingInfo.RequestedTxID, contractingInfo.BurnedConstAmount, contractingInfo.RedeemAmount, instType)
}

func (bc *BlockChain) processDividendSubmitInst(inst []string) error {
	ds, err := metadata.ParseDividendSubmitActionValue(inst[1])
	if err != nil {
		return err
	}

	// Store current list of token holders to local state
	_, holders, amounts, err := bc.GetAmountPerAccount(ds.TokenID)
	if err != nil {
		return err
	}
	forDCB := ds.TokenID.IsEqual(&common.DCBTokenID)
	return bc.config.DataBase.StoreDividendReceiversForID(ds.DividendID, forDCB, holders, amounts)
}

// ProcessStandAloneInstructions processes all stand-alone instructions in block (e.g., DividendSubmit, Salary)
func (bc *BlockChain) ProcessStandAloneInstructions(block *ShardBlock) error {
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue
		}
		switch inst[0] {
		case strconv.Itoa(metadata.DividendSubmitMeta):
			if err := bc.processDividendSubmitInst(inst); err != nil {
				return err
			}
		}
	}
	return nil
}

//func (bc *BlockChain) UpdateDividendPayout(block *Block) error {
//	for _, tx := range block.Transactions {
//		switch tx.GetMetadataType() {
//		case metadata.DividendMeta:
//			{
//				tx := tx.(*transaction.Tx)
//				meta := tx.Metadata.(*metadata.Dividend)
//				if tx.Proof == nil {
//					return errors.New("Miss output in tx")
//				}
//				for _, _ = range tx.Proof.OutputCoins {
//					keySet := cashec.KeySet{
//						PaymentAddress: meta.PaymentAddress,
//					}
//					vouts, err := bc.GetUnspentTxCustomTokenVout(keySet, meta.TokenID)
//					if err != nil {
//						return err
//					}
//					for _, vout := range vouts {
//						txHash := vout.GetTxCustomTokenID()
//						err := bc.config.DataBase.UpdateRewardAccountUTXO(meta.TokenID, keySet.PaymentAddress.Pk, &txHash, vout.GetIndex())
//						if err != nil {
//							return err
//						}
//					}
//				}
//			}
//		}
//	}
//	return nil
//}

// func (bc *BlockChain) ProcessCMBTxs(block *Block) error {
// 	for _, tx := range block.Transactions {
// 		switch tx.GetMetadataType() {
// 		case metadata.CMBInitRequestMeta:
// 			{
// 				err := bc.processCMBInitRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitResponseMeta:
// 			{
// 				err := bc.processCMBInitResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitRefundMeta:
// 			{
// 				err := bc.processCMBInitRefund(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBDepositSendMeta:
// 			{
// 				err := bc.processCMBDepositSend(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawRequestMeta:
// 			{
// 				err := bc.processCMBWithdrawRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawResponseMeta:
// 			{
// 				err := bc.processCMBWithdrawResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}

// 	// Penalize late response for cmb withdraw request
// 	return bc.findLateWithdrawResponse(uint64(block.Header.Height))
// }
