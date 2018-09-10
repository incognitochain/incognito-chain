package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
)

// Response is the general form of a JSON-RPC response.  The type of the Result
// field varies from one command to the next, so it is implemented as an
// interface.  The ID field has to be a pointer for Go to put a null in it when
// empty.
type Response struct {
	Result json.RawMessage  `json:"result"`
	Error  *common.RPCError `json:"error"`
	ID     *interface{}     `json:"id"`
}

// NewResponse returns a new JSON-RPC response object given the provided id,
// marshalled result, and RPC error.  This function is only provided in case the
// caller wants to construct raw responses for some reason.
//
// Typically callers will instead want to create the fully marshalled JSON-RPC
// response to send over the wire with the MarshalResponse function.
func NewResponse(id interface{}, marshalledResult []byte, rpcErr *common.RPCError) (*Response, error) {
	if !IsValidIDType(id) {
		str := fmt.Sprintf("the id of type '%T' is invalid", id)
		return nil, common.MakeError(common.ErrInvalidType, str)
	}

	pid := &id
	return &Response{
		Result: marshalledResult,
		Error:  rpcErr,
		ID:     pid,
	}, nil
}

// IsValidIDType checks that the ID field (which can go in any of the JSON-RPC
// requests, responses, or notifications) is valid.  JSON-RPC 1.0 allows any
// valid JSON type.  JSON-RPC 2.0 (which coind follows for some parts) only
// allows string, number, or null, so this function restricts the allowed types
// to that list.  This function is only provided in case the caller is manually
// marshalling for some reason.    The functions which accept an ID in this
// package already call this function to ensure the provided id is valid.
func IsValidIDType(id interface{}) bool {
	switch id.(type) {
	case int, int8, int16, int32, int64,
	uint, uint8, uint16, uint32, uint64,
	float32, float64,
	string,
	nil:
		return true
	default:
		return false
	}
}

// MarshalResponse marshals the passed id, result, and RPCError to a JSON-RPC
// response byte slice that is suitable for transmission to a JSON-RPC client.
func MarshalResponse(id interface{}, result interface{}, rpcErr *common.RPCError) ([]byte, error) {
	marshalledResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	response, err := NewResponse(id, marshalledResult, rpcErr)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&response)
}

// GetBlockChainInfoResult models the data returned from the getblockchaininfo
// command.
type GetBlockChainInfoResult struct {
	Chain                string  `json:"BlockChain"`
	Blocks               int     `json:"Blocks"`
	Headers              int32   `json:"Headers"`
	BestBlockHash        string  `json:"BestBlockHash"`
	Difficulty           uint32  `json:"Difficulty"`
	MedianTime           int64   `json:"MedianTime"`
	VerificationProgress float64 `json:"VerificationProgress,omitempty"`
	Pruned               bool    `json:"Pruned"`
	PruneHeight          int32   `json:"PruneHeight,omitempty"`
	ChainWork            string  `json:"ChainWork,omitempty"`
	//SoftForks            []*SoftForkDescription              `json:"softforks"`
	//Bip9SoftForks        map[string]*Bip9SoftForkDescription `json:"bip9_softforks"`
}

// ListUnspentResult models a successful response from the listunspent request.
type ListUnspentResult struct {
	/*TxID          string  `json:"TxID"`
	Vout          int     `json:"Vout"`
	Address       string  `json:"Address"`
	Account       string  `json:"Account"`
	ScriptPubKey  string  `json:"ScriptPubKey"`
	RedeemScript  string  `json:"RedeemScript,omitempty"`
	Amount        float64 `json:"Amount"`
	Confirmations int64   `json:"Confirmations"`
	Spendable     bool    `json:"Spendable"`
	TxOutType     string  `json:"TxOutType"`*/
	ListUnspentResultItems map[string][]ListUnspentResultItem `json:"ListUnspentResultItems"`
}

type ListUnspentResultItem struct {
	TxId          string          `json:"TxId"`
	JoinSplitDesc []JoinSplitDesc `json:"JoinSplitDesc"`
}

type JoinSplitDesc struct {
	Commitments [][]byte `json:"Commitments"`
	Amount      uint64   `json:"Amount"`
	Anchor      []byte   `json:"Anchor"`
}

// end

type GetHeaderResult struct {
	BlockNum  int                    `json:"blocknum"`
	BlockHash string                 `json:"blockhash"`
	Header    blockchain.BlockHeader `json:"header"`
}

type ListAccounts struct {
	Accounts map[string]float64 `json:"Accounts"`
}

type GetAddressesByAccount struct {
	Addresses [] string
}
