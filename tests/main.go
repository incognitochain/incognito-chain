package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var (
	shard0 []*Client
	shard1 []*Client
	beacon []*Client
)

func main() {
	var nodefile = "sample-config.json"
	var testcase = "testcase.json"
	var err error
	// -1: beacon
	// 0, 1: shard
	var shardIndex = []int{-1, 0, 1}
	var nodeFileData = make(map[string]interface{})
	var testcaseData = make(map[string]interface{})
	if len(os.Args) == 3 {
		nodefile = os.Args[2]
	}
	testcaseBytes, err := ioutil.ReadFile(testcase)
	if err != nil {
		panic("Failed to get config file")
	}
	err = json.Unmarshal(testcaseBytes, &testcaseData)
	nodefileBytes, err := ioutil.ReadFile(nodefile)
	if err != nil {
		panic("Failed to get config file")
	}

	err = json.Unmarshal(nodefileBytes, &nodeFileData)
	if err != nil {
		panic("Failed to marshal config file")
	}
	for _, shard := range shardIndex {
		if nodes, ok := nodeFileData[strconv.Itoa(shard)]; ok {
			nodeInfos, ok := nodes.(map[string]interface{})
			if !ok {
				panic("Failed to read data from config file")
			}
			for _, nodeInfo := range nodeInfos {
				nodeInfoData, ok := nodeInfo.(map[string]interface{})
				if !ok {
					panic("Failed to read data from config file")
				}
				client := &Client{
					Host: nodeInfoData["host"].(string),
					Port: nodeInfoData["port"].(string),
				}
				if shard == -1 {
					beacon = append(beacon, client)
				}
				if shard == 0 {
					shard0 = append(shard0, client)
				}
				if shard == 1 {
					shard1 = append(shard1, client)
				}
			}
		}
	}
	switch os.Args[1] {
	case "all":
		fmt.Println("Test all")
	case "client":
		tempInitTestcase, ok := testcaseData["client"]
		if !ok {
			log.Println("Failed to get init testcase")
			os.Exit(0)
		}
		var initTestcase = []string{}
		for _, value := range tempInitTestcase.([]interface{}) {
			temp, ok := value.(string)
			if !ok {
				log.Println("Failed to get init testcase")
				os.Exit(1)
			}
			initTestcase = append(initTestcase, temp)
		}

		log.Println("Begin to run Init Testcase")
		for _, initTestcaseName := range initTestcase {
			cmd := exec.Command("go", "test", "-run", initTestcaseName)
			msg, err := cmd.Output()
			if err != nil {
				log.Printf("Failed to run test %+v, err %+v \n", initTestcaseName, err)
			} else {
				log.Printf("%+v Message: %+v \n", initTestcaseName, string(msg))
			}
		}
	case "blockchain":
		log.Println("Begin to run testcase Blockchain")
	case "transaction":
		tempInitTestcase, ok := testcaseData["transaction"]
		if !ok {
			log.Println("Failed to get transaction testcase")
			os.Exit(0)
		}
		var initTestcase = []string{}
		for _, value := range tempInitTestcase.([]interface{}) {
			temp, ok := value.(string)
			if !ok {
				log.Println("Failed to get transaction testcase")
				os.Exit(1)
			}
			initTestcase = append(initTestcase, temp)
		}
		
		log.Println("Begin to run Transaction Testcase")
		for _, initTestcaseName := range initTestcase {
			cmd := exec.Command("go", "test", "-run", initTestcaseName)
			msg, err := cmd.Output()
			if err != nil {
				log.Printf("Failed to run test %+v, err %+v \n", initTestcaseName, err)
			} else {
				log.Printf("%+v Message: %+v \n", initTestcaseName, string(msg))
			}
		}
	case "crossshard":
		tempInitTestcase, ok := testcaseData["crossshard"]
		if !ok {
			log.Println("Failed to get crossshard testcase")
			os.Exit(0)
		}
		var initTestcase = []string{}
		for _, value := range tempInitTestcase.([]interface{}) {
			temp, ok := value.(string)
			if !ok {
				log.Println("Failed to get crossshard testcase")
				os.Exit(1)
			}
			initTestcase = append(initTestcase, temp)
		}
		
		log.Println("Begin to run Crossshard Testcase")
		for _, initTestcaseName := range initTestcase {
			cmd := exec.Command("go", "test", "-run", initTestcaseName)
			msg, err := cmd.Output()
			if err != nil {
				log.Printf("Failed to run test %+v, err %+v \n", initTestcaseName, err)
			} else {
				log.Printf("%+v Message: %+v \n", initTestcaseName, string(msg))
			}
		}
	default:
		log.Println("Please choose the right test to run")
	}
}
