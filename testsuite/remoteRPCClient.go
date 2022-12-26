package devframework //This file is auto generated. Please do not change if you dont know what you are doing
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/testsuite/rpcclient"
)

type RemoteRPCClient struct {
	Endpoint string
}

func (r *RemoteRPCClient) IsInstantFinality(chainID int) (res bool, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "isinstantfinality",
		"params":  []interface{}{chainID},
		"id":      1,
	})
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(err.Error())
	}
	resp := struct {
		Result bool
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetStateDB(checkpoint string, cid int, dbType int, offset uint64, f func([]byte)) error {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbootstrapstatedb",
		"params":  []interface{}{checkpoint, cid, dbType, offset},
		"id":      1,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//TODO: stream body and then parse
	return nil
}

func (r *RemoteRPCClient) CreateRawTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createtransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy},
		"id":      1,
	})
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}

	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(err.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res, errors.New(err.Error())
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, err
}

func (r *RemoteRPCClient) GetAllViewDetail(chainID int) (res []jsonresult.GetViewResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getallviewdetail",
		"params":  []interface{}{chainID},
		"id":      1,
	})
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(err.Error())
	}
	resp := struct {
		Result []jsonresult.GetViewResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetMiningInfo() (res *jsonresult.GetMiningInfoResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getmininginfo",
		"params":  []interface{}{},
		"id":      1,
	})
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(err.Error())
	}
	resp := struct {
		Result *jsonresult.GetMiningInfoResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetPortalShieldingRequestStatus(tx string) (res *metadata.PortalShieldingRequestStatus, err error) {
	panic("implement me")
}

func (r *RemoteRPCClient) CreateAndSendTxWithPortalV4UnshieldRequest(privatekey string, tokenID string, amount string, paymentAddress string, remoteAddress string) (res jsonresult.CreateTransactionTokenResult, err error) {
	panic("implement me")
}

func (r *RemoteRPCClient) GetPortalUnshieldRequestStatus(tx string) (res *metadata.PortalUnshieldRequestStatus, err error) {
	panic("implement me")
}

func (r *RemoteRPCClient) CreateAndSendTokenInitTransaction(param rpcclient.PdexV3InitTokenParam) (jsonresult.CreateTransactionTokenResult, error) {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxMintNft(privatekeys string) error {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxAddLiquidity(privatekey string, param rpcclient.PdexV3AddLiquidityParam) error {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxWithdrawLiquidity(privatekey, poolairID, nftID, shareAmount string) error {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxModifyParams(privatekey string, newParams rpcclient.PdexV3Params) {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxStake(privatekey, stakingPoolID, nftID, amount string) error {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxUnstake(privatekey, stakingPoolID, nftID, amount string) error {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxAddTrade(privatekey string, param rpcclient.PdexV3TradeParam) error {
	panic("implement me")
}

func (r *RemoteRPCClient) Pdexv3_TxAddOrder(privatekey string, params rpcclient.PdexV3AddOrderParam) error {
	panic("implement me")
}

func (r *RemoteRPCClient) SendFinishSync(mining, cpk string, sid float64) error {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "sendfinishsync",
		"params":  []interface{}{mining, cpk, sid},
		"id":      1,
	})
	if rpcERR != nil {
		return errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return errors.New(rpcERR.Error())
	}
	resp := struct {
		Result bool
		Error  *ErrMsg
	}{}
	//fmt.Println(string(body))
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return errors.New(err.Error())
	}
	return err

}
func (r *RemoteRPCClient) CreateConvertCoinVer1ToVer2Transaction(privateKey string) (err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createconvertcoinver1tover2transaction",
		"params":  []interface{}{privateKey, -1},
		"id":      1,
	})
	if err != nil {
		return errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return errors.New(rpcERR.Error())
	}
	resp := struct {
		Result bool
		Error  *ErrMsg
	}{}
	//fmt.Println(string(body))
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return errors.New(err.Error())
	}
	return err
}

func (r *RemoteRPCClient) CreateAndSendTXShieldingRequest(privateKey string, incAddr string, tokenID string, proof string) (res jsonresult.CreateTransactionResult, err error) {
	panic("implement me")
}

type ErrMsg struct {
	Code       int
	Message    string
	StackTrace string
}

func (r *RemoteRPCClient) sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post(r.Endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *RemoteRPCClient) GetBlocksFromHeight(shardID int, from uint64, num int) (res interface{}, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getblocksfromheight",
		"params":  []interface{}{shardID, from, num},
		"id":      1,
	})
	if err != nil {
		return res, err
	}
	body, err := r.sendRequest(requestBody)

	if err != nil {
		return res, err
	}

	if shardID == -1 {
		resp := struct {
			Result []types.BeaconBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	} else {
		resp := struct {
			Result []types.ShardBlock
			Error  *ErrMsg
		}{}
		err = json.Unmarshal(body, &resp)
		if resp.Error != nil && resp.Error.StackTrace != "" {
			return res, errors.New(resp.Error.StackTrace)
		}
		if err != nil {
			return res, err
		}
		return resp.Result, nil
	}
}

func (r *RemoteRPCClient) GetMempoolInfo() (res *jsonresult.GetMempoolInfo, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"params":  []interface{}{},
		"method":  "getmempoolinfo",
		"id":      1,
	})
	if rpcERR != nil {
		return nil, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return nil, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result *jsonresult.GetMempoolInfo
		Error  *ErrMsg
	}{}
	//fmt.Println(string(body))
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return nil, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return nil, errors.New(err.Error())
	}
	return resp.Result, nil
}

