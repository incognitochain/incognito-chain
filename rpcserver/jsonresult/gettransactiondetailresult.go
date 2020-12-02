package jsonresult

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/transaction"
	"time"
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

	IsPrivacy       bool              `json:"IsPrivacy"`
	Proof           *zkp.PaymentProof `json:"Proof"`
	ProofDetail     ProofDetail       `json:"ProofDetail"`
	InputCoinPubKey string            `json:"InputCoinPubKey"`
	SigPubKey       string            `json:"SigPubKey,omitempty"` // 64 bytes
	Sig             string            `json:"Sig,omitempty"`       // 64 bytes

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

func NewTransactionDetail(tx basemeta.Transaction, blockHash *common.Hash, blockHeight uint64, index int, shardID byte) (*TransactionDetail, error) {
	var result *TransactionDetail
	blockHashStr := ""
	if blockHash != nil {
		blockHashStr = blockHash.String()
	}
	switch tx.GetType() {
	case common.TxNormalType, common.TxRewardType, common.TxReturnStakingType:
		{
			tempTx := tx.(*transaction.Tx)
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
			if result.Proof != nil && len(result.Proof.GetInputCoins()) > 0 && result.Proof.GetInputCoins()[0].CoinDetails.GetPublicKey() != nil {
				result.InputCoinPubKey = base58.Base58Check{}.Encode(result.Proof.GetInputCoins()[0].CoinDetails.GetPublicKey().ToBytesS(), common.ZeroByte)
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
			if result.Proof != nil && len(result.Proof.GetInputCoins()) > 0 && result.Proof.GetInputCoins()[0].CoinDetails.GetPublicKey() != nil {
				result.InputCoinPubKey = base58.Base58Check{}.Encode(result.Proof.GetInputCoins()[0].CoinDetails.GetPublicKey().ToBytesS(), common.ZeroByte)
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
	InputCoins  []*CoinDetail
	OutputCoins []*CoinDetail
}

func (proofDetail *ProofDetail) ConvertFromProof(proof *zkp.PaymentProof) {
	proofDetail.InputCoins = make([]*CoinDetail, 0)
	for _, input := range proof.GetInputCoins() {
		in := CoinDetail{
			CoinDetails: Coin{},
		}
		if input.CoinDetails != nil {
			in.CoinDetails.Value = input.CoinDetails.GetValue()
			in.CoinDetails.Info = base58.Base58Check{}.Encode(input.CoinDetails.GetInfo(), 0x0)
			if input.CoinDetails.GetCoinCommitment() != nil {
				in.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(input.CoinDetails.GetCoinCommitment().ToBytesS(), 0x0)
			}
			if input.CoinDetails.GetRandomness() != nil {
				in.CoinDetails.Randomness = *input.CoinDetails.GetRandomness()
			}
			if input.CoinDetails.GetSNDerivator() != nil {
				in.CoinDetails.SNDerivator = *input.CoinDetails.GetSNDerivator()
			}
			if input.CoinDetails.GetSerialNumber() != nil {
				in.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(input.CoinDetails.GetSerialNumber().ToBytesS(), 0x0)
			}
			if input.CoinDetails.GetPublicKey() != nil {
				in.CoinDetails.PublicKey = base58.Base58Check{}.Encode(input.CoinDetails.GetPublicKey().ToBytesS(), 0x0)
			}
		}
		proofDetail.InputCoins = append(proofDetail.InputCoins, &in)
	}

	for _, output := range proof.GetOutputCoins() {
		out := CoinDetail{
			CoinDetails: Coin{},
		}
		if output.CoinDetails != nil {
			out.CoinDetails.Value = output.CoinDetails.GetValue()
			out.CoinDetails.Info = base58.Base58Check{}.Encode(output.CoinDetails.GetInfo(), 0x0)
			if output.CoinDetails.GetCoinCommitment() != nil {
				out.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(output.CoinDetails.GetCoinCommitment().ToBytesS(), 0x0)
			}
			if output.CoinDetails.GetRandomness() != nil {
				out.CoinDetails.Randomness = *output.CoinDetails.GetRandomness()
			}
			if output.CoinDetails.GetSNDerivator() != nil {
				out.CoinDetails.SNDerivator = *output.CoinDetails.GetSNDerivator()
			}
			if output.CoinDetails.GetSerialNumber() != nil {
				out.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(output.CoinDetails.GetSerialNumber().ToBytesS(), 0x0)
			}
			if output.CoinDetails.GetPublicKey() != nil {
				out.CoinDetails.PublicKey = base58.Base58Check{}.Encode(output.CoinDetails.GetPublicKey().ToBytesS(), 0x0)
			}
			if output.CoinDetailsEncrypted != nil {
				out.CoinDetailsEncrypted = base58.Base58Check{}.Encode(output.CoinDetailsEncrypted.Bytes(), 0x0)
			}
		}
		proofDetail.OutputCoins = append(proofDetail.OutputCoins, &out)
	}
}

type CoinDetail struct {
	CoinDetails          Coin
	CoinDetailsEncrypted string
}

type Coin struct {
	PublicKey      string
	CoinCommitment string
	SNDerivator    privacy.Scalar
	SerialNumber   string
	Randomness     privacy.Scalar
	Value          uint64
	Info           string
}
