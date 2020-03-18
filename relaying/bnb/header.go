package bnb

import (
	"encoding/json"
	"fmt"
	"github.com/tendermint/tendermint/types"
)

func ParseHeaderFromJson(object string) (*types.Header, *BNBRelayingError) {
	header := types.Header{}
	err := json.Unmarshal([]byte(object), &header)
	if err != nil {
		fmt.Printf("err unmarshal: %+v\n", err)
	}

	//fmt.Printf("header: %+v\n", header)
	//fmt.Printf("header: %+v\n", header)

	return &header, nil
}

func ParseCommitFromJson(object string) (*types.Commit, *BNBRelayingError) {
	commit := types.Commit{}
	err := json.Unmarshal([]byte(object), &commit)
	if err != nil {
		fmt.Printf("err unmarshal: %+v\n", err)
	}

	//fmt.Printf("commit: %+v\n", commit)
	//fmt.Printf("commit: %+v\n", commit.Precommits[0].Signature)
	return &commit, nil
}