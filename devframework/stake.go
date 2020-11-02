package devframework

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

const (
	stakeShardAmount   int = 1750000000000
	stakeBeaceonAmount int = stakeShardAmount * 3
)

type StakingTxParam struct {
	SenderPrk   string
	MinerPrk    string
	RewardAddr  string
	StakeShard  bool
	AutoRestake bool
}

type StopStakingParam struct {
	SenderPrk string
	MinerPrk  string
}

func (sim *SimulationEngine) API_CreateTxStaking(stakeMeta StakingTxParam) (*jsonresult.CreateTransactionResult, error) {
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
		stakeMeta.RewardAddr = wl.Base58CheckSerialize(wallet.PaymentAddressType)
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
	minerPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	candidateWallet, err := wallet.Base58CheckDeserialize(minerPayment)
	if err != nil || candidateWallet == nil {
		fmt.Println(stakeMeta.MinerPrk, wl.KeySet.PaymentAddress, minerPayment)
		fmt.Println(err, candidateWallet)
		panic(0)
	}
	burnAddr := sim.bc.GetBurningAddress(sim.bc.BeaconChain.GetFinalViewHeight())
	txResp, err := sim.rpc_createandsendstakingtransaction(stakeMeta.SenderPrk, map[string]interface{}{burnAddr: float64(stakeAmount)}, 1, 0, map[string]interface{}{
		"StakingType":                  float64(stakingType),
		"CandidatePaymentAddress":      minerPayment,
		"PrivateSeed":                  privateSeed,
		"RewardReceiverPaymentAddress": stakeMeta.RewardAddr,
		"AutoReStaking":                stakeMeta.AutoRestake,
	})
	if err != nil {
		return nil, err
	}
	return &txResp, nil
}

func (sim *SimulationEngine) CreateTxStopAutoStake(stopStakeMeta StopStakingParam) (*jsonresult.CreateTransactionResult, error) {
	if stopStakeMeta.MinerPrk == "" {
		stopStakeMeta.MinerPrk = stopStakeMeta.SenderPrk
	}
	wl, err := wallet.Base58CheckDeserialize(stopStakeMeta.MinerPrk)
	if err != nil {
		return nil, err
	}
	privateSeedBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
	minerPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	burnAddr := sim.bc.GetBurningAddress(sim.bc.BeaconChain.GetFinalViewHeight())

	txResp, err := sim.rpc_createandsendstopautostakingtransaction(stopStakeMeta.SenderPrk, map[string]interface{}{burnAddr: float64(0)}, 1, 0, map[string]interface{}{
		"StopAutoStakingType":     float64(127),
		"CandidatePaymentAddress": minerPayment,
		"PrivateSeed":             privateSeed,
	})
	if err != nil {
		return nil, err
	}
	return &txResp, nil
}

func (sim *SimulationEngine) GetRewardAmount(paymentAddress string) (map[string]uint64, error) {
	result, err := sim.rpc_getrewardamount(paymentAddress)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (sim *SimulationEngine) WithdrawReward(privateKey string, paymentAddress string) (*jsonresult.CreateTransactionResult, error) {

	txResp, err := sim.rpc_withdrawreward(privateKey, 0, 0, map[string]interface{}{
		"PaymentAddress": paymentAddress, "TokenID": "0000000000000000000000000000000000000000000000000000000000000004", "Version": 0,
	})
	if err != nil {
		return nil, err
	}
	return &txResp, nil
}
