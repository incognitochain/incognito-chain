# JSON RPC Result

## Get Blockchain Current Infomation
`$ curl --user username:password --data-binary '{"jsonrpc":"1.0","id":"curltext","method":"getblockchaininfo","params":[]}' -H 'content-type:text/plain;' http://159.65.142.153:8332`
```
{
"result":
    {
        "chain":"main",
        "blocks":579338,
        "headers":579338,
        "bestblockhash":"0000000000000000001a963eef5d0e1b1e4d42517a881a3a63a5898bf14de4a0",
        "difficulty":7459680720542.296,
        "mediantime":1559726144,
        "verificationprogress":0.999999247406822,
        "initialblockdownload":false,
        "chainwork":"0000000000000000000000000000000000000000068941730368880f64550e87",
        "size_on_disk":253339619699,
        "pruned":false,
        "softforks":[{"id":"bip34","version":2,"reject":{"status":true}},{"id":"bip66","version":3,"reject":{"status":true}},{"id":"bip65","version":4,"reject":{"status":true}}],
        "bip9_softforks":{"csv":{"status":"active","startTime":1462060800,"timeout":1493596800,"since":419328},
        "segwit":{"status":"active","startTime":1479168000,"timeout":1510704000,"since":481824}},"warnings":""
    },
 "error":null,
 "id":"curltext"
 }
```
## Get Best Block Height
`$ curl --user username:password --data-binary '{"jsonrpc":"1.0","id":"curltext","method":"getblockcount","params":[]}' -H 'content-type:text/plain;' http://159.65.142.153:8332`
```$
{"result":579339,"error":null,"id":"curltext"}
```
