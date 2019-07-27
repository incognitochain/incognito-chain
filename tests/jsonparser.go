package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)
type step struct {
	input struct {
		name string
		params []interface{}
		isWait bool
		wait time.Duration
		conn string
	}
	output struct {
		error struct {
			isNil bool
			code int
			message string
		}
		response map[string]interface{}
	}
}
func newStep() *step {
	step := &step{}
	step.input.name = ""
	step.input.params = []interface{}{}
	step.input.wait = time.Duration(0*time.Second)
	step.input.isWait = false
	step.input.conn = "http"
	step.output.error.isNil = true
	step.output.response = make(map[string]interface{})
	return step
}

func readfile(filename string) ([]*step, error) {
	var  (
		err error
		ok bool
		data []byte
		tests []map[string]interface{}
		sc []*step
	)
	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return sc, err
	}
	err = json.Unmarshal(data, &tests)
	log.Println(tests)
	sc,ok = parseScenarios(tests)
	if !ok {
		return sc, fmt.Errorf("Parse file %+v error", filename)
	}
	return sc, nil
}

func parseScenarios(tests []map[string]interface{}) ([]*step,bool) {
	sc := []*step{}
	for _, tests := range tests {
		step := newStep()
		if inputData, ok := tests["input"]; !ok {
			return sc, false
		} else {
			if input, ok := inputData.(map[string]interface{}); !ok {
				return sc, false
			} else {
				step.input.name = input["command"].(string)
				if params, ok := input["params"]; !ok {
					return sc, false
				} else {
					//paramsBytes, err := json.Marshal(params)
					//if err != nil {
					//	return sc, false
					//}
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
						if err, ok := errData.(map[string]interface{}); !ok {
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
		sc = append(sc, step)
	}
	return sc, true
}