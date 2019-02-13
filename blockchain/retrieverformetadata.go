package blockchain

import (
	"bytes"
	"fmt"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

func (bc *BlockChain) GetDatabase() database.DatabaseInterface {
	return bc.config.DataBase
}

func (bc *BlockChain) GetTxChainHeight(tx metadata.Transaction) (uint64, error) {
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	return bc.GetChainHeight(shardID), nil
}

func (bc *BlockChain) GetChainHeight(shardID byte) uint64 {
	return bc.BestState.Shard[shardID].ShardHeight
}

func (bc *BlockChain) GetBeaconHeight() uint64 {
	return bc.BestState.Beacon.BeaconHeight
}

func (bc *BlockChain) GetBoardPubKeys(boardType byte) [][]byte {
	if boardType == common.DCBBoard {
		return bc.GetDCBBoardPubKeys()
	} else {
		return bc.GetGOVBoardPubKeys()
	}
}

func (bc *BlockChain) GetDCBBoardPubKeys() [][]byte {
	pubkeys := [][]byte{}
	for _, addr := range bc.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress {
		pubkeys = append(pubkeys, addr.Pk[:])
	}
	return pubkeys
}

func (bc *BlockChain) GetGOVBoardPubKeys() [][]byte {
	pubkeys := [][]byte{}
	for _, addr := range bc.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress {
		pubkeys = append(pubkeys, addr.Pk[:])
	}
	return pubkeys
}
func (bc *BlockChain) GetBoardPaymentAddress(boardType byte) []privacy.PaymentAddress {
	if boardType == common.DCBBoard {
		return bc.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress
	}
	return bc.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress
}

func ListPubKeyFromListPayment(listPaymentAddresses []privacy.PaymentAddress) [][]byte {
	pubKeys := make([][]byte, 0)
	for _, i := range listPaymentAddresses {
		pubKeys = append(pubKeys, i.Pk)
	}
	return pubKeys
}

func (bc *BlockChain) GetDCBParams() params.DCBParams {
	return bc.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams
}

func (bc *BlockChain) GetGOVParams() params.GOVParams {
	return bc.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams
}

//// Loan
func (bc *BlockChain) GetLoanReq(loanID []byte) (*common.Hash, error) {
	key := getLoanRequestKeyBeacon(loanID)
	reqHash, ok := bc.BestState.Beacon.Params[key]
	if !ok {
		return nil, errors.Errorf("Loan request with ID %x not found", loanID)
	}
	resp, err := common.NewHashFromStr(reqHash)
	return resp, err
}

// GetLoanResps returns all responses of a given loanID
func (bc *BlockChain) GetLoanResps(loanID []byte) ([][]byte, []metadata.ValidLoanResponse, error) {
	key := getLoanResponseKeyBeacon(loanID)
	senders := [][]byte{}
	responses := []metadata.ValidLoanResponse{}
	if data, ok := bc.BestState.Beacon.Params[key]; ok {
		lrds, err := parseLoanResponseValueBeacon(data)
		if err != nil {
			return nil, nil, err
		}
		for _, lrd := range lrds {
			senders = append(senders, lrd.SenderPubkey)
			responses = append(responses, lrd.Response)
		}
	}
	return senders, responses, nil
}

func (bc *BlockChain) GetLoanPayment(loanID []byte) (uint64, uint64, uint64, error) {
	return bc.config.DataBase.GetLoanPayment(loanID)
}

func (bc *BlockChain) GetLoanRequestMeta(loanID []byte) (*metadata.LoanRequest, error) {
	reqHash, err := bc.GetLoanReq(loanID)
	if err != nil {
		return nil, err
	}
	_, _, _, txReq, err := bc.GetTransactionByHash(reqHash)
	if err != nil {
		return nil, err
	}
	requestMeta := txReq.GetMetadata().(*metadata.LoanRequest)
	return requestMeta, nil
}

//// Dividends
func (bc *BlockChain) GetLatestDividendProposal(forDCB bool) (id, amount uint64) {
	return bc.BestState.Beacon.GetLatestDividendProposal(forDCB)
}

func (bc *BlockChain) GetDividendReceiversForID(dividendID uint64, forDCB bool) ([]privacy.PaymentAddress, []uint64, bool, error) {
	return bc.config.DataBase.GetDividendReceiversForID(dividendID, forDCB)
}

//// Crowdsales
func (bc *BlockChain) parseProposalCrowdsaleData(proposalTxHash *common.Hash, saleID []byte) *params.SaleData {
	var saleData *params.SaleData
	_, _, _, proposalTx, err := bc.GetTransactionByHash(proposalTxHash)
	if err == nil {
		proposalMeta := proposalTx.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
		fmt.Printf("[db] proposal cs data: %+v\n", proposalMeta)
		for _, data := range proposalMeta.DCBParams.ListSaleData {
			fmt.Printf("[db] data ptr: %p, data: %+v\n", &data, data)
			if bytes.Equal(data.SaleID, saleID) {
				saleData = &data
				saleData.SetProposalTxHash(*proposalTxHash)
				break
			}
		}
	}
	return saleData
}

func (bc *BlockChain) GetCrowdsaleData(saleID []byte) (*params.SaleData, error) {
	key := getSaleDataKeyBeacon(saleID)
	if value, ok := bc.BestState.Beacon.Params[key]; ok {
		saleData, err := parseSaleDataValueBeacon(value)
		if err != nil {
			return nil, err
		}
		return saleData, nil
	} else {
		return nil, errors.New("Error getting SaleData from beacon best state")
	}
}

func (bc *BlockChain) GetAllCrowdsales() ([]*params.SaleData, error) {
	saleDataList := []*params.SaleData{}
	saleIDs, proposalTxHashes, buyingAmounts, sellingAmounts, err := bc.config.DataBase.GetAllCrowdsales()
	if err == nil {
		for i, hash := range proposalTxHashes {
			saleData := bc.parseProposalCrowdsaleData(&hash, saleIDs[i])
			if saleData != nil {
				saleData.BuyingAmount = buyingAmounts[i]
				saleData.SellingAmount = sellingAmounts[i]
			}
			saleDataList = append(saleDataList, saleData)
		}
	}
	return saleDataList, err
}

//// CMB
func (bc *BlockChain) GetCMB(mainAccount []byte) (privacy.PaymentAddress, []privacy.PaymentAddress, uint64, *common.Hash, uint8, uint64, error) {
	reserveAcc, members, capital, hash, state, fine, err := bc.config.DataBase.GetCMB(mainAccount)
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

func (bc *BlockChain) GetCMBResponse(mainAccount []byte) ([][]byte, error) {
	return bc.config.DataBase.GetCMBResponse(mainAccount)
}

func (bc *BlockChain) GetDepositSend(contractID []byte) ([]byte, error) {
	return bc.config.DataBase.GetDepositSend(contractID)
}

func (bc *BlockChain) GetWithdrawRequest(contractID []byte) ([]byte, uint8, error) {
	return bc.config.DataBase.GetWithdrawRequest(contractID)
}
