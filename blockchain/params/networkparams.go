package params

import (
	"encoding/hex"
	"fmt"

	"github.com/ninjadotorg/constant/common"
)

// Todo: @0xjackalope, @0xbunyip Check logic in Hash and Validate and rpcfunction because other will change params struct without modified these function
type SellingBonds struct {
	BondName       string
	BondSymbol     string
	TotalIssue     uint64
	BondsToSell    uint64
	BondPrice      uint64 // in Constant unit
	Maturity       uint64
	BuyBackPrice   uint64 // in Constant unit
	StartSellingAt uint64 // start selling bonds at block height
	SellingWithin  uint64 // selling bonds within n blocks
}

type SellingGOVTokens struct {
	TotalIssue      uint64
	GOVTokensToSell uint64
	GOVTokenPrice   uint64 // in Constant unit
	StartSellingAt  uint64 // start selling gov tokens at block height
	SellingWithin   uint64 // selling tokens within n blocks
}

func (self SellingBonds) GetID() *common.Hash {
	record := fmt.Sprintf("%d", self.Maturity)
	record += fmt.Sprintf("%d", self.BuyBackPrice)
	record += fmt.Sprintf("%d", self.StartSellingAt)
	temp := common.DoubleHashH([]byte(record))
	bondIDBytesWithPrefix := append(common.BondTokenID[0:8], temp[8:]...)
	result := &common.Hash{}
	result.SetBytes(bondIDBytesWithPrefix)
	return result
}

func NewSellingBonds(
	bondName string,
	bondSymbol string,
	totalIssue uint64,
	bondsToSell uint64,
	bondPrice uint64,
	maturity uint64,
	buyBackPrice uint64,
	startSellingAt uint64,
	sellingWithin uint64,
) *SellingBonds {
	return &SellingBonds{
		BondName:       bondName,
		BondSymbol:     bondSymbol,
		TotalIssue:     totalIssue,
		BondsToSell:    bondsToSell,
		BondPrice:      bondPrice,
		Maturity:       maturity,
		BuyBackPrice:   buyBackPrice,
		StartSellingAt: startSellingAt,
		SellingWithin:  sellingWithin,
	}
}

func NewSellingGOVTokens(
	totalIssue uint64,
	govTokensToSell uint64,
	govTokenPrice uint64,
	startSellingAt uint64,
	sellingWithin uint64,
) *SellingGOVTokens {
	return &SellingGOVTokens{
		TotalIssue:      totalIssue,
		GOVTokensToSell: govTokensToSell,
		GOVTokenPrice:   govTokenPrice,
		StartSellingAt:  startSellingAt,
		SellingWithin:   sellingWithin,
	}
}

func NewSellingBondsFromJson(data interface{}) *SellingBonds {
	sellingBondsData := data.(map[string]interface{})
	sellingBonds := NewSellingBonds(
		sellingBondsData["BondName"].(string),
		sellingBondsData["BondSymbol"].(string),
		uint64(sellingBondsData["TotalIssue"].(float64)),
		uint64(sellingBondsData["BondsToSell"].(float64)),
		uint64(sellingBondsData["BondPrice"].(float64)),
		uint64(sellingBondsData["Maturity"].(float64)),
		uint64(sellingBondsData["BuyBackPrice"].(float64)),
		uint64(sellingBondsData["StartSellingAt"].(float64)),
		uint64(sellingBondsData["SellingWithin"].(float64)),
	)
	return sellingBonds
}

func NewSellingGOVTokensFromJson(data interface{}) *SellingGOVTokens {
	sellingGOVTokensData := data.(map[string]interface{})
	sellingGOVTokens := NewSellingGOVTokens(
		uint64(sellingGOVTokensData["TotalIssue"].(float64)),
		uint64(sellingGOVTokensData["GOVTokensToSell"].(float64)),
		uint64(sellingGOVTokensData["GOVTokenPrice"].(float64)),
		uint64(sellingGOVTokensData["StartSellingAt"].(float64)),
		uint64(sellingGOVTokensData["SellingWithin"].(float64)),
	)
	return sellingGOVTokens
}

