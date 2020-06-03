package jsonresult

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/transaction"
)

type TransactionDetail struct {
	BlockHash   string `json:"BlockHash"`
	BlockHeight uint64 `json:"BlockHeight"`
	TxSize      uint64 `json:"TxSize"`
	Index       uint64 `json:"Index"`
	ShardID     byte   `json:"ShardID"`
	Hash        string `json:"Hash"`
	Version     int8   `json:"Version"`
	Type        string `json:"Type"` // Transaction type
	LockTime    string `json:"LockTime"`
	Fee         uint64 `json:"Fee"` // Fee applies: always consant
	Image       string `json:"Image"`

	IsPrivacy       bool          `json:"IsPrivacy"`
	Proof           privacy.Proof `json:"Proof"`
	ProofDetail     ProofDetail   `json:"ProofDetail"`
	InputCoinPubKey string        `json:"InputCoinPubKey"`
	SigPubKey       string        `json:"SigPubKey,omitempty"` // 64 bytes
	Sig             string        `json:"Sig,omitempty"`       // 64 bytes

	Metadata                      string      `json:"Metadata"`
	CustomTokenData               string      `json:"CustomTokenData"`
	PrivacyCustomTokenID          string      `json:"PrivacyCustomTokenID"`
	PrivacyCustomTokenName        string      `json:"PrivacyCustomTokenName"`
	PrivacyCustomTokenSymbol      string      `json:"PrivacyCustomTokenSymbol"`
	PrivacyCustomTokenData        string      `json:"PrivacyCustomTokenData"`
	PrivacyCustomTokenProofDetail ProofDetail `json:"PrivacyCustomTokenProofDetail"`
	PrivacyCustomTokenIsPrivacy   bool        `json:"PrivacyCustomTokenIsPrivacy"`
	PrivacyCustomTokenFee         uint64      `json:"PrivacyCustomTokenFee"`

	IsInMempool bool `json:"IsInMempool"`
	IsInBlock   bool `json:"IsInBlock"`

	Info string `json:"Info"`
}

