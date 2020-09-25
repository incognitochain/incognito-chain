package transaction

import (
	"github.com/incognitochain/incognito-chain/transaction/metadata"
	"github.com/incognitochain/incognito-chain/transaction/privacy"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver1"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver2"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

var Logger = &utils.Logger

func BuildCoinBaseTxByCoinID(params *BuildCoinBaseTxByCoinIDParams) (metadata.Transaction, error) {
	paymentInfo := &privacy.PaymentInfo{
		PaymentAddress: *params.payToAddress,
		Amount: params.amount,
	}
	otaCoin, err := privacy.NewCoinFromPaymentInfo(paymentInfo)
	params.meta.SetSharedRandom(otaCoin.GetSharedRandom().ToBytesS())

	if err != nil {
		utils.Logger.Log.Errorf("Cannot get new coin from amount and receiver")
		return nil, err
	}
	switch params.txType {
	case NormalCoinType:
		tx := new(TxVersion2)
		err = tx.InitTxSalary(otaCoin, params.payByPrivateKey, params.transactionStateDB, params.meta)
		return tx, err
	case CustomTokenPrivacyType:
		var propertyID [common.HashSize]byte
		copy(propertyID[:], params.coinID[:])
		propID := common.Hash(propertyID)
		tx := new(TxTokenVersion2)
		err = tx.InitTxTokenSalary(otaCoin, params.payByPrivateKey, params.transactionStateDB,
			params.meta, &propID, params.coinName)
		return tx, err
	}
	return nil, nil
}

// Used to parse json
type txJsonDataVersion struct {
	Version int8 `json:"Version"`
	Type    string
}

// For PRV and the Fee inside TokenTx
func NewTransactionFromJsonBytes(data []byte) (metadata.Transaction, error) {
	//fmt.Println(string(data))
	txJsonVersion := new(txJsonDataVersion)
	if err := json.Unmarshal(data, txJsonVersion); err != nil {
		return nil, err
	}
	if txJsonVersion.Type == common.TxConversionType || txJsonVersion.Type == common.TxTokenConversionType {
		if txJsonVersion.Version == int8(utils.TxConversionVersion12Number) {
				tx := new(TxVersion2)
				if err := json.Unmarshal(data, tx); err != nil {
					return nil, err
				}
				return tx, nil
		} else {
			return nil, errors.New("Cannot new txConversion from jsonBytes, type is incorrect.")
		}
	} else {
		switch txJsonVersion.Version {
		case int8(TxVersion1Number), int8(TxVersion0Number):
			tx := new(TxVersion1)
			if err := json.Unmarshal(data, tx); err != nil {
				return nil, err
			}
			return tx, nil
		case int8(TxVersion2Number):
			tx := new(TxVersion2)
			if err := json.Unmarshal(data, tx); err != nil {
				return nil, err
			}
			return tx, nil
		default:
			return nil, errors.New("Cannot new tx from jsonBytes, version is incorrect")
		}
	}
}

// Return token transaction from bytes
func NewTransactionTokenFromJsonBytes(data []byte) (tx_generic.TransactionToken, error) {
	txJsonVersion := new(txJsonDataVersion)
	if err := json.Unmarshal(data, txJsonVersion); err != nil {
		return nil, err
	}

	if txJsonVersion.Type == common.TxTokenConversionType {
		if txJsonVersion.Version == TxConversionVersion12Number {
			tx := new(TxTokenVersion2)
			if err := json.Unmarshal(data, tx); err != nil {
				return nil, err
			}
			return tx, nil
		} else {
			return nil, errors.New("Cannot new txTokenConversion from jsonBytes, version is incorrect")
		}
	} else {
		switch txJsonVersion.Version {
		case int8(TxVersion1Number), TxVersion0Number:
			tx := new(TxTokenVersion1)
			if err := json.Unmarshal(data, tx); err != nil {
				return nil, err
			}
			return tx, nil
		case int8(TxVersion2Number):
			tx := new(TxTokenVersion2)
			if err := json.Unmarshal(data, tx); err != nil {
				return nil, err
			}
			return tx, nil
		default:
			return nil, errors.New("Cannot new txToken from bytes because version is incorrect")
		}
	}
}

func (txToken *TxTokenBase) UnmarshalJSON(data []byte) error {
	var err error
	if txToken.Tx, err = NewTransactionFromJsonBytes(data); err != nil {
		return err
	}
	temp := &struct {
		TxTokenData TxTokenData `json:"TxTokenPrivacyData"`
	}{}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	TxTokenDataJson, err := json.MarshalIndent(temp.TxTokenData, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(UnexpectedError, err)
	}
	err = json.Unmarshal(TxTokenDataJson, &txToken.TxTokenData)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}

	// TODO: hotfix, remove when fixed this issue
	if txToken.Tx.GetMetadata() != nil && txToken.Tx.GetMetadata().GetType() == 81 {
		if txToken.TxTokenData.Amount == 37772966455153490 {
			txToken.TxTokenData.Amount = 37772966455153487
		}
	}
	return nil
}

