package ppos

import (
	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash/blockchain"
	"github.com/ninjadotorg/cash/common/base58"
	"github.com/ninjadotorg/cash/wire"
	"github.com/ninjadotorg/cash/cashec"
	"github.com/ninjadotorg/cash/common"
)

func (self *Engine) OnRequestSign(msgBlock *wire.MessageRequestBlockSign) {
	block := &msgBlock.Block
	err := self.validatePreSignBlockSanity(block)
	if err != nil {
		invalidBlockMsg := &wire.MessageInvalidBlock{
			Reason:    err.Error(),
			BlockHash: block.Hash().String(),
			ChainID:   block.Header.ChainID,
			Validator: base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)),
		}
		dataByte, _ := invalidBlockMsg.JsonSerialize()
		invalidBlockMsg.ValidatorSig, err = self.signData(dataByte)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		Logger.log.Critical("Invalid block msg", invalidBlockMsg)
		err = self.config.Server.PushMessageToAll(invalidBlockMsg)
		if err != nil {
			Logger.log.Error(err)
			return
		}
		return
	}

	sig, err := self.signData([]byte(block.Hash().String()))
	if err != nil {
		Logger.log.Error("Can't sign block ", err)
		// ??? something went terribly wrong
		return
	}
	blockSigMsg := wire.MessageBlockSig{
		BlockHash:    block.Hash().String(),
		Validator:    base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)),
		ValidatorSig: sig,
	}
	peerID, err := peer2.IDB58Decode(msgBlock.SenderID)
	if err != nil {
		Logger.log.Error("ERROR", msgBlock.SenderID, peerID, err)
	}
	Logger.log.Info(block.Hash().String(), blockSigMsg)
	err = self.config.Server.PushMessageToPeer(&blockSigMsg, peerID)
	if err != nil {
		Logger.log.Error(err)
	}
	return
}

func (self *Engine) OnBlockReceived(block *blockchain.Block) {
	if self.config.BlockChain.BestState[block.Header.ChainID].Height < block.Height {
		if exists, _ := self.config.BlockChain.BlockExists(block.Hash()); !exists {
			err := self.validateBlockSanity(block)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			self.UpdateChain(block)
		}
	}
	return
}

func (self *Engine) OnBlockSigReceived(blockHash string, validator string, sig string) {
	Logger.log.Info("Received a block signature")
	self.cBlockSig <- blockSig{
		BlockHash:    blockHash,
		Validator:    validator,
		ValidatorSig: sig,
	}
	return
}

func (self *Engine) OnInvalidBlockReceived(blockHash string, chainID byte, reason string) {
	// leave empty for now
	Logger.log.Error(blockHash, chainID, reason)
	return
}

func (self *Engine) OnChainStateReceived(msg *wire.MessageChainState) {
	chainInfo := msg.ChainInfo.(map[string]interface{})
	for i, v := range self.validatedChainsHeight.Heights {
		if chainInfo["ChainsHeight"] != nil {
			if v < int(chainInfo["ChainsHeight"].([]interface{})[i].(float64)) {
				self.knownChainsHeight.Heights[i] = int(chainInfo["ChainsHeight"].([]interface{})[i].(float64))
				lastBlockHash := self.config.BlockChain.BestState[i].BestBlockHash.String()
				getBlkMsg := &wire.MessageGetBlocks{
					LastBlockHash: lastBlockHash,
				}
				Logger.log.Info("Send " + getBlkMsg.MessageType() + " to " + msg.SenderID)
				peerID, err := peer2.IDB58Decode(msg.SenderID)
				if err != nil {
					Logger.log.Error(err)
					continue
				}
				self.config.Server.PushMessageToPeer(getBlkMsg, peerID)
			}
		} else {
			Logger.log.Error("ChainsHeight is empty!")
		}
	}
	return
}

