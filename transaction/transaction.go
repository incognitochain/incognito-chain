package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/tx_ver1"
	txtypes "github.com/incognitochain/incognito-chain/transaction/types"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

// Logger is the logger instance for this package
var Logger = &utils.Logger

type TxVersion1 = tx_ver1.Tx
type Tx = txtypes.Tx
type TxTokenVersion1 = tx_ver1.TxToken
type TxToken = txtypes.TxToken
type TransactionToken = tx_generic.TransactionToken
type TokenParam = tx_generic.TokenParam
type TxTokenParams = tx_generic.TxTokenParams
type TxTokenData = tx_generic.TxTokenData
type TxSigPubKeyVer2 = txtypes.SigPubKey

// BuildCoinBaseTxByCoinID is used to create a salary transaction.
// It must take its own defined parameter struct.
// Deprecated: this is not used in v2 code
func BuildCoinBaseTxByCoinID(params *BuildCoinBaseTxByCoinIDParams) (metadata.Transaction, error) {
	p := privacy.NewCoinParams().FromPaymentInfo(&privacy.PaymentInfo{
		PaymentAddress: *params.payToAddress,
		Amount:         params.amount,
	})
	p.CoinPrivacyType = privacy.CoinPrivacyTypeMint
	otaCoin, err := privacy.NewCoinFromPaymentInfo(p)
	if err != nil {
		utils.Logger.Log.Errorf("Cannot get new coin from amount and receiver")
		return nil, err
	}
	params.meta.SetSharedRandom(otaCoin.GetSharedRandom().ToBytesS())

	switch params.txType {
	case utils.NormalCoinType:
		tx := new(Tx)
		err = tx.InitTxSalary(otaCoin, params.payByPrivateKey, params.transactionStateDB, params.meta)
		return tx, err
	case utils.CustomTokenPrivacyType:
		var propertyID [common.HashSize]byte
		copy(propertyID[:], params.coinID[:])
		propID := common.Hash(propertyID)
		tx := new(TxToken)
		err = tx.InitTxTokenSalary(otaCoin, params.payByPrivateKey, params.transactionStateDB,
			params.meta, &propID, params.coinName)
		return tx, err
	}
	return nil, nil
}

// TxSalaryOutputParams is a helper struct for "mint"-type transactions.
// It first tries to create a TX in version 2; if that fails, it falls back to the older version.
// The receiver is defined by either the ReceiverAddress field (for non-privacy use), or the fields PublicKey & TxRandom combined.
type TxSalaryOutputParams struct {
	Amount          uint64
	ReceiverAddress *privacy.PaymentAddress
	PublicKey       *privacy.Point
	TxRandom        *privacy.TxRandom
	TokenID         *common.Hash
	Info            []byte
	Type            string
}

func (pr TxSalaryOutputParams) getVersion() int {
	// can create mint TX iff the version is 1 or 2
	if pr.PublicKey != nil && pr.TxRandom != nil {
		return 2
	}
	if _, err := metadata.AssertPaymentAddressAndTxVersion(*pr.ReceiverAddress, 2); err == nil {
		return 2
	}
	if _, err := metadata.AssertPaymentAddressAndTxVersion(*pr.ReceiverAddress, 1); err != nil {
		Logger.Log.Errorf("AssertPaymentAddressAndTxVersion error: %v", err)
		return 0
	}
	return 1
}

func (pr TxSalaryOutputParams) generateOutputCoin() (*privacy.CoinV2, error) {
	var err error
	var outputCoin *privacy.CoinV2
	isPRV := (pr.TokenID == nil) || (*pr.TokenID == common.PRVCoinID)
	if pr.ReceiverAddress == nil {
		outputCoin = privacy.NewCoinFromAmountAndTxRandomBytes(pr.Amount, pr.PublicKey, pr.TxRandom, pr.Info)
	} else {
		p := privacy.NewCoinParams().FromPaymentInfo(&privacy.PaymentInfo{Amount: pr.Amount, PaymentAddress: *pr.ReceiverAddress})
		p.CoinPrivacyType = privacy.CoinPrivacyTypeMint
		outputCoin, err = privacy.NewCoinFromPaymentInfo(p)
		if err != nil {
			return nil, err
		}
	}
	// for salary TX, tokenID is never blinded; so when making coin for token transactions, we set an unblinded asset tag
	if !isPRV {
		err := outputCoin.SetPlainTokenID(pr.TokenID)
		if err != nil {
			return nil, err
		}
	}
	return outputCoin, err
}

