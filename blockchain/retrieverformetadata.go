package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

func (blockchain *BlockChain) GetDatabase() database.DatabaseInterface {
	return blockchain.config.DataBase
}

func (blockchain *BlockChain) GetTxChainHeight(tx metadata.Transaction) (uint64, error) {
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	return blockchain.GetChainHeight(shardID), nil
}

func (blockchain *BlockChain) GetChainHeight(shardID byte) uint64 {
	return blockchain.BestState.Shard[shardID].ShardHeight
}

func (blockchain *BlockChain) GetBeaconHeight() uint64 {
	return blockchain.BestState.Beacon.BeaconHeight
}

func (blockchain *BlockChain) GetBoardPubKeys(boardType metadata.BoardType) [][]byte {
	if boardType == metadata.DCBBoard {
		return blockchain.GetDCBBoardPubKeys()
	} else {
		return blockchain.GetGOVBoardPubKeys()
	}
}

func (blockchain *BlockChain) GetDCBBoardPubKeys() [][]byte {
	pubkeys := [][]byte{}
	for _, addr := range blockchain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress {
		pubkeys = append(pubkeys, addr.Pk[:])
	}
	return pubkeys
}

func (blockchain *BlockChain) GetGOVBoardPubKeys() [][]byte {
	pubkeys := [][]byte{}
	for _, addr := range blockchain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress {
		pubkeys = append(pubkeys, addr.Pk[:])
	}
	return pubkeys
}

func (blockchain *BlockChain) GetBoardPaymentAddress(boardType metadata.BoardType) []privacy.PaymentAddress {
	if boardType == metadata.DCBBoard {
		return blockchain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress
	}
	return blockchain.BestState.Beacon.StabilityInfo.GOVGovernor.BoardPaymentAddress
}

func ListPubKeyFromListPayment(listPaymentAddresses []privacy.PaymentAddress) [][]byte {
	pubKeys := make([][]byte, 0)
	for _, i := range listPaymentAddresses {
		pubKeys = append(pubKeys, i.Pk)
	}
	return pubKeys
}

func (blockchain *BlockChain) GetDCBParams() params.DCBParams {
	return blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams
}

func (blockchain *BlockChain) GetGOVParams() params.GOVParams {
	return blockchain.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams
}

func (blockchain *BlockChain) GetLoanReq(loanID []byte) (*common.Hash, error) {
	key := getLoanRequestKeyBeacon(loanID)
	reqHash, ok := blockchain.BestState.Beacon.Params[key]
	if !ok {
		return nil, errors.Errorf("Loan request with ID %x not found", loanID)
	}
	resp, err := common.NewHashFromStr(reqHash)
	return resp, err
}

