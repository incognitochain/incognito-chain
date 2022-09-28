package rpcclient //This file is auto generated. Please do not change if you dont know what you are doing

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type PdexV3InitTokenParam struct {
	PrivateKey  string `json:"PrivateKey"`
	TokenName   string `json:"TokenName"`
	TokenSymbol string `json:"TokenSymbol"`
	Amount      uint64 `json:"Amount"`
}

type PdexV3AddLiquidityParam struct {
	NftID             string
	TokenID           string
	PoolPairID        string
	PairHash          string
	ContributedAmount string
	Amplifier         string
}

type PdexV3Params struct {
	DefaultFeeRateBPS               string            `json:"DefaultFeeRateBPS"`
	FeeRateBPS                      map[string]string `json:"FeeRateBPS"`
	PRVDiscountPercent              string            `json:"PRVDiscountPercent"`
	TradingProtocolFeePercent       string            `json:"TradingProtocolFeePercent"`
	TradingStakingPoolRewardPercent string            `json:"TradingStakingPoolRewardPercent"`
	PDEXRewardPoolPairsShare        map[string]string `json:"PDEXRewardPoolPairsShare"`
	StakingPoolsShare               map[string]string `json:"StakingPoolsShare"`
	StakingRewardTokens             []string          `json:"StakingRewardTokens"`
	MintNftRequireAmount            string            `json:"MintNftRequireAmount"`
	MaxOrdersPerNft                 string            `json:"MaxOrdersPerNft"`
}

type PdexV3AddOrderParam struct {
	PoolPairID          string `json:"PoolPairID"`
	TokenToSell         string `json:"TokenToSell"`
	TokenToBuy          string `json:"TokenToBuy"`
	NftID               string `json:"NftID"`
	SellAmount          string `json:"SellAmount"`
	MinAcceptableAmount string `json:"MinAcceptableAmount"`
}
type PdexV3TradeParam struct {
	TradePath           []string `json:"TradePath"`
	TokenToSell         string   `json:"TokenToSell"`
	TokenToBuy          string   `json:"TokenToBuy"`
	SellAmount          string   `json:"SellAmount"`
	MinAcceptableAmount string   `json:"MinAcceptableAmount"`
	TradingFee          uint64   `json:"TradingFee"`
	FeeInPRV            bool     `json:"FeeInPRV"`
}

type ClientInterface interface {
	GetBalanceByPrivateKey(privateKey string) (uint64, error)
	GetListPrivacyCustomTokenBalance(privateKey string) (jsonresult.ListCustomTokenBalance, error)
	GetRewardAmount(paymentAddress string) (map[string]uint64, error)
	WithdrawReward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendStopAutoStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendUnStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error)
	CreateAndSendReDelegateTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, redelegateInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error)
	CreateAndSendTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (jsonresult.CreateTransactionResult, error)
	CreateAndSendPrivacyCustomTokenTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	CreateAndSendTxWithWithdrawalReqV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendTxWithPTokenTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	CreateAndSendTxWithPTokenCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	CreateAndSendTxWithPRVTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendTxWithPRVCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendTxWithPTokenContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)
	CreateAndSendTxWithPRVContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	GetPDEState(data map[string]interface{}) (jsonresult.CurrentPDEState, error)
	GetBeaconBestState() (jsonresult.GetBeaconBestState, error)
	GetShardBestState(sid int) (jsonresult.GetShardBestState, error)
	GetTransactionByHash(transactionHash string) (*jsonresult.TransactionDetail, error)
	GetPrivacyCustomToken(tokenStr string) (*jsonresult.GetCustomToken, error)
	GetBurningAddress(beaconHeight float64) (string, error)
	GetPublicKeyRole(publicKey string, detail bool) (interface{}, error)
	GetBlockChainInfo() (*jsonresult.GetBlockChainInfoResult, error)
	GetCandidateList() (*jsonresult.CandidateListsResult, error)
	GetCommitteeList() (*jsonresult.CommitteeListsResult, error)
	GetBlockHash(chainID float64, height float64) ([]common.Hash, error)
	RetrieveBlock(hash string, verbosity string) (*jsonresult.GetShardBlockResult, error)
	RetrieveBlockByHeight(shardID float64, height float64, verbosity string) ([]*jsonresult.GetShardBlockResult, error)
	RetrieveBeaconBlock(hash string) (*jsonresult.GetBeaconBlockResult, error)
	RetrieveBeaconBlockByHeight(height float64) ([]*jsonresult.GetBeaconBlockResult, error)
	GetRewardAmountByEpoch(shard float64, epoch float64) (uint64, error)
	DefragmentAccount(privateKey string, maxValue float64, fee float64, privacy float64) (jsonresult.CreateTransactionResult, error)
	DefragmentAccountToken(privateKey string, receiver map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (jsonresult.CreateTransactionTokenResult, error)

	SubmitKey(privateKey string) (bool, error)
	CreateConvertCoinVer1ToVer2Transaction(privateKey string) error
	SendFinishSync(mining, cpk string, sid float64) error
	GetMempoolInfo() (res *jsonresult.GetMempoolInfo, err error)
	CreateAndSendTXShieldingRequest(privateKey string, incAddr string, tokenID string, proof string) (res jsonresult.CreateTransactionResult, err error)
	GetPortalShieldingRequestStatus(tx string) (res *metadata.PortalShieldingRequestStatus, err error)
	CreateAndSendTxWithPortalV4UnshieldRequest(privatekey string, tokenID string, amount string, paymentAddress string, remoteAddress string) (res jsonresult.CreateTransactionTokenResult, err error)
	GetPortalUnshieldRequestStatus(tx string) (res *metadata.PortalUnshieldRequestStatus, err error)
	//GetPortalV4State(beaconheight string) (rpcserver.CurrentPortalState, error)
	//GetPortalSignedRawTransaction(batchID string) (rpcserver.GetSignedTxResult, error)

	//pdex v3
	CreateAndSendTokenInitTransaction(param PdexV3InitTokenParam) (jsonresult.CreateTransactionTokenResult, error)
	Pdexv3_TxMintNft(privatekeys string) error
	//Pdexv3_GetState(data map[string]interface{}) (*jsonresult.Pdexv3State, error)
	Pdexv3_TxAddLiquidity(privatekey string, param PdexV3AddLiquidityParam) error
	Pdexv3_TxWithdrawLiquidity(privatekey, poolairID, nftID, shareAmount string) error
	Pdexv3_TxModifyParams(privatekey string, newParams PdexV3Params)
	Pdexv3_TxStake(privatekey, stakingPoolID, nftID, amount string) error
	Pdexv3_TxUnstake(privatekey, stakingPoolID, nftID, amount string) error
	Pdexv3_TxAddTrade(privatekey string, param PdexV3TradeParam) error
	Pdexv3_TxAddOrder(privatekey string, params PdexV3AddOrderParam) error
}
