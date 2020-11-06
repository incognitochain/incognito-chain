package rpcclient

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

type ClientInterface interface {
	RPC_createandsendtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (jsonresult.CreateTransactionResult, error)
	RPC_getbalancebyprivatekey(privateKey string) (uint64, error)
	RPC_getlistprivacycustomtokenbalance(privateKey string) (jsonresult.ListCustomTokenBalance, error)
	RPC_getrewardamount(paymentaddress string) (map[string]uint64, error)
	RPC_withdrawreward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendstakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendstopautostakingtransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopstakeinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_getcommitteelist(empty string) (jsonresult.CommitteeListsResult, error)
	RPC_createandsendprivacycustomtokentransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithwithdrawalreqv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithpdefeewithdrawalreq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithptokentradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithptokencrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithprvtradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithprvcrosspooltradereq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_createandsendtxwithptokencontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	RPC_createandsendtxwithprvcontributionv2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqinfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	RPC_getpdestate(data map[string]interface{}) (jsonresult.CurrentPDEState, error)
}

type RPCClient struct {
	client ClientInterface
}

func NewRPCClient(client ClientInterface) *RPCClient {
	rpc := &RPCClient{
		client: client,
	}
	return rpc
}

func (r *RPCClient) API_CreateAndSendTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtransaction(privateKey, receivers, fee, privacy)
	return &result, err
}

func (r *RPCClient) API_CreateAndSendPrivacyCustomTokenTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}, p1 string, pPrivacy float64) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendprivacycustomtokentransaction(privateKey, receivers, fee, privacy, tokenInfo, p1, pPrivacy)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithWithdrawalReqV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithwithdrawalreqv2(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithpdefeewithdrawalreq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithPTokenTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendtxwithptokentradereq(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithPTokenCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendtxwithptokencrosspooltradereq(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithPRVTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithprvtradereq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithPRVCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithprvcrosspooltradereq(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithPTokenContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	result, err := r.client.RPC_createandsendtxwithptokencontributionv2(privateKey, receivers, fee, privacy, reqInfo, "", 0)
	return &result, err
}
func (r *RPCClient) API_CreateAndSendTxWithPRVContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	result, err := r.client.RPC_createandsendtxwithprvcontributionv2(privateKey, receivers, fee, privacy, reqInfo)
	return &result, err
}
func (r *RPCClient) API_GetPDEState(beaconHeight float64) (jsonresult.CurrentPDEState, error) {
	result, err := r.client.RPC_getpdestate(map[string]interface{}{"BeaconHeight": beaconHeight})
	return result, err
}

func (r *RPCClient) API_GetBalance(privateKey string) (map[string]uint64, error) {
	tokenList := make(map[string]uint64)
	prv, _ := r.client.RPC_getbalancebyprivatekey(privateKey)
	tokenList["PRV"] = prv

	tokenBL, _ := r.client.RPC_getlistprivacycustomtokenbalance(privateKey)
	for _, token := range tokenBL.ListCustomTokenBalance {
		tokenList[token.TokenID] = token.Amount

	}
	return tokenList, nil
}

const (
	stakeShardAmount   int = 1750000000000
	stakeBeaceonAmount int = stakeShardAmount * 3
)

type StakingTxParam struct {
	CommitteeKey *incognitokey.CommitteePublicKey
	BurnAddr     string
	StakerPrk    string
	MinerPrk     string
	RewardAddr   string
	StakeShard   bool
	AutoRestake  bool
	Name         string
}

type StopStakingParam struct {
	BurnAddr  string
	SenderPrk string
	MinerPrk  string
}

func (r *RPCClient) API_CreateAndSendStakingTransaction(stakeMeta StakingTxParam) (*jsonresult.CreateTransactionResult, error) {
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
		wl, err := wallet.Base58CheckDeserialize(stakeMeta.StakerPrk)
		if err != nil {
			return nil, err
		}
		stakeMeta.RewardAddr = wl.Base58CheckSerialize(wallet.PaymentAddressType)
	}

	if stakeMeta.MinerPrk == "" {
		stakeMeta.MinerPrk = stakeMeta.StakerPrk
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
	burnAddr := stakeMeta.BurnAddr
	txResp, err := r.client.RPC_createandsendstakingtransaction(stakeMeta.StakerPrk, map[string]interface{}{burnAddr: float64(stakeAmount)}, 1, 0, map[string]interface{}{
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

func (r *RPCClient) API_CreateTxStopAutoStake(stopStakeMeta StopStakingParam) (*jsonresult.CreateTransactionResult, error) {
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

	burnAddr := stopStakeMeta.BurnAddr

	txResp, err := r.client.RPC_createandsendstopautostakingtransaction(stopStakeMeta.SenderPrk, map[string]interface{}{burnAddr: float64(0)}, 1, 0, map[string]interface{}{
		"StopAutoStakingType":     float64(127),
		"CandidatePaymentAddress": minerPayment,
		"PrivateSeed":             privateSeed,
	})
	if err != nil {
		return nil, err
	}
	return &txResp, nil
}

func (r *RPCClient) API_GetRewardAmount(paymentAddress string) (map[string]float64, error) {
	result := make(map[string]float64)
	rs, err := r.client.RPC_getrewardamount(paymentAddress)
	if err != nil {
		return nil, err
	}
	for token, amount := range rs {
		result[token] = float64(amount) / 1e9
	}
	return result, nil
}

func (r *RPCClient) API_WithdrawReward(privateKey string, paymentAddress string) (*jsonresult.CreateTransactionResult, error) {

	txResp, err := r.client.RPC_withdrawreward(privateKey, nil, 0, 0, map[string]interface{}{
		"PaymentAddress": paymentAddress, "TokenID": "0000000000000000000000000000000000000000000000000000000000000004", "Version": 0,
	})
	if err != nil {
		return nil, err
	}
	return &txResp, nil
}