// GetLoanResps returns all responses of a given loanID
func (blockchain *BlockChain) GetLoanResps(loanID []byte) ([][]byte, []metadata.ValidLoanResponse, error) {
	key := getLoanResponseKeyBeacon(loanID)
	senders := [][]byte{}
	responses := []metadata.ValidLoanResponse{}
	if data, ok := blockchain.BestState.Beacon.Params[key]; ok {
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

func (blockchain *BlockChain) GetLoanPayment(loanID []byte) (uint64, uint64, uint64, error) {
	return blockchain.config.DataBase.GetLoanPayment(loanID)
}

func (blockchain *BlockChain) GetLoanRequestMeta(loanID []byte) (*metadata.LoanRequest, error) {
	reqHash, err := blockchain.GetLoanReq(loanID)
	if err != nil {
		return nil, err
	}
	_, _, _, txReq, err := blockchain.GetTransactionByHash(reqHash)
	if err != nil {
		return nil, err
	}
	requestMeta := txReq.GetMetadata().(*metadata.LoanRequest)
	return requestMeta, nil
}

//// Dividends
func (blockchain *BlockChain) GetLatestDividendProposal(forDCB bool) (id, amount uint64) {
	return blockchain.BestState.Beacon.GetLatestDividendProposal(forDCB)
}

func (blockchain *BlockChain) GetDividendReceiversForID(dividendID uint64, forDCB bool) ([]privacy.PaymentAddress, []uint64, bool, error) {
	return blockchain.config.DataBase.GetDividendReceiversForID(dividendID, forDCB)
}

//// Crowdsales
func (blockchain *BlockChain) parseProposalCrowdsaleData(proposalTxHash *common.Hash, saleID []byte) *params.SaleData {
	var saleData *params.SaleData
	_, _, _, proposalTx, err := blockchain.GetTransactionByHash(proposalTxHash)
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

func (blockchain *BlockChain) GetCrowdsaleData(saleID []byte) (*params.SaleData, error) {
	key := getSaleDataKeyBeacon(saleID)
	if value, ok := blockchain.BestState.Beacon.Params[key]; ok {
		saleData, err := parseSaleDataValueBeacon(value)
		if err != nil {
			return nil, err
		}
		return saleData, nil
	} else {
		return nil, errors.New("Error getting SaleData from beacon best state")
	}
}

func (blockchain *BlockChain) GetAllCrowdsales() ([]*params.SaleData, error) {
	saleDataList := []*params.SaleData{}
	saleIDs, proposalTxHashes, buyingAmounts, sellingAmounts, err := blockchain.config.DataBase.GetAllCrowdsales()
	if err == nil {
		for i, hash := range proposalTxHashes {
			saleData := blockchain.parseProposalCrowdsaleData(&hash, saleIDs[i])
			if saleData != nil {
				saleData.BuyingAmount = buyingAmounts[i]
				saleData.SellingAmount = sellingAmounts[i]
			}
			saleDataList = append(saleDataList, saleData)
		}
	}
	return saleDataList, err
}

//// Reserve
func (blockchain *BlockChain) GetAssetPrice(assetID *common.Hash) uint64 {
	return blockchain.BestState.Beacon.getAssetPrice(*assetID)
}

//// CMB
func (blockchain *BlockChain) GetCMB(mainAccount []byte) (privacy.PaymentAddress, []privacy.PaymentAddress, uint64, *common.Hash, uint8, uint64, error) {
	reserveAcc, members, capital, hash, state, fine, err := blockchain.config.DataBase.GetCMB(mainAccount)
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

func (blockchain *BlockChain) GetCMBResponse(mainAccount []byte) ([][]byte, error) {
	return blockchain.config.DataBase.GetCMBResponse(mainAccount)
}

func (blockchain *BlockChain) GetDepositSend(contractID []byte) ([]byte, error) {
	return blockchain.config.DataBase.GetDepositSend(contractID)
}

func (blockchain *BlockChain) GetWithdrawRequest(contractID []byte) ([]byte, uint8, error) {
	return blockchain.config.DataBase.GetWithdrawRequest(contractID)
}

func (blockchain *BlockChain) GetAllCommitteeValidatorCandidate() (map[byte][]string, map[byte][]string, []string, []string, []string, []string, []string, []string) {
	beaconBestState := BestStateBeacon{}
	temp, err := blockchain.config.DataBase.FetchBeaconBestState()
	if err != nil {
		panic("Can't Fetch Beacon BestState")
	} else {
		if err := json.Unmarshal(temp, &beaconBestState); err != nil {
			Logger.log.Error(err)
			panic("Fail to unmarshal Beacon BestState")
		}
	}
	SC := beaconBestState.ShardCommittee
	SPV := beaconBestState.ShardPendingValidator
	BC := beaconBestState.BeaconCommittee
	BPV := beaconBestState.BeaconPendingValidator
	CBWFCR := beaconBestState.CandidateBeaconWaitingForCurrentRandom
	CBWFNR := beaconBestState.CandidateBeaconWaitingForNextRandom
	CSWFCR := beaconBestState.CandidateShardWaitingForCurrentRandom
	CSWFNR := beaconBestState.CandidateShardWaitingForNextRandom
	return SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR
}
