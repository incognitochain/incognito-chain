package ppos

import (
	"bytes"
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/cashec"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/common/base58"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

func (self *Engine) ValidateTxList(txList []transaction.Transaction) error {
	for _, tx := range txList {
		if tx.ValidateTransaction() == false {
			return errTxIsWrong
		}
	}
	return nil
}

func (self *Engine) ValidateCommitteeSigs(blockHash []byte, committeeSigs map[string]string) error {
	validatedSigs := 0
	for validator, sig := range committeeSigs {
		decPubkey, _, err := base58.Base58Check{}.Decode(validator)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		k := cashec.KeySetSealer{
			SpublicKey: decPubkey,
		}
		decSig, _, err := base58.Base58Check{}.Decode(sig)
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
	self.config.BlockChain.UpdateMerkleTreeForBlock(newTree, block)
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
	if len(blockBytes) > MAX_BLOCKSIZE {
		return errBlockSizeExceed
	}
	return nil
}
