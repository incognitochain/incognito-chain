package main

import (
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/cash/benchmark/api"
	"github.com/ninjadotorg/cash/wallet"
)


type RPC struct {
	endpoint string
}

/**
InitRPC to make instance call blockchain rpc
 */
func InitRPC(endpoint string) *RPC {
	rpc := &RPC{endpoint}
	return rpc
}

/**
GetAccountAddress used to get account address by wallet, if not exist auto create
 */
func (rpc *RPC) GetAccountAddress(params string) (error, *wallet.KeySerializedData) {
	args := map[string]interface{}{
		"jsonrpc": "1.0",
		"method": "getaccessaddress",
		"params": params,
	}
	err, response := api.Post(rpc.endpoint, args)

	if err != nil {
		return err, nil
	}

	var data map[string]interface{}
	json.Unmarshal(response, &data)
	if data["Error"].(string) != "" {
		return errors.New(data["Error"].(string)), nil
	}

	ksd := &wallet.KeySerializedData{}
	json.Unmarshal(response, ksd)

	return nil, ksd
}

/**
DumpPrivateKey used to get private key of address
 */
func (rpc *RPC) DumpPrivateKey(params string) (error, string) {
	args := map[string]interface{}{
		"jsonrpc": "1.0",
		"method": "dumpprivkey",
		"params": params,
	}
	err, response := api.Post(rpc.endpoint, args)

	if err != nil {
		return err, ""
	}

	var data map[string]interface{}
	json.Unmarshal(response, &data)

	if data["Error"].(string) != "" {
		return errors.New(data["Error"].(string)), ""
	}

	return nil, data["Result"].(string)
}

/**
SendMany used to send coin to user
*/
func (rpc *RPC) SendMany(fromPrvKey string, toPubKey string, value float64) (error, string) {
	args := map[string]interface{}{
		"jsonrpc": "1.0",
		"method": "sendmany",
		"params": []interface{}{
			fromPrvKey,
			map[string]interface{}{
				toPubKey: value,
			},
			-1,
			8,
		},
	}
	err, response := api.Post(rpc.endpoint, args)

	if err != nil {
		return err, ""
	}

	var data map[string]interface{}
	json.Unmarshal(response, &data)

	if data["Error"].(string) != "" {
		return errors.New(data["Error"].(string)), ""
	}

	return nil, (data["Result"].(map[string]interface{}))["TxID"].(string)
}




