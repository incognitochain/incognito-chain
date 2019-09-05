package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type step struct {
	client *Client
	input  struct {
		name        string
		fromContext bool
		params      []interface{}
		isWait      bool
		wait        time.Duration
		conn        string
	}
	output struct {
		error struct {
			isNil   bool
			code    int
			message string
		}
		response interface{}
	}
	store map[string]string
}
type scenarios struct {
	steps   []*step
	context map[string]interface{}
}

func newStep() *step {
	step := &step{}
	step.client = newClient()
	step.input.name = ""
	step.input.params = []interface{}{}
	step.input.wait = time.Duration(0 * time.Second)
	step.input.isWait = false
	step.input.conn = "http"
	step.output.error.isNil = true
	step.store = make(map[string]string)
	return step
}
func newScenarios() *scenarios {
	return &scenarios{
		steps:   []*step{},
		context: make(map[string]interface{}),
	}
}
func readfile(filename string) (*scenarios, error) {
	var (
		err      error
		ok       bool
		data     []byte
		testcase []map[string]interface{}
		sc       *scenarios
	)
	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return sc, err
	}
	err = json.Unmarshal(data, &testcase)
	sc, ok = parseScenarios(testcase)
	if !ok {
		return sc, fmt.Errorf("Parse file %+v error", filename)
	}
	return sc, nil
}

func parseScenarios(tests []map[string]interface{}) (*scenarios, bool) {
	sc := newScenarios()
	env := os.Getenv("ENV")
	nodeList, err := readNodeConfig(env)
	if err != nil {
		return sc, false
	}
	for _, tests := range tests {
		step := newStep()
		if nodeData, ok := tests["node"]; !ok {
			return sc, false
		} else {
			node, ok := nodeData.(string)
			if !ok {
				return sc, false
			}
			c, ok := nodeList[node]
			if !ok {
				return sc, false
			}
			step.client = c
		}
		if inputData, ok := tests["input"]; !ok {
			return sc, false
		} else {
			if input, ok := inputData.(map[string]interface{}); !ok {
				return sc, false
			} else {
				step.input.name = input["command"].(string)
				if fromContext, ok := input["context"]; !ok {
					step.input.fromContext = false
				} else {
					step.input.fromContext = fromContext.(bool)
				}
				if params, ok := input["params"]; !ok {
					return sc, false
				} else {
					step.input.params = params.([]interface{})
				}
				if wait, ok := input["wait"]; !ok {
					step.input.isWait = false
				} else {
					step.input.isWait = true
					step.input.wait = time.Second * time.Duration(int64(wait.(float64)))
				}
				if conn, ok := input["type"]; !ok {
					step.input.conn = "http"
				} else {
					step.input.conn = conn.(string)
				}
			}
		}
		if outputData, ok := tests["output"]; !ok {
			return sc, false
		} else {
			if output, ok := outputData.(map[string]interface{}); !ok {
				return sc, false
			} else {
				if errData, ok := output["error"]; !ok {
					return sc, false
				} else {
					if errData == nil {
						step.output.error.isNil = true
					} else {
						if err, ok := errData.(map[string]interface{}); ok {
							step.output.error.isNil = false
							if code, ok := err["code"]; !ok {
								return sc, false
							} else {
								step.output.error.code = int(code.(float64))
							}
							if message, ok := err["message"]; !ok {
								return sc, false
							} else {
								step.output.error.message = message.(string)
							}
						}
					}
				}
				if response, ok := output["response"]; !ok {
					step.output.response = make(map[string]interface{})
				} else {
					step.output.response = response
				}
			}
		}
		if storeData, ok := tests["store"]; ok {
			if store, ok := storeData.(map[string]interface{}); ok {
				for key, value := range store {
					if _, ok := value.(string); !ok {
						return sc, false
					} else {
						step.store[key] = value.(string)
					}
				}
			} else {
				// not return object => store all data in Result of Response
				if store, ok := storeData.(string); !ok {
					return sc, false
				} else {
					step.store[store] = ""
				}
			}
		}
		sc.steps = append(sc.steps, step)
	}
	return sc, true

}
func readNodeConfig(env string) (map[string]*Client, error) {
	var nodeList = make(map[string]*Client)
	var fileNodeInterface = make(map[string]interface{})
	var fileName = ""
	switch env {
	case "testnet":
		fileName = "./testsconfig/testnet-config.json"
	case "burn":
		fileName = "./testsconfig/burn-config.json"
	default:
		fileName = "./testsconfig/sample-config.json"
	}
	fileNodeData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(fileNodeData, &fileNodeInterface)
	if err != nil {
		return nil, err
	}
	for shardNodeKey, shardNodeValueInterface := range fileNodeInterface {
		clientKey := ""
		switch shardNodeKey {
		case "-1":
			clientKey = "beacon"
		case "0":
			clientKey = "shard0-"
		case "1":
			clientKey = "shard1-"
		}
		shardNodeValue, ok := shardNodeValueInterface.(map[string]interface{})
		if !ok {
			return nil, ParseNodeConfigFailedError
		}
		for nodeKey, nodeValueInterface := range shardNodeValue {
			clientNodeKey := clientKey + nodeKey
			nodeValue, ok := nodeValueInterface.(map[string]interface{})
			if !ok {
				return nil, ParseFailedError
			}
			host, ok := nodeValue["host"].(string)
			if !ok {
				return nil, ParseHostError
			}
			rpcport, ok := nodeValue["port"].(string)
			if !ok {
				return nil, ParsePortError
			}
			wsport, ok := nodeValue["ws"].(string)
			if !ok {
				return nil, ParseHostError
			}
			c := newClientWithFullInform(host, rpcport, wsport)
			nodeList[clientNodeKey] = c
		}
	}
	return nodeList, nil
}

/*
	Type
	- Number: float64
	- String: string
	- Boolean: bool
	- Array: []interface
	- Object: map[string]interface{}
*/
func parseResult(responseResult json.RawMessage) interface{} {
	var (
		number  float64
		str     string
		boolean bool
		array   []interface{}
		obj     = make(map[string]interface{})
	)
	if err := json.Unmarshal(responseResult, &number); err == nil {
		return number
	}
	if err := json.Unmarshal(responseResult, &str); err == nil {
		return str
	}
	if err := json.Unmarshal(responseResult, &boolean); err == nil {
		return boolean
	}
	if err := json.Unmarshal(responseResult, &array); err == nil {
		return array
	}
	if err := json.Unmarshal(responseResult, &obj); err == nil {
		return obj
	}
	return nil
}
