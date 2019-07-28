package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		response map[string]interface{}
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
	step.output.response = make(map[string]interface{})
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
	log.Println(testcase)
	sc, ok = parseScenarios(testcase)
	if !ok {
		return sc, fmt.Errorf("Parse file %+v error", filename)
	}
	return sc, nil
}

func parseScenarios(tests []map[string]interface{}) (*scenarios, bool) {
	sc := newScenarios()
	for _, tests := range tests {
		step := newStep()
		if nodeData, ok := tests["node"]; !ok {
			return sc, false
		} else {
			if node, ok := nodeData.(map[string]interface{}); !ok {
				return sc, false
			} else {
				host := node["host"].(string)
				port := node["port"].(string)
				step.client = newClientWithHost(host, port)
			}
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
					return sc, false
				} else {
					step.output.response = response.(map[string]interface{})
				}
			}
		}
		if storeData, ok := tests["store"]; ok {
			if store, ok := storeData.(map[string]interface{}); !ok {
				return sc, false
			} else {
				for key, value := range store {
					if _, ok := value.(string); !ok {
						return sc, false
					} else {
						step.store[key] = value.(string)
					}
				}
			}
		}
		sc.steps = append(sc.steps, step)
	}
	return sc, true
}
