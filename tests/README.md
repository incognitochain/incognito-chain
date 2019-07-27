# Intergration Test

## Params
- Second Os Params: testcase name
- Third Os Params: config file
## Rpc Method Prototype
```$xslt
type ExampleReponse struct {
	F1 string
	F2 int
}
// example response => json result
func (client *Client) getExampleRpc(p1 string, p2 int) (result *ExampleReponse, err *rpcserver.RPCError) {
	res, rpcError := makeRPCRequest(client.Host, client.Port, getMethodName(), p1, p2)
	if rpcError != nil {
		return nil, rpcError
	}
	errUnMarshal := json.Unmarshal(res.Result, &result)
	if errUnMarshal != nil {
		return nil, rpcserver.NewRPCError(rpcserver.ErrNetwork, err)
	}
	return result, res.Error
}

``` 