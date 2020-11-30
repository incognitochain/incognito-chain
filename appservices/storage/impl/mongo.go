package impl

import (
	"context"
	"github.com/incognitochain/incognito-chain/appservices/storage"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
	"github.com/incognitochain/incognito-chain/appservices/storage/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const (
	DataBaseName = "Incognito"
	//Beacon
	BeaconStateCollection = "BeaconState"

	//Shard
	ShardState = "ShardState"

	//PDE Collections
	PDEShare	= "PDEShare"
	PDEPoolForPair = "PDEPoolForPair"
	PDETradingFee = "PDETradingFee"
	WaitingPDEContribution ="WaitingPDEContribution"

	//Portal Collections

	Custodian = "Custodian"
	WaitingPortingRequest = "WaitingPortingRequest"
	FinalExchangeRates = "FinalExchangeRates"
	WaitingRedeemRequest = "WaitingRedeemRequest"
	MatchedRedeemRequest = "MatchedRedeemRequest"
	LockedCollateral = "LockedCollateral"


)

var ctx = context.TODO()

func init ()  {
	log.Printf("Init mongodb")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	mongoDBDriver := &mongoDBDriver{client: client, beaconStateStorer: nil}

	err = storage.AddDBDriver(storage.MONGODB, mongoDBDriver)

	if err != nil {
		log.Fatal(err)
	}

}

type mongoDBDriver struct {
	client *mongo.Client
	//Beacon
	beaconStateStorer *mongoBeaconStateStorer

	//Shard
	shardStateStorer *mongoShardStateStorer

	//PDE
	pdeShareStorer *mongoPDEShareStorer
	pdePoolForPairStorer *mongoPDEPoolForPairStorer
	pdeTradingFeeStorer *mongoPDETradingFeeStorer
	waitingPDEContributionStorer *mongoWaitingPDEContributionStorer

	//Portal
	custodianStorer *mongoCustodianStorer
	waitingPortingRequestStorer *mongoWaitingPortingRequestStorer
	finalExchangeRatesStorer *mongoFinalExchangeRatesStorer
	waitingRedeemRequestStorer *mongoWaitingRedeemRequestStorer
	matchedRedeemRequestStorer *mongoMatchedRedeemRequestStorer
	lockedCollateralStorer *mongoLockedCollateralStorer
}
//Beacon
type mongoBeaconStateStorer struct {
	collection *mongo.Collection
}

//Shard
type mongoShardStateStorer struct {
	collection *mongo.Collection
}

//PDE
type mongoPDEShareStorer struct {
	collection *mongo.Collection
}

type mongoPDEPoolForPairStorer struct {
	collection *mongo.Collection
}

type mongoPDETradingFeeStorer struct {
	collection *mongo.Collection
}

type mongoWaitingPDEContributionStorer struct {
	collection *mongo.Collection
}

//Portal
type mongoCustodianStorer struct {
	collection *mongo.Collection
}
type mongoWaitingPortingRequestStorer struct {
	collection *mongo.Collection
}
type mongoFinalExchangeRatesStorer struct {
	collection *mongo.Collection
}
type mongoWaitingRedeemRequestStorer struct {
	collection *mongo.Collection
}
type mongoMatchedRedeemRequestStorer struct {
	collection *mongo.Collection
}
type mongoLockedCollateralStorer struct {
	collection *mongo.Collection
}

//Beacon Get Storer
func (mongo *mongoDBDriver) GetBeaconStorer() repository.BeaconStateStorer  {
	if mongo.beaconStateStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(BeaconStateCollection)
		mongo.beaconStateStorer = &mongoBeaconStateStorer{collection: collection}
	}
	return mongo.beaconStateStorer
}

//Shard Get Storer

func (mongo *mongoDBDriver) GetShardStorer() repository.ShardStateStorer  {
	if mongo.shardStateStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(ShardState)
		mongo.shardStateStorer = &mongoShardStateStorer{collection: collection}
	}
	return mongo.shardStateStorer
}

//PDE Get Storer
func (mongo *mongoDBDriver) GetPDEShareStorer () repository.PDEShareStorer  {
	if mongo.pdeShareStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(PDEShare)
		mongo.pdeShareStorer = &mongoPDEShareStorer{collection: collection}
	}
	return mongo.pdeShareStorer
}

func (mongo *mongoDBDriver) GetPDEPoolForPairStorer () repository.PDEPoolForPairStateStorer  {
	if mongo.pdePoolForPairStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(PDEPoolForPair)
		mongo.pdePoolForPairStorer = &mongoPDEPoolForPairStorer{collection: collection}
	}
	return mongo.pdePoolForPairStorer
}

func (mongo *mongoDBDriver) GetPDETradingFeeStorer () repository.PDETradingFeeStorer  {
	if mongo.pdeTradingFeeStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(PDETradingFee)
		mongo.pdeTradingFeeStorer = &mongoPDETradingFeeStorer{collection: collection}
	}
	return mongo.pdeTradingFeeStorer
}

func (mongo *mongoDBDriver) GetWaitingPDEContributionStorer () repository.WaitingPDEContributionStorer  {
	if mongo.waitingPDEContributionStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(WaitingPDEContribution)
		mongo.waitingPDEContributionStorer = &mongoWaitingPDEContributionStorer{collection: collection}
	}
	return mongo.waitingPDEContributionStorer
}


//Portal Get Storer

