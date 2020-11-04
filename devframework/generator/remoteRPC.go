package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

const RPC_TEMPLATE = `func (sim *SimulationEngine) rpc_%APINAME%(%APIPARAMS%) (%APIRESULT%) {
	%API_REQUEST%
	%API_REQUEST_RESPONSE%
	%API_RETURN%
}`

const RPC_REQUEST_TEMPLATE = `requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "%API_REQ_NAME%",
		"params":   []interface{}{%API_REQ_PARAMS%},
		"id":      1,
	})
	if err != nil {
		return %API_RET_ERR%
	}
	body, err := sendRequest(requestBody)
	if err != nil {
		return %API_RET_ERR%
	}`

const RPC_RESPONSE_TEMPLATE = `resp := struct {
		Result  %API_RES_TYPE%
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return %API_RET_ERR%
	}`

func client() {
	fd, _ := os.Open("api.json")
	b, _ := ioutil.ReadAll(fd)

	var apis []API
	json.Unmarshal(b, &apis)

	apiF, _ := os.OpenFile("../rpc.go", os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_TRUNC, 0666)
	apiF.Truncate(0)
	apiF.WriteString(`package devframework
import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)`)

	for _, api := range apis {
		apiname := api.Name
		apiparams := []string{}
		for _, param := range api.Params {
			apiparams = append(apiparams, param.Name+" "+param.Type)
		}

		apierrreturn := "err"
		apiresult := "error"
		if api.Result != "" {
			apiresult = api.Result + ",error"
			apierrreturn = "nil,err"
		}

		reqparams := []string{}
		for _, param := range api.Params {
			reqparams = append(reqparams, param.Name)
		}

		//build request body
		reqstr := strings.Replace(RPC_REQUEST_TEMPLATE, "%API_REQ_NAME%", strings.ToLower(apiname), -1)
		reqstr = strings.Replace(reqstr, "%API_REQ_PARAMS%", strings.Join(reqparams, ","), -1)
		reqstr = strings.Replace(reqstr, "%API_RET_ERR%", apierrreturn, -1)

		//build request response
		resstr := "_=body"
		if api.Result != "" {
			resstr = strings.Replace(RPC_RESPONSE_TEMPLATE, "%API_RES_TYPE%", api.Result, -1)
			resstr = strings.Replace(resstr, "%API_RET_ERR%", apierrreturn, -1)
		}

		//build return
		retstr := "return err"
		if api.Result != "" {
			retstr = "return resp.Result,err"
		}
		//build function
		fgen := strings.Replace(RPC_TEMPLATE, "%APINAME%", apiname, -1)
		fgen = strings.Replace(fgen, "%APIPARAMS%", strings.Join(apiparams, ","), -1)
		fgen = strings.Replace(fgen, "%APIRESULT%", apiresult, -1)
		fgen = strings.Replace(fgen, "%RPC_REQUEST%", reqstr, -1)
		fgen = strings.Replace(fgen, "%RPC_REQUEST_RESPONSE%", resstr, -1)
		fgen = strings.Replace(fgen, "%RPC_RETURN%", retstr, -1)
		apiF.WriteString("\n" + fgen)

	}
	apiF.Close()
	fd.Close()
}
