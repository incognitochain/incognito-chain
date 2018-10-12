package ppos

import (
	peer2 "github.com/libp2p/go-libp2p-peer"
	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common/base58"
	"github.com/ninjadotorg/cash-prototype/wire"
)

func (self *Engine) OnRequestSign(msgBlock *wire.MessageRequestSign) {
	block := &msgBlock.Block
	err := self.validatePreSignBlock(block)
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
		Logger.log.Critical("OHSHITT", err)
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
		if _, _, err := self.config.BlockChain.GetBlockHeightByBlockHash(block.Hash()); err != nil {
			err := self.validateBlock(block)
			if err != nil {
				Logger.log.Error(err)
				return
			}
			isMainChain, ok, err := self.config.BlockChain.ProcessBlock(block)
			_ = isMainChain
			_ = ok
			if err != nil {
				Logger.log.Error(err)
				return
			}
			self.UpdateChain(block)
			err = self.config.FeeEstimator[block.Header.ChainID].RegisterBlock(block)
			if err != nil {
				Logger.log.Error(err)
			}
			self.knownChainsHeight.Lock()
			if self.knownChainsHeight.Heights[block.Header.ChainID] < int(block.Height) {
				self.knownChainsHeight.Heights[block.Header.ChainID] = int(block.Height)
			}
			self.knownChainsHeight.Unlock()
		}
	} else {
		//save block to cache
	}
	return
}

func (self *Engine) OnBlockSigReceived(blockHash string, validator string, sig string) {
	Logger.log.Info("Received a block signature")
	self.blockSigCh <- blockSig{
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
	// fmt.Println(msg)
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
					continue
				}
				self.config.Server.PushMessageToPeer(getBlkMsg, peerID)
			}
		} else {
			Logger.log.Error("what the ...")
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
		CurrentCommittee:  self.currentCommittee,
		CandidateListHash: "",
		ChainsHeight:      self.validatedChainsHeight.Heights,
	}
	peerID, _ := peer2.IDB58Decode(msg.SenderID)
	self.config.Server.PushMessageToPeer(newMsg, peerID)
	return
}

func (self *Engine) OnCandidateProposal() {

}

func (self *Engine) OnCandidateVote() {

}

func (self *Engine) OnCandidateRequestTx() {

}
