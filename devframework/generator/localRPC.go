package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/incognitochain/incognito-chain/rpcserver"
)

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Note string `json:"note"`
}
type API struct {
	Name   string  `json:"name"`
	Params []Param `json:"params"`
	Result string  `json:"result"`
}

const APITEMPLATE = `func (r *LocalRPCClient) RPC_%API_NAME%(%API_PARAMS%) (%API_RESULT%) {
	httpServer := r.rpcServer.HttpServer
	c := rpcserver.%HANDLER%["%API_NAME%"]
	resI, rpcERR := c(httpServer, []interface{}{%API_PARAM_REQ%}, nil)
	if rpcERR != nil {
		%API_ERR%
	}
	%API_RETURN%
}`

func main() {
	fd, _ := os.Open("apispec.go")
	b, _ := ioutil.ReadAll(fd)

	apis := strings.Split(string(b), "\n")

	apiF, _ := os.OpenFile("../localRPCClient.go", os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_TRUNC, 0666)
	apiF.Truncate(0)
	apiF.WriteString(`package devframework
//This file is auto generated. Please do not change if you dont know what you are doing
import (
	"errors"
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)
type LocalRPCClient struct {
	rpcServer *rpcserver.RpcServer
}`)
	//
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
		apiResultType := ""
		if len(strings.Split(res[0][3], ",")) > 1 {
			apiResultType = strings.Split(res[0][3], ",")[0]
		}

		rpchandler := "HttpHandler"
		if _, ok := rpcserver.HttpHandler[apiName]; !ok {
			if _, ok := rpcserver.LimitedHttpHandler[apiName]; !ok {
				panic("no method name in rpc request" + apiName)
			} else {
				rpchandler = "LimitedHttpHandler"
			}
		}
		//build return
		retstr := "_ = resI \n return nil"
		resstr := "error"
		errstr := "return errors.New(rpcERR.Error())"
		if apiResultType != "" {
			retstr = "return resI.(" + apiResultType + "),nil"
			resstr = "res " + apiResultType + ",err error"
			errstr = "return res,errors.New(rpcERR.Error())"
		}
		//build function
		fgen := strings.Replace(APITEMPLATE, "%API_NAME%", apiName, -1)
		fgen = strings.Replace(fgen, "%API_PARAMS%", res[0][2], -1)
		fgen = strings.Replace(fgen, "%API_RESULT%", resstr, -1)
		fgen = strings.Replace(fgen, "%API_PARAM_REQ%", strings.Join(apiParams, ","), -1)
		fgen = strings.Replace(fgen, "%API_ERR%", errstr, -1)
		fgen = strings.Replace(fgen, "%API_RETURN%", retstr, -1)
		fgen = strings.Replace(fgen, "%HANDLER%", rpchandler, -1)

		apiF.WriteString("\n" + fgen)

	}
	apiF.Close()
	fd.Close()
}
