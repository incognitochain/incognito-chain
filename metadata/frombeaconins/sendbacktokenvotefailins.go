package frombeaconins

import (
	"encoding/json"
	"strconv"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
)

type TxSendBackTokenVoteFailIns struct {
	PaymentAddress privacy.PaymentAddress
	Amount         uint64
	PropertyID     common.Hash
}

func (txSendBackTokenVoteFailIns *TxSendBackTokenVoteFailIns) GetStringFormat() ([]string, error) {
	content, err := json.Marshal(txSendBackTokenVoteFailIns)
	if err != nil {
		return nil, err
	}
	shardID := GetShardIDFromPaymentAddressBytes(txSendBackTokenVoteFailIns.PaymentAddress)
	return []string{
		strconv.Itoa(metadata.SendBackTokenVoteFailMeta),
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
) (metadata.Transaction, error) {
	tx := NewSendBackTokenVoteFailTx(
		minerPrivateKey,
		db,
		txSendBackTokenVoteFailIns.PaymentAddress,
		txSendBackTokenVoteFailIns.Amount,
		txSendBackTokenVoteFailIns.PropertyID,
	)
	return tx, nil
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
	minerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	propertyID common.Hash,
) metadata.Transaction {
	//==== TODO: create token params
	// customTokenParamTx := &transaction.CustomTokenParamTx{
	// PropertyID: propertyID.String(),
	// PropertyName
	// PropertySymbol
	// Amount
	// TokenTxType: transaction.CustomTokenMint,
	// Receiver
	// }
	//====
	// TODO: CALL DB
	// data, err := blockchain.config.DataBase.ListCustomToken()
	// if err != nil {
	// 	return nil, err
	// }
	// result := make(map[common.Hash]transaction.TxCustomToken)
	// for _, txData := range data {
	// 	hash := common.Hash{}
	// 	hash.SetBytes(txData)
	// 	_, blockHash, index, tx, err := blockchain.GetTransactionByHash(&hash)
	// 	_ = blockHash
	// 	_ = index
	// 	if err != nil {
	// 		return nil, NewBlockChainError(UnExpectedError, err)
	// 	}
	// 	txCustomToken := tx.(*transaction.TxCustomToken)
	// 	result[txCustomToken.TxTokenData.PropertyID] = *txCustomToken
	// }
	//=======
	// TODO: Init tx custom token
	// tx := &transaction.TxCustomToken{}
	// err := tx.Init(
	// 	&senderKeySet.PrivateKey,
	// 	nil,
	// 	nil,
	// 	0,
	// 	customTokenParamTx,
	// 	listCustomTokens,
	// 	*rpcServer.config.Database,
	// 	metaData,
	// 	hasPrivacy,
	// 	shardIDSender,
	// )
	//=======
	// txTokenVout := transaction.TxTokenVout{
	// 	Value:          amount,
	// 	PaymentAddress: paymentAddress,
	// }
	// newTx := transaction.TxCustomToken{
	// 	TxTokenData: transaction.TxTokenData{
	// 		Type:       transaction.CustomTokenTransfer,
	// 		Amount:     amount,
	// 		PropertyID: propertyID,
	// 		Vins:       []transaction.TxTokenVin{},
	// 		Vouts:      []transaction.TxTokenVout{txTokenVout},
	// 	},
	// }
	// newTx.SetMetadata(metadata.NewSendBackTokenVoteFailMetadata())
	//Create: CustomTokenParamTx
	return nil
}
