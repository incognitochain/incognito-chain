package frombeaconins

import (
	"encoding/json"
	"github.com/constant-money/constant-chain/transaction"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
)

type TxSendBackTokenVoteFailIns struct {
	BoardType      common.BoardType
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	PropertyID     common.Hash
}

func NewTxSendBackTokenVoteFailIns(boardType common.BoardType, paymentAddress privacy.PaymentAddress, amount uint64, propertyID common.Hash) *TxSendBackTokenVoteFailIns {
	return &TxSendBackTokenVoteFailIns{BoardType: boardType, PaymentAddress: paymentAddress, Amount: amount, PropertyID: propertyID}
}

func (txSendBackTokenVoteFailIns *TxSendBackTokenVoteFailIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txSendBackTokenVoteFailIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(txSendBackTokenVoteFailIns.PaymentAddress)
	return []string{
		strconv.Itoa(metadata.SendBackTokenVoteBoardFailMeta),
		strconv.Itoa(int(shardID)),
		string(content),
	}, nil
}

func GetShardIDFromPaymentAddressBytes(paymentAddress privacy.PaymentAddress) byte {
	lastByte := paymentAddress.Pk[len(paymentAddress.Pk)-1]
	return common.GetShardIDFromLastByte(lastByte)
}

func (txSendBackTokenVoteFailIns *TxSendBackTokenVoteFailIns) BuildTransaction(
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) (metadata.Transaction, error) {
	return NewSendBackTokenVoteFailTx(
		txSendBackTokenVoteFailIns.BoardType,
		minerPrivateKey,
		db,
		txSendBackTokenVoteFailIns.PaymentAddress,
		txSendBackTokenVoteFailIns.Amount,
		bcr,
		shardID,
	)
}

func NewSendBackTokenVoteFailIns(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	propertyID common.Hash,
) *TxSendBackTokenVoteFailIns {
	return &TxSendBackTokenVoteFailIns{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		PropertyID:     propertyID,
	}
}

func NewSendBackTokenVoteFailTx(
	boardType common.BoardType,
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) (metadata.Transaction, error) {

	//create token params
	customTokenParamTx := mintDCBTokenParam
	if boardType == common.GOVBoard {
		customTokenParamTx = mintGOVTokenParam
	}
	customTokenParamTx.Amount = amount

	//CALL DB
	listCustomTokens, err := GetListCustomTokens(db, bcr)
	if err != nil {
		return nil, err
	}
	//Init tx custom token
	paymentInfo := privacy.PaymentInfo{
		PaymentAddress: paymentAddress,
		Amount:         amount,
	}
	txCustom := &transaction.TxCustomToken{}
	err = txCustom.Init(
		minerPrivateKey,
		[]*privacy.PaymentInfo{&paymentInfo},
		nil,
		0,
		&customTokenParamTx,
		listCustomTokens,
		db,
		metadata.NewSendBackTokenVoteFailMetadata(),
		false,
		shardID,
	)
	return txCustom, err
}

func GetListCustomTokens(
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
) (map[common.Hash]transaction.TxCustomToken, error) {
	data, err := db.ListCustomToken()
	if err != nil {
		return nil, err
	}
	result := make(map[common.Hash]transaction.TxCustomToken)
	for _, txData := range data {
		hash := common.Hash{}
		hash.SetBytes(txData)
		_, blockHash, index, tx, err := bcr.GetTransactionByHash(&hash)
		_ = blockHash
		_ = index
		if err != nil {
			return nil, err
		}
		txCustomToken := tx.(*transaction.TxCustomToken)
		result[txCustomToken.TxTokenData.PropertyID] = *txCustomToken
	}
	return result, nil
}
