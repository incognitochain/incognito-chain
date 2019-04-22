package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	privacy "github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/wallet"
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

func (blockchain *BlockChain) GetBoardPubKeys(boardType common.BoardType) [][]byte {
	if boardType == common.DCBBoard {
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

func (blockchain *BlockChain) GetBoardPaymentAddress(boardType common.BoardType) []privacy.PaymentAddress {
	if boardType == common.DCBBoard {
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

func (blockchain *BlockChain) GetDCBParams() component.DCBParams {
	return blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams
}

func (blockchain *BlockChain) GetGOVParams() component.GOVParams {
	return blockchain.BestState.Beacon.StabilityInfo.GOVConstitution.GOVParams
}

//// Loan
func (blockchain *BlockChain) GetLoanReq(loanID []byte) (*common.Hash, error) {
	key := getLoanRequestKeyBeacon(loanID)
	reqHash, ok := blockchain.BestState.Beacon.Params[key]
	if !ok {
		return nil, errors.Errorf("Loan request with ID %x not found", loanID)
	}
	return common.NewHashFromStr(reqHash)
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

func (blockchain *BlockChain) GetLoanWithdrawed(loanID []byte) (bool, error) {
	return blockchain.config.DataBase.GetLoanWithdrawed(loanID)
}

//// Crowdsales
func (blockchain *BlockChain) parseProposalCrowdsaleData(proposalTxHash *common.Hash, saleID []byte) *component.SaleData {
	var saleData *component.SaleData
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

// GetProposedCrowdsale returns SaleData from BeaconBestState; BuyingAmount and SellingAmount might be outdated, the rest is ok to use
func (blockchain *BlockChain) GetSaleData(saleID []byte) (*component.SaleData, error) {
	saleRaw, err := blockchain.config.DataBase.GetSaleData(saleID)
	if err != nil {
		return nil, err
	}
	var sale *component.SaleData
	err = json.Unmarshal(saleRaw, sale)
	return sale, err
}

func (blockchain *BlockChain) GetDCBBondInfo(bondID *common.Hash) (uint64, uint64) {
	return blockchain.config.DataBase.GetDCBBondInfo(bondID)
}

func (blockchain *BlockChain) GetAllSaleData() ([]*component.SaleData, error) {
	data, err := blockchain.config.DataBase.GetAllSaleData()
	if err != nil {
		return nil, err
	}
	sales := []*component.SaleData{}
	for _, saleRaw := range data {
		var sale *component.SaleData
		err := json.Unmarshal(saleRaw, sale)
		if err != nil {
			return nil, err
		}
		sales = append(sales, sale)
	}
	return sales, nil
}

func (blockchain *BlockChain) CrowdsaleExisted(saleID []byte) bool {
	_, err := blockchain.config.DataBase.GetSaleData(saleID)
	if err != nil {
		return true
	}
	return false
}

// GetDCBAvailableAsset returns number of token left accounted for all on-going crowdsales
func (blockchain *BlockChain) GetDCBAvailableAsset(assetID *common.Hash) uint64 {
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	vouts, err := blockchain.GetUnspentTxCustomTokenVout(keyWalletDCBAccount.KeySet, assetID)
	if err != nil {
		return 0
	}
	tokenLeft := uint64(0)
	for _, vout := range vouts {
		tokenLeft += vout.Value
	}

	sales, _ := blockchain.GetAllSaleData()
	for _, sale := range sales {
		if !sale.Buy && sale.BondID.IsEqual(assetID) && sale.EndBlock < blockchain.GetBeaconHeight() {
			if sale.Amount >= tokenLeft {
				tokenLeft = 0
			} else {
				tokenLeft -= sale.Amount
			}
		}
	}
	return tokenLeft
}

//// Reserve
func (blockchain *BlockChain) GetAssetPrice(assetID *common.Hash) uint64 {
	return blockchain.BestState.Beacon.getAssetPrice(*assetID)
}

//// Trade bonds
func (blockchain *BlockChain) GetAllTrades() []*component.TradeBondWithGOV {
	return blockchain.BestState.Beacon.StabilityInfo.DCBConstitution.DCBParams.TradeBonds
}

func (blockchain *BlockChain) GetTradeActivation(tradeID []byte) (*common.Hash, bool, bool, uint64, error) {
	return blockchain.config.DataBase.GetTradeActivation(tradeID)
}

// GetLatestTradeActivation returns trade activation from local state if exist, otherwise get from current proposal
func (blockchain *BlockChain) GetLatestTradeActivation(tradeID []byte) (*common.Hash, bool, bool, uint64, error) {
	bondID, buy, activated, amount, err := blockchain.config.DataBase.GetTradeActivation(tradeID)
	if err == nil {
		return bondID, buy, activated, amount, nil
	}
	for _, trade := range blockchain.GetAllTrades() {
		if bytes.Equal(trade.TradeID, tradeID) {
			activated := false
			return trade.BondID, trade.Buy, activated, trade.Amount, nil
		}
	}
	return nil, false, false, 0, errors.New("no trade found")
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
