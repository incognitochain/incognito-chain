package transaction

import (
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

var allowedAsset = []string{common.AssetTypeCoin, common.AssetTypeBond}

type SaleData struct {
	SaleID []byte // Unique id of the crowdsale to store in db
	BondID []byte // in case either base or quote asset is bond

	BuyingAsset  string
	SellingAsset string
	Price        uint64
}

type TxCrowdsale struct {
	*TxCustomToken

	*SaleData
}

// Hash returns the hash of all fields of the transaction
func (tx TxCrowdsale) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of txtoken
	record += tx.TxTokenData.PropertyName
	record += tx.TxTokenData.PropertySymbol
	record += strconv.Itoa(tx.TxTokenData.Type)
	record += strconv.Itoa(int(tx.TxTokenData.Amount))

	// add more hash of crowdsale
	record += string(tx.SaleID)
	record += tx.BuyingAsset + tx.SellingAsset
	record += fmt.Sprint(tx.Price)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func isAllowed(assetType string, allowed []string) bool {
	for _, t := range allowed {
		if assetType == t {
			return true
		}
	}
	return false
}

// ValidateTransaction ...
func (tx *TxCrowdsale) ValidateTransaction() bool {
	// validate for normal tx
	if tx.Tx.ValidateTransaction() {
		if tx.SellingAsset == common.AssetTypeBond {
			bondID := ""
			if len(tx.TxTokenData.Vouts) > 0 {
				bondID = tx.TxTokenData.Vouts[0].BondID
			}

			for _, vout := range tx.TxTokenData.Vouts {
				if vout.BondID != bondID {
					return false
				}
			}
		}

		// TODO(@0xbunyip): get Vout from Vin and check as well
		//		for _, vin := range tx.TxTokenData.Vins {
		//			if vin.BondID != bondID {
		//				return false
		//			}
		//		}

		// Check if crowdsale assets are valid
		if !isAllowed(tx.BuyingAsset, allowedAsset) || !isAllowed(tx.SellingAsset, allowedAsset) {
			return false
		}

		// TODO, verify signature
		return true
	}
	return false
}

// CreateTxCrowdsale ...
func CreateTxCrowdsale(
	senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx, // All Vins and Vouts must have the same bondID
	listCustomToken map[common.Hash]TxCustomToken,
	saleData *SaleData,
) (*TxCrowdsale, error) {
	txCustom, err := CreateTxCustomToken(
		senderKey,
		paymentInfo,
		rts,
		usableTx,
		commitments,
		fee,
		senderChainID,
		tokenParams,
		listCustomToken,
	)
	if err != nil {
		return nil, err
	}

	tx := &TxCrowdsale{
		TxCustomToken: txCustom,
		SaleData:      saleData,
	}

	if !tx.ValidateTransaction() {
		return nil, fmt.Errorf("Created tx is invalid")
	}

	// TODO: sign tx
	return tx, nil
}
