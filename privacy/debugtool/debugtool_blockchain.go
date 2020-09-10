package debugtool

import (
	"errors"
)

func (this *DebugTool) GetBlockchainInfo() ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := `{
		"jsonrpc":"1.0",
		"method":"getblockchaininfo",
		"params": "",
		"id":1
	}`
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) GetBestBlockHash() ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := `{
		"jsonrpc":"1.0",
		"method":"getbestblockhash",
		"params": "",
		"id":1
	}`
	return this.SendPostRequestWithQuery(query)
}

func (this *DebugTool) GetRawMempool() ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := `{
		"jsonrpc": "1.0",
		"method": "getrawmempool",
		"params": "",
		"id": 1
	}`
	return this.SendPostRequestWithQuery(query)
}