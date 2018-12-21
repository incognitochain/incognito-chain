package constantpos

import (
	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/wire"
)

// func (self *Engine) sendBlockMsg(block *blockchain.BlockV2) {
// 	blockMsg, err := wire.MakeEmptyMessage(wire.CmdBlock)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return
// 	}
// 	blockMsg.(*wire.MessageBlock).Block = *block
// 	self.config.Server.PushMessageToAll(blockMsg)
// }

// func (self *Engine) OnRequestSign(msgBlock *wire.MessageBlockSigReq) {
// 	block := &msgBlock.Block
// 	err := self.validatePreSignBlockSanity(block)
// 	if err != nil {
// 		invalidBlockMsg := &wire.MessageInvalidBlock{
// 			Reason:    err.Error(),
// 			BlockHash: block.Hash().String(),
// 			shardID:   block.Header.shardID,
// 			Validator: base58.Base58Check{}.Encode(self.config.UserKeySet.PaymentAddress.Pk, byte(0x00)),
// 		}
// 		dataByte, _ := invalidBlockMsg.JsonSerialize()
// 		invalidBlockMsg.ValidatorSig, err = self.config.UserKeySet.SignBase58(dataByte)
// 		if err != nil {
// 			Logger.log.Error(err)
// 			return
// 		}
// 		Logger.log.Critical("Invalid block msg", invalidBlockMsg)
// 		err = self.config.Server.PushMessageToAll(invalidBlockMsg)
// 		if err != nil {
// 			Logger.log.Error(err)
// 			return
// 		}
// 		return
// 	}

// 	sig, err := self.config.UserKeySet.SignBase58([]byte(block.Hash().String()))
// 	if err != nil {
// 		Logger.log.Error("Can't sign block ", err)
// 		// TODO something went terribly wrong
// 		return
// 	}
// 	blockSigMsg := wire.MessageBlockSig{
// 		Validator: base58.Base58Check{}.Encode(self.config.UserKeySet.PaymentAddress.Pk, byte(0x00)),
// 		BlockSig:  sig,
// 	}
// 	peerID, err := libp2p.IDB58Decode(msgBlock.SenderID)
// 	if err != nil {
// 		Logger.log.Error("ERROR", msgBlock.SenderID, peerID, err)
// 	}
// 	Logger.log.Info(block.Hash().String(), blockSigMsg)
// 	err = self.config.Server.PushMessageToPeer(&blockSigMsg, peerID)
// 	if err != nil {
// 		Logger.log.Error(err)
// 	}
// 	return
// }

// func (self *Engine) OnBlockReceived(block *blockchain.BlockV2) {
// 	// if self.config.BlockChain.BestState[block.Header.shardID].Height < block.Header.Height {
// 	// 	exists, err := self.config.BlockChain.BlockExists(block.Hash())
// 	// 	if err != nil {
// 	// 		Logger.log.Error(err)
// 	// 		return
// 	// 	} else {
// 	// 		if !exists {
// 	// 			err := self.validateBlockSanity(block)
// 	// 			if err != nil {
// 	// 				Logger.log.Error(err)
// 	// 				return
// 	// 			}
// 	// 			self.UpdateChain(block)
// 	// 		}
// 	// 	}
// 	// }
// 	return
// }

func (self *Engine) OnBFTPropose(msg *wire.MessageBFTPropose) {
	return
}

func (self *Engine) OnBFTPrepare(msg *wire.MessageBFTPrepare) {
	return

}

func (self *Engine) OnBFTCommit(msg *wire.MessageBFTCommit) {
	return

}

func (self *Engine) OnBFTReply(msg *wire.MessageBFTReply) {
	return

}

func (self *Engine) OnInvalidBlockReceived(blockHash string, shardID byte, reason string) {
	// leave empty for now
	Logger.log.Error(blockHash, shardID, reason)
	return
}