func (self *Engine) OnGetChainState(msg *wire.MessageGetChainState) {
	newMsg, err := wire.MakeEmptyMessage(wire.CmdChainState)
	if err != nil {
		return
	}
	newMsg.(*wire.MessageChainState).ChainInfo = ChainInfo{
		CurrentCommittee:        self.GetCommittee(),
		CandidateListMerkleHash: "",
		ChainsHeight:            self.validatedChainsHeight.Heights,
	}
	peerID, _ := peer2.IDB58Decode(msg.SenderID)
	self.config.Server.PushMessageToPeer(newMsg, peerID)
	return
}

func (self *Engine) OnCandidateProposal() {

}

func (self *Engine) OnCandidateVote() {

}

func (self *Engine) sendBlockMsg(block *blockchain.Block) {
	blockMsg, err := wire.MakeEmptyMessage(wire.CmdBlock)
	if err != nil {
		Logger.log.Error(err)
		return
	}
	blockMsg.(*wire.MessageBlock).Block = *block
	self.config.Server.PushMessageToAll(blockMsg)
}

func (self *Engine) OnRequestSwap(msg *wire.MessageRequestSwap) {
	Logger.log.Info("Received a MessageRequestSwap")

	peerIDs := self.config.Server.GetPeerIDsFromPublicKey(msg.SealerPbk)
	if len(peerIDs) == 0 {
		return
	}

	senderID := base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00))

	rawBytes := []byte{}
	rawBytes = append(rawBytes, []byte(msg.RequesterPbk)...)
	rawBytes = append(rawBytes, msg.ChainID)
	rawBytes = append(rawBytes, []byte(msg.SealerPbk)...)

	sig, err := self.signData(rawBytes)
	if err != nil {
		Logger.log.Error("Can't sign swap ", err)
		return
	}
	messageSigMsg := wire.MessageSignSwap{
		SenderID:     senderID,
		RequesterPbk: msg.RequesterPbk,
		Validator:    base58.Base58Check{}.Encode(self.config.ValidatorKeySet.SpublicKey, byte(0x00)),
		ValidatorSig: sig,
	}
	peerID, err := peer2.IDB58Decode(msg.SenderID)
	if err != nil {
		Logger.log.Error("ERROR", msg.SenderID, peerID, err)
	}
	err = self.config.Server.PushMessageToPeer(&messageSigMsg, peerID)
	if err != nil {
		Logger.log.Error(err)
	}

	return
}

func (self *Engine) OnSignSwap(msg *wire.MessageSignSwap) {
	Logger.log.Info("Received a MessageSignSwap")
	self.cSwapSig <- swapSig{
		RequesterPbk:    msg.RequesterPbk,
		ChainID:         msg.ChainID,
		SealerPublicKey: msg.SealerPbk,
		Validator:       msg.Validator,
		ValidatorSig:    msg.ValidatorSig,
	}
	return
}

func (self *Engine) OnUpdateSwap(msg *wire.MessageUpdateSwap) {
	Logger.log.Info("Received a MessageUpdateSwap")

	committee := make([]string, common.TotalValidators)
	copy(committee, self.GetCommittee())

	if common.IndexOfStr(msg.SealerPbk, committee) >= 0 {
		Logger.log.Error("ERROR OnUpdateSwap is existed committee")
		return
	}

	//versify signatures
	rawBytes := []byte{}
	rawBytes = append(rawBytes, []byte(msg.RequesterPbk)...)
	rawBytes = append(rawBytes, msg.ChainID)
	rawBytes = append(rawBytes, []byte(msg.SealerPbk)...)
	cLeader := 0
	for leaderPbk, leaderSig := range msg.Signatures {
		if common.IndexOfStr(leaderPbk, committee) >= 0 {
			err := cashec.ValidateDataB58(leaderPbk, leaderSig, rawBytes)
			if err != nil {
				Logger.log.Error("ERROR OnUpdateSwap", leaderPbk, err)
				return
			}
		}
		cLeader += 1
	}
	if cLeader < common.TotalValidators / 2 {
		Logger.log.Error("ERROR OnUpdateSwap not enough signatures")
		return
	}
	//TODO update committee list
	self.updateCommittee(msg.SealerPbk, msg.ChainID)

	return
}
