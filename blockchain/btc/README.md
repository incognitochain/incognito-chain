# Download bitcoin and run node
## Download Bitcoin Core client: 
https://bitcoin.org/en/download
## Run Bitcoin Core Client via this script
`$ bitcoind -server -rest -rpcport=8332 -rpcallowip=0.0.0.0/0 -rpcbind=0.0.0.0 -rpcuser=[username] -rpcpassword=[password]`
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
## Get Block By Hash
`$ curl --user username:password --data-binary '{"jsonrpc":"1.0","id":"curltext","method":"getblock","params":["000000000000000000210a7be63100bf18ccd43ea8c3e536c476d8d339baa1d9"]}' -H 'content-type:text/plain;' http://159.65.142.153:8332`
```$xslt
{
    "result":
        {
        "hash":"000000000000000000210a7be63100bf18ccd43ea8c3e536c476d8d339baa1d9",
        "confirmations":2,
        "strippedsize":954627,
        "size":1129664,
        "weight":3993545,
        "height":579340,
        "version":541065216,
        "versionHex":"20400000",
        "merkleroot":"d7448a85151667c19b645b37472c5aae5207f43a75d7e85008068238ee7a8314",
        "tx":[...],
        "time":1559728975,
        "mediantime":1559727448,
        "nonce":2544736069,
        "bits":"1725bb76",
        "difficulty":7459680720542.296,
        "chainwork":"000000000000000000000000000000000000000006894f04c037721ae5d25931",
        "nTx":2112,
        "previousblockhash":"0000000000000000000438eb8120c7b7c0fb1e2b8044beb888d07d3cf848a253",
        "nextblockhash":"00000000000000000002862fcb70eaf9186f8b9e3b5516e03c6dd333ad5910b8"
    },
    "error":null,
    "id":"curltext"
}
        
        
```
