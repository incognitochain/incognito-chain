package main

import (
	"testing"
)

func TestSwapBridge(t *testing.T) {
	// Remember to set shardID to 1 in calculateCandidateShardID
	_, err := executeTest("./testsdata/bridge/swapbridge.json")
	checkError(t, err)
}

func TestSwapBeacon(t *testing.T) {
	_, err := executeTest("./testsdata/bridge/swapbeacon.json")
	checkError(t, err)
}
