package blockchain

import (
	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy"
)

func (self *BlockChain) GetDatabase() database.DatabaseInterface {
	return self.config.DataBase
}

func (self *BlockChain) GetTxChainHeight(tx metadata.Transaction) (uint64, error) {
	shardID, err := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
	if err != nil {
		return 0, err
	}
	return self.GetChainHeight(shardID), nil
}

func (self *BlockChain) GetChainHeight(shardID byte) uint64 {
	// return uint64(self.BestState[shardID].BestBlock.Header.Height)
	return 0
}

func (self *BlockChain) GetBoardPubKeys(boardType string) [][]byte {
	if boardType == "dcb" {
		return self.GetDCBBoardPubKeys()
	} else {
		return self.GetGOVBoardPubKeys()
	}
}

func (self *BlockChain) GetDCBBoardPubKeys() [][]byte {
	// return self.BestState[0].BestBlock.Header.DCBGovernor.BoardPubKeys
	return nil
}

func (self *BlockChain) GetGOVBoardPubKeys() [][]byte {
	// return self.BestState[0].BestBlock.Header.GOVGovernor.BoardPubKeys
	return nil
}
func (self *BlockChain) GetBoardPaymentAddress(boardType string) []privacy.PaymentAddress {
	if boardType == "dcb" {
		// return self.BestState[0].BestBlock.Header.DCBGovernor.BoardPaymentAddress
		return []privacy.PaymentAddress{}
	} else {
		// return self.BestState[0].BestBlock.Header.GOVGovernor.BoardPaymentAddress
		return []privacy.PaymentAddress{}
	}

}

func ListPubKeyFromListPayment(listPaymentAddresses []privacy.PaymentAddress) [][]byte {
	pubKeys := make([][]byte, 0)
	for _, i := range listPaymentAddresses {
		pubKeys = append(pubKeys, i.Pk)
	}
	return pubKeys
}

func (self *BlockChain) GetDCBParams() params.DCBParams {
	// return self.BestState[0].BestBlock.Header.DCBConstitution.DCBParams
	return params.DCBParams{}
}

func (self *BlockChain) GetGOVParams() params.GOVParams {
	// return self.BestState[0].BestBlock.Header.GOVConstitution.GOVParams
	return params.GOVParams{}
}

func (self *BlockChain) GetLoanTxs(loanID []byte) ([][]byte, error) {
	// return self.config.DataBase.GetLoanTxs(loanID)
	return nil, nil
}

func (self *BlockChain) GetLoanPayment(loanID []byte) (uint64, uint64, uint64, error) {
	// return self.config.DataBase.GetLoanPayment(loanID)
	return 0, 0, 0, nil
}

func (self *BlockChain) GetCrowdsaleData(saleID []byte) (*params.SaleData, error) {
	endBlock, buyingAsset, buyingAmount, sellingAsset, sellingAmount, err := self.config.DataBase.GetCrowdsaleData(saleID)
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
