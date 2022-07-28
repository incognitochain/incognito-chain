package rpcclient

import (
	"fmt"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/testsuite/account"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
)

const defaultTxFee = float64(5)

type RPCClient struct {
	Client ClientInterface
}

func NewRPCClient(client ClientInterface) *RPCClient {
	rpc := &RPCClient{
		Client: client,
	}
	return rpc
}

func mapUintToInterface(m map[string]uint64) map[string]interface{} {
	mfl := make(map[string]interface{})
	for a, v := range m {
		mfl[a] = float64(v)
	}
	return mfl
}

func (r *RPCClient) API_SubmitKey(privateKey string) (bool, error) {
	b, err := r.Client.SubmitKey(privateKey)
	time.Sleep(time.Millisecond)
	return b, err
}

func (r *RPCClient) API_CreateAndSendTokenInitTransaction(privateKey, tokenName, tokenSymbol string, amount uint64) (jsonresult.CreateTransactionTokenResult, error) {
	param := PdexV3InitTokenParam{
		privateKey,
		tokenName,
		tokenSymbol,
		amount,
	}
	res, err := r.Client.CreateAndSendTokenInitTransaction(param)
	time.Sleep(time.Millisecond)
	return res, err
}

func (r *RPCClient) API_PdexV3AddLiquididty(privatekey, nftID, tokenID, poolPairID, pairHash, contributedAmount, amplifier string) error {
	param := PdexV3AddLiquidityParam{
		nftID,
		tokenID,
		poolPairID,
		pairHash,
		contributedAmount,
		amplifier,
	}
	err := r.Client.Pdexv3_TxAddLiquidity(privatekey, param)
	time.Sleep(time.Millisecond)
	return err
}
func (r *RPCClient) API_Pdexv3AddOrder(
	privatekey, poolPairID, tokenToSell, tokenToBuy, nftID, sellAmount, min string,
) error {
	param := PdexV3AddOrderParam{
		poolPairID, tokenToSell, tokenToBuy, nftID, sellAmount, min,
	}
	err := r.Client.Pdexv3_TxAddOrder(privatekey, param)
	time.Sleep(time.Millisecond)
	return err
}

func (r *RPCClient) API_Pdexv3Trade(privatekey string, param PdexV3TradeParam) error {
	err := r.Client.Pdexv3_TxAddTrade(privatekey, param)
	time.Sleep(time.Millisecond)
	return err
}

func (r *RPCClient) API_MintNFT(privateKey string) error {
	err := r.Client.Pdexv3_TxMintNft(privateKey)
	time.Sleep(time.Millisecond)
	return err
}

//func (r *RPCClient) API_GetPDESV3tate(beaconheight uint64) (*jsonresult.Pdexv3State, error) {
//	res, err := r.Client.Pdexv3_GetState(map[string]interface{}{"BeaconHeight": float64(beaconheight), "Filter": map[string]interface{}{}})
//	time.Sleep(time.Millisecond)
//	return res, err
//}

func (r *RPCClient) API_CreateConvertCoinVer1ToVer2Transaction(privateKey string) error {
	err := r.Client.CreateConvertCoinVer1ToVer2Transaction(privateKey)
	time.Sleep(time.Millisecond)
	return err
}

func (r *RPCClient) API_SendTxPRV(privateKey string, receivers map[string]uint64, fee int64, privacy bool) (*jsonresult.CreateTransactionResult, error) {
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}

	result, err := r.Client.CreateAndSendTransaction(privateKey, mapUintToInterface(receivers), float64(fee), privacyTx)
	time.Sleep(time.Millisecond)
	return &result, err
}

