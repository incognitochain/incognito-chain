package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"strconv"
)

// PortalRedeemRequest - portal user redeem requests to get public token by burning ptoken
// metadata - redeem request - create normal tx with this metadata
type PortalExpiredWaitingPortingReq struct {
	MetadataBase
	UniquePortingID      string
	ExpiredByLiquidation bool
}

// PortalExpiredWaitingPortingReqContent - Beacon builds a new instruction with this content after detecting user haven't sent public token to custodian
// It will be appended to beaconBlock
type PortalExpiredWaitingPortingReqContent struct {
	MetadataBase
	UniquePortingID      string
	ExpiredByLiquidation bool
	ShardID              byte
}

// PortalExpiredWaitingPortingReqStatus - Beacon tracks status of custodian liquidation into db
type PortalExpiredWaitingPortingReqStatus struct {
	Status               byte
	UniquePortingID      string
	ShardID              byte
	ExpiredByLiquidation bool
	ExpiredBeaconHeight  uint64
}

func NewPortalExpiredWaitingPortingReq(
	metaType int,
	uniquePortingID string,
	expiredByLiquidation bool,
) (*PortalExpiredWaitingPortingReq, error) {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	liquidCustodianMeta := &PortalExpiredWaitingPortingReq{
		UniquePortingID:      uniquePortingID,
		ExpiredByLiquidation: expiredByLiquidation,
	}
	liquidCustodianMeta.MetadataBase = metadataBase
	return liquidCustodianMeta, nil
}

func (expiredPortingReq PortalExpiredWaitingPortingReq) ValidateTxWithBlockChain(
	txr Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	db *statedb.StateDB,
) (bool, error) {
	return true, nil
}

func (expiredPortingReq PortalExpiredWaitingPortingReq) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, txr Transaction) (bool, bool, error) {
	return true, true, nil
}

func (expiredPortingReq PortalExpiredWaitingPortingReq) ValidateMetadataByItself() bool {
	return expiredPortingReq.Type == PortalExpiredWaitingPortingReqMeta
}

func (expiredPortingReq PortalExpiredWaitingPortingReq) Hash() *common.Hash {
	record := expiredPortingReq.MetadataBase.Hash().String()
	record += expiredPortingReq.UniquePortingID
	record += strconv.FormatBool(expiredPortingReq.ExpiredByLiquidation)
	// final hash
	hash := common.HashH([]byte(record))
	return &hash
}

func (expiredPortingReq *PortalExpiredWaitingPortingReq) CalculateSize() uint64 {
	return calculateSize(expiredPortingReq)
}
