package bft

import (
	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/wire"
)

//
//type ProposeMsg struct {
//	ChainKey   string
//	Block      consensus.BlockInterface
//	ContentSig string
//	Pubkey     string
//	Timestamp  int64
//	RoundKey   string
//}
//
//type PrepareMsg struct {
//	ChainKey   string
//	IsOk       bool
//	Pubkey     string
//	ContentSig string
//	BlkHash    string
//	RoundKey   string
//	Timestamp  int64
//}

func MakeBFTProposeMsg(block, chainkey, roundkey string, userKeySet *cashec.KeySet) (wire.Message, error) {
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFTPropose)
	msg.(*wire.MessageBFTProposeV2).Block = block
	msg.(*wire.MessageBFTProposeV2).ChainKey = chainkey
	msg.(*wire.MessageBFTProposeV2).Pubkey = userKeySet.GetPublicKeyB58()
	msg.(*wire.MessageBFTProposeV2).RoundKey = roundkey
	err := msg.(*wire.MessageBFTProposeV2).SignMsg(userKeySet)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

func MakeBFTPrepareMsg(isOK bool, chainKey, blkHash, roundKey string, userKeySet *cashec.KeySet) (wire.Message, error) {
	msg, _ := wire.MakeEmptyMessage(wire.CmdBFTPrepare)
	msg.(*wire.MessageBFTPrepareV2).IsOk = isOK
	msg.(*wire.MessageBFTPrepareV2).ChainKey = chainKey
	msg.(*wire.MessageBFTPrepareV2).BlkHash = blkHash
	msg.(*wire.MessageBFTPrepareV2).Pubkey = userKeySet.GetPublicKeyB58()
	msg.(*wire.MessageBFTPrepareV2).RoundKey = roundKey
	err := msg.(*wire.MessageBFTProposeV2).SignMsg(userKeySet)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