func (r *RPCClient) API_SendTxCreateCustomToken(privateKey string, receiverPaymentAddress string, privacy bool, tokenName string, tokenSymbol string, tokenAmount uint64) (*jsonresult.CreateTransactionTokenResult, error) {
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.CreateAndSendPrivacyCustomTokenTransaction(privateKey, nil, defaultTxFee, privacyTx, map[string]interface{}{
		"Privacy":     true,
		"TokenID":     "",
		"TokenName":   tokenName,
		"TokenSymbol": tokenSymbol,
		"TokenFee":    float64(0),
		"TokenTxType": float64(0),
		"TokenAmount": float64(tokenAmount),
		"TokenReceivers": map[string]interface{}{
			receiverPaymentAddress: float64(tokenAmount),
		},
	}, "", privacyTx)
	return &result, err
}

func (r *RPCClient) API_SendTxCustomToken(privateKey string, tokenID string, receivers map[string]uint64, fee int64, privacy bool) (*jsonresult.CreateTransactionTokenResult, error) {
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.CreateAndSendPrivacyCustomTokenTransaction(privateKey, mapUintToInterface(receivers), float64(fee), privacyTx, map[string]interface{}{
		"Privacy":        true,
		"TokenID":        tokenID,
		"TokenName":      "",
		"TokenSymbol":    "",
		"TokenFee":       float64(0),
		"TokenTxType":    float64(1),
		"TokenAmount":    float64(0),
		"TokenReceivers": receivers,
	}, "", privacyTx)
	return &result, err
}