type SaleData struct {
	SaleID   []byte // Unique id of the crowdsale to store in db
	EndBlock uint64

	BuyingAsset     common.Hash
	BuyingAmount    uint64
	DefaultBuyPrice uint64

	SellingAsset     common.Hash
	SellingAmount    uint64
	DefaultSellPrice uint64
}

func NewSaleData(
	saleID []byte,
	endBlock uint64,
	buyingAsset *common.Hash,
	buyingAmount uint64,
	defaultBuyPrice uint64,
	sellingAsset *common.Hash,
	sellingAmount uint64,
	defaultSellPrice uint64,
) *SaleData {
	return &SaleData{
		SaleID:           saleID,
		EndBlock:         endBlock,
		BuyingAsset:      *buyingAsset,
		BuyingAmount:     buyingAmount,
		DefaultBuyPrice:  defaultBuyPrice,
		SellingAsset:     *sellingAsset,
		SellingAmount:    sellingAmount,
		DefaultSellPrice: defaultSellPrice,
	}
}

func NewSaleDataFromJson(data interface{}) *SaleData {
	saleDataData := data.(map[string]interface{})
	saleIDStr := saleDataData["SaleID"].(string)
	saleID, errSale := hex.DecodeString(saleIDStr)

	buyingAssetStr := saleDataData["BuyingAsset"].(string)
	buyingAsset, errBuy := common.Hash{}.NewHashFromStr(buyingAssetStr)

	sellingAssetStr := saleDataData["SellingAsset"].(string)
	sellingAsset, errSell := common.Hash{}.NewHashFromStr(sellingAssetStr)
	if errSale != nil || errBuy != nil || errSell != nil {
		return nil
	}

	saleData := NewSaleData(
		saleID,
		uint64(saleDataData["EndBlock"].(float64)),
		buyingAsset,
		uint64(saleDataData["BuyingAmount"].(float64)),
		uint64(saleDataData["DefaultBuyPrice"].(float64)),
		sellingAsset,
		uint64(saleDataData["SellingAmount"].(float64)),
		uint64(saleDataData["DefaultSellPrice"].(float64)),
	)
	return saleData
}

type RefundInfo struct {
	ThresholdToLargeTx uint64
	RefundAmount       uint64
}

func NewRefundInfo(
	thresholdToLargeTx uint64,
	refundAmount uint64,
) *RefundInfo {
	return &RefundInfo{
		ThresholdToLargeTx: thresholdToLargeTx,
		RefundAmount:       refundAmount,
	}
}

func NewRefundInfoFromJson(data interface{}) *RefundInfo {
	refundInfoData := data.(map[string]interface{})
	refundInfo := NewRefundInfo(
		uint64(refundInfoData["ThresholdToLargeTx"].(float64)),
		uint64(refundInfoData["RefundAmount"].(float64)),
	)
	return refundInfo
}

type SaleDCBTokensByUSDData struct {
	Amount   uint64
	EndBlock uint64
}

func NewSaleDCBTokensByUSDData(amount uint64, endBlock uint64) *SaleDCBTokensByUSDData {
	return &SaleDCBTokensByUSDData{Amount: amount, EndBlock: endBlock}
}

func NewSaleDCBTokensByUSDDataFromJson(data interface{}) *SaleDCBTokensByUSDData {
	saleDCBTokensByUSDDataData := data.(map[string]interface{})
	saleDCBTokensByUSDData := NewSaleDCBTokensByUSDData(
		uint64(saleDCBTokensByUSDDataData["Amount"].(float64)),
		uint64(saleDCBTokensByUSDDataData["EndBlock"].(float64)),
	)
	return saleDCBTokensByUSDData
}