func NewTransactionDetail(tx metadata.Transaction, blockHash *common.Hash, blockHeight uint64, index int, shardID byte) (*TransactionDetail, error) {
	var result *TransactionDetail
	blockHashStr := ""
	if blockHash != nil {
		blockHashStr = blockHash.String()
	}
	switch tx.GetType() {
	case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
		{
			tempTx := tx.(*transaction.TxBase)
			result = &TransactionDetail{
				BlockHash:   blockHashStr,
				BlockHeight: blockHeight,
				Index:       uint64(index),
				TxSize:      tx.GetTxActualSize(),
				ShardID:     shardID,
				Hash:        tx.Hash().String(),
				Version:     tempTx.Version,
				Type:        tempTx.Type,
				LockTime:    time.Unix(tempTx.LockTime, 0).Format(common.DateOutputFormat),
				Fee:         tempTx.Fee,
				IsPrivacy:   tempTx.IsPrivacy(),
				Proof:       tempTx.Proof,
				SigPubKey:   base58.Base58Check{}.Encode(tempTx.SigPubKey, 0x0),
				Sig:         base58.Base58Check{}.Encode(tempTx.Sig, 0x0),
				Info:        string(tempTx.Info),
			}
			inputCoins := result.Proof.GetInputCoins()
			if result.Proof != nil && len(inputCoins) > 0 && inputCoins[0].GetPublicKey() != nil {
				result.InputCoinPubKey = base58.Base58Check{}.Encode(inputCoins[0].GetPublicKey().ToBytesS(), common.ZeroByte)
			}
			if tempTx.Metadata != nil {
				metaData, _ := json.MarshalIndent(tempTx.Metadata, "", "\t")
				result.Metadata = string(metaData)
			}
			if result.Proof != nil {
				result.ProofDetail.ConvertFromProof(result.Proof)
			}
		}
	case common.TxCustomTokenPrivacyType:
		{
			tempTx := tx.(*transaction.TxCustomTokenPrivacy)
			result = &TransactionDetail{
				BlockHash:                blockHashStr,
				BlockHeight:              blockHeight,
				Index:                    uint64(index),
				TxSize:                   tempTx.GetTxActualSize(),
				ShardID:                  shardID,
				Hash:                     tx.Hash().String(),
				Version:                  tempTx.Version,
				Type:                     tempTx.Type,
				LockTime:                 time.Unix(tempTx.LockTime, 0).Format(common.DateOutputFormat),
				Fee:                      tempTx.Fee,
				Proof:                    tempTx.Proof,
				SigPubKey:                base58.Base58Check{}.Encode(tempTx.SigPubKey, 0x0),
				Sig:                      base58.Base58Check{}.Encode(tempTx.Sig, 0x0),
				Info:                     string(tempTx.Info),
				IsPrivacy:                tempTx.IsPrivacy(),
				PrivacyCustomTokenSymbol: tempTx.TxPrivacyTokenData.PropertySymbol,
				PrivacyCustomTokenName:   tempTx.TxPrivacyTokenData.PropertyName,
				PrivacyCustomTokenID:     tempTx.TxPrivacyTokenData.PropertyID.String(),
				PrivacyCustomTokenFee:    tempTx.TxPrivacyTokenData.TxNormal.Fee,
			}
			inputCoins := result.Proof.GetInputCoins()
			if result.Proof != nil && len(inputCoins) > 0 && inputCoins[0].GetPublicKey() != nil {
				result.InputCoinPubKey = base58.Base58Check{}.Encode(inputCoins[0].GetPublicKey().ToBytesS(), common.ZeroByte)
			}
			tokenData, _ := json.MarshalIndent(tempTx.TxPrivacyTokenData, "", "\t")
			result.PrivacyCustomTokenData = string(tokenData)
			if tempTx.Metadata != nil {
				metaData, _ := json.MarshalIndent(tempTx.Metadata, "", "\t")
				result.Metadata = string(metaData)
			}
			if result.Proof != nil {
				result.ProofDetail.ConvertFromProof(result.Proof)
			}
			result.PrivacyCustomTokenIsPrivacy = tempTx.TxPrivacyTokenData.TxNormal.IsPrivacy()
			if tempTx.TxPrivacyTokenData.TxNormal.Proof != nil {
				result.PrivacyCustomTokenProofDetail.ConvertFromProof(tempTx.TxPrivacyTokenData.TxNormal.Proof)
			}
		}
	default:
		{
			return nil, errors.New("Tx type is invalid")
		}
	}
	return result, nil
}

type ProofDetail struct {
	InputCoins  []CoinRPC
	OutputCoins []CoinRPC
}

func (proofDetail *ProofDetail) ConvertFromProof(proof privacy.Proof) {
	inputCoins := proof.GetInputCoins()
	outputCoins := proof.GetOutputCoins()

	proofDetail.InputCoins = make([]CoinRPC, len(inputCoins))
	for i, input := range inputCoins {
		proofDetail.InputCoins[i] = ParseCoinRPCInput(input)
	}

	proofDetail.OutputCoins = make([]CoinRPC, len(outputCoins))
	for i, output := range outputCoins {
		proofDetail.OutputCoins[i] = ParseCoinRPCOutput(output)
	}
}

func ParseCoinRPCInput(inputCoin coin.PlainCoin) CoinRPC {
	var coinrpc CoinRPC
	if inputCoin.GetVersion() == 1 {
		coinrpc = new(CoinRPCV1)
	} else {
		coinrpc = new(CoinRPCV2)
	}
	return coinrpc.SetInputCoin(inputCoin)
}

func ParseCoinRPCOutput(outputCoin coin.Coin) CoinRPC {
	var coinrpc CoinRPC
	if outputCoin.GetVersion() == 1 {
		coinrpc = new(CoinRPCV1)
	} else {
		coinrpc = new(CoinRPCV2)
	}
	return coinrpc.SetOutputCoin(outputCoin)
}

