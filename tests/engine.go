package main

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"reflect"
)

var (
	ErrExpectNoError = errors.New("Expect Error is null")
	ErrExpectError = errors.New("Expect Error is not null")
	ErrWrongCode = errors.New("Wrong Error Code")
	ErrWrongMessage = errors.New("Wrong Error Message")
	ErrResponseNotFound = errors.New("Expected Response Not Found")
	ErrWrongExpectedResponse = errors.New("Wrong Expected Response")
	ErrNetworkError = errors.New("No Error and Response from Server")
	ErrAssertionData = errors.New("Assertion type failure")
	ErrContextNotFound = errors.New("Key in context not found")
	ErrWantedKeyNotFound = errors.New("Wanted Key Not Found in Response")
)
func executeTest(filename string) (map[string]interface{},error) {
	var rpcError *rpcserver.RPCError
	var result = make(map[string]interface{})
	scenarios, err := readfile(filename)
	if err != nil {
		return result, err
	}
	for _, step := range scenarios.steps {
		//command := Command[step.input.name]
		if !step.input.isWait {
			var params []interface{}
			if step.input.fromContext {
				for _, value := range step.input.params {
					if contextKey, ok := value.(string); !ok {
						return result, fmt.Errorf("%+v, expect %+v is %+v", ErrAssertionData, value, "string")
					} else {
						if contextValue, ok := scenarios.context[contextKey]; !ok {
							return result, fmt.Errorf("%+v, key %+v", ErrContextNotFound, contextKey)
						} else {
							params = append(params, contextValue)
						}
					}
				}
			} else {
				params = append(params, step.input.params...)
			}
			result, rpcError = makeRPCRequestV2(step.client, step.input.name, params)
			//data, err := command(step.client, step.input.params)
			if rpcError != nil && rpcError.Code == rpcserver.GetErrorCode(rpcserver.ErrNetwork) {
				return result, err
			}
			// check error
			if step.output.error.isNil {
				if rpcError != nil {
				    return result, fmt.Errorf("%+v, get %+v, %+v", ErrExpectNoError, rpcError.Code, rpcError.Message)
				}
			} else {
				if rpcError == nil {
					return result, fmt.Errorf("%+v, but null", ErrExpectError)
				}
				if step.output.error.code != rpcError.Code {
					return result, fmt.Errorf("%+v, get %+v", ErrWrongCode, rpcError.Code)
				}
				if step.output.error.message != rpcError.Message {
					return result, fmt.Errorf("%+v, get %+v", ErrWrongMessage, rpcError.Message)
				}
			}
			// check output
			// if output is empty list then continue
			//result = data.(map[string]interface{})
			for key, expectedResponse := range step.output.response {
				if returnedResponse, ok := result[key]; !ok {
					return result, ErrResponseNotFound
				} else {
					if !reflect.DeepEqual(expectedResponse, returnedResponse) {
						return result, fmt.Errorf("%+v, get %+v", ErrWrongExpectedResponse, returnedResponse)
					}
				}
			}
			for contextKey, resultKey := range step.store {
				if resultValue, ok := result[resultKey]; !ok {
					return result, fmt.Errorf("%+v, key %+v", ErrWantedKeyNotFound, resultKey)
				} else {
					scenarios.context[contextKey] = resultValue
				}
			}
		}
	}
	return result, rpcError
}