func (r *RPCClient) API_SendTxWithWithdrawalReqV2(privateKey string, receivers map[string]uint64, fee int64, privacy bool, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.CreateAndSendTxWithWithdrawalReqV2(privateKey, mapUintToInterface(receivers), float64(fee), privacyTx, reqInfo)
	return &result, err
}
func (r *RPCClient) API_SendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]uint64, fee int64, privacy bool, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey, mapUintToInterface(receivers), float64(fee), privacyTx, reqInfo)
	return &result, err
}
func (r *RPCClient) API_SendTxWithPTokenTradeReq(privateKey string, receivers map[string]uint64, fee int64, privacy bool, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionTokenResult, error) {
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.CreateAndSendTxWithPTokenTradeReq(privateKey, mapUintToInterface(receivers), float64(fee), privacyTx, reqInfo, "", 0)
	return &result, err
}
func (r *RPCClient) API_SendTxWithPTokenCrossPoolTradeReq(acount account.Account, tokenID string, buyTokenID string, sellAmount int, miniumBuyAmount int) (*jsonresult.CreateTransactionTokenResult, error) {
	burnAddr, err := r.Client.GetBurningAddress(float64(0))
	if err != nil {
		return nil, err
	}
	reqInfo := map[string]interface{}{
		"Privacy":     true,
		"TokenID":     tokenID,
		"TokenTxType": float64(1),
		"TokenName":   "",
		"TokenSymbol": "",
		"TokenAmount": strconv.Itoa(sellAmount),
		"TokenReceivers": map[string]interface{}{
			burnAddr: strconv.Itoa(sellAmount),
		},
		"TokenFee":            "0",
		"TokenIDToBuyStr":     buyTokenID,
		"TokenIDToSellStr":    tokenID,
		"SellAmount":          strconv.Itoa(sellAmount),
		"MinAcceptableAmount": strconv.Itoa(miniumBuyAmount),
		"TradingFee":          "1",
		"TraderAddressStr":    acount.PaymentAddress,
		"TxVersion":           float64(1),
	}
	result, err := r.Client.CreateAndSendTxWithPTokenCrossPoolTradeReq(acount.PrivateKey, map[string]interface{}{burnAddr: "1"}, -1, 0, reqInfo, "", 0)
	return &result, err
}
func (r *RPCClient) API_SendTxWithPRVTradeReq(privateKey string, receivers map[string]uint64, fee int64, privacy bool, reqInfo map[string]interface{}) (*jsonresult.CreateTransactionResult, error) {
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.CreateAndSendTxWithPRVTradeReq(privateKey, mapUintToInterface(receivers), float64(fee), privacyTx, reqInfo)
	return &result, err
}
func (r *RPCClient) API_SendTxWithPRVCrossPoolTradeReq(account account.Account, buyTokenID string, sellAmount int, miniumBuyAmount int) (*jsonresult.CreateTransactionResult, error) {
	burnAddr, err := r.Client.GetBurningAddress(float64(0))
	if err != nil {
		return nil, err
	}
	reqInfo := map[string]interface{}{
		"TokenIDToBuyStr":     buyTokenID,
		"TokenIDToSellStr":    "0000000000000000000000000000000000000000000000000000000000000004",
		"SellAmount":          strconv.Itoa(sellAmount),
		"MinAcceptableAmount": strconv.Itoa(miniumBuyAmount),
		"TradingFee":          "0",
		"TraderAddressStr":    account.PaymentAddress,
		"TxVersion":           float64(1),
	}
	result, err := r.Client.CreateAndSendTxWithPRVCrossPoolTradeReq(account.PrivateKey, map[string]interface{}{
		burnAddr: strconv.Itoa(sellAmount),
	}, -1, -1, reqInfo)
	return &result, err
}
func (r *RPCClient) API_SendTxWithPTokenContributionV2(account account.Account, tokenID string, tokenAmount int, pairID string) (*jsonresult.CreateTransactionTokenResult, error) {
	burnAddr, err := r.Client.GetBurningAddress(float64(1))
	if err != nil {
		return nil, err
	}
	reqInfo := map[string]interface{}{
		"Privacy":     true,
		"TokenID":     tokenID,
		"TokenTxType": float64(1),
		"TokenName":   "",
		"TokenSymbol": "",
		"TokenAmount": strconv.Itoa(tokenAmount),
		"TokenReceivers": map[string]interface{}{
			burnAddr: strconv.Itoa(tokenAmount),
		},
		"TokenFee":              "0",
		"PDEContributionPairID": pairID,
		"ContributorAddressStr": account.PaymentAddress,
		"ContributedAmount":     strconv.Itoa(tokenAmount),
		"TokenIDStr":            tokenID,
		"TxVersion":             float64(1),
	}
	result, err := r.Client.CreateAndSendTxWithPTokenContributionV2(account.PrivateKey, nil, defaultTxFee, 0, reqInfo, "", 0)
	return &result, err
}
func (r *RPCClient) API_SendTxWithPRVContributionV2(account account.Account, prvAmount int, pairID string) (*jsonresult.CreateTransactionResult, error) {
	burnAddr, err := r.Client.GetBurningAddress(float64(0))
	if err != nil {
		return nil, err
	}
	reqInfo := map[string]interface{}{
		"PDEContributionPairID": pairID,
		"ContributorAddressStr": account.PaymentAddress,
		"ContributedAmount":     strconv.Itoa(prvAmount),
		"TokenIDStr":            "0000000000000000000000000000000000000000000000000000000000000004",
	}
	result, err := r.Client.CreateAndSendTxWithPRVContributionV2(account.PrivateKey, map[string]interface{}{burnAddr: strconv.Itoa(prvAmount)}, -1, 0, reqInfo)
	return &result, err
}
func (r *RPCClient) API_GetPDEState(beaconHeight float64) (jsonresult.CurrentPDEState, error) {
	result, err := r.Client.GetPDEState(map[string]interface{}{"BeaconHeight": beaconHeight})
	return result, err
}

func (sim *RPCClient) SendPRV(args ...interface{}) (string, error) {
	var sender string
	var receivers = make(map[string]uint64)
	for i, arg := range args {
		if i == 0 {
			sender = arg.(account.Account).PrivateKey
		} else {
			switch arg.(type) {
			default:
				if i%2 == 1 {
					amount, ok := args[i+1].(uint64)
					if !ok {
						amountF64 := args[i+1].(float64)
						amount = uint64(amountF64)
					}
					receivers[arg.(account.Account).PaymentAddress] = amount
				}
			}
		}
	}

	res, err := sim.API_SendTxPRV(sender, receivers, -1, true)
	if err != nil {
		fmt.Println(err)
	}
	return res.TxID, nil
}

