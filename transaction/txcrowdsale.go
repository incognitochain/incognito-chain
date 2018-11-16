package transaction

import (
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

var allowedAsset = []string{common.AssetTypeCoin, common.AssetTypeBond}

type SaleData struct {
	SaleID []byte // Unique id of the crowdsale to store in db
	BondID []byte // in case either base or quote asset is bond

	BaseAsset     string
	QuoteAsset    string
	Price         uint64
	EscrowAccount client.PaymentAddress
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
	record += tx.BaseAsset + tx.QuoteAsset
	record += fmt.Sprint(tx.Price)
	record += string(tx.EscrowAccount.Apk[:])

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
		// Check if all tokens are of the same kind
		bondID := ""
		if len(tx.TxTokenData.Vouts) > 0 {
			bondID = tx.TxTokenData.Vouts[0].BondID
		}

		// TODO(@0xbunyip): get Vout from Vin and check as well
		//		for _, vin := range tx.TxTokenData.Vins {
		//			if vin.BondID != bondID {
		//				return false
		//			}
		//		}

		for _, vout := range tx.TxTokenData.Vouts {
			if vout.BondID != bondID {
				return false
			}
		}

		// Check if crowdsale assets are valid
		if !isAllowed(tx.BaseAsset, allowedAsset) || !isAllowed(tx.QuoteAsset, allowedAsset) {
			return false
		}

		// TODO, verify signature
		return true
	}
	return false
}

// CreateTxCrowdsale ...
func CreateTxCrowdsale(
	senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
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