func (self *Engine) OnChainStateReceived(msg *wire.MessageChainState) {
	// chainInfo := msg.ChainInfo.(map[string]interface{})
	// for i, v := range self.validatedChainsHeight.Heights {
	// 	if chainInfo["ChainsHeight"] != nil {
	// 		if v < int(chainInfo["ChainsHeight"].([]interface{})[i].(float64)) {
	// 			self.knownChainsHeight.Heights[i] = int(chainInfo["ChainsHeight"].([]interface{})[i].(float64))
	// 			lastBlockHash := self.config.BlockChain.BestState[i].BestBlockHash.String()
	// 			getBlkMsg := &wire.MessageGetBlocks{
	// 				LastBlockHash: lastBlockHash,
	// 			}
	// 			Logger.log.Info("Send " + getBlkMsg.MessageType() + " to " + msg.SenderID)
	// 			peerID, err := libp2p.IDB58Decode(msg.SenderID)
	// 			if err != nil {
	// 				Logger.log.Error(err)
	// 				continue
	// 			}
	// 			self.config.Server.PushMessageToPeer(getBlkMsg, peerID)
	// 		}
	// 	} else {
	// 		Logger.log.Error("ChainsHeight is empty!")
	// 	}
	// }
	return
}

func (self *Engine) OnGetChainState(msg *wire.MessageGetChainState) {
	// newMsg, err := wire.MakeEmptyMessage(wire.CmdChainState)
	// if err != nil {
	// 	return
	// }
	// newMsg.(*wire.MessageChainState).ChainInfo = ChainInfo{
	// 	CurrentCommittee:        self.Committee().GetCommittee(),
	// 	CandidateListMerkleHash: common.EmptyString,
	// 	ChainsHeight:            self.validatedChainsHeight.Heights,
	// }
	// peerID, _ := libp2p.IDB58Decode(msg.SenderID)
	// self.config.Server.PushMessageToPeer(newMsg, peerID)
	return
}

func (self *Engine) OnBeaconStateReceived(msg *wire.MessageChainState) {

}

func (self *Engine) OnGetBeaconState(msg *wire.MessageChainState) {

}

// func (self *Engine) OnSwapRequest(msg *wire.MessageSwapRequest) {
// 	Logger.log.Info("Received a MessageSwapRequest")

// 	if msg.LockTime > time.Now().Unix() {
// 		return
// 	}

// 	committee := self.Committee().GetCommittee()

// 	if common.IndexOfStr(msg.Requester, committee) < 0 {
// 		Logger.log.Error("ERROR OnSwapRequest is not existed committee")
// 		return
// 	}

// 	// rawBytes := self.getRawBytesForSwap(msg.LockTime, msg.Requester, msg.shardID, msg.Candidate)
// 	// // TODO check requester signature
// 	// err := cashec.ValidateDataB58(msg.RequesterPbk, msg.RequesterSig, rawBytes)
// 	// if err != nil {
// 	// 	Logger.log.Info("Received a MessageSwapRequest validate error", err)
// 	// 	return
// 	// }
// 	err := msg.Verify()
// 	if err != nil {
// 		Logger.log.Info("Received a MessageSwapRequest validate error", err)
// 		return
// 	}
// 	// validate condition for swap

// 	peerIDs := self.config.Server.GetPeerIDsFromPublicKey(msg.Candidate)
// 	if len(peerIDs) == 0 {
// 		return
// 	}

// 	sig, err := self.config.UserKeySet.SignBase58(msg.GetMsgByte())
// 	if err != nil {
// 		Logger.log.Error("Can't sign swap ", err)
// 		return
// 	}
// 	messageSigMsg, err := wire.MakeEmptyMessage(wire.CmdSwapSig)
// 	if err != nil {
// 		return
// 	}
// 	// messageSigMsg.(*wire.MessageSwapSig).LockTime = msg.LockTime
// 	// messageSigMsg.(*wire.MessageSwapSig).RequesterPbk = msg.RequesterPbk
// 	// messageSigMsg.(*wire.MessageSwapSig).shardID = msg.shardID
// 	// messageSigMsg.(*wire.MessageSwapSig).ProducerPbk = msg.ProducerPbk
// 	messageSigMsg.(*wire.MessageSwapSig).Validator = base58.Base58Check{}.Encode(self.config.UserKeySet.PaymentAddress.Pk, byte(0x00))
// 	messageSigMsg.(*wire.MessageSwapSig).SwapSig = sig

