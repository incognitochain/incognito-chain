package transaction

import (
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

var allowedAsset = []string{common.AssetTypeCoin, common.AssetTypeBond}

type TxCrowdsale struct {
	*TxCustomToken

	BaseAsset     string
	QuoteAsset    string
	Price         uint64
	EscrowAccount client.PaymentAddress
}

// CreateEmptyCrowdsaleTx - return an init custom token transaction
func CreateEmptyCrowdsaleTx() (*TxCrowdsale, error) {
	emptyTx, err := CreateEmptyTx(common.TxCrowdsale)

	if err != nil {
		return nil, err
	}

	txToken := TxTokenData{}

	txCrowdsale := &TxCrowdsale{
		TxCustomToken: &TxCustomToken{
			Tx:          *emptyTx,
			TxTokenData: txToken,
		},
	}
	return txCrowdsale, nil
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
	record += tx.BaseAsset + tx.QuoteAsset
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
		if !isAllowed(tx.BaseAsset, allowedAsset) || !isAllowed(tx.QuoteAsset, allowedAsset) {
			return false
		}

		// TODO, verify signature
		return true
	}
	return false
}

// CreateTxCrowdsale ...
func CreateTxCrowdsale(senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx,
	listCustomToken map[common.Hash]TxCustomToken,
	baseAsset string,
	quoteAsset string,
	price uint64,
) (*TxCrowdsale, error) {
	txCustom, err := buildTxCustomToken(
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

	// TODO(@0xbunyip): sign on full crowdsale token
	// Sign tx
	txCustom, err = SignPrivacyTxCustomToken(txCustom)
	if err != nil {
		return nil, err
	}

	tx := &TxCrowdsale{
		TxCustomToken: txCustom,
		BaseAsset:     baseAsset,
		QuoteAsset:    quoteAsset,
		Price:         price,
	}

	if !isAllowed(tx.BaseAsset, allowedAsset) || !isAllowed(tx.QuoteAsset, allowedAsset) {
		return nil, fmt.Errorf("Asset type is incorrect")
	}

	return tx, nil
}
