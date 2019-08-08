# Intergration Test
## Params
- Second Os Params: testcase name
- Third Os Params: config file
## Test by Json
Sample Json:
```json
{
 "node": "node-key",
 "input": {
  "command": "...",
  "context": false,
  "params": [],
  "type": "http",
  "wait": "60"
 },
 "output": {
  "error": null,
  "response": {}
 },
 "store": {
  "variable-key": "variable-value"
 }    
}
```
### Node Structure
- node-key is host, rpcport, wsport loaded from websocket.
- Example key:
    + Key -1: beacon
    + Key 0: shard0-
    + Key 1: shard1-
    ...
### Input Structure:
- command: name of command (ex: gettransactionbyhash)
- context:
    + true: get value for params from scenarios context, get by key
    + false (by default): get value from input params
- params: params of command (ex: "0xe4afb36e5a99c20cbd5835a1312fc1b5fd65dbe7d36eb992f1dcfcfa8b64c796")
    + if context is true: input params is keys in context
    + if context is false: input params is value from json file
- type: connection type to node, http or ws (websocket)
- wait: wait time of this command, counted in second. This command may wait for n seconds or executed again and again in n second until the return output has `null` error. 
### Output Structure:
- Error:
    1. null
    2. Expect code and message, example:
```json
    {
      "code": -1,
      "message": "unexpected error"
    }
``` 
- Response: compatible type with RPC result 
    + Key-Value: expected value with key value type, Check return result and find expected value by key then compare the value (response must be subset of result)
    + Number, String, Boolean, Array: must be deeply equal with RPC result 

If response is `{}` then ignore expected response
Example:
- Expected response
```json
{
  "balance": 1000
}
```
- Returned response
```json
{
  "paymentaddress": "1Uv34F64ktQkX1eyd6YEG8KTENV8W5w48LRsi6oqqxVm65uvcKxEAzL2dp5DDJTqAQA7HANfQ1enKXCh2EvVdvBftko6GtGnjSZ1KqJhi",
  "balance": 1000
}
```
Returned response may have many information but as long as it meet the expected response then it pass the test. In another word, expected response is subset of returned response
### Store Structure:
- Value to be stored must be a value from returned output
- This key-value pair will be stored in context of scenarios
 
#### Example
```json
[
  {
    "node": {
      "host": "http://localhost",
      "port": "9334"
    },
    "input": {
      "command": "createandsendtransaction",
      "context": false,
      "params": [
        "112t8rtTwTgp4QKJ7rP2p5TyqtFjKYxeFHCUumTwuH4NbCAk7g7H1MvH5eDKyy6N5wvT1FVVLoPrUzrAKKzJeHcCrc2BoSJfTvkDobVSmSZe",
        {
          "1Uv34F64ktQkX1eyd6YEG8KTENV8W5w48LRsi6oqqxVm65uvcKxEAzL2dp5DDJTqAQA7HANfQ1enKXCh2EvVdvBftko6GtGnjSZ1KqJhi": 1000
        },
        10,
        0
      ]
    },
    "output": {
      "error": null,
      "response": {}
    },
    "store": {
      "TxID1": "TxID"
      // returned response must contain key TxID
      // get value by key TxID then stored in scenarios context with key TxID1
      // if TxID can't be found from returned response then testcase is failed  
    }
  },
  {
    "node": {
      "host": "http://localhost",
      "port": "9334"
    },
    "input": {
      "command": "gettransactionbyhash",
      "context": true,
      // get input params from scenarios context
      "params": ["TxID1"],
      // get value from scenarios context by key TxID1 as input params 
      "type": "ws"
    },
    "output": {
      "error": null,
      "response": {}
    }
  }
]  
```
## Test by Coding
### Rpc Method Prototype
```$xslt

type ExampleReponse struct {
	F1 string
	F2 int
}
// example response => json result
func (client *Client) getExampleRpc(p1 string, p2 int) (result *ExampleReponse, err *rpcserver.RPCError) {
	res, rpcError := makeRPCRequest(client.Host, client.Port, getMethodName(), p1, p2)
	err = handleResponse(res.Result, rpcError, &result)
	return result, err
}

func handleResponse(resResult json.RawMessage, rpcError *rpcserver.RPCError, resultObj interface{}) *rpcserver.RPCError {
	if rpcError != nil {
		return rpcError
	}
	errUnMarshal := json.Unmarshal(resResult, resultObj)
	if errUnMarshal != nil {
		//TODO: unmarshal error
		return rpcserver.NewRPCError(rpcserver.ErrNetwork, errUnMarshal)
	}
	return nil
}
``` 
### Sample Command and Params List: 
#### Create and Send Transaction:
```
command: createandsendtransaction
params: [ "112t8rsURTpYQMp3978j2nvYXTbuMa9H7MfLTA4PCJoxyweZNWRR3beMEtsoLBBbc473Bv8NE3uKUXcVA2Jnh6sPhTEnFfmQEpY8opeFytoM",
        { "1NHp2EKw7ALdXUzBfoRJvKrBBM9nkejyDcHVPvUjDcWRyG22dHHyiBKQGL1c":5250000000000},
        2,
        0,
        64 ]
```
#### Get Transaction By Hash:
```
command: gettranssactionbyhash
params: ["0xe4afb36e5a99c20cbd5835a1312fc1b5fd65dbe7d36eb992f1dcfcfa8b64c796"]
```

