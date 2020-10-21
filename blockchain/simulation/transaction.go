package main

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

const (
	stakeShardAmount   int = 1750000000000
	stakeBeaceonAmount int = stakeShardAmount * 3
)

func (sim *simInstance) CreateTxStaking(stakeMeta CreateStakingTx) (*jsonresult.CreateTransactionResult, error) {
	stakeAmount := 0
	stakingType := 0
	if stakeMeta.StakeShard {
		stakeAmount = stakeShardAmount
		stakingType = 63
	} else {
		stakeAmount = stakeBeaceonAmount
		stakingType = 64
	}

	if stakeMeta.RewardAddr == "" {
		wl, err := wallet.Base58CheckDeserialize(stakeMeta.SenderPrk)
		if err != nil {
			return nil, err
		}
		stakeMeta.RewardAddr = base58.Base58Check{}.Encode(wl.KeySet.PaymentAddress.Pk, common.ZeroByte)
	}

	if stakeMeta.MinerPrk == "" {
		stakeMeta.MinerPrk = stakeMeta.SenderPrk
	}
	wl, err := wallet.Base58CheckDeserialize(stakeMeta.MinerPrk)
	if err != nil {
		return nil, err
	}
	privateSeedBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
	minerPayment := base58.Base58Check{}.Encode(wl.KeySet.PaymentAddress.Pk, common.ZeroByte)

	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendstakingtransaction",
		"params": []interface{}{stakeMeta.SenderPrk, map[string]int{"12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA": stakeAmount}, -1, 0, map[string]interface{}{
			"StakingType":                  stakingType,
			"CandidatePaymentAddress":      minerPayment,
			"PrivateSeed":                  privateSeed,
			"RewardReceiverPaymentAddress": stakeMeta.RewardAddr,
			"AutoReStaking":                stakeMeta.AutoRestake,
		}},
		"id": 1,
	})
	if err != nil {
		return nil, err
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return nil, err
	}
	txResp := struct {
		Result jsonresult.CreateTransactionResult
	}{}
	err = json.Unmarshal(body, &txResp)
	if err != nil {
		return nil, err
	}
	return &txResp.Result, nil
}
