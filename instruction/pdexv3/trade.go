package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

type Content interface {
	GetType() int
	GetStatus() string
}

type Action struct {
	Content
	RequestTxID common.Hash `json:"RequestTxID"`
	shardID     byte
}

func (acn Action) ShardID() byte { return acn.shardID }

func (acn *Action) FromStrings(source []string) error {
	if len(source) != 4 {
		return errors.New("Invalid action length")
	}
	err := json.Unmarshal([]byte(source[3]), acn)
	if err != nil {
		return err
	}

	// Content must be of the correct concrete type before entering FromStrings()
	if mdType, err := strconv.Atoi(source[1]); err != nil || mdType != int(acn.GetType()) {
		return fmt.Errorf("Metadata type mismatch")
	}

	shardID, err := strconv.Atoi(source[2])
	acn.shardID = byte(shardID)

	return nil
}

func (acn *Action) Strings() []string {
	result := []string{strconv.Itoa(acn.GetType()), acn.GetStatus(), strconv.Itoa(int(acn.shardID))}
	jsonBytes, _ := json.Marshal(acn)
	result = append(result, string(jsonBytes))
	return result
}
