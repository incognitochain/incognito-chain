package blockchain

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	privacy "github.com/ninjadotorg/constant/privacy"
)

func (self *BlockChain) GetDatabase() database.DatabaseInterface {
	return self.config.DataBase
}

func (self *BlockChain) GetHeight() int32 {
	return self.BestState[0].BestBlock.Header.Height
}

func (self *BlockChain) GetChainHeight(chainID byte) int32 {
	return self.BestState[chainID].BestBlock.Header.Height
}

func (self *BlockChain) GetDCBBoardPubKeys() [][]byte {
	return self.BestState[0].BestBlock.Header.DCBGovernor.BoardPubKeys
}

func (self *BlockChain) GetGOVBoardPubKeys() [][]byte {
	return self.BestState[0].BestBlock.Header.GOVGovernor.BoardPubKeys
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

func (self *BlockChain) GetCrowdsaleData(saleID []byte) (*params.SaleData, error) {
	endBlock, buyingAsset, buyingAmount, sellingAsset, sellingAmount, err := self.config.DataBase.LoadCrowdsaleData(saleID)
	var saleData *params.SaleData
	if err != nil {
		saleData = &params.SaleData{
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

func (self *BlockChain) GetCMB(mainAccount []byte) (privacy.PaymentAddress, []privacy.PaymentAddress, uint64, *common.Hash, uint8, uint64, error) {
	reserveAcc, members, capital, hash, state, fine, err := self.config.DataBase.GetCMB(mainAccount)
	if err != nil {
		return privacy.PaymentAddress{}, nil, 0, nil, 0, 0, err
	}

	memberAddresses := []privacy.PaymentAddress{}
	for _, member := range members {
		memberAddress := (&privacy.PaymentAddress{}).SetBytes(member)
		memberAddresses = append(memberAddresses, *memberAddress)
	}

	txHash, _ := (&common.Hash{}).NewHash(hash)
	reserve := (&privacy.PaymentAddress{}).SetBytes(reserveAcc)
	return *reserve, memberAddresses, capital, txHash, state, fine, nil
}

func (self *BlockChain) GetCMBResponse(mainAccount []byte) ([][]byte, error) {
	return self.config.DataBase.GetCMBResponse(mainAccount)
}

func (self *BlockChain) GetDepositSend(contractID []byte) ([]byte, error) {
	return self.config.DataBase.GetDepositSend(contractID)
}

func (self *BlockChain) GetWithdrawRequest(contractID []byte) ([]byte, uint8, error) {
	return self.config.DataBase.GetWithdrawRequest(contractID)
}
