This package provide list api feature to interact with blockchain node
Using http server to provide a api service

Http client --(http request post)--> RPC http server
       <-----(json http response) --


- Post body request:
```json
{
    "jsonrpc": "1.0",
    "method": "__command_name__",
    "params": __json_data_format__,
    "id": __integer_number__
}
```

- Response format:
```json
{
    "Result": __json_data_format,
    "Error": {
        "code": __error_code__,
        "message": __error_message__,
    },
    "Id": __integer_number__
}
```

- List common rpc command, client doesn't need to provide limited username/password to call:
  - getblockchaininfo
  - listtransactions
  - createrawtransaction
  - sendtransaction
  - getnumberofcoinsandbonds
  - createactionparamstransaction
  - votecandidate
  - getheader
  
- List limited rpc command:
  - listaccounts
  - getaccount
  - getaddressesbyaccount
  - getaccountaddress
  - dumpprivkey
  - importaccount
  - listunspent