func (mongo *mongoDBDriver) GetCustodianStorer () repository.CustodianStorer  {
	if mongo.custodianStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(Custodian)
		mongo.custodianStorer = &mongoCustodianStorer{collection: collection}
	}
	return mongo.custodianStorer
}

func (mongo *mongoDBDriver) GetWaitingPortingRequestStorer () repository.WaitingPortingRequestStorer  {
	if mongo.waitingPortingRequestStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(WaitingPortingRequest)
		mongo.waitingPortingRequestStorer = &mongoWaitingPortingRequestStorer{collection: collection}
	}
	return mongo.waitingPortingRequestStorer
}

func (mongo *mongoDBDriver) GetFinalExchangeRatesStorer () repository.FinalExchangeRatesStorer  {
	if mongo.finalExchangeRatesStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(FinalExchangeRates)
		mongo.finalExchangeRatesStorer = &mongoFinalExchangeRatesStorer{collection: collection}
	}
	return mongo.finalExchangeRatesStorer
}



func (mongo *mongoDBDriver) GetWaitingRedeemRequestStorer  () repository.WaitingRedeemRequestStorer   {
	if mongo.waitingRedeemRequestStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(WaitingRedeemRequest)
		mongo.waitingRedeemRequestStorer = &mongoWaitingRedeemRequestStorer{collection: collection}
	}
	return mongo.waitingRedeemRequestStorer
}

func (mongo *mongoDBDriver) GetMatchedRedeemRequestStorer () repository.MatchedRedeemRequestStorer  {
	if mongo.matchedRedeemRequestStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(MatchedRedeemRequest)
		mongo.matchedRedeemRequestStorer = &mongoMatchedRedeemRequestStorer{collection: collection}
	}
	return mongo.matchedRedeemRequestStorer
}

func (mongo *mongoDBDriver) GetLockedCollateralStorer () repository.LockedCollateralStorer  {
	if mongo.lockedCollateralStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(LockedCollateral)
		mongo.lockedCollateralStorer = &mongoLockedCollateralStorer{collection: collection}
	}
	return mongo.lockedCollateralStorer
}


//Store Beacon
func (beaconStorer *mongoBeaconStateStorer) StoreBeaconState (ctx context.Context, beaconState model.BeaconState) error {
	_, err := beaconStorer.collection.InsertOne(ctx, beaconState)
	return err
}

//Store Shar
func (shardStateStorer *mongoShardStateStorer) StoreShardState (ctx context.Context, shardState model.ShardState) error {
	_, err := shardStateStorer.collection.InsertOne(ctx, shardState)
	return err
}

//Store PDE
func (pdeShareStorer *mongoPDEShareStorer) StorePDEShare (ctx context.Context, pdeShare model.PDEShare) error {
	_, err := pdeShareStorer.collection.InsertOne(ctx, pdeShare)
	return err
}

func (pdePoolForPairStorer *mongoPDEPoolForPairStorer) StorePDEPoolForPairState (ctx context.Context, pdePoolForPair model.PDEPoolForPair) error {
	_, err := pdePoolForPairStorer.collection.InsertOne(ctx, pdePoolForPair)
	return err
}

func (pdePoolForPairStorer *mongoPDETradingFeeStorer) StorePDETradingFee (ctx context.Context, pdeTradingFee model.PDETradingFee) error {
	_, err := pdePoolForPairStorer.collection.InsertOne(ctx, pdeTradingFee)
	return err
}


func (waitingPDEContributionStorer *mongoWaitingPDEContributionStorer) StoreWaitingPDEContribution (ctx context.Context, waitingPDEContribution model.WaitingPDEContribution) error {
	_, err := waitingPDEContributionStorer.collection.InsertOne(ctx, waitingPDEContribution)
	return err
}

//Store Portal
func (custodianStorer *mongoCustodianStorer) StoreCustodian (ctx context.Context, custodian model.Custodian) error {
	_, err := custodianStorer.collection.InsertOne(ctx, custodian)
	return err
}

func (waitingPortingRequestStorer *mongoWaitingPortingRequestStorer) StoreWaitingPortingRequest (ctx context.Context, waitingPortingRequest model.WaitingPortingRequest) error {
	_, err := waitingPortingRequestStorer.collection.InsertOne(ctx, waitingPortingRequest)
	return err
}


func (finalExchangeRatesStorer *mongoFinalExchangeRatesStorer) 	StoreFinalExchangeRates(ctx context.Context, finalExchangeRates model.FinalExchangeRates) error {
	_, err := finalExchangeRatesStorer.collection.InsertOne(ctx, finalExchangeRates)
	return err
}


func (waitingRedeemRequestStorer *mongoWaitingRedeemRequestStorer) StoreWaitingRedeemRequest (ctx context.Context, redeemRequest model.RedeemRequest) error {
	_, err := waitingRedeemRequestStorer.collection.InsertOne(ctx, redeemRequest)
	return err
}

func (matchedRedeemRequestStorer *mongoMatchedRedeemRequestStorer) StoreMatchedRedeemRequest (ctx context.Context, redeemRequest model.RedeemRequest) error {
	_, err := matchedRedeemRequestStorer.collection.InsertOne(ctx, redeemRequest)
	return err
}

func (lockedCollateralStorer *mongoLockedCollateralStorer) StoreLockedCollateral (ctx context.Context, lockedCollateral model.LockedCollateral) error {
	_, err := lockedCollateralStorer.collection.InsertOne(ctx, lockedCollateral)
	return err
}















