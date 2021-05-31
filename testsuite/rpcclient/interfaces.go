package rpcclient //This file is auto generated. Please do not change if you dont know what you are doing

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type ClientInterface interface {
	GetBalanceByPrivateKey(privateKey string) (uint64, error)
	GetListPrivacyCustomTokenBalance(privateKey string) (jsonresult.ListCustomTokenBalance, error)
	GetRewardAmount(paymentAddress string) (map[string]uint64, error)
	WithdrawReward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
	CreateAndSendStopAutoStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (jsonresult.CreateTransactionResult, error)
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

	SubmitKey(privateKey string) error
	CreateConvertCoinVer1ToVer2Transaction(privateKey string) error
	GetMempoolInfo() (res *jsonresult.GetMempoolInfo, err error)
	CreateAndSendTXShieldingRequest(privateKey string, incAddr string, tokenID string, proof string) (res jsonresult.CreateTransactionResult, err error)
	GetPortalShieldingRequestStatus(tx string) (res *metadata.PortalShieldingRequestStatus, err error)
}
