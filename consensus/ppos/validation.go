package ppos

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/common/base58"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/wire"
)

func (self *Engine) ValidateTxList(txList []transaction.Transaction) error {
	for _, tx := range txList {
		if tx.ValidateTransaction() == false {
			return errTxIsWrong
		}
	}
	return nil
}

func (self *Engine) ValidateCommitteeSigs(blockHash []byte, committee []string, sigs []string) error {
	validatedSigs := 0
	for idx, validator := range committee {
		decPubkey, _, err := base58.Base58Check{}.Decode(validator)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		k := cashec.KeySetSealer{
			SpublicKey: decPubkey,
		}
		decSig, _, err := base58.Base58Check{}.Decode(sigs[idx])
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		isValidSignature, err := k.Verify(blockHash, decSig)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if isValidSignature == false {
			continue
		}
		validatedSigs++
	}

	if validatedSigs < 10 {
		return errNotEnoughSigs
	}
	return nil
}

func (self *Engine) ValidateMerkleRootCommitments(block *blockchain.Block) error {
	rtOld := self.config.BlockChain.BestState[block.Header.ChainID].BestBlock.Header.MerkleRootCommitments.CloneBytes()
	newTree := self.config.BlockChain.BestState[block.Header.ChainID].CmTree.MakeCopy()
	Logger.log.Infof("[validateblock] old tree rt: %x\n", newTree.GetRoot(common.IncMerkleTreeHeight))
	blockchain.UpdateMerkleTreeForBlock(newTree, block)
	rt := newTree.GetRoot(common.IncMerkleTreeHeight)
	Logger.log.Infof("[validateblock] updated tree rt: %x\n", rt)
	if !bytes.Equal(rt[:], block.Header.MerkleRootCommitments.CloneBytes()) {
		Logger.log.Errorf("MerkleRootCommitments diff!! \n%x\n%x\n%x", rtOld, rt[:], block.Header.MerkleRootCommitments[:])
		for _, blockTx := range block.Transactions {
			if blockTx.GetType() == common.TxNormalType {
				tx, ok := blockTx.(*transaction.Tx)
				if ok == false {
					Logger.log.Errorf("Transaction in block not valid")
				}

				for _, desc := range tx.Descs {
					for _, cm := range desc.Commitments {
						Logger.log.Infof("%x", cm[:])
					}
				}
			}
		}
		return errMerkleRootCommitments
	}
	return nil
}

func (self *Engine) CheckBlockSize(block *blockchain.Block) error {
	blockBytes, err := json.Marshal(*block)
	if err != nil {
		return err
	}
	if len(blockBytes) > common.MAX_BLOCKSIZE {
		return errBlockSizeExceed
	}
	return nil
}

func (self *Engine) IsEnoughData(block *blockchain.Block) error {
	if self.validatedChainsHeight.Heights[block.Header.ChainID] == (int(block.Height) - 1) {
		notFullySync := false
		for i := 0; i < common.TOTAL_VALIDATORS; i++ {
			if self.validatedChainsHeight.Heights[i] < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
				notFullySync = true
				getBlkMsg := &wire.MessageGetBlocks{
					LastBlockHash: self.config.BlockChain.BestState[i].BestBlockHash.String(),
				}
				peerIDs := self.config.Server.GetPeerIdsFromPublicKey(block.ChainLeader)
				if len(peerIDs) != 0 {
					Logger.log.Info("Send getblock to "+peerIDs[0], block.ChainLeader)
					self.config.Server.PushMessageToPeer(getBlkMsg, peerIDs[0])
				} else {
					Logger.log.Error("Validator's peer not found!", block.ChainLeader)
				}
			}
		}
		if notFullySync {
			timer := time.NewTimer(common.MAX_SYNC_CHAINS_TIME * time.Second)
			<-timer.C
			for i := 0; i < common.TOTAL_VALIDATORS; i++ {
				if int(self.config.BlockChain.BestState[i].Height) < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
					return errChainNotFullySynced
				}
			}
		}
	} else {
		return errChainNotFullySynced
	}
	return nil
}

func (self *Engine) validateBlockSanity(block *blockchain.Block) error {
	// 1. Check whether we acquire enough data to validate this block
	err := self.IsEnoughData(block)
	if err != nil {
		return err
	}

	// 2. Check block size
	err = self.CheckBlockSize(block)
	if err != nil {
		return err
	}

	// 3. Check signature of the block leader for block header
	headerBytes, _ := json.Marshal(block.Header)
	err = cashec.ValidateDataB58(block.ChainLeader, block.ChainLeaderSig, headerBytes)
	if err != nil {
		return err
	}

	// 5. Validate MerkleRootCommitments
	err = self.ValidateMerkleRootCommitments(block)
	if err != nil {
		return err
	}

	// 6. Validate committee member signatures
	err = self.ValidateCommitteeSigs([]byte(block.Hash().String()), block.Header.Committee, block.Header.BlockCommitteeSigs)
	if err != nil {
		return err
	}
	// 7. Validate transactions
	return self.ValidateTxList(block.Transactions)

}

func (self *Engine) validatePreSignBlockSanity(block *blockchain.Block) error {
	// 1. Check whether we acquire enough data to validate this block
	err := self.IsEnoughData(block)
	if err != nil {
		return err
	}

	// 2. Check block size
	err = self.CheckBlockSize(block)
	if err != nil {
		return err
	}

	// 3. Check signature of the block leader for block header
	headerBytes, _ := json.Marshal(block.Header)
	err = cashec.ValidateDataB58(block.ChainLeader, block.ChainLeaderSig, headerBytes)
	if err != nil {
		return err
	}

	// 4. Check whether we acquire enough data to validate this block
	err = self.IsEnoughData(block)
	if err != nil {
		return err
	}

	// 5. Validate MerkleRootCommitments
	err = self.ValidateMerkleRootCommitments(block)
	if err != nil {
		return err
	}

	// 7. Validate transactions
	return self.ValidateTxList(block.Transactions)
}