func (sim *RPCClient) ShowBalance(acc account.Account) {
	res, err := sim.API_GetBalance(acc)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}

func (r *RPCClient) API_GetBeaconBestState() (jsonresult.GetBeaconBestState, error) {
	result, err := r.Client.GetBeaconBestState()
	return result, err
}

func (r *RPCClient) API_GetShardBestState(sid int) (jsonresult.GetShardBestState, error) {
	result, err := r.Client.GetShardBestState(sid)
	return result, err
}

func (r *RPCClient) API_GetTransactionHash(h string) (jsonresult.TransactionDetail, error) {
	result, err := r.Client.GetTransactionByHash(h)
	return *result, err
}

func (r *RPCClient) API_GetPrivacyCustomToken(h string) (*jsonresult.GetCustomToken, error) {
	result, err := r.Client.GetPrivacyCustomToken(h)
	return result, err
}

func (r *RPCClient) API_GetBalance(acc account.Account) (map[string]uint64, error) {
	tokenList := make(map[string]uint64)
	prv, _ := r.Client.GetBalanceByPrivateKey(acc.PrivateKey)
	tokenList["PRV"] = prv

	tokenBL, _ := r.Client.GetListPrivacyCustomTokenBalance(acc.PrivateKey)
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
	StakerPrk string
	MinerPrk  string
}

func (r *RPCClient) Stake(acc account.Account) (*jsonresult.CreateTransactionResult, error) {
	stake1 := StakingTxParam{
		BurnAddr:    "12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA",
		StakerPrk:   acc.PrivateKey,
		StakeShard:  true,
		AutoRestake: true,
	}
	return r.API_SendTxStaking(stake1)
}

