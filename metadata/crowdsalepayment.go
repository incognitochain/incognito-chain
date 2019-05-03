package metadata

import (
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

type CrowdsalePayment struct {
	SaleID []byte

	MetadataBase
}

func (csRes *CrowdsalePayment) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, shardID byte, db database.DatabaseInterface) (bool, error) {
	// Check if sale exists
	sale, err := bcr.GetSaleData(csRes.SaleID) // okay to use unsynced data since we only use immutable fields
	if err != nil {
		return false, err
	}

	// TODO(@0xbunyip): check if sending address is DCB's
	//keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	//if !bytes.Equal(txr.GetSigPubKey(), keyWalletDCBAccount.KeySet.PaymentAddress.Pk[:]) {
	//	return false, fmt.Errorf("Crowdsale payment must send asset from DCB address")
	//}

	// TODO(@0xbunyip): check double spending for coinbase CST tx?
	if !sale.Buy {
		// Check if sent from DCB address
		// check double spending if selling bond
		return true, nil
	}
	return false, nil
}

func (csRes *CrowdsalePayment) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	if len(csRes.SaleID) == 0 {
		return false, false, errors.New("Wrong request info's SaleID")
	}
	return false, true, nil
}

func (csRes *CrowdsalePayment) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (csRes *CrowdsalePayment) Hash() *common.Hash {
	record := string(csRes.SaleID)

	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (csRes *CrowdsalePayment) CalculateSize() uint64 {
	return calculateSize(csRes)
}

func (csRes *CrowdsalePayment) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instUsed []int,
	shardID byte,
	tx Transaction,
	bcr BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	fmt.Printf("[db] verifying crowdsale payment tx\n")
	idx := -1
	for i, inst := range insts {
		if instUsed[i] > 0 || inst[0] != strconv.Itoa(CrowdsalePaymentMeta) || inst[1] != strconv.Itoa(int(shardID)) {
			continue
		}
		cpi, err := component.ParseCrowdsalePaymentInstruction(inst[2])
		if err != nil {
			continue
		}
		unique, pk, amount, assetID := tx.GetTransferData()
		txData := component.CrowdsalePaymentInstruction{
			PaymentAddress: privacy.PaymentAddress{Pk: pk},
			Amount:         amount,
			AssetID:        *assetID,
			SaleID:         nil, // no need to check these last fields
			SentAmount:     0,
			UpdateSale:     false,
		}
		if unique && txData.Compare(cpi) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false, errors.Errorf("no instruction found for CrowdsalePayment tx %s", tx.Hash().String())
	}

	instUsed[idx] += 1
	fmt.Printf("[db] inst %d matched\n", idx)
	return true, nil
}

func (csRes *CrowdsalePayment) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return true
}
