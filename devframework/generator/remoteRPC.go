package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

const RPC_TEMPLATE = `
func (r *RemoteRPCClient) RPC_%APINAME%(%APIPARAMS%) (%APIRESULT%) {
	%API_REQUEST%
	%API_REQUEST_RESPONSE%
	%API_RETURN%
}
`

const RPC_REQUEST_TEMPLATE = `requestBody, rpcERR := json.Marshal(map[string]interface{}{
		"jsonrpc": "1.0",
		"method":  "%API_REQ_NAME%",
		"params":   []interface{}{%API_REQ_PARAMS%},
		"id":      1,
	})
	if err != nil {
		%API_RET_ERR%
	}
	body, err := r.sendRequest(requestBody)
	if err != nil {
		%API_RET_ERR%
	}`

const RPC_RESPONSE_TEMPLATE = `resp := struct {
		Result  %API_RES_TYPE%
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		%API_RET_ERR%
	}`

func main() {
	fd, _ := os.Open("apispec.go")
	b, _ := ioutil.ReadAll(fd)
	apis := strings.Split(string(b), "\n")

	apiF, _ := os.OpenFile("../remoteRPCClient.go", os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_TRUNC, 0666)
	apiF.Truncate(0)
	apiF.WriteString(`package devframework //This file is auto generated. Please do not change if you dont know what you are doing
import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type RemoteRPCClient struct {
	endpoint string
}

func (s *RemoteRPCClient) sendRequest(requestBody []byte) ([]byte, error) {
	resp, err := http.Post(s.endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
`)

	for _, api := range apis {
		regex := regexp.MustCompile(`([^ ]+)\((.+)\)[ ]*\(([ ]*([^, ]*)(error|,error|, error))\)`)

		res := regex.FindAllStringSubmatch(api, -1)
		if strings.Trim(api, " ") == "" {
			continue
		}
		if len(res) <= 0 {
			continue
		}

		apiName := strings.Trim(strings.ToLower(res[0][1]), "\t ")
		apiParams := []string{}

		for _, param := range strings.Split(res[0][2], ",") {
			trimParam := strings.Trim(param, " ")
			regex := regexp.MustCompile(`(.+) (.+)`)
			paramStruct := regex.FindAllStringSubmatch(trimParam, -1)
			apiParams = append(apiParams, paramStruct[0][1])
		}

		errstr := "return errors.New(rpcERR.Error())"
		hasResType := false
		resType := ""
		if len(strings.Split(res[0][3], ",")) > 1 {
			errstr = "return res,errors.New(rpcERR.Error())"
			hasResType = true
			resType = strings.Split(res[0][3], ",")[0]
		}

		//build request body
		reqstr := strings.Replace(RPC_REQUEST_TEMPLATE, "%API_REQ_NAME%", strings.ToLower(apiName), -1)
		reqstr = strings.Replace(reqstr, "%API_REQ_PARAMS%", strings.Join(apiParams, ","), -1)
		reqstr = strings.Replace(reqstr, "%API_RET_ERR%", errstr, -1)

		//build request response
		resstr := "_=body"
		if hasResType {
			resstr = strings.Replace(RPC_RESPONSE_TEMPLATE, "%API_RES_TYPE%", resType, -1)
			resstr = strings.Replace(resstr, "%API_RET_ERR%", errstr, -1)
		}

		//build return
		apiresult := "err error"
		retstr := "return err"
		if hasResType {
			retstr = "return resp.Result,err"
			apiresult = "res " + res[0][4] + ",err error"
		}

		//build function
		fgen := strings.Replace(RPC_TEMPLATE, "%APINAME%", apiName, -1)
		fgen = strings.Replace(fgen, "%APIPARAMS%", res[0][2], -1)
		fgen = strings.Replace(fgen, "%APIRESULT%", apiresult, -1)
		fgen = strings.Replace(fgen, "%API_REQUEST%", reqstr, -1)
		fgen = strings.Replace(fgen, "%API_REQUEST_RESPONSE%", resstr, -1)
		fgen = strings.Replace(fgen, "%API_RETURN%", retstr, -1)
		apiF.WriteString("\n" + fgen)
	}
	apiF.Close()
	fd.Close()
}