// 	peerID, err := libp2p.IDB58Decode(msg.SenderID)
// 	if err != nil {
// 		Logger.log.Error("ERROR", msg.SenderID, peerID, err)
// 		return
// 	}
// 	self.config.Server.PushMessageToPeer(messageSigMsg, peerID)

// 	return
// }

// func (self *Engine) OnSwapSig(msg *wire.MessageSwapSig) {
// 	Logger.log.Info("Received a MessageSwapSig")
// 	self.cSwapSig <- swapSig{
// 		Validator: msg.Validator,
// 		SwapSig:   msg.SwapSig,
// 	}
// 	return
// }

func (self *Engine) OnSwapUpdate(msg *wire.MessageSwapUpdate) {
	// Logger.log.Info("Received a MessageSwapUpdate")

	// if msg.LockTime > time.Now().Unix() {
	// 	return
	// }

	// committee := self.Committee().GetCommittee()

	// if common.IndexOfStr(msg.Candidate, committee) >= 0 {
	// 	Logger.log.Error("ERROR OnSwapUpdate is existed committee")
	// 	return
	// }

	// //verify signatures
	// rawBytes := getRawBytesForSwap(msg.LockTime, msg.Requester, msg.shardID, msg.Candidate)
	// cLeader := 0
	// for leaderPbk, leaderSig := range msg.Signatures {
	// 	if common.IndexOfStr(leaderPbk, committee) >= 0 {
	// 		err := cashec.ValidateDataB58(leaderPbk, leaderSig, rawBytes)
	// 		if err != nil {
	// 			Logger.log.Error("ERROR OnSwapUpdate", leaderPbk, err)
	// 			continue
	// 		}
	// 	} else {
	// 		continue
	// 	}
	// 	cLeader++
	// }
	// if cLeader < common.TotalValidators/2 {
	// 	Logger.log.Error("ERROR OnSwapUpdate not enough signatures")
	// 	return
	// }
	// //TODO update committee list
	// self.Committee().UpdateCommittee(msg.Candidate, msg.shardID)

	// return
}

func MakeMsgBFTPropose(aggregatedSig string, validatorsIdx []int, block blockchain.BFTBlockInterface) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTPropose)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBFTPropose).Block = block
	msg.(*wire.MessageBFTPropose).AggregatedSig = aggregatedSig
	msg.(*wire.MessageBFTPropose).ValidatorsIdx = validatorsIdx
	return msg, nil
}

func MakeMsgBFTPrepare(Ri []byte, pubkey string) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTPrepare)
	if err != nil {
		Logger.log.Error(err)

		return msg, err
	}
	msg.(*wire.MessageBFTPrepare).Ri = Ri
	msg.(*wire.MessageBFTPrepare).Pubkey = pubkey
	return msg, nil
}

func MakeMsgBFTCommit(commitSig string, R string, validatorsIdx []int, pubkey string) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTCommit)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBFTCommit).CommitSig = commitSig
	msg.(*wire.MessageBFTCommit).R = R
	msg.(*wire.MessageBFTCommit).ValidatorsIdx = validatorsIdx
	msg.(*wire.MessageBFTCommit).Pubkey = pubkey
	return msg, nil
}

func MakeMsgBFTReply(aggregatedSig string, validatorsIdx []int) (wire.Message, error) {
	msg, err := wire.MakeEmptyMessage(wire.CmdBFTReply)
	if err != nil {
		Logger.log.Error(err)
		return msg, err
	}
	msg.(*wire.MessageBFTReply).AggregatedSig = aggregatedSig
	msg.(*wire.MessageBFTReply).ValidatorsIdx = validatorsIdx
	return msg, nil
}
