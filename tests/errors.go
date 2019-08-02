package main

import "errors"

//Error
var (
	ParseFailedError               = errors.New("Failed to parse result")
	UnexpectedError                = errors.New("Expect Error is null")
	ExpectedError                  = errors.New("Expect Error is not null")
	WrongReturnedErrorCodeError    = errors.New("Wrong Returned Error Code")
	WrongReturnedErrorMessageError = errors.New("Wrong Error Message")
	ResponseNotFoundError          = errors.New("Expected Response Not Found")
	WrongExpectedResponseError     = errors.New("Wrong Expected Response")
	RPCNetworkError                = errors.New("No Error and Response from Server")
	DataAssertionError             = errors.New("Assertion type failure")
	ContextNotFoundError           = errors.New("Key in context not found")
	WantedKeyNotFoundError         = errors.New("Wanted Key Not Found in Response")
	ResultAndResponseTypeError     = errors.New("RPC Result And Response Type are Not Compatible")
	ParseNodeConfigFailedError     = errors.New("Failed To Parse Node Data From Config")
	ParseHostError                 = errors.New("Failed To Parse host Data From Config")
	ParsePortError                 = errors.New("Failed To Parse port Data From Config")
	ParseWsDataError               = errors.New("Failed To Parse Websocket Data From Config")
)
