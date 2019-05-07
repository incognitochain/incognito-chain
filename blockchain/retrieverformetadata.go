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
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
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
	sale := &component.SaleData{}
	err = json.Unmarshal(saleRaw, &sale)
	return sale, err
}

func (blockchain *BlockChain) GetDCBBondInfo(bondID *common.Hash) (uint64, uint64) {
	return blockchain.config.DataBase.GetDCBBondInfo(bondID)
}

func (blockchain *BlockChain) GetAllSaleData() ([]*component.SaleData, error) {
	data, err := blockchain.config.DataBase.GetAllSaleData()
	if err == lvdberr.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	sales := []*component.SaleData{}
	for _, saleRaw := range data {
		sale := &component.SaleData{}
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
	return err == nil
}

// GetDCBBondInfo returns amount of bonds owned by DCB that is free to trade
// (not being hold up in other trades/crowdsales)
func (blockchain *BlockChain) GetDCBFreeBond(bondID *common.Hash) uint64 {
	amount, _ := blockchain.GetDCBBondInfo(bondID)

	// Subtract amounts from ongoing crowdsales that are selling the same bond
	sales, _ := blockchain.GetAllSaleData()
	for _, sale := range sales {
		if sale.Buy || !sale.BondID.IsEqual(bondID) || sale.EndBlock < blockchain.GetBeaconHeight() {
			continue
		}
		if sale.Amount >= amount {
			amount = 0
		} else {
			amount -= sale.Amount
		}
	}

	// Subtract amounts from ongoing trades that proposed selling the same bonds to GOV
	trades := blockchain.GetAllTrades()
	for _, t := range trades {
		if t.Buy || !t.BondID.IsEqual(bondID) {
			continue
		}
		if t.Amount >= amount {
			amount = 0
		} else {
			amount -= t.Amount
		}
	}
	return amount
}

//// Reserve
func (blockchain *BlockChain) GetAssetPrice(assetID *common.Hash) uint64 {
	return blockchain.BestState.Beacon.GetAssetPrice(*assetID)
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
