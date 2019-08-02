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
	ErrNetworkError                = errors.New("No Error and Response from Server")
	ErrAssertionData               = errors.New("Assertion type failure")
	ErrContextNotFound             = errors.New("Key in context not found")
	ErrWantedKeyNotFound           = errors.New("Wanted Key Not Found in Response")
	ErrResultAndResponseType       = errors.New("RPC Result And Response Type are Not Compatible")
	ErrParseNodeConfigFailed       = errors.New("Failed To Parse Node Data From Config")
	ErrParseHost                   = errors.New("Failed To Parse host Data From Config")
	ErrParsePort                   = errors.New("Failed To Parse port Data From Config")
	ErrParseWs                     = errors.New("Failed To Parse Websocket Data From Config")
)
