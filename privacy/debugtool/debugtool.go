package debugtool
import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)
type DebugTool struct {
	url string
}
func (this *DebugTool) InitMainnet() *DebugTool {
	if this == nil {
		this = new(DebugTool)
	}
	this.url = "https://mainnet.incognito.org/fullnode"
	return this
}
func (this *DebugTool) InitTestnet() *DebugTool {
	if this == nil {
		this = new(DebugTool)
	}
	this.url = "http://51.161.119.66:9334"
	return this
}
func (this *DebugTool) InitLocal(port string) *DebugTool {
	if this == nil {
		this = new(DebugTool)
	}
	this.url = "http://127.0.0.1:" + port
	return this
}
func (this *DebugTool) InitDevNet() *DebugTool {
	if this == nil {
		this = new(DebugTool)
	}
	this.url = "http://54.39.158.106:9334"
	return this
}
func (this *DebugTool) SendPostRequestWithQuery(query string) ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", this.url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		return body, nil
	}
}
type RPCError struct {
	Code       int    `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
	StackTrace string `json:"StackTrace"`
	err error `json:"Err"`
}
type JsonResponse struct {
	Id      *interface{}         `json:"Id"`
	Result  json.RawMessage      `json:"Result"`
	Error   *RPCError `json:"Error"`
	Params  interface{}          `json:"Params"`
	Method  string               `json:"Method"`
	Jsonrpc string               `json:"Jsonrpc"`
}
type GetBlockChainInfoResult struct {
	ChainName    string                   `json:"ChainName"`
	BestBlocks   map[int]GetBestBlockItem `json:"BestBlocks"`
	ActiveShards int                      `json:"ActiveShards"`
}
type GetBestBlockItem struct {
	Height              uint64 `json:"Height"`
	Hash                string `json:"Hash"`
	TotalTxs            uint64 `json:"TotalTxs"`
	BlockProducer       string `json:"BlockProducer"`
	ValidationData      string `json:"ValidationData"`
	Epoch               uint64 `json:"Epoch"`
	Time                int64  `json:"Time"`
	RemainingBlockEpoch uint64 `json:"RemainingBlockEpoch"`
	EpochBlock          uint64 `json:"EpochBlock"`
}
func ParseResponse(respondInBytes []byte) (*JsonResponse, error) {
	var respond JsonResponse
	err := json.Unmarshal(respondInBytes, &respond)
	if err != nil {
		return nil, err
	}
	return &respond, nil
}
func (this *DebugTool) GetBestBlock() ([]byte, error) {
	if len(this.url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	query := `{
		"jsonrpc":"1.0",
		"method":"getbestblock",
		"params": "",
		"id":1
	}`
	return this.SendPostRequestWithQuery(query)
}
func (this *DebugTool) GetBestBeaconHeight() (uint64, error){
	b, err := this.GetBestBlock()
	if err != nil{
		return 0, err
	}
	respond, err := ParseResponse(b)
	if err != nil{
		return 0, err
	}
	var res json.RawMessage
	err = json.Unmarshal(respond.Result, &res)
	if err != nil{
		return 0, err
	}
	var blockChainInfo GetBlockChainInfoResult
	err = json.Unmarshal(res, &blockChainInfo)
	if err != nil{
		return 0, err
	}
	bestBeaconBlock, ok := blockChainInfo.BestBlocks[-1]
	if !ok{
		return 0, errors.New("cannot get best beacon block")
	}
	return bestBeaconBlock.Height, nil
}