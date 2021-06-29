package main

import (
	"os"
	"time"
)

var (
	shard0UrlListWithBeacon = []string{
		"http://localhost:9334",
		"http://localhost:9335",
		"http://localhost:9336",
		"http://localhost:9337",
		"http://localhost:9331",
		"http://localhost:9332",
		"http://localhost:9333",
		"http://localhost:9349",
		"http://localhost:9350",
		"http://localhost:9351",
		"http://localhost:9352",
		"http://localhost:9353",
	}
	shard1UrlListWithBeacon = []string{
		"http://localhost:9338",
		"http://localhost:9339",
		"http://localhost:9340",
		"http://localhost:9341",
		"http://localhost:9342",
		"http://localhost:9343",
		"http://localhost:9344",
		"http://localhost:9345",
		"http://localhost:9350",
		"http://localhost:9351",
		"http://localhost:9352",
		"http://localhost:9353",
	}
	shard0UrlList = []string{
		"http://localhost:9334",
		"http://localhost:9335",
		"http://localhost:9336",
		"http://localhost:9337",
		"http://localhost:9331",
		"http://localhost:9332",
		"http://localhost:9333",
		"http://localhost:9349",
	}
	shard1UrlList = []string{
		"http://localhost:9338",
		"http://localhost:9339",
		"http://localhost:9340",
		"http://localhost:9341",
		"http://localhost:9342",
		"http://localhost:9343",
		"http://localhost:9344",
		"http://localhost:9345",
	}
)

func main() {
	if os.Args[1] == "submit" || os.Args[1] == "all" {
		submitKeyShard0_0()
		submitKeyShard0_1()
		submitKeyShard0_2()
		submitKeyShard0_3()
		submitKeyShard1_1()
		submitKeyShard1_2()
		submitKeyShard1_3()
	}
	if os.Args[1] == "convert" || os.Args[1] == "all" {
		convertShard0_0()
		convertShard0_1()
		convertShard0_2()
		convertShard0_3()
		convertShard1_1()
		convertShard1_2()
		convertShard1_3()
	}
	if os.Args[1] == "send-tx" || os.Args[1] == "all" {
		sendTransactionFromTestnetGenesisKeyFromShard0_0()
		sendTransactionFromTestnetGenesisKeyFromShard0_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_0()
		sendTransactionFromTestnetGenesisKeyFromShard1_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_2()
		sendTransactionToShard1()
	}
	if os.Args[1] == "flush-tx" {
		flushTx()
	}
}

func flushTx() {
	ticker := time.Tick(500 * time.Millisecond)
	for _ = range ticker {
		sendTransactionFromTestnetGenesisKeyFromShard0_0()
		sendTransactionFromTestnetGenesisKeyFromShard0_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_0()
		sendTransactionFromTestnetGenesisKeyFromShard1_1()
		sendTransactionFromTestnetGenesisKeyFromShard1_2()
		sendTransactionToShard1()
	}
}
