package blsbft

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/consensus/multisigschemes/bls"
	"github.com/incognitochain/incognito-chain/wire"
)

type BFTPropose struct {
	Block json.RawMessage
}

type BFTAgree struct {
	RoundKey  string
	Validator string
	Sig       string
}

func MakeBFTProposeMsg(block []byte, chainKey string, userKeySet *bls.KeySet) (wire.Message, error) {
	var proposeCtn BFTPropose
	proposeCtn.Block = block
	proposeCtnBytes, err := json.Marshal(proposeCtn)
	if err != nil {
		return nil, err
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = proposeCtnBytes
	return msg, nil
}

func MakeBFTAgreeMsg(userPubKey, chainKey, sig, roundKey string) (wire.Message, error) {
	var agreeCtn BFTAgree
	agreeCtn.RoundKey = roundKey
	agreeCtn.Validator = userPubKey
	agreeCtn.Sig = sig
	agreeCtnBytes, err := json.Marshal(agreeCtn)
	if err != nil {
		return nil, err
	}
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFT)
	msg.(*wire.MessageBFT).ChainKey = chainKey
	msg.(*wire.MessageBFT).Content = agreeCtnBytes
	return msg, nil
}

func (e *BLSBFT) ProcessBFTMsg(msg *wire.MessageBFT) {
	return
}