type OracleNetwork struct {
	OraclePubKeys          [][]byte
	WrongTimesAllowed      uint8
	Quorum                 uint8
	AcceptableErrorMargin  uint32
	UpdateFrequency        uint32
	OracleRewardMultiplier uint8
}

func NewOracleNetwork(oraclePubKeys [][]byte, wrongTimesAllowed uint8, quorum uint8, acceptableErrorMargin uint32, updateFrequency uint32, oracleRewardMultiplier uint8) *OracleNetwork {
	return &OracleNetwork{OraclePubKeys: oraclePubKeys, WrongTimesAllowed: wrongTimesAllowed, Quorum: quorum, AcceptableErrorMargin: acceptableErrorMargin, UpdateFrequency: updateFrequency, OracleRewardMultiplier: oracleRewardMultiplier}
}

func NewOracleNetworkFromJson(data interface{}) *OracleNetwork {
	oracleNetworkData := data.(map[string]interface{})

	oraclePubKeysInterface := common.InterfaceSlice(oracleNetworkData["OraclePubKeys"])
	if oraclePubKeysInterface == nil {
		panic("oraclePubKey")
	}
	oraclePubKeys := make([][]byte, 0)
	for _, i := range oraclePubKeysInterface {
		oraclePubKeys = append(oraclePubKeys, common.SliceInterfaceToSliceByte(common.InterfaceSlice(i)))
	}

	oracleNetwork := NewOracleNetwork(
		oraclePubKeys,
		uint8(oracleNetworkData["WrongTimesAllowed"].(float64)),
		uint8(oracleNetworkData["Quorum"].(float64)),
		uint32(oracleNetworkData["AcceptableErrorMargin"].(float64)),
		uint32(oracleNetworkData["UpdateFrequency"].(float64)),
		uint8(oracleNetworkData["OracleRewardMultiplier"].(float64)),
	)
	return oracleNetwork
}

func (saleData *SaleData) Hash() *common.Hash {
	record := string(saleData.SaleID)
	record += string(saleData.EndBlock)
	record += saleData.BuyingAsset.String()
	record += string(saleData.BuyingAmount)
	record += string(saleData.DefaultBuyPrice)
	record += saleData.SellingAsset.String()
	record += string(saleData.SellingAmount)
	record += string(saleData.DefaultSellPrice)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sellingBonds *SellingBonds) Hash() *common.Hash {
	record := sellingBonds.BondName
	record += sellingBonds.BondSymbol
	record += string(sellingBonds.BondsToSell)
	record += string(sellingBonds.BondPrice)
	record += string(sellingBonds.Maturity)
	record += string(sellingBonds.BuyBackPrice)
	record += string(sellingBonds.StartSellingAt)
	record += string(sellingBonds.SellingWithin)
	record += string(sellingBonds.TotalIssue)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sellingGOVTokens *SellingGOVTokens) Hash() *common.Hash {
	record := string(sellingGOVTokens.TotalIssue)
	record += string(sellingGOVTokens.GOVTokensToSell)
	record += string(sellingGOVTokens.GOVTokenPrice)
	record += string(sellingGOVTokens.StartSellingAt)
	record += string(sellingGOVTokens.SellingWithin)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (refundInfo *RefundInfo) Hash() *common.Hash {
	record := string(refundInfo.ThresholdToLargeTx)
	record += string(refundInfo.RefundAmount)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (sdt *SaleDCBTokensByUSDData) Hash() *common.Hash {
	record := string(sdt.Amount)
	record += string(sdt.EndBlock)
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (on *OracleNetwork) Hash() *common.Hash {
	record := string(on.WrongTimesAllowed)
	record += string(on.Quorum)
	record += string(on.AcceptableErrorMargin)
	record += string(on.UpdateFrequency)
	for _, oraclePk := range on.OraclePubKeys {
		record += string(oraclePk)
	}
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
