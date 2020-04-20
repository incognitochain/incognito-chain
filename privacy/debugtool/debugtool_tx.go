package debugtool

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Parse from byte to AutoTxByHash
func ParseAutoTxHashFromBytes(b []byte) (*AutoTxByHash, error) {
	data := new(AutoTxByHash)
	err := json.Unmarshal(b, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Query the RPC server then return the AutoTxByHash
func (this *DebugTool) getAutoTxByHash(txHash string) (*AutoTxByHash, error) {
	if len(this.url) == 0 {
		return nil, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := fmt.Sprintf(`{
		"jsonrpc":"1.0",
		"method":"gettransactionbyhash",
		"params":["%s"],
		"id":1
	}`, txHash)
	b, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	autoTx, txError := ParseAutoTxHashFromBytes(b)
	if txError != nil {
		return nil, err
	}
	return autoTx, nil
}

// Get only the proof of transaction requiring the txHash
func (this *DebugTool) GetProofTransactionByHash(txHash string) (string, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.Proof, nil
}

// Get only the Sig of transaction requiring the txHash
func (this *DebugTool) GetSigTransactionByHash(txHash string) (string, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.Sig, nil
}

// Get only the BlockHash of transaction requiring the txHash
func (this *DebugTool) GetBlockHashTransactionByHash(txHash string) (string, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.BlockHash, nil
}

// Get only the BlockHeight of transaction requiring the txHash
func (this *DebugTool) GetBlockHeightTransactionByHash(txHash string) (int, error) {
	tx, err := this.getAutoTxByHash(txHash)
	if err != nil {
		return -1, err
	}
	return tx.Result.BlockHeight, nil
}

// Get the whole result of rpc call 'gettransactionbyhash'
func (this *DebugTool) GetTransactionByHash(txHash string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := fmt.Sprintf(`{
		"jsonrpc":"1.0",
		"method":"gettransactionbyhash",
		"params":["%s"],
		"id":1
	}`, txHash)
	return this.SendPostRequestWithQuery(query)
}