func (r *RemoteRPCClient) SendRawTransaction(data string) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "sendtransaction",
		"params":  []interface{}{data},
		"id":      1,
	})
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(err.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res, errors.New(err.Error())
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, err
}

func (r *RemoteRPCClient) SubmitKey(privateKey string) (res bool, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "submitkey",
		"params":  []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, err
	}
	resp := struct {
		Result bool
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) AuthorizedSubmitKey(privateKey string) (res bool, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "authorizedsubmitkey",
		"params":  []interface{}{privateKey, "0c3d46946bbf99c8213dd7f6c640ed6433bdc056a5b68e7e80f5525311b0ca11", 0, true},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result bool
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetBalanceByPrivateKey(privateKey string) (res uint64, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbalancebyprivatekey",
		"params":  []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result uint64
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetListPrivacyCustomTokenBalance(privateKey string) (res jsonresult.ListCustomTokenBalance, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getlistprivacycustomtokenbalance",
		"params":  []interface{}{privateKey},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.ListCustomTokenBalance
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetRewardAmount(paymentAddress string) (res map[string]uint64, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getrewardamount",
		"params":  []interface{}{paymentAddress},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result map[string]uint64
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) WithdrawReward(privateKey string, receivers map[string]interface{}, amount float64, privacy float64, info map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "withdrawreward",
		"params":  []interface{}{privateKey, receivers, amount, privacy, info},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) StopAutoStake(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createunstaketransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy, stakeInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(string(body))
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, err
}

func (r *RemoteRPCClient) Unstake(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createunstaketransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy, stakeInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(string(body))
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, err
}

func (r *RemoteRPCClient) AddStake(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendaddstakingtransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy, stakeInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(string(body))
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendstakingtransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy, stakeInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return res, err
	}

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendStopAutoStakingTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, stopStakeInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendstopautostakingtransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy, stopStakeInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendPrivacyCustomTokenTransaction(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, tokenInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendprivacycustomtokentransaction",
		"params":  []interface{}{privateKey, receivers, fee, privacy, tokenInfo, p1, pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionTokenResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithWithdrawalReqV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithwithdrawalreqv2",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithPDEFeeWithdrawalReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithpdefeewithdrawalreq",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithPTokenTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokentradereq",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo, p1, pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionTokenResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithPTokenCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokencrosspooltradereq",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo, p1, pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionTokenResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithPRVTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvtradereq",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithPRVCrossPoolTradeReq(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvcrosspooltradereq",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithPTokenContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithptokencontributionv2",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo, p1, pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionTokenResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) CreateAndSendTxWithPRVContributionV2(privateKey string, receivers map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtxwithprvcontributionv2",
		"params":  []interface{}{privateKey, receivers, fee, privacy, reqInfo},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetPDEState(data map[string]interface{}) (res jsonresult.CurrentPDEState, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getpdestate",
		"params":  []interface{}{data},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CurrentPDEState
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetBeaconBestState() (res jsonresult.GetBeaconBestState, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbeaconbeststate",
		"params":  []interface{}{},
		"id":      1,
	})
	if rpcERR != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(err.Error())
	}
	resp := struct {
		Result jsonresult.GetBeaconBestState
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetShardBestState(sid int) (res jsonresult.GetShardBestState, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getshardbeststate",
		"params":  []interface{}{sid},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.GetShardBestState
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

type TXDetail struct {
	*jsonresult.TransactionDetail
	Proof                         interface{}
	ProofDetail                   interface{}
	PrivacyCustomTokenProofDetail interface{}
}

func (r *RemoteRPCClient) GetTransactionByHash(transactionHash string) (res *TXDetail, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "gettransactionbyhash",
		"params":  []interface{}{transactionHash},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}

	resp := struct {
		Result *TXDetail
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetPrivacyCustomToken(tokenStr string) (res *jsonresult.GetCustomToken, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getprivacycustomtoken",
		"params":  []interface{}{tokenStr},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result *jsonresult.GetCustomToken
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetBurningAddress(beaconHeight float64) (res string, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getburningaddress",
		"params":  []interface{}{beaconHeight},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result string
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetPublicKeyRole(publicKey string, detail bool) (res interface{}, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getpublickeyrole",
		"params":  []interface{}{publicKey, detail},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result interface{}
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetBlockChainInfo() (res *jsonresult.GetBlockChainInfoResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getblockchaininfo",
		"params":  []interface{}{},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result *jsonresult.GetBlockChainInfoResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetCandidateList() (res *jsonresult.CandidateListsResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getcandidatelist",
		"params":  []interface{}{},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result *jsonresult.CandidateListsResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

type ConsensusRule struct {
	VoteRule          string `json:"vote_rule,omitempty"`
	CreateRule        string `json:"create_rule,omitempty"`
	HandleVoteRule    string `json:"handle_vote_rule,omitempty"`
	HandleProposeRule string `json:"handle_propose_rule,omitempty"`
	InsertRule        string `json:"insert_rule,omitempty"`
	ValidatorRule     string `json:"validator_rule,omitempty"`
}

func (r *RemoteRPCClient) SetConsensusRule(rules ConsensusRule) (res map[string]interface{}, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "setconsensusrule",
		"params":  []interface{}{rules},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}

	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result map[string]interface{}
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(err.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetCommitteeList() (res *jsonresult.CommitteeListsResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getcommitteelist",
		"params":  []interface{}{},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result *jsonresult.CommitteeListsResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetBlockHash(chainID float64, height float64) (res []common.Hash, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getblockhash",
		"params":  []interface{}{chainID, height},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result []common.Hash
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) RetrieveBlock(hash string, verbosity string) (res *jsonresult.GetShardBlockResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "retrieveblock",
		"params":  []interface{}{hash, verbosity},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result *jsonresult.GetShardBlockResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) RetrieveBlockByHeight(shardID float64, height float64, verbosity string) (res []*jsonresult.GetShardBlockResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "retrieveblockbyheight",
		"params":  []interface{}{shardID, height, verbosity},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result []*jsonresult.GetShardBlockResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) RetrieveBeaconBlock(hash string) (res *jsonresult.GetBeaconBlockResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "retrievebeaconblock",
		"params":  []interface{}{hash},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result *jsonresult.GetBeaconBlockResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) RetrieveBeaconBlockByHeight(height float64) (res []*jsonresult.GetBeaconBlockResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "retrievebeaconblockbyheight",
		"params":  []interface{}{height},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result []*jsonresult.GetBeaconBlockResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetRewardAmountByEpoch(shard float64, epoch float64) (res uint64, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getrewardamountbyepoch",
		"params":  []interface{}{shard, epoch},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result uint64
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) DefragmentAccount(privateKey string, maxValue float64, fee float64, privacy float64) (res jsonresult.CreateTransactionResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "defragmentaccount",
		"params":  []interface{}{privateKey, maxValue, fee, privacy},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) DefragmentAccountToken(privateKey string, receiver map[string]interface{}, fee float64, privacy float64, reqInfo map[string]interface{}, p1 string, pPrivacy float64) (res jsonresult.CreateTransactionTokenResult, err error) {
	requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "defragmentaccounttoken",
		"params":  []interface{}{privateKey, receiver, fee, privacy, reqInfo, p1, pPrivacy},
		"id":      1,
	})
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	resp := struct {
		Result jsonresult.CreateTransactionTokenResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return res, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return res, errors.New(rpcERR.Error())
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) PreparePRVForTest(
	privateKey string, receivers map[string]interface{},
) (res *jsonresult.CreateTransactionResult, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "createandsendtransaction",
		"params":  []interface{}{privateKey, receivers, -1, 0},
		"id":      1,
	})
	if err != nil {

	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return nil, err
	}
	resp := struct {
		Result *jsonresult.CreateTransactionResult
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil && resp.Error.StackTrace != "" {
		return nil, fmt.Errorf(resp.Error.StackTrace)
	}
	return resp.Result, err
}

func (r *RemoteRPCClient) GetCommitteeState(height uint64, hash string) (*jsonresult.CommiteeState, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getcommitteestate",
		"params": []interface{}{
			height,
			hash,
		},
		"id": 1,
	})
	if err != nil {
		return nil, err
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return nil, err
	}
	resp := struct {
		Result *jsonresult.CommiteeState
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return nil, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return nil, err
	}
	return resp.Result, err
}

/*func (r *RemoteRPCClient) GetShardStakerInfo(height uint64, stakerPubkey string) (*statedb.ShardStakerInfo, error) {*/
/*requestBody, err := json.Marshal(map[string]interface{}{*/
/*"jsonrpc": "1.0",*/
/*"method":  "getshardstakerinfo",*/
/*"params": []interface{}{*/
/*height,*/
/*stakerPubkey,*/
/*},*/
/*"id": 1,*/
/*})*/
/*if err != nil {*/
/*return nil, err*/
/*}*/
/*body, err := r.sendRequest(requestBody)*/
/*if err != nil {*/
/*return nil, err*/
/*}*/
/*resp := struct {*/
/*Result *statedb.ShardStakerInfo*/
/*Error  *ErrMsg*/
/*}{}*/
/*err = json.Unmarshal(body, &resp)*/

/*if resp.Error != nil && resp.Error.StackTrace != "" {*/
/*return nil, errors.New(resp.Error.StackTrace)*/
/*}*/

/*if err != nil {*/
/*return nil, err*/
/*}*/
/*return resp.Result, err*/
/*}*/

func (r *RemoteRPCClient) GetBeaconStakerInfo(height uint64, hash string) (*statedb.BeaconStakerInfo, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "getbeaconstakerinfo",
		"params": []interface{}{
			height,
			hash,
		},
		"id": 1,
	})
	if err != nil {
		return nil, err
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		return nil, err
	}
	resp := struct {
		Result *statedb.BeaconStakerInfo
		Error  *ErrMsg
	}{}
	err = json.Unmarshal(body, &resp)

	if resp.Error != nil && resp.Error.StackTrace != "" {
		return nil, errors.New(resp.Error.StackTrace)
	}

	if err != nil {
		return nil, err
	}
	return resp.Result, err
}