// BuildTXSalary is called from the parameter struct to actually create a "mint" transaction. It takes parameters:
//
//   - private key : to sign the TX
//   - db pointer : to get other coins for the ring signature
//   - metadata maker (optional): a closure that returns a metadata object (this allows making the metadata based on the output coin)
func (pr TxSalaryOutputParams) BuildTxSalary(privateKey *privacy.PrivateKey, stateDB *statedb.StateDB, mdMaker func(privacy.Coin) metadata.Metadata) (metadata.Transaction, error) {
	var res metadata.Transaction
	isPRV := (pr.TokenID == nil) || (*pr.TokenID == common.PRVCoinID)
	switch pr.getVersion() {
	case 1:
		if isPRV {
			temp := new(TxVersion1)
			err := temp.InitTxSalary(pr.Amount, pr.ReceiverAddress, privateKey, stateDB, mdMaker(nil))
			if err != nil {
				Logger.Log.Errorf("Cannot build Tx Salary v1. Err: %v", err)
				return nil, err
			}
			// handle for "return staking" TX; need to set the transaction type
			if pr.Type == common.TxReturnStakingType {
				temp.Type = common.TxReturnStakingType
			}
			res = temp
		} else {
			shardID := common.GetShardIDFromLastByte(pr.ReceiverAddress.Pk[len(pr.ReceiverAddress.Pk)-1])
			tokenParams := &tx_generic.TokenParam{
				PropertyID:     pr.TokenID.String(),
				PropertyName:   "",
				PropertySymbol: "",
				Amount:         pr.Amount,
				TokenTxType:    CustomTokenInit,
				Receiver:       []*privacy.PaymentInfo{{PaymentAddress: *pr.ReceiverAddress, Amount: pr.Amount}},
				TokenInput:     []privacy.PlainCoin{},
				Mintable:       true,
			}
			temp := new(TxTokenVersion1)
			err := temp.Init(
				tx_generic.NewTxTokenParams(privateKey,
					[]*privacy.PaymentInfo{},
					nil,
					0,
					tokenParams,
					stateDB,
					mdMaker(nil),
					false,
					false,
					shardID,
					[]byte{},
					nil,
				))
			if err != nil {
				Logger.Log.Errorf("Cannot build Tx Token Salary v1. Err: %v", err)
				return nil, err
			}
			res = temp
		}
	case 2:
		otaCoin, err := pr.generateOutputCoin()
		if err != nil {
			Logger.Log.Errorf("Cannot create coin for TX salary. Err: %v", err)
			return nil, err
		}
		md := mdMaker(otaCoin)

		if isPRV {
			temp := new(Tx)
			err = temp.InitTxSalary(otaCoin, privateKey, stateDB, md)
			if err != nil {
				Logger.Log.Errorf("Cannot build Tx Salary v2. Err: %v", err)
				return nil, err
			}
			// handle for "return staking" TX; need to set the transaction type
			if pr.Type == common.TxReturnStakingType {
				temp.Type = common.TxReturnStakingType
				temp.Sig = nil
				temp.SigPubKey = nil
				if temp.Sig, temp.SigPubKey, err = tx_generic.SignNoPrivacy(privateKey, temp.Hash()[:]); err != nil {
					return nil, utils.NewTransactionErr(utils.SignTxError, err)
				}
				// valid, err = temp.ValidateTxSalary(stateDB)
				// Logger.Log.Debugf("Verify Salary TX for return staking : %v, %v", valid, err)
			}
			res = temp
		} else {
			temp := new(TxToken)
			err = temp.InitTxTokenSalary(otaCoin, privateKey, stateDB, md, pr.TokenID, "")
			if err != nil {
				Logger.Log.Errorf("Cannot build Tx Token Salary v2. Err: %v", err)
				return nil, err
			}
			res = temp
		}
	default:
		Logger.Log.Errorf("Cannot build Tx Salary - invalid parameters %v", pr)
		return nil, fmt.Errorf("cannot build Tx Salary - invalid parameters")
	}
	jsb, _ := json.MarshalIndent(res, "", "\t")
	Logger.Log.Infof("Built new salary transaction %s", string(jsb))
	return res, nil
}

// Used to parse json
type txJsonDataVersion struct {
	Version int8 `json:"Version"`
	Type    string
}

// NewTransactionFromJsonBytes parses a transaction from raw JSON, assuming it is a PRV transfer.
// This is a legacy function; it will be replaced by DeserializeTransactionJSON.
func NewTransactionFromJsonBytes(data []byte) (metadata.Transaction, error) {
	choices, err := DeserializeTransactionJSON(data)
	if err != nil {
		return nil, err
	}
	if choices.Version1 != nil {
		return choices.Version1, nil
	}
	if choices.Version2 != nil {
		return choices.Version2, nil
	}
	return nil, fmt.Errorf("cannot parse TX as PRV transaction")
}

