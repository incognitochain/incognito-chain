package bft

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/wire"
)

func MakeBFTProposeMsg(block, chainkey, roundkey string, userKeySet *incognitokey.KeySet) (wire.Message, error) {
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

func MakeBFTPrepareMsg(isOK bool, chainKey, blkHash, roundKey string, userKeySet *incognitokey.KeySet) (wire.Message, error) {
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
