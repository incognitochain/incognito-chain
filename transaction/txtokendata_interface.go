package transaction

import (
	"fmt"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"strconv"
)

func (txData *TxPrivacyTokenData) UnmarshalJSON(data []byte) error {
	// For rolling version
	type Alias TxPrivacyTokenData
	temp := &struct {
		TxNormal    []byte
		*Alias
	}{
		Alias: (*Alias)(txData),
	}
	err := json.Unmarshal(data, temp)
	if err != nil {
		Logger.Log.Error("UnmarshalJSON tx", string(data))
		return NewTransactionErr(UnexpectedError, err)
	}

	if txData.TxNormal, err = NewTransactionFromJsonBytes(temp.TxNormal); err != nil {
		Logger.Log.Error(err)
		return err
	}
	return nil
}

type TxPrivacyTokenData struct {
	// TxNormal is the normal transaction, it will never be token transaction
	TxNormal metadata.Transaction
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string

	Type     int    // action type
	Mintable bool   // default false
	Amount   uint64 // init amount
}

func (txData TxPrivacyTokenData) GetPropertyID() common.Hash { return txData.PropertyID }

func (txData *TxPrivacyTokenData) SetPropertyID(propID common.Hash) { txData.PropertyID = propID }

func (txData TxPrivacyTokenData) GetPropertyName() string { return txData.PropertyName }

func (txData *TxPrivacyTokenData) SetPropertyName(propertyName string) { txData.PropertyName = propertyName }

func (txData TxPrivacyTokenData) GetPropertySymbol() string { return txData.PropertySymbol }

func (txData *TxPrivacyTokenData) SetPropertySymbol(propertySymbol string)  { txData.PropertySymbol = propertySymbol }

func (txData TxPrivacyTokenData) GetType() int { return txData.Type }

func (txData *TxPrivacyTokenData) SetType(t int) { txData.Type = t }

func (txData TxPrivacyTokenData) IsMintable() bool { return txData.Mintable }

func (txData *TxPrivacyTokenData) SetMintable(mintable bool) { txData.Mintable = mintable }

func (txData TxPrivacyTokenData) GetAmount() uint64 { return txData.Amount }

func (txData *TxPrivacyTokenData) SetAmount(amount uint64) { txData.Amount = amount }

func (txData TxPrivacyTokenData) String() string {
	record := txData.PropertyName
	record += txData.PropertySymbol
	record += fmt.Sprintf("%d", txData.Amount)
	if txData.TxNormal.GetProof() != nil {
		inputCoins := txData.TxNormal.GetProof().GetInputCoins()
		outputCoins := txData.TxNormal.GetProof().GetOutputCoins()
		for _, out := range outputCoins {
			publicKeyBytes := []byte{}
			if out.GetPublicKey() != nil {
				publicKeyBytes = out.GetPublicKey().ToBytesS()
			}
			record += string(publicKeyBytes)
			record += strconv.FormatUint(out.GetValue(), 10)
		}
		for _, in := range inputCoins {
			publicKeyBytes := []byte{}
			if in.GetPublicKey() != nil {
				publicKeyBytes = in.GetPublicKey().ToBytesS()
			}
			record += string(publicKeyBytes)
			record += strconv.FormatUint(in.GetValue(), 10)
		}
	}
	return record
}

// Hash - return hash of custom token data, be used as Token ID
func (txData TxPrivacyTokenData) Hash() (*common.Hash, error) {
	point := operation.HashToPoint([]byte(txData.String()))
	hash := new(common.Hash)
	err := hash.SetBytes(point.ToBytesS())
	if err != nil {
		return nil, err
	}
	return hash, nil
}

type TxTokenInterface interface {
	// GET/SET FUNCTION
	GetVersion() int8
	SetVersion(int8)
	GetTxFee() uint64
	SetTxFee(uint64)
	GetType() string
	SetType(string)
	GetInfo() []byte
	SetInfo([]byte)
	GetSigPubKey() []byte
	SetSigPubKey([]byte)
	GetSig() []byte
	SetSig([]byte)
	GetProof() privacy.Proof
	SetProof(privacy.Proof)

	GetTxPrivacyTokenData() TxPrivacyTokenData
	SetTxPrivacyTokenData(TxPrivacyTokenData)
	GetTxBase() metadata.Transaction
	SetTxBase(metadata.Transaction)

	// =================== FUNCTIONS THAT GET STUFF AND REQUIRE SOME CODING ===================
	GetTxActualSize() uint64
	GetReceivers() ([][]byte, []uint64)
	GetTransferData() (bool, []byte, uint64, *common.Hash)

	GetReceiverData() ([]*privacy.Point, []*coin.TxRandom, []uint64, error)
	GetTxMintData() (bool, []byte, []byte,  uint64, *common.Hash, error)
	GetTxBurnData(metadata.ChainRetriever, uint64) (bool, []byte, uint64, *common.Hash, error)

	GetMetadataFromVinsTx(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) (metadata.Metadata, error)
	ListSerialNumbersHashH() []common.Hash
	String() string
	Hash() *common.Hash
	CalculateTxValue() uint64

	// =================== FUNCTION THAT CHECK STUFFS  ===================
	CheckTxVersion(int8) bool
	CheckAuthorizedSender([]byte) (bool, error)
	ShouldSignMetaData() bool
	IsSalaryTx() bool
	IsPrivacy() bool
	IsCoinsBurning(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) bool

	// =================== FUNCTIONS THAT VALIDATE STUFFS ===================
	ValidateTxSalary(*statedb.StateDB) (bool, error)
	ValidateTxWithCurrentMempool(metadata.MempoolRetriever) error
	ValidateSanityData(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) (bool, error)
	ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error
	ValidateDoubleSpendWithBlockchain(byte, *statedb.StateDB, *common.Hash) error
	ValidateTxByItself(bool, *statedb.StateDB, *statedb.StateDB, metadata.ChainRetriever, byte, bool, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) (bool, error)
	ValidateType() bool
	ValidateTransaction(bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash, bool, bool) (bool, error)
	VerifyMinerCreatedTxBeforeGettingInBlock([]metadata.Transaction, []int, [][]string, []int, byte, metadata.ChainRetriever, *metadata.AccumulatedValues, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) (bool, error)

	// Init Transaction, the input should be params such as: TxPrivacyInitParams
	Init(interface{}) error
	// Verify the init function above, which verify zero knowledge proof and signatures
	Verify(bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash, bool, bool) (bool, error)
}