type CoinRPC interface {
	SetInputCoin(coin.PlainCoin) CoinRPC
	SetOutputCoin(coin.Coin) CoinRPC
}

func EncodeBase58Check(b []byte) string {
	return base58.Base58Check{}.Encode(b, 0x0)
}

func OperationPointPtrToBase58(point *operation.Point) string {
	if point != nil {
		return EncodeBase58Check(point.ToBytesS())
	}
	return ""
}

func OperationScalarPtrToScalar(scalar *operation.Scalar) operation.Scalar {
	if scalar != nil {
		return *scalar
	}
	return *new(operation.Scalar).FromUint64(0)
}

func (c *CoinRPCV1) SetInputCoin(inputCoin coin.PlainCoin) CoinRPC {
	coinv1 := inputCoin.(*coin.PlainCoinV1)

	c.Version = coinv1.GetVersion()
	c.PublicKey = OperationPointPtrToBase58(coinv1.GetPublicKey())
	c.Commitment = OperationPointPtrToBase58(coinv1.GetCommitment())
	c.SNDerivator = OperationScalarPtrToScalar(coinv1.GetSNDerivator())
	c.KeyImage = OperationPointPtrToBase58(coinv1.GetKeyImage())
	c.Randomness = OperationScalarPtrToScalar(coinv1.GetRandomness())
	c.Value = coinv1.GetValue()
	c.Info = EncodeBase58Check(coinv1.GetInfo())
	return c
}

func (c *CoinRPCV1) SetOutputCoin(inputCoin coin.Coin) CoinRPC {
	coinv1 := inputCoin.(*coin.CoinV1)

	c.Version = coinv1.GetVersion()
	c.PublicKey = OperationPointPtrToBase58(coinv1.GetPublicKey())
	c.Commitment = OperationPointPtrToBase58(coinv1.GetCommitment())
	c.SNDerivator = OperationScalarPtrToScalar(coinv1.GetSNDerivator())
	c.KeyImage = OperationPointPtrToBase58(coinv1.GetKeyImage())
	c.Randomness = OperationScalarPtrToScalar(coinv1.GetRandomness())
	c.Value = coinv1.CoinDetails.GetValue()
	c.Info = EncodeBase58Check(coinv1.GetInfo())
	c.CoinDetailsEncrypted = EncodeBase58Check(coinv1.CoinDetailsEncrypted.Bytes())
	return c
}

func (c *CoinRPCV2) SetInputCoin(inputCoin coin.PlainCoin) CoinRPC {
	return c.SetOutputCoin(inputCoin.(coin.Coin))
}

func (c *CoinRPCV2) SetOutputCoin(inputCoin coin.Coin) CoinRPC {
	coinv2 := inputCoin.(*coin.CoinV2)

	c.Version = coinv2.GetVersion()
	c.Index = coinv2.GetIndex()
	c.Info = EncodeBase58Check(coinv2.GetInfo())
	c.PublicKey = OperationPointPtrToBase58(coinv2.GetPublicKey())
	c.Commitment = OperationPointPtrToBase58(coinv2.GetCommitment())
	c.KeyImage = OperationPointPtrToBase58(coinv2.GetKeyImage())
	c.TxRandom = OperationPointPtrToBase58(coinv2.GetTxRandomPoint())
	c.Amount = OperationScalarPtrToScalar(coinv2.GetAmount())
	c.Randomness = OperationScalarPtrToScalar(coinv2.GetRandomness())
	return c
}

type CoinRPCV1 struct {
	Version              uint8
	PublicKey            string
	Commitment           string
	SNDerivator          privacy.Scalar
	KeyImage             string
	Randomness           privacy.Scalar
	Value                uint64
	Info                 string
	CoinDetailsEncrypted string
}

type CoinRPCV2 struct {
	Version    uint8
	Index      uint32
	Info       string
	PublicKey  string
	Commitment string
	KeyImage   string
	TxRandom   string

	Amount 		 operation.Scalar
	Randomness   operation.Scalar
}
