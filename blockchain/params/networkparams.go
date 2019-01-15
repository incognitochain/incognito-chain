package params

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
)

// Todo: @0xjackalope, @0xbunyip Check logic in Hash2 and Validate and rpcfunction because other will change params struct without modified these function
type SellingBonds struct {
	BondName       string
	BondSymbol     string
	TotalIssue     uint64
	BondsToSell    uint64
	BondPrice      uint64 // in Constant unit
	Maturity       uint32
	BuyBackPrice   uint64 // in Constant unit
	StartSellingAt uint32 // start selling bonds at block height
	SellingWithin  uint32 // selling bonds within n blocks
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
	maturity uint32,
	buyBackPrice uint64,
	startSellingAt uint32,
	sellingWithin uint32,
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

func NewSellingBondsFromJson(data interface{}) *SellingBonds {
	sellingBondsData := data.(map[string]interface{})
	sellingBonds := NewSellingBonds(
		sellingBondsData["BondName"].(string),
		sellingBondsData["BondSymbol"].(string),
		uint64(sellingBondsData["TotalIssue"].(float64)),
		uint64(sellingBondsData["BondsToSell"].(float64)),
		uint64(sellingBondsData["BondPrice"].(float64)),
		uint32(sellingBondsData["Maturity"].(float64)),
		uint64(sellingBondsData["BuyBackPrice"].(float64)),
		uint32(sellingBondsData["StartSellingAt"].(float64)),
		uint32(sellingBondsData["SellingWithin"].(float64)),
	)
	return sellingBonds
}

type SaleData struct {
	SaleID   []byte // Unique id of the crowdsale to store in db
	EndBlock int32

	BuyingAsset  []byte
	BuyingAmount uint64 // TODO(@0xbunyip): change to big.Int

	SellingAsset  []byte
	SellingAmount uint64
}

func NewSaleData(saleID []byte, endBlock int32, buyingAsset []byte, buyingAmount uint64, sellingAsset []byte, sellingAmount uint64) *SaleData {
	return &SaleData{
		SaleID:        saleID,
		EndBlock:      endBlock,
		BuyingAsset:   buyingAsset,
		BuyingAmount:  buyingAmount,
		SellingAsset:  sellingAsset,
		SellingAmount: sellingAmount,
	}
}

func NewSaleDataFromJson(data interface{}) *SaleData {
	saleDataData := data.(map[string]interface{})
	saleData := NewSaleData(
		[]byte(saleDataData["SaleID"].(string)),
		int32(saleDataData["EndBlock"].(float64)),
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData["BuyingAsset"])),
		uint64(saleDataData["BuyingAmount"].(float64)),
		common.SliceInterfaceToSliceByte(common.InterfaceSlice(saleDataData["SellingAsset"])),
		uint64(saleDataData["SellingAmount"].(float64)),
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
	EndBlock int32
}

func NewSaleDCBTokensByUSDData(amount uint64, endBlock int32) *SaleDCBTokensByUSDData {
	return &SaleDCBTokensByUSDData{Amount: amount, EndBlock: endBlock}
}

func NewSaleDCBTokensByUSDDataFromJson(data interface{}) *SaleDCBTokensByUSDData {
	saleDCBTokensByUSDDataData := data.(map[string]interface{})
	saleDCBTokensByUSDData := NewSaleDCBTokensByUSDData(
		uint64(saleDCBTokensByUSDDataData["Amount"].(float64)),
		int32(saleDCBTokensByUSDDataData["EndBlock"].(float64)),
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
	record := ""
	for _, i := range saleData.SaleID {
		record += string(i)
	}
	for _, i := range saleData.BuyingAsset {
		record += string(i)
	}
	record += string(saleData.EndBlock)
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
