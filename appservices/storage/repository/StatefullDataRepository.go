package repository

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
"github.com/incognitochain/incognito-chain/common"
)

type PDEShareStorer interface {
	StorePDEShare (ctx context.Context,  beaconState model.PDEShare) error
}

type PDEShareRetriver interface {
	GetPDEShareByBeaconHash (hash common.Hash, offset uint, limit uint) []model.PDEShare
	GetPDEShareByBeaconHeight(height uint64, offset uint, limit uint) []model.PDEShare
	GetLatestPDEShareState (offset uint, limit uint) []model.PDEShare
}

type PDEShareRepository interface {
	PDEShareStorer
	PDEShareRetriver
}

type PDEPoolForPairStateStorer interface {
	StorePDEPoolForPairState (ctx context.Context, pdePoolForPair model.PDEPoolForPair) error
}

type PDEPoolForPairRetriver interface {
	GetPDEShareByBeaconHash (hash common.Hash, offset uint, limit uint) []model.PDEPoolForPair
	GetPDEShareByBeaconHeight(height uint64, offset uint, limit uint) []model.PDEPoolForPair
	GetLatestPDEShareState (offset uint, limit uint) []model.PDEPoolForPair

}

type PDEPoolForPairRepository interface {
	PDEPoolForPairStateStorer
	PDEPoolForPairRetriver
}

type PDETradingFeeStorer interface {
	StorePDETradingFee (ctx context.Context, pdeTradingFee model.PDETradingFee) error
}

type PDETradingFeeRetriver interface {
	GetPDETradingFeeByBeaconHash (hash common.Hash, offset uint, limit uint) []model.PDETradingFee
	GetPDETradingFeeByBeaconHeight(height uint64, offset uint, limit uint) []model.PDETradingFee
	GetLatestPDETradingFeeState (offset uint, limit uint) []model.PDETradingFee

}

type PDEFeeRepository interface {
	PDETradingFeeStorer
	PDETradingFeeRetriver
}

type WaitingPDEContributionStorer interface {
	StoreWaitingPDEContribution (ctx context.Context, waitingPDEContribution model.WaitingPDEContribution) error
}

type WaitingPDEContributionRetriver interface {
	GetWaitingPDEContributionByBeaconHash (hash common.Hash, offset uint, limit uint) []model.WaitingPDEContribution
	GetWaitingPDEContributionByBeaconHeight(height uint64, offset uint, limit uint) []model.WaitingPDEContribution
	GetLatestWaitingPDEContributionState (offset uint, limit uint) []model.WaitingPDEContribution

}

type WaitingPDEContributionRepository interface {
	WaitingPDEContributionStorer
	WaitingPDEContributionRetriver
}

type CustodianStorer interface {
	StoreCustodian (ctx context.Context, custodian model.Custodian) error
}

type CustodianRetriver interface {
	GetCustodianByBeaconHash (hash common.Hash, offset uint, limit uint) []model.Custodian
	GetCustodianByBeaconHeight(height uint64, offset uint, limit uint) []model.Custodian
	GetLatestCustodianState (offset uint, limit uint) []model.Custodian

}

type CustodianRepository interface {
	CustodianStorer
	CustodianRetriver
}


type WaitingPortingRequestStorer interface {
	StoreWaitingPortingRequest (ctx context.Context, waitingPortingRequest model.WaitingPortingRequest) error
}

type WaitingPortingRequestRetriver interface {
	GetWaitingPortingRequestByBeaconHash (hash common.Hash, offset uint, limit uint) []model.WaitingPortingRequest
	GetWaitingPortingRequestByBeaconHeight(height uint64, offset uint, limit uint) []model.WaitingPortingRequest
	GetLatestWaitingPortingRequestState (offset uint, limit uint) []model.WaitingPortingRequest

}

type WaitingPortingRequestRepository interface {
	WaitingPortingRequestStorer
	WaitingPortingRequestRetriver
}


type FinalExchangeRatesStorer interface {
	StoreFinalExchangeRates(ctx context.Context, finalExchangeRates model.FinalExchangeRate) error
}

type FinalExchangeRatesRetriver interface {
	GetFinalExchangeRatesByBeaconHash (hash common.Hash, offset uint, limit uint) []model.FinalExchangeRate
	GetFinalExchangeRatesByBeaconHeight(height uint64, offset uint, limit uint) []model.FinalExchangeRate
	GetLatestFinalExchangeRatesState (offset uint, limit uint) []model.FinalExchangeRate

}

type FinalExchangeRatesRepository interface {
	FinalExchangeRatesStorer
	FinalExchangeRatesRetriver
}

type WaitingRedeemRequestStorer interface {
	StoreWaitingRedeemRequest (ctx context.Context, redeemRequest model.RedeemRequest) error
}

type WaitingRedeemRequestRetriver interface {
	GetWaitingRedeemRequestByBeaconHash (hash common.Hash, offset uint, limit uint) []model.RedeemRequest
	GetWaitingRedeemRequestByBeaconHeight(height uint64, offset uint, limit uint) []model.RedeemRequest
	GetWaitingRedeemRequestState (offset uint, limit uint) []model.RedeemRequest

}

type WaitingRedeemRequestRepository interface {
	WaitingRedeemRequestStorer
	WaitingRedeemRequestRetriver
}

type MatchedRedeemRequestStorer interface {
	StoreMatchedRedeemRequest (ctx context.Context, redeemRequest model.RedeemRequest) error
}

type MatchedRedeemRequestRetriver interface {
	GetMatchedRedeemRequestByBeaconHash (hash common.Hash, offset uint, limit uint) []model.RedeemRequest
	GetMatchedRedeemRequestByBeaconHeight(height uint64, offset uint, limit uint) []model.RedeemRequest
	GetLatestMatchedRedeemRequestState (offset uint, limit uint) []model.RedeemRequest

}

type MatchedRedeemRequestRepository interface {
	MatchedRedeemRequestStorer
	MatchedRedeemRequestRetriver
}

type LockedCollateralStorer interface {
	StoreLockedCollateral (ctx context.Context, lockedCollateral model.LockedCollateral) error
}

type LockedCollateralRetriver interface {
	GetLockedCollateralByBeaconHash (hash common.Hash, offset uint, limit uint) []model.LockedCollateral
	GetLockedCollateralByBeaconHeight(height uint64, offset uint, limit uint) []model.LockedCollateral
	GetLockedCollateralState (offset uint, limit uint) []model.LockedCollateral

}



type LockedCollateralRepository interface {
	LockedCollateralStorer
	LockedCollateralRetriver
}


type TokenStateStorer interface {
	StoreTokenState (ctx context.Context, tokenState model.TokenState) error
}


type CommitteeRewardStateStorer interface {
	StoreCommitteeRewardState (ctx context.Context, tokenState model.CommitteeRewardState) error
}