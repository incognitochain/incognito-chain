package devframework

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type LocalRPCClient struct {
	rpcServer *rpcserver.RpcServer
}

func (r *LocalRPCClient) GetMempoolInfo() (res *jsonresult.GetMempoolInfo, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getmempoolinfo"]
	resI, rpcERR := c(httpServer, []interface{}{}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.GetMempoolInfo), nil
}

func (r *LocalRPCClient) SubmitKey(privateKey string) (err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.LimitedHttpHandler["submitkey"]
	_, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	return nil
}

func (r *LocalRPCClient) CreateConvertCoinVer1ToVer2Transaction(privateKey string) (err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createconvertcoinver1tover2transaction"]
	_, rpcERR := c(httpServer, []interface{}{privateKey, float64(-1)}, nil)
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	return nil
}

func (r *LocalRPCClient) GetBalanceByPrivateKey(privateKey string) (res uint64, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.LimitedHttpHandler["getbalancebyprivatekey"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(uint64), nil
}
func (r *LocalRPCClient) GetListPrivacyCustomTokenBalance(privateKey string) (res jsonresult.ListCustomTokenBalance, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getlistprivacycustomtokenbalance"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.ListCustomTokenBalance), nil
}
func (r *LocalRPCClient) GetRewardAmount(paymentAddress string) (res map[string]uint64, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getrewardamount"]
	resI, rpcERR := c(httpServer, []interface{}{paymentAddress}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(map[string]uint64), nil
}
func (r *LocalRPCClient) WithdrawReward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["withdrawreward"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, amount, privacy, info}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendstakingtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, stakeInfo}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendStopAutoStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendstopautostakingtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, stopStakeInfo}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendPrivacyCustomTokenTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendprivacycustomtokentransaction"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, tokenInfo, p1, pPrivacy}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithWithdrawalReqV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithwithdrawalreqv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithpdefeewithdrawalreq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithPTokenTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokentradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo, p1, pPrivacy}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithPTokenCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokencrosspooltradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo, p1, pPrivacy}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithPRVTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvtradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithPRVCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvcrosspooltradereq"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithPTokenContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithptokencontributionv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo, p1, pPrivacy}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult), nil
}
func (r *LocalRPCClient) CreateAndSendTxWithPRVContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["createandsendtxwithprvcontributionv2"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receivers, fee, privacy, reqInfo}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) GetPDEState(data map[string]interface{}) (res jsonresult.CurrentPDEState, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getpdestate"]
	resI, rpcERR := c(httpServer, []interface{}{data}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CurrentPDEState), nil
}
func (r *LocalRPCClient) GetBeaconBestState() (res jsonresult.GetBeaconBestState, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getbeaconbeststate"]
	resI, rpcERR := c(httpServer, []interface{}{}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.GetBeaconBestState), nil
}
func (r *LocalRPCClient) GetShardBestState(sid int) (res jsonresult.GetShardBestState, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getshardbeststate"]
	resI, rpcERR := c(httpServer, []interface{}{sid}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.GetShardBestState), nil
}
func (r *LocalRPCClient) GetTransactionByHash(transactionHash string) (res *jsonresult.TransactionDetail, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["gettransactionbyhash"]
	resI, rpcERR := c(httpServer, []interface{}{transactionHash}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.TransactionDetail), nil
}
func (r *LocalRPCClient) GetPrivacyCustomToken(tokenStr string) (res *jsonresult.GetCustomToken, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getprivacycustomtoken"]
	resI, rpcERR := c(httpServer, []interface{}{tokenStr}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.GetCustomToken), nil
}
func (r *LocalRPCClient) GetBurningAddress(beaconHeight float64) (res string, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getburningaddress"]
	resI, rpcERR := c(httpServer, []interface{}{beaconHeight}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(string), nil
}
func (r *LocalRPCClient) GetPublicKeyRole(publicKey string, detail bool) (res interface{}, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getpublickeyrole"]
	resI, rpcERR := c(httpServer, []interface{}{publicKey, detail}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(interface{}), nil
}
func (r *LocalRPCClient) GetBlockChainInfo() (res *jsonresult.GetBlockChainInfoResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getblockchaininfo"]
	resI, rpcERR := c(httpServer, []interface{}{}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.GetBlockChainInfoResult), nil
}
func (r *LocalRPCClient) GetCandidateList() (res *jsonresult.CandidateListsResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getcandidatelist"]
	resI, rpcERR := c(httpServer, []interface{}{}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.CandidateListsResult), nil
}
func (r *LocalRPCClient) GetCommitteeList() (res *jsonresult.CommitteeListsResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getcommitteelist"]
	resI, rpcERR := c(httpServer, []interface{}{}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.CommitteeListsResult), nil
}
func (r *LocalRPCClient) GetBlockHash(chainID float64, height float64) (res []common.Hash, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getblockhash"]
	resI, rpcERR := c(httpServer, []interface{}{chainID, height}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.([]common.Hash), nil
}
func (r *LocalRPCClient) RetrieveBlock(hash string, verbosity string) (res *jsonresult.GetShardBlockResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["retrieveblock"]
	resI, rpcERR := c(httpServer, []interface{}{hash, verbosity}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.GetShardBlockResult), nil
}
func (r *LocalRPCClient) RetrieveBlockByHeight(shardID float64, height float64, verbosity string) (res []*jsonresult.GetShardBlockResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["retrieveblockbyheight"]
	resI, rpcERR := c(httpServer, []interface{}{shardID, height, verbosity}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.([]*jsonresult.GetShardBlockResult), nil
}
func (r *LocalRPCClient) RetrieveBeaconBlock(hash string) (res *jsonresult.GetBeaconBlockResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["retrievebeaconblock"]
	resI, rpcERR := c(httpServer, []interface{}{hash}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(*jsonresult.GetBeaconBlockResult), nil
}
func (r *LocalRPCClient) RetrieveBeaconBlockByHeight(height float64) (res []*jsonresult.GetBeaconBlockResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["retrievebeaconblockbyheight"]
	resI, rpcERR := c(httpServer, []interface{}{height}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.([]*jsonresult.GetBeaconBlockResult), nil
}
func (r *LocalRPCClient) GetRewardAmountByEpoch(shard float64, epoch float64) (res uint64, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["getrewardamountbyepoch"]
	resI, rpcERR := c(httpServer, []interface{}{shard, epoch}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(uint64), nil
}
func (r *LocalRPCClient) DefragmentAccount(privateKey string, maxValue float64, fee float64, privacy float64) (res jsonresult.CreateTransactionResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["defragmentaccount"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, maxValue, fee, privacy}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionResult), nil
}
func (r *LocalRPCClient) DefragmentAccountToken(privateKey string, receiver map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.HttpHandler["defragmentaccounttoken"]
	resI, rpcERR := c(httpServer, []interface{}{privateKey, receiver, fee, privacy, reqInfo, p1, pPrivacy}, nil)
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resI.(jsonresult.CreateTransactionTokenResult), nil
}