func (txData *TxTokenData) UnmarshalJSON(data []byte) error {
	type Alias TxTokenData
	temp := &struct {
		TxNormal json.RawMessage
		*Alias
	}{
		Alias: (*Alias)(txData),
	}
	err := json.Unmarshal(data, temp)
	if err != nil {
		Logger.Log.Error("UnmarshalJSON tx", string(data))
		return utils.NewTransactionErr(utils.UnexpectedError, err)
	}

	if txData.TxNormal, err = NewTransactionFromJsonBytes(temp.TxNormal); err != nil {
		Logger.Log.Error(err)
		return err
	}
	return nil
}

type BuildCoinBaseTxByCoinIDParams struct {
	payToAddress       *privacy.PaymentAddress
	amount             uint64
	txRandom           *privacy.TxRandom
	payByPrivateKey    *privacy.PrivateKey
	transactionStateDB *statedb.StateDB
	bridgeStateDB      *statedb.StateDB
	meta               *metadata.WithDrawRewardResponse
	coinID             common.Hash
	txType             int
	coinName           string
	shardID            byte
}

func NewBuildCoinBaseTxByCoinIDParams(payToAddress *privacy.PaymentAddress,
	amount uint64,
	payByPrivateKey *privacy.PrivateKey,
	stateDB *statedb.StateDB,
	meta *metadata.WithDrawRewardResponse,
	coinID common.Hash,
	txType int,
	coinName string,
	shardID byte,
	bridgeStateDB *statedb.StateDB) *BuildCoinBaseTxByCoinIDParams {
	params := &BuildCoinBaseTxByCoinIDParams{
		transactionStateDB: stateDB,
		bridgeStateDB:      bridgeStateDB,
		shardID:            shardID,
		meta:               meta,
		amount:             amount,
		coinID:             coinID,
		coinName:           coinName,
		payByPrivateKey:    payByPrivateKey,
		payToAddress:       payToAddress,
		txType:             txType,
	}
	return params
}

func NewTransactionFromParams(params *tx_generic.TxPrivacyInitParams) (metadata.Transaction, error) {
	inputCoins := params.inputCoins
	ver, err := tx_generic.GetTxVersionFromCoins(inputCoins)
	if err != nil {
		return nil, err
	}

	if ver == 1 {
		return new(TxVersion1), nil
	} else if ver == 2 {
		return new(TxVersion2), nil
	}
	return nil, errors.New("Cannot create new transaction from params, ver is wrong")
}

func NewTransactionTokenFromParams(params *tx_generic.TxTokenParams) (TransactionToken, error) {
	inputCoins := params.inputCoin
	ver, err := tx_generic.GetTxVersionFromCoins(inputCoins)
	if err != nil {
		return nil, err
	}

	if ver == 1 {
		return new(TxTokenVersion1), nil
	} else if ver == 2 {
		return new(TxTokenVersion2), nil
	}
	return nil, errors.New("Something is wrong when NewTransactionFromParams")
}

func GetTxTokenDataFromTransaction(tx metadata.Transaction) *TxTokenData {
	if tx.GetType() != common.TxCustomTokenPrivacyType && tx.GetType() != common.TxTokenConversionType {
		return nil
	}
	if tx.GetVersion() == utils.TxVersion1Number {
		txTemp := tx.(*TxTokenVersion1)
		return &txTemp.TxTokenData
	} else if tx.GetVersion() == utils.TxVersion2Number || tx.GetVersion() == utils.TxConversionVersion12Number {
		txTemp := tx.(*TxTokenVersion2)
		return &txTemp.TxTokenData
	}
	return nil
}