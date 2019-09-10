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
	//var nodefile = "sample-config.json"
	var testcase = "testcase.json"
	var err error
	// -1: beacon
	// 0, 1: shard
	var shardIndex = []int{-1, 0, 1}
	var nodeFileData = make(map[string]interface{})
	var testcaseData = make(map[string]interface{})
	testcaseBytes, err := ioutil.ReadFile(testcase)
	if err != nil {
		panic("Failed to get config file")
	}
	err = json.Unmarshal(testcaseBytes, &testcaseData)
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
					host: nodeInfoData["host"].(string),
					port: nodeInfoData["port"].(string),
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
				log.Println("Failed to load init testcase")
				os.Exit(1)
			}
			initTestcase = append(initTestcase, temp)
		}

		log.Println("Begin to run Init Testcase")
		for _, initTestcaseName := range initTestcase {
			cmd := exec.Command("go", "test", "-run", initTestcaseName)
			msg, err := cmd.Output()
			if err != nil {
				panic("Failed to run test " + initTestcaseName + " : " + string(msg))
			} else {
				log.Printf("%+v Message: %+v \n", initTestcaseName, string(msg))
			}
		}
	case "blockchain":
		log.Println("Begin to run testcase Blockchain")
	case "transaction":
		tempTransactionTestcase, ok := testcaseData["transaction"]
		if !ok {
			log.Println("Failed to load transaction testcase")
			os.Exit(0)
		}
		var transactionTestcase = []string{}
		for _, value := range tempTransactionTestcase.([]interface{}) {
			temp, ok := value.(string)
			if !ok {
				log.Println("Failed to load transaction testcase")
				os.Exit(1)
			}
			transactionTestcase = append(transactionTestcase, temp)
		}
		log.Println("Begin to run Transaction Testcase")
		for _, transactionTestcaseName := range transactionTestcase {
			log.Printf("Begin to run %+v Testcase \n", transactionTestcaseName)
			cmd := exec.Command("go", "test", "-run", transactionTestcaseName)
			msg, err := cmd.Output()
			if err != nil {
				panic("Failed to run test " + transactionTestcaseName + " : " + string(msg))
			} else {
				log.Printf("%+v Message: %+v \n", transactionTestcaseName, string(msg))
			}
		}
	case "crossshard":
		tempCrossShardTestcase, ok := testcaseData["crossshard"]
		if !ok {
			log.Println("Failed to load crossshard testcase")
			os.Exit(0)
		}
		var crossShardTestcase = []string{}
		for _, value := range tempCrossShardTestcase.([]interface{}) {
			temp, ok := value.(string)
			if !ok {
				log.Println("Failed to get crossshard testcase")
				os.Exit(1)
			}
			crossShardTestcase = append(crossShardTestcase, temp)
		}

		log.Println("Begin to run Crossshard Testcase")
		for _, crossShardTestcaseName := range crossShardTestcase {
			log.Printf("Begin to run %+v Testcase \n", crossShardTestcaseName)
			cmd := exec.Command("go", "test", "-run", crossShardTestcaseName)
			msg, err := cmd.Output()
			if err != nil {
				panic("Failed to run test " + crossShardTestcaseName + " : " + string(msg))
			} else {
				log.Printf("%+v Message: %+v \n", crossShardTestcaseName, string(msg))
			}
		}
	case "stake":
		tempStakeTestcase, ok := testcaseData["stake"]
		if !ok {
			log.Println("Failed to get stake testcase")
			os.Exit(0)
		}
		var stakeTestcase = []string{}
		for _, value := range tempStakeTestcase.([]interface{}) {
			temp, ok := value.(string)
			if !ok {
				log.Println("Failed to get stake testcase")
				os.Exit(1)
			}
			stakeTestcase = append(stakeTestcase, temp)
		}

		log.Println("Begin to run Stake Testcase")
		for _, stakeTestcaseName := range stakeTestcase {
			log.Printf("Begin to run %+v Testcase \n", stakeTestcaseName)
			cmd := exec.Command("go", "test", "-run", stakeTestcaseName)
			msg, err := cmd.Output()
			if err != nil {
				panic("Failed to run test " + stakeTestcaseName + " : " + string(msg))
			} else {
				log.Printf("%+v Message: %+v \n", stakeTestcaseName, string(msg))
			}
		}
	default:
		log.Println("Please choose the right test to run")
	}
}
