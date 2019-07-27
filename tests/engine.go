package main

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"reflect"
)

var (
	ErrExpectNoError = errors.New("Expect Error is null")
	ErrWrongCode = errors.New("Wrong Error Code")
	ErrWrongMessage = errors.New("Wrong Error Message")
	ErrResponseNotFound = errors.New("Expected Response Not Found")
	ErrWrongExpectedResponse = errors.New("Wrong Expected Response")
	ErrNetworkError = errors.New("No Error and Response from Server")
)
func executeTest(filename string) (map[string]interface{},error) {
	var returnedErr error
	response := make(map[string]interface{})
	scenarios, err := readfile(filename)
	if err != nil {
		return response, err
	}
	for _, step := range scenarios {
		command := Command[step.input.name]
		if !step.input.isWait {
			data, err := command(step.client, step.input.params)
			returnedErr = err
			if err != nil && err.Code == rpcserver.GetErrorCode(rpcserver.ErrNetwork) {
				return response, err
			}
			// check error
			if step.output.error.isNil {
				if err != nil {
				    return response, fmt.Errorf("%+v, get %+v, %+v", ErrExpectNoError, err.Code, err.Message)
				}
			} else {
				if step.output.error.code != err.Code {
					return response, fmt.Errorf("%+v, get %+v", ErrWrongCode, err.Code)
				}
				if step.output.error.message != err.Message {
					return response, fmt.Errorf("%+v, get %+v", ErrWrongMessage, err.Message)
				}
			}
			// check output
			// if output is empty list then continue
			response = data.(map[string]interface{})
			for key, expectedResponse := range step.output.response {
				if returnedResponse, ok := response[key]; !ok {
					return response, ErrResponseNotFound
				} else {
					if !reflect.DeepEqual(expectedResponse, returnedResponse) {
						return response, fmt.Errorf("%+v, get %+v", ErrWrongExpectedResponse, returnedResponse)
					}
				}
			}
		}
	}
	return response, returnedErr
}