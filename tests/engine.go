package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func executeTest(filename string) (interface{}, error) {
	var rpcError *rpcservice.RPCError
	//var result = make(map[string]interface{})
	var rpcResult interface{}
	scenarios, err := readfile(filename)
	if err != nil {
		return nil, err
	}
	for index, step := range scenarios.steps {
		var params []interface{}
		if step.input.fromContext {
			for _, value := range step.input.params {
				if contextKey, ok := value.(string); !ok {
					return nil, fmt.Errorf("%+v, expect %+v is %+v", DataAssertionError, value, "string")
				} else {
					if contextValue, ok := scenarios.context[contextKey]; !ok {
						return nil, fmt.Errorf("%+v, key %+v", ContextNotFoundError, contextKey)
					} else {
						params = append(params, contextValue)
					}
				}
			}
		} else {
			params = append(params, step.input.params...)
		}
		if step.input.conn == "ws" {
			if step.input.wait.Seconds() == 0 {
				step.input.wait = defaultTimeout
			}
			rpcResult, rpcError = makeWsRequest(step.client, step.input.name, step.input.wait, params...)
		} else {
			if step.input.isWait {
				<-time.Tick(step.input.wait)
			}
			rpcResult, rpcError = makeRPCRequestJson(step.client, step.input.name, params...)
		}
		//data, err := command(step.client, step.input.params)
		if rpcError != nil && rpcError.Code == rpcservice.GetErrorCode(rpcservice.NetworkError) {
			return rpcResult, rpcError
		}
		// check error
		if step.output.error.isNil {
			if rpcError != nil {
				return rpcResult, fmt.Errorf("%+v, get %+v, %+v", UnexpectedError, rpcError.Code, rpcError.Message)
			}
		} else {
			if rpcError == nil {
				return rpcResult, fmt.Errorf("%+v, but null", ExpectedError)
			}
			if step.output.error.code != rpcError.Code {
				return rpcResult, fmt.Errorf("%+v, get %+v", WrongReturnedErrorCodeError, rpcError.Code)
			}
			if step.output.error.message != rpcError.Message {
				return rpcResult, fmt.Errorf("%+v, get %+v", WrongReturnedErrorMessageError, rpcError.Message)
			}
		}
		// check output
		// if output is empty list then continue
		if result, ok := rpcResult.(map[string]interface{}); ok {
			if response, ok := step.output.response.(map[string]interface{}); ok {
				for key, expectedResponse := range response {
					if returnedResponse, ok := result[key]; !ok {
						return rpcResult, ResponseNotFoundError
					} else {
						if !reflect.DeepEqual(expectedResponse, returnedResponse) {
							return rpcResult, fmt.Errorf("%+v, expected %+v, get %+v", WrongExpectedResponseError, expectedResponse, returnedResponse)
						}
					}
				}
				for contextKey, resultKey := range step.store {
					if resultValue, ok := result[resultKey]; !ok {
						return rpcResult, fmt.Errorf("%+v, key %+v", WantedKeyNotFoundError, resultKey)
					} else {
						scenarios.context[contextKey] = resultValue
					}
				}
			} else {
				return rpcResult, fmt.Errorf("%+v, result %+v, response %+v", ResultAndResponseTypeError, reflect.TypeOf(rpcResult), reflect.TypeOf(response))
			}
		} else {
			if !reflect.DeepEqual(rpcResult, step.output.response) {
				return rpcResult, fmt.Errorf("%+v, result %+v, type %+v; response %+v, type %+v", WrongExpectedResponseError, rpcResult, reflect.TypeOf(rpcResult), step.output.response, reflect.TypeOf(step.output.response))
			}
		}
		log.Printf("Testcase %+v, pass step %+v, command %+v", filename, index+1, step.input.name)
	}
	return rpcResult, rpcError
}
