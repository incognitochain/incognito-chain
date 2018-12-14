package ppos

import (
	"encoding/json"
	"time"

	"github.com/ninjadotorg/constant/blockchain"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wire"
)

func (self *Engine) ValidateTxList(txList []metadata.Transaction) error {
	for _, tx := range txList {
		err := self.ValidateSpecTxWithBlockChain(tx)
		if err != nil {
			return err
		}
		if self.ValidateSpecTxByItSelf(tx) == false {
			return NewConsensusError(ErrTxIsWrong, nil)
		}
	}
	return nil
}

// Check tx with blockchain
func (self *Engine) ValidateSpecTxWithBlockChain(tx metadata.Transaction) error {
	// get chainID of tx
	chainID, err := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
	if err != nil {
		return err
	}
	// return self.config.MemPool.ValidateTxWithBlockChain(tx, chainID)
	return tx.ValidateTxWithBlockChain(self.config.BlockChain, chainID, self.config.BlockChain.GetDatabase())
}

// Checl spec tx by it self
func (self *Engine) ValidateSpecTxByItSelf(tx metadata.Transaction) bool {
	// get chainID of tx
	// return self.config.MemPool.ValidateTxByItSelf(tx)
	chainID, _ := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
	return tx.ValidateTxByItself(tx.IsPrivacy(), self.config.BlockChain.GetDatabase(), self.config.BlockChain, chainID)
}

func (self *Engine) ValidateCommitteeSigs(blockHash []byte, committee []string, sigs []string) error {
	validatedSigs := 0
	for idx, validator := range committee {
		decPubkey, _, err := base58.Base58Check{}.Decode(validator)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		k := cashec.KeySet{
			PaymentAddress: privacy.PaymentAddress{
				Pk: decPubkey,
			},
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
		return NewConsensusError(ErrNotEnoughSigs, nil)
	}
	return nil
}

/*func (self *Engine) ValidateMerkleRootCommitments(block *blockchain.Block) error {
	rtOld := self.config.BlockChain.BestState[block.Header.ChainID].BestBlock.Header.MerkleRootCommitments.CloneBytes()
	newTree := self.config.BlockChain.BestState[block.Header.ChainID].CmTree.MakeCopy()
	Logger.log.Infof("[validateblock] old tree rt: %x\n", newTree.GetRoot(common.IncMerkleTreeHeight))
	err := blockchain.UpdateMerkleTreeForBlock(newTree, block)
	if err != nil {
		return err
	}
	rt := newTree.GetRoot(common.IncMerkleTreeHeight)
	Logger.log.Infof("[validateblock] updated tree rt: %x\n", rt)
	if !bytes.Equal(rt[:], block.Header.MerkleRootCommitments.CloneBytes()) {
		Logger.log.Errorf("MerkleRootCommitments diff!! \n%x\n%x\n%x", rtOld, rt[:], block.Header.MerkleRootCommitments[:])
		for _, blockTx := range block.Transactions {
			if blockTx.GetType() == common.TxNormalType || blockTx.GetType() == common.TxSalaryType {
				tx, ok := blockTx.(*transaction.TxNormal)
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
		return NewConsensusError(ErrMerkleRootCommitments, nil)
	}
	return nil
}*/

func (self *Engine) CheckBlockSize(block *blockchain.Block) error {
	blockBytes, err := json.Marshal(*block)
	if err != nil {
		return err
	}
	if len(blockBytes) > common.MaxBlockSize {
		return NewConsensusError(ErrBlockSizeExceed, nil)
	}
	return nil
}

func (self *Engine) IsEnoughData(block *blockchain.Block) error {
	if self.validatedChainsHeight.Heights[block.Header.ChainID] == (int(block.Header.Height) - 1) {
		notFullySync := false
		for i := 0; i < common.TotalValidators; i++ {
			if self.validatedChainsHeight.Heights[i] < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
				notFullySync = true
				getBlkMsg := &wire.MessageGetBlocks{
					LastBlockHash: self.config.BlockChain.BestState[i].BestBlockHash.String(),
				}
				go func(chainLeader string) {
					peerIDs := self.config.Server.GetPeerIDsFromPublicKey(chainLeader)
					if len(peerIDs) != 0 {
						Logger.log.Info("Send getblock to "+peerIDs[0], chainLeader)
						self.config.Server.PushMessageToPeer(getBlkMsg, peerIDs[0])
					} else {
						Logger.log.Error("Validator's peer not found!", chainLeader)
					}
				}(block.BlockProducer)
			}
		}
		if notFullySync {
			timer := time.NewTimer(common.MaxSyncChainTime * time.Second)
			<-timer.C
			for i := 0; i < common.TotalValidators; i++ {
				if self.validatedChainsHeight.Heights[i] < (block.Header.ChainsHeight[i]) && (i != int(block.Header.ChainID)) {
					return NewConsensusError(ErrChainNotFullySynced, nil)
				}
			}
		}
	} else {
		return NewConsensusError(ErrChainNotFullySynced, nil)
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
	err = cashec.ValidateDataB58(block.BlockProducer, block.BlockProducerSig, headerBytes)
	if err != nil {
		return err
	}

	// 4. ValidateTransaction committee member signatures
	err = self.ValidateCommitteeSigs([]byte(block.Hash().String()), block.Header.Committee, block.Header.BlockCommitteeSigs)
	if err != nil {
		return err
	}

	// 5. ValidateTransaction MerkleRootCommitments
	/*err = self.ValidateMerkleRootCommitments(block)
	if err != nil {
		return err
	}*/

	// 6. Validate transactions
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

	// 3. Check signature of the block leader for block hash
	err = cashec.ValidateDataB58(block.BlockProducer, block.Header.BlockCommitteeSigs[block.Header.ChainID], []byte(block.Hash().String()))
	if err != nil {
		return err
	}

	// 4. ValidateTransaction MerkleRootCommitments
	/*err = self.ValidateMerkleRootCommitments(block)
	if err != nil {
		return err
	}*/

	// 5. ValidateTransaction transactions
	return self.ValidateTxList(block.Transactions)
}

// func (self *Engine) ValidateSalary(block )
