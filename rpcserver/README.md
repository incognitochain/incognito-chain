This package provide list api feature to interact with blockchain node
Using http server to provide a api service

Http client --(http request post)--> RPC http server
       <-----(json http response) --


Post body request:
{
    "jsonrpc": "1.0",
    "method": "__method_name__",
    "params": __json_data_format__,
    "id": __integer_number__
}

Response format:
{
    "Result": __json_data_format,
    "Error": {
        "code": __error_code__,
        "message": __error_message__,
    },
    "Id": __integer_number__
}


