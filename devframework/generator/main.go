package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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

const APITEMPLATE = `func (sim *SimulationEngine) rpc_%API_NAME%(%API_PARAMS%) (%API_RESULT%) {
	httpServer := sim.rpcServer.HttpServer
	c := rpcserver.HttpHandler["%API_NAME%"]
	resI, err := c(httpServer, []interface{}{%API_PARAM_REQ%}, nil)
	if err != nil {
		%API_ERR%
	}
	%API_RETURN%
}`

func main() {
	fd, _ := os.Open("api.spec")
	b, _ := ioutil.ReadAll(fd)

	apis := strings.Split(string(b), "\n")

	apiF, _ := os.OpenFile("../rpc.go", os.O_CREATE|os.O_RDWR|os.O_APPEND|os.O_TRUNC, 0666)
	apiF.Truncate(0)
	apiF.WriteString(`package devframework

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)`)
	//
	for _, api := range apis {
		regex := regexp.MustCompile(`(.+)\((.+)\)[ ]*\([ ]*([^, ]*)(error|,error)\)`)
		res := regex.FindAllStringSubmatch(api, -1)
		if strings.Trim(api, " ") == "" {
			continue
		}
		fmt.Println(res[0])
		apiName := strings.ToLower(res[0][1])
		apiParams := []string{}
		for _, param := range strings.Split(res[0][2], ",") {
			trimParam := strings.Trim(param, " ")
			regex := regexp.MustCompile(`(.+) (.+)`)
			paramStruct := regex.FindAllStringSubmatch(trimParam, -1)
			apiParams = append(apiParams, paramStruct[0][1])
		}
		apiResultType := ""
		if len(res[0]) == 5 {
			apiResultType = strings.Trim(res[0][3], " ")
		}

		//fmt.Println(apiName)
		//fmt.Println(strings.Join(apiParams, ","))
		//fmt.Println(apiResultType)

		//build return
		retstr := "_ = resI \n return err"
		resstr := "error"
		errstr := "return err"
		if apiResultType != "" {
			retstr = "return resI.(" + apiResultType + "),err"
			resstr = apiResultType + ",error"
			errstr = "return nil,err"
		}
		//build function
		fgen := strings.Replace(APITEMPLATE, "%API_NAME%", apiName, -1)
		fgen = strings.Replace(fgen, "%API_PARAMS%", res[0][2], -1)
		fgen = strings.Replace(fgen, "%API_RESULT%", resstr, -1)
		fgen = strings.Replace(fgen, "%API_PARAM_REQ%", strings.Join(apiParams, ","), -1)
		fgen = strings.Replace(fgen, "%API_ERR%", errstr, -1)
		fgen = strings.Replace(fgen, "%API_RETURN%", retstr, -1)
		apiF.WriteString("\n" + fgen)

	}
	apiF.Close()
	fd.Close()
}