func (r *RPCClient) API_SendTxStaking(stakeMeta StakingTxParam) (*jsonresult.CreateTransactionResult, error) {
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

	//fmt.Println(burnAddr)
	//fmt.Println(stakingType)
	//fmt.Println(stakeAmount)
	//fmt.Println(stakeMeta.StakerPrk)
	//fmt.Println(minerPayment)
	//fmt.Println(privateSeed)
	//fmt.Println(stakeMeta.RewardAddr)
	//fmt.Println(stakeMeta.AutoRestake)

	txResp, err := r.Client.CreateAndSendStakingTransaction(stakeMeta.StakerPrk, map[string]interface{}{burnAddr: float64(stakeAmount)}, 1, 0, map[string]interface{}{
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

func (r *RPCClient) API_SendTxStopAutoStake(stopStakeMeta StopStakingParam) (*jsonresult.CreateTransactionResult, error) {
	if stopStakeMeta.MinerPrk == "" {
		stopStakeMeta.MinerPrk = stopStakeMeta.StakerPrk
	}
	wl, err := wallet.Base58CheckDeserialize(stopStakeMeta.MinerPrk)
	if err != nil {
		return nil, err
	}
	privateSeedBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
	minerPayment := wl.Base58CheckSerialize(wallet.PaymentAddressType)

	burnAddr := stopStakeMeta.BurnAddr

	txResp, err := r.Client.CreateAndSendStopAutoStakingTransaction(stopStakeMeta.StakerPrk, map[string]interface{}{burnAddr: float64(0)}, 1, 0, map[string]interface{}{
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
	rs, err := r.Client.GetRewardAmount(paymentAddress)
	if err != nil {
		return nil, err
	}
	for token, amount := range rs {
		result[token] = float64(amount) / 1e9
	}
	return result, nil
}

func (r *RPCClient) API_SendTxWithdrawReward(privateKey string, paymentAddress string) (*jsonresult.CreateTransactionResult, error) {

	txResp, err := r.Client.WithdrawReward(privateKey, nil, 10, 0, map[string]interface{}{
		"PaymentAddress": paymentAddress, "TokenID": "0000000000000000000000000000000000000000000000000000000000000004", "Version": 0,
	})
	if err != nil {
		return nil, err
	}
	return &txResp, nil
}

func (r *RPCClient) API_GetPublicKeyRole(miningPublicKey string) (role string, chainID int) {
	result, err := r.Client.GetPublicKeyRole("bls:"+miningPublicKey, true)
	if err != nil {
		return "", -2
	}
	switch result.(*struct {
		Role    int
		ShardID int
	}).Role {
	case -1:
		role = ""
		break
	case 0:
		role = "waiting"
		break
	case 1:
		role = "pending"
		break
	case 2:
		role = "committee"
		break
	}
	return role, result.(*struct {
		Role    int
		ShardID int
	}).ShardID
}

func (r *RPCClient) API_GetBlockChainInfo() (*jsonresult.GetBlockChainInfoResult, error) {
	result, err := r.Client.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RPCClient) API_GetCandidateList() (*jsonresult.CandidateListsResult, error) {
	result, err := r.Client.GetCandidateList()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RPCClient) API_GetCommitteeList() (*jsonresult.CommitteeListsResult, error) {
	result, err := r.Client.GetCommitteeList()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RPCClient) API_GetBlockHash(chainID int, height uint64) ([]common.Hash, error) {
	result, err := r.Client.GetBlockHash(float64(chainID), float64(height))
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (r *RPCClient) API_RetrieveShardBlock(hash string, verbosity string) (*jsonresult.GetShardBlockResult, error) {
	// var result *jsonresult.GetShardBlockResult
	result, err := r.Client.RetrieveBlock(hash, verbosity)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (r *RPCClient) API_RetrieveShardBlockByHeight(shardID byte, height uint64, verbosity string) ([]*jsonresult.GetShardBlockResult, error) {
	result := []*jsonresult.GetShardBlockResult{}
	result, err := r.Client.RetrieveBlockByHeight(float64(shardID), float64(height), verbosity)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (r *RPCClient) API_RetrieveBeaconBlock(hash string) (*jsonresult.GetBeaconBlockResult, error) {
	var result *jsonresult.GetBeaconBlockResult
	result, err := r.Client.RetrieveBeaconBlock(hash)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (r *RPCClient) API_RetrieveBeaconBlockByHeight(height uint64) ([]*jsonresult.GetBeaconBlockResult, error) {
	result := []*jsonresult.GetBeaconBlockResult{}
	result, err := r.Client.RetrieveBeaconBlockByHeight(float64(height))
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (r *RPCClient) API_GetRewardAmountByEpoch(shardID byte, epoch uint64) (uint64, error) {
	return r.Client.GetRewardAmountByEpoch(float64(shardID), float64(epoch))
}
func (r *RPCClient) API_DefragmentAccountPRV(privateKey string, maxValue uint64, fee uint64, privacy bool) (*jsonresult.CreateTransactionResult, error) {
	var result jsonresult.CreateTransactionResult
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.DefragmentAccount(privateKey, float64(maxValue), float64(fee), privacyTx)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func (r *RPCClient) API_DefragmentAccountToken(privateKey string, tokenID string, fee uint64, privacy bool) (*jsonresult.CreateTransactionTokenResult, error) {
	var result jsonresult.CreateTransactionTokenResult
	privacyTx := float64(0)
	if privacy {
		privacyTx = 1
	}
	result, err := r.Client.DefragmentAccountToken(privateKey, map[string]interface{}{}, float64(fee), privacyTx, map[string]interface{}{
		"Privacy":     true,
		"TokenID":     tokenID,
		"TokenName":   "",
		"TokenSymbol": "",
		"TokenTxType": 1,
		"TokenAmount": 0,
		"TokenFee":    0,
	}, "", privacyTx)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
