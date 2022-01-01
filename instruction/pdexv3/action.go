package pdexv3

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

type Content interface {
	GetType() int
	GetStatus() int
}

type Action struct {
	Content
	requestTxID common.Hash
	shardID     byte
}

func NewAction(c Content, txID common.Hash, shardID byte) *Action {
	return &Action{
		Content:     c,
		requestTxID: txID,
		shardID:     shardID,
	}
}

func (acn *Action) ShardID() byte            { return acn.shardID }
func (acn *Action) RequestTxID() common.Hash { return acn.requestTxID }

func (acn *Action) FromStringSlice(source []string) error {
	if len(source) != 5 {
		return fmt.Errorf("Invalid action length %d, expect 5", len(source))
	}
	err := json.Unmarshal([]byte(source[4]), acn)
	if err != nil {
		return err
	}

	// Content must be of the correct concrete type before entering FromStrings()
	if mdType, err := strconv.Atoi(source[0]); err != nil || mdType != int(acn.GetType()) {
		return fmt.Errorf("Metadata type mismatch")
	}
	if status, err := strconv.Atoi(source[1]); err != nil || status != int(acn.GetStatus()) {
		return fmt.Errorf("Metadata status mismatch")
	}

	if h, err := acn.requestTxID.NewHashFromStr(source[3]); err != nil {
		return fmt.Errorf("Invalid RequestTxID %v", source[3])
	} else {
		acn.requestTxID = *h
	}

	if res, err := strconv.Atoi(source[2]); err != nil {
		return fmt.Errorf("Invalid shardID %v", source[2])
	} else {
		acn.shardID = byte(res)
	}

	return nil
}

func (acn *Action) StringSlice() []string {
	result := []string{strconv.Itoa(acn.GetType()), strconv.Itoa(acn.GetStatus()),
		strconv.Itoa(int(acn.shardID)), acn.requestTxID.String()}
	jsonBytes, _ := json.Marshal(acn)
	result = append(result, string(jsonBytes))
	return result
}