// NewTransactionTokenFromJsonBytes parses a transaction from raw JSON, assuming it is a pToken transfer.
// This is a legacy function; it will be replaced by DeserializeTransactionJSON.
func NewTransactionTokenFromJsonBytes(data []byte) (tx_generic.TransactionToken, error) {
	choices, err := DeserializeTransactionJSON(data)
	if err != nil {
		return nil, err
	}
	if choices.TokenVersion1 != nil {
		return choices.TokenVersion1, nil
	}
	if choices.TokenVersion2 != nil {
		return choices.TokenVersion2, nil
	}
	return nil, fmt.Errorf("cannot parse TX as token transaction")
}

// TxChoice is a helper struct for parsing transactions of all types from JSON.
// After parsing succeeds, one of its fields will have the TX object; others will be nil.
// This can be used to assert the transaction type.
type TxChoice struct {
	Version1      *TxVersion1      `json:"TxVersion1,omitempty"`
	TokenVersion1 *TxTokenVersion1 `json:"TxTokenVersion1,omitempty"`
	Version2      *Tx              `json:"TxVersion2,omitempty"`
	TokenVersion2 *TxToken         `json:"TxTokenVersion2,omitempty"`
}

// ToTx returns a generic transaction from a TxChoice object.
// Use this when the underlying TX type is irrelevant.
func (ch *TxChoice) ToTx() metadata.Transaction {
	if ch == nil {
		return nil
	}
	// `choice` struct only ever contains 1 non-nil field
	if ch.Version1 != nil {
		return ch.Version1
	}
	if ch.Version2 != nil {
		return ch.Version2
	}
	if ch.TokenVersion1 != nil {
		return ch.TokenVersion1
	}
	if ch.TokenVersion2 != nil {
		return ch.TokenVersion2
	}
	return nil
}

// DeserializeTransactionJSON parses a transaction from raw JSON into a TxChoice object.
// It covers all transaction types.
func DeserializeTransactionJSON(data []byte) (*TxChoice, error) {
	result := &TxChoice{}
	holder := make(map[string]interface{})
	err := json.Unmarshal(data, &holder)
	if err != nil {
		return nil, err
	}
	_, isTokenTx := holder["TxTokenPrivacyData"]
	_, hasVersionOutside := holder["Version"]
	var verHolder txJsonDataVersion
	// unmarshalling error here corresponds to the `else` block below
	_ = json.Unmarshal(data, &verHolder)
	if hasVersionOutside {
		switch verHolder.Version {
		case utils.TxVersion1Number:
			if isTokenTx {
				// token ver 1
				result.TokenVersion1 = &TxTokenVersion1{}
				err := json.Unmarshal(data, result.TokenVersion1)
				return result, err
			}
			// tx ver 1
			result.Version1 = &TxVersion1{}
			err := json.Unmarshal(data, result.Version1)
			return result, err
		case utils.TxVersion2Number, utils.TxVersion3Number: // the same as utils.TxConversionVersion12Number
			if isTokenTx {
				// rejected
				return nil, fmt.Errorf("error unmarshalling TX from JSON : misplaced version")
			}
			// tx ver 2
			result.Version2 = &Tx{}
			err := json.Unmarshal(data, result.Version2)
			return result, err
		default:
			return nil, fmt.Errorf("error unmarshalling TX from JSON : wrong version of %d", verHolder.Version)
		}
	} else {
		if isTokenTx {
			// token ver 2
			result.TokenVersion2 = &TxToken{}
			err := json.Unmarshal(data, result.TokenVersion2)
			return result, err
		}
		return nil, fmt.Errorf("error unmarshalling TX from JSON")
	}

}

// BuildCoinBaseTxByCoinIDParams defines the parameters for BuildCoinBaseTxByCoinID
type BuildCoinBaseTxByCoinIDParams struct {
	payToAddress       *privacy.PaymentAddress
	amount             uint64
	txRandom           *privacy.TxRandom //nolint:structcheck,unused
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

// NewTransactionFromParams is a helper that returns a new transaction whose version matches the parameter object.
// The result is a PRV-transfer transaction that only has the version ready, nothing else.
// One needs to call ".Init(params)" after this to get a valid TX.
func NewTransactionFromParams(params *tx_generic.TxPrivacyInitParams) (metadata.Transaction, error) {
	inputCoins := params.InputCoins
	ver, err := tx_generic.GetTxVersionFromCoins(inputCoins)
	if err != nil {
		return nil, err
	}

	if ver == 1 {
		return new(TxVersion1), nil
	} else if ver == 2 {
		return new(Tx), nil
	}
	return nil, fmt.Errorf("cannot create new transaction from params, ver is wrong")
}

// NewTransactionTokenFromParams is a helper that returns a new transaction whose version matches the parameter object.
// The result is a pToken transaction that only has the version ready, nothing else.
// One needs to call ".Init(params)" after this to get a valid TX.
func NewTransactionTokenFromParams(params *tx_generic.TxTokenParams) (tx_generic.TransactionToken, error) {
	inputCoins := params.InputCoin
	ver, err := tx_generic.GetTxVersionFromCoins(inputCoins)
	if err != nil {
		return nil, err
	}

	switch ver {
	case 1:
		return new(TxTokenVersion1), nil
	case 2:
		return new(TxToken), nil
	default:
		return nil, fmt.Errorf("invalid version for NewTransactionFromParams")
	}
}

// GetTxTokenDataFromTransaction is an alternative getter for the TokenData field.
// It takes the most generic TX interface and casts it to a token transaction.
// Upon any error, it returns nil.
func GetTxTokenDataFromTransaction(tx metadata.Transaction) *tx_generic.TxTokenData {
	if tx.GetType() != common.TxCustomTokenPrivacyType && tx.GetType() != common.TxTokenConversionType {
		return nil
	}
	switch tx_specific := tx.(type) {
	case *TxTokenVersion1:
		// txTemp := tx.(*TxTokenVersion1)
		return &tx_specific.TxTokenData
	// } else if tx.GetVersion() == utils.TxVersion2Number || tx.GetVersion() == utils.TxConversionVersion12Number {
	case *TxToken:
		// txTemp := tx.(*TxTokenVersion2)
		res := tx_specific.GetTxTokenData()
		return &res
	default:
		return nil
	}
}

// GetFullBurnData is a helper that indicates whether or not a TX is burning any coin.
// Returned values beyond the first will contain the coins (PRV or pToken) being burnt and its token ID.
func GetFullBurnData(tx metadata.Transaction) (bool, coin.Coin, coin.Coin, *common.Hash, error) {

	switch tx.GetType() {
	case common.TxNormalType:
		isBurned, burnPrv, _, err := tx.GetTxBurnData()
		if err != nil || !isBurned {
			return false, nil, nil, nil, err
		}
		return true, burnPrv, nil, nil, err
	case common.TxCustomTokenPrivacyType:
		if txTmp, ok := tx.(TransactionToken); ok {
			isBurnToken, burnToken, burnedTokenID, err1 := txTmp.GetTxBurnData()
			isBurnPrv, burnPrv, _, err2 := txTmp.GetTxBase().GetTxBurnData()

			if err1 != nil && err2 != nil {
				return false, nil, nil, nil, fmt.Errorf("%v and %v", err1, err2)
			}
			return isBurnPrv || isBurnToken, burnToken, burnPrv, burnedTokenID, nil
		}
		return false, nil, nil, nil, fmt.Errorf("tx is not tp")
	default:
		return false, nil, nil, nil, nil
	}
}

type OTADeclaration = metadata.OTADeclaration

func GetOTADeclarationsFromTx(tx metadata.Transaction) []OTADeclaration {
	return tx_generic.GetOTADeclarationsFromTx(tx)
}

func WithPrivacy(vE metadata.ValidationEnviroment) *tx_generic.ValidationEnv {
	return tx_generic.WithPrivacy(vE)
}

func WithNoPrivacy(vE metadata.ValidationEnviroment) *tx_generic.ValidationEnv {
	return tx_generic.WithNoPrivacy(vE)
}

func WithShardID(vE metadata.ValidationEnviroment, sID int) *tx_generic.ValidationEnv {
	return tx_generic.WithShardID(vE, sID)
}

func WithShardHeight(vE metadata.ValidationEnviroment, sHeight uint64) *tx_generic.ValidationEnv {
	return tx_generic.WithShardHeight(vE, sHeight)
}

func WithBeaconHeight(vE metadata.ValidationEnviroment, bcHeight uint64) *tx_generic.ValidationEnv {
	return tx_generic.WithBeaconHeight(vE, bcHeight)
}

func WithConfirmedTime(vE metadata.ValidationEnviroment, confirmedTime int64) *tx_generic.ValidationEnv {
	return tx_generic.WithConfirmedTime(vE, confirmedTime)
}

func WithType(vE metadata.ValidationEnviroment, t string) *tx_generic.ValidationEnv {
	return tx_generic.WithType(vE, t)
}

func WithAct(vE metadata.ValidationEnviroment, act int) *tx_generic.ValidationEnv {
	return tx_generic.WithAct(vE, act)
}
