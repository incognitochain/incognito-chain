package blockchain

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/voting"
)

func (self *BlockChain) GetDatabase() database.DatabaseInterface {
	return self.config.DataBase
}

func (self *BlockChain) GetHeight() int32 {
	return self.BestState[0].BestBlock.Header.Height
}

func (self *BlockChain) GetDCBBoardPubKeys() [][]byte {
	return self.BestState[0].BestBlock.Header.DCBGovernor.DCBBoardPubKeys
}

func (self *BlockChain) GetGOVBoardPubKeys() [][]byte {
	return self.BestState[0].BestBlock.Header.GOVGovernor.GOVBoardPubKeys
}

func (self *BlockChain) GetDCBParams() params.DCBParams {
	return self.BestState[0].BestBlock.Header.DCBConstitution.DCBParams
}

func (self *BlockChain) GetGOVParams() params.GOVParams {
	return self.BestState[0].BestBlock.Header.GOVConstitution.GOVParams
}

func (self *BlockChain) GetLoanTxs(loanID []byte) ([][]byte, error) {
	return self.config.DataBase.GetLoanTxs(loanID)
}

func (self *BlockChain) GetLoanPayment(loanID []byte) (uint64, uint64, uint32, error) {
	return self.config.DataBase.GetLoanPayment(loanID)
}

func (self *BlockChain) GetCrowdsaleTxs(requestTxHash []byte) ([][]byte, error) {
	return self.config.DataBase.GetCrowdsaleTxs(requestTxHash)
}

func (self *BlockChain) GetCrowdsaleData(saleID []byte) (*voting.SaleData, error) {
	endBlock, buyingAsset, buyingAmount, sellingAsset, sellingAmount, err := self.config.DataBase.LoadCrowdsaleData(saleID)
	var saleData *voting.SaleData
	if err != nil {
		saleData = &voting.SaleData{
			SaleID:        saleID,
			EndBlock:      endBlock,
			BuyingAsset:   buyingAsset,
			BuyingAmount:  buyingAmount,
			SellingAsset:  sellingAsset,
			SellingAmount: sellingAmount,
		}
	}
	return saleData, err
}

func (self *BlockChain) GetCMB(mainAccount []byte) ([]privacy.PaymentAddress, uint64, *common.Hash, uint8, error) {
	members, capital, hash, state, err := self.config.DataBase.GetCMB(mainAccount)
	if err != nil {
		return nil, 0, nil, 0, err
	}

	memberAddresses := []privacy.PaymentAddress{}
	for _, member := range members {
		memberAddress := (&privacy.PaymentAddress{}).FromBytes(member)
		memberAddresses = append(memberAddresses, *memberAddress)
	}

	txHash, _ := (&common.Hash{}).NewHash(hash)
	return memberAddresses, capital, txHash, state, nil
}

func (self *BlockChain) GetCMBResponse(mainAccount []byte) ([][]byte, error) {
	return self.config.DataBase.GetCMBResponse(mainAccount)
}
