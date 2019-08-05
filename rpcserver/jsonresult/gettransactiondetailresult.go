package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"math/big"
)

type TransactionDetail struct {
	BlockHash   string `json:"BlockHash"`
	BlockHeight uint64 `json:"BlockHeight"`
	Index       uint64 `json:"index"`
	ShardID     byte   `json:"shardID"`
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

	Metadata               string `json:"Metadata"`
	CustomTokenData        string `json:"CustomTokenData"`
	PrivacyCustomTokenData string `json:"PrivacyCustomTokenData"`

	IsInMempool bool `json:"IsInMempool"`
	IsInBlock   bool `json:"IsInBlock"`

	Info string `json:"Info"`
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
			in.CoinDetails.Value = input.CoinDetails.Value
			in.CoinDetails.Info = base58.Base58Check{}.Encode(input.CoinDetails.Info, 0x0)
			if input.CoinDetails.CoinCommitment != nil {
				in.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(input.CoinDetails.CoinCommitment.Compress(), 0x0)
			}
			if input.CoinDetails.Randomness != nil {
				in.CoinDetails.Randomness = *input.CoinDetails.Randomness
			}
			if input.CoinDetails.SNDerivator != nil {
				in.CoinDetails.SNDerivator = *input.CoinDetails.SNDerivator
			}
			if input.CoinDetails.SerialNumber != nil {
				in.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(input.CoinDetails.SerialNumber.Compress(), 0x0)
			}
			if input.CoinDetails.PublicKey != nil {
				in.CoinDetails.PublicKey = base58.Base58Check{}.Encode(input.CoinDetails.PublicKey.Compress(), 0x0)
			}
		}
		proofDetail.InputCoins = append(proofDetail.InputCoins, &in)
	}

	for _, output := range proof.GetOutputCoins() {
		out := CoinDetail{
			CoinDetails: Coin{},
		}
		if output.CoinDetails != nil {
			out.CoinDetails.Value = output.CoinDetails.Value
			out.CoinDetails.Info = base58.Base58Check{}.Encode(output.CoinDetails.Info, 0x0)
			if output.CoinDetails.CoinCommitment != nil {
				out.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(output.CoinDetails.CoinCommitment.Compress(), 0x0)
			}
			if output.CoinDetails.Randomness != nil {
				out.CoinDetails.Randomness = *output.CoinDetails.Randomness
			}
			if output.CoinDetails.SNDerivator != nil {
				out.CoinDetails.SNDerivator = *output.CoinDetails.SNDerivator
			}
			if output.CoinDetails.SerialNumber != nil {
				out.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(output.CoinDetails.SerialNumber.Compress(), 0x0)
			}
			if output.CoinDetails.PublicKey != nil {
				out.CoinDetails.PublicKey = base58.Base58Check{}.Encode(output.CoinDetails.PublicKey.Compress(), 0x0)
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
	SNDerivator    big.Int
	SerialNumber   string
	Randomness     big.Int
	Value          uint64
	Info           string
}
