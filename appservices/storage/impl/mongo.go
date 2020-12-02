package impl

import (
	"context"
	"fmt"
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

	//Transaction
	Transaction = "Transaction"

	//InputCoin
	InputCoin = "InputCoin"

	//OutputCoin
	OutputCoin = "OutputCoin"

	//Commitment
	Commitment= "Commitment"

	//TokenState
	TokenState = "TokenState"

	//PDE Collections
	PDEShare               = "PDEShare"
	PDEPoolForPair         = "PDEPoolForPair"
	PDETradingFee          = "PDETradingFee"
	WaitingPDEContribution = "WaitingPDEContribution"

	//Portal Collections

	Custodian             = "Custodian"
	WaitingPortingRequest = "WaitingPortingRequest"
	FinalExchangeRates    = "FinalExchangeRates"
	WaitingRedeemRequest  = "WaitingRedeemRequest"
	MatchedRedeemRequest  = "MatchedRedeemRequest"
	LockedCollateral      = "LockedCollateral"
)

var ctx = context.TODO()

func init() {
	log.Printf("Init mongodb")
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017/")
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


	//Transaction
	transactionStorer *mongoTransactionStorer

	//InputCoin
	inputCoinStorer *mongoInputCoinStorer

	//InputCoin
	outputCoinStorer *mongoOutputCoinStorer

	//Commitment
	commitmentStorer *mongoCommitmentStorer

	//TokenState
	tokenStateStorer *mongoTokenStateStorer

	//PDE
	pdeShareStorer               *mongoPDEShareStorer
	pdePoolForPairStorer         *mongoPDEPoolForPairStorer
	pdeTradingFeeStorer          *mongoPDETradingFeeStorer
	waitingPDEContributionStorer *mongoWaitingPDEContributionStorer

	//Portal
	custodianStorer             *mongoCustodianStorer
	waitingPortingRequestStorer *mongoWaitingPortingRequestStorer
	finalExchangeRatesStorer    *mongoFinalExchangeRatesStorer
	waitingRedeemRequestStorer  *mongoWaitingRedeemRequestStorer
	matchedRedeemRequestStorer  *mongoMatchedRedeemRequestStorer
	lockedCollateralStorer      *mongoLockedCollateralStorer
}

//Beacon
type mongoBeaconStateStorer struct {
	collection *mongo.Collection
}

//Shard
type mongoShardStateStorer struct {
	mongo 		*mongoDBDriver
	prefix     string
	collection [256]*mongo.Collection
}

//Transactions
type mongoTransactionStorer struct {
	mongo 		*mongoDBDriver
	prefix     string
	collection *mongo.Collection
}

//InputCoin
type mongoInputCoinStorer struct {
	mongo 		*mongoDBDriver
	prefix     string
	collection *mongo.Collection
}

//OutputCoin
type mongoOutputCoinStorer struct {
	mongo 		*mongoDBDriver
	prefix     string
	collection *mongo.Collection
}


//Commitment
type mongoCommitmentStorer struct {
	mongo 		*mongoDBDriver
	prefix     string
	collection [256]*mongo.Collection
}

//Token State
type mongoTokenStateStorer struct {
	mongo 		*mongoDBDriver
	prefix     string
	collection [256]*mongo.Collection
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
func (mongo *mongoDBDriver) GetBeaconStorer() repository.BeaconStateStorer {
	if mongo.beaconStateStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(BeaconStateCollection)
		mongo.beaconStateStorer = &mongoBeaconStateStorer{collection: collection}
	}
	return mongo.beaconStateStorer
}

//Shard Get Storer

func (mongo *mongoDBDriver) GetShardStorer() repository.ShardStateStorer {
	if mongo.shardStateStorer == nil {
		mongo.shardStateStorer = &mongoShardStateStorer{prefix: ShardState, mongo: mongo}
	}
	return mongo.shardStateStorer
}

//Get Transaction Storer

func (mongo *mongoDBDriver)  GetTransactionStorer() repository.TransactionStorer {
	if mongo.transactionStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(Transaction)
		mongo.transactionStorer = &mongoTransactionStorer{prefix: Transaction, mongo: mongo, collection: collection}
	}
	return mongo.transactionStorer

}

//Get InputCoin Storer

func (mongo *mongoDBDriver)  GetInputCoinStorer() repository.InputCoinStorer {
	if mongo.inputCoinStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(InputCoin)
		mongo.inputCoinStorer = &mongoInputCoinStorer{prefix: InputCoin, mongo: mongo, collection: collection}
	}
	return mongo.inputCoinStorer

}

//Get OutputCoin Storer

func (mongo *mongoDBDriver)  GetOutputCoinStorer() repository.OutputCoinStorer {
	if mongo.outputCoinStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(OutputCoin)
		mongo.outputCoinStorer = &mongoOutputCoinStorer{prefix: OutputCoin, mongo: mongo, collection: collection}
	}
	return mongo.outputCoinStorer

}

//Get Commitment Storer
func (mongo *mongoDBDriver) GetCommitmentStorer() repository.CommitmentStorer {
	if mongo.commitmentStorer == nil {
		mongo.commitmentStorer = &mongoCommitmentStorer{prefix: Commitment, mongo: mongo}
	}
	return mongo.commitmentStorer
}

//Get TokenState Storer
func (mongo *mongoDBDriver) GetTokenStateStorer() repository.TokenStateStorer {
	if mongo.tokenStateStorer == nil {
		mongo.tokenStateStorer = &mongoTokenStateStorer{prefix: TokenState, mongo: mongo}
	}
	return mongo.tokenStateStorer
}

//PDE Get Storer
func (mongo *mongoDBDriver) GetPDEShareStorer() repository.PDEShareStorer {
	if mongo.pdeShareStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(PDEShare)
		mongo.pdeShareStorer = &mongoPDEShareStorer{collection: collection}
	}
	return mongo.pdeShareStorer
}

func (mongo *mongoDBDriver) GetPDEPoolForPairStateStorer() repository.PDEPoolForPairStateStorer {
	if mongo.pdePoolForPairStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(PDEPoolForPair)
		mongo.pdePoolForPairStorer = &mongoPDEPoolForPairStorer{collection: collection}
	}
	return mongo.pdePoolForPairStorer
}

func (mongo *mongoDBDriver) GetPDETradingFeeStorer() repository.PDETradingFeeStorer {
	if mongo.pdeTradingFeeStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(PDETradingFee)
		mongo.pdeTradingFeeStorer = &mongoPDETradingFeeStorer{collection: collection}
	}
	return mongo.pdeTradingFeeStorer
}

func (mongo *mongoDBDriver) GetWaitingPDEContributionStorer() repository.WaitingPDEContributionStorer {
	if mongo.waitingPDEContributionStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(WaitingPDEContribution)
		mongo.waitingPDEContributionStorer = &mongoWaitingPDEContributionStorer{collection: collection}
	}
	return mongo.waitingPDEContributionStorer
}

//Portal Get Storer

func (mongo *mongoDBDriver) GetCustodianStorer() repository.CustodianStorer {
	if mongo.custodianStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(Custodian)
		mongo.custodianStorer = &mongoCustodianStorer{collection: collection}
	}
	return mongo.custodianStorer
}

func (mongo *mongoDBDriver) GetWaitingPortingRequestStorer() repository.WaitingPortingRequestStorer {
	if mongo.waitingPortingRequestStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(WaitingPortingRequest)
		mongo.waitingPortingRequestStorer = &mongoWaitingPortingRequestStorer{collection: collection}
	}
	return mongo.waitingPortingRequestStorer
}

func (mongo *mongoDBDriver) GetFinalExchangeRatesStorer() repository.FinalExchangeRatesStorer {
	if mongo.finalExchangeRatesStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(FinalExchangeRates)
		mongo.finalExchangeRatesStorer = &mongoFinalExchangeRatesStorer{collection: collection}
	}
	return mongo.finalExchangeRatesStorer
}

func (mongo *mongoDBDriver) GetWaitingRedeemRequestStorer() repository.WaitingRedeemRequestStorer {
	if mongo.waitingRedeemRequestStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(WaitingRedeemRequest)
		mongo.waitingRedeemRequestStorer = &mongoWaitingRedeemRequestStorer{collection: collection}
	}
	return mongo.waitingRedeemRequestStorer
}

func (mongo *mongoDBDriver) GetMatchedRedeemRequestStorer() repository.MatchedRedeemRequestStorer {
	if mongo.matchedRedeemRequestStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(MatchedRedeemRequest)
		mongo.matchedRedeemRequestStorer = &mongoMatchedRedeemRequestStorer{collection: collection}
	}
	return mongo.matchedRedeemRequestStorer
}

func (mongo *mongoDBDriver) GetLockedCollateralStorer() repository.LockedCollateralStorer {
	if mongo.lockedCollateralStorer == nil {
		collection := mongo.client.Database(DataBaseName).Collection(LockedCollateral)
		mongo.lockedCollateralStorer = &mongoLockedCollateralStorer{collection: collection}
	}
	return mongo.lockedCollateralStorer
}

//Store Beacon
func (beaconStorer *mongoBeaconStateStorer) StoreBeaconState(ctx context.Context, beaconState model.BeaconState) error {
	_, err := beaconStorer.collection.InsertOne(ctx, beaconState)
	return err
}

//Store Shard
func (shardStateStorer *mongoShardStateStorer) StoreShardState(ctx context.Context, shardState model.ShardState) error {
	if shardStateStorer.collection[shardState.ShardID]  == nil {
		collectionName := fmt.Sprintf("%s-%d", shardStateStorer.prefix, shardState.ShardID)
		shardStateStorer.collection[shardState.ShardID] =
		 	shardStateStorer.mongo.client.Database(DataBaseName).Collection(collectionName)
	}
	_, err := shardStateStorer.collection[shardState.ShardID].InsertOne(ctx, shardState)
	return err
}

//Store Transaction
func (transactionStorer *mongoTransactionStorer) StoreTransaction (ctx context.Context, transaction model.Transaction) error {
	_, err := transactionStorer.collection.InsertOne(ctx, transaction)
	return err
}

//Store InputCoin
func (inputCoinStorer *mongoInputCoinStorer) StoreInputCoin (ctx context.Context, inputCoin model.InputCoin) error {
	_, err := inputCoinStorer.collection.InsertOne(ctx, inputCoin)
	return err
}

//Store OutputCoin
func (outputCoinStorer *mongoOutputCoinStorer) StoreOutputCoin (ctx context.Context, outputCoin model.OutputCoin) error {
	_, err := outputCoinStorer.collection.InsertOne(ctx, outputCoin)
	return err
}

//Store Commitment
func (commitmentStorer *mongoCommitmentStorer) StoreCommitment(ctx context.Context, commitment model.Commitment) error {
	if commitmentStorer.collection[commitment.ShardId]  == nil {

		collectionName := fmt.Sprintf("%s-%d", commitmentStorer.prefix, commitment.ShardId)
		fmt.Printf("Collection Name %s", collectionName)
		commitmentStorer.collection[commitment.ShardId] =
			commitmentStorer.mongo.client.Database(DataBaseName).Collection(collectionName)
	}
	_, err := commitmentStorer.collection[commitment.ShardId].InsertOne(ctx, commitment)
	return err
}

//Store PDE
func (pdeShareStorer *mongoPDEShareStorer) StorePDEShare(ctx context.Context, pdeShare model.PDEShare) error {
	_, err := pdeShareStorer.collection.InsertOne(ctx, pdeShare)
	return err
}

func (pdePoolForPairStorer *mongoPDEPoolForPairStorer) StorePDEPoolForPairState(ctx context.Context, pdePoolForPair model.PDEPoolForPair) error {
	_, err := pdePoolForPairStorer.collection.InsertOne(ctx, pdePoolForPair)
	return err
}

func (pdePoolForPairStorer *mongoPDETradingFeeStorer) StorePDETradingFee(ctx context.Context, pdeTradingFee model.PDETradingFee) error {
	_, err := pdePoolForPairStorer.collection.InsertOne(ctx, pdeTradingFee)
	return err
}

func (waitingPDEContributionStorer *mongoWaitingPDEContributionStorer) StoreWaitingPDEContribution(ctx context.Context, waitingPDEContribution model.WaitingPDEContribution) error {
	_, err := waitingPDEContributionStorer.collection.InsertOne(ctx, waitingPDEContribution)
	return err
}

//Store Portal
func (custodianStorer *mongoCustodianStorer) StoreCustodian(ctx context.Context, custodian model.Custodian) error {
	_, err := custodianStorer.collection.InsertOne(ctx, custodian)
	return err
}

func (waitingPortingRequestStorer *mongoWaitingPortingRequestStorer) StoreWaitingPortingRequest(ctx context.Context, waitingPortingRequest model.WaitingPortingRequest) error {
	_, err := waitingPortingRequestStorer.collection.InsertOne(ctx, waitingPortingRequest)
	return err
}

func (finalExchangeRatesStorer *mongoFinalExchangeRatesStorer) StoreFinalExchangeRates(ctx context.Context, finalExchangeRates model.FinalExchangeRate) error {
	_, err := finalExchangeRatesStorer.collection.InsertOne(ctx, finalExchangeRates)
	return err
}

func (waitingRedeemRequestStorer *mongoWaitingRedeemRequestStorer) StoreWaitingRedeemRequest(ctx context.Context, redeemRequest model.RedeemRequest) error {
	_, err := waitingRedeemRequestStorer.collection.InsertOne(ctx, redeemRequest)
	return err
}

func (matchedRedeemRequestStorer *mongoMatchedRedeemRequestStorer) StoreMatchedRedeemRequest(ctx context.Context, redeemRequest model.RedeemRequest) error {
	_, err := matchedRedeemRequestStorer.collection.InsertOne(ctx, redeemRequest)
	return err
}

func (lockedCollateralStorer *mongoLockedCollateralStorer) StoreLockedCollateral(ctx context.Context, lockedCollateral model.LockedCollateral) error {
	_, err := lockedCollateralStorer.collection.InsertOne(ctx, lockedCollateral)
	return err
}

//Store Token State data

func (tokenStateStorer *mongoTokenStateStorer)StoreTokenState (ctx context.Context, tokenState model.TokenState) error {
	if tokenStateStorer.collection[tokenState.ShardID]  == nil {
		collectionName := fmt.Sprintf("%s-%d", tokenStateStorer.prefix, tokenState.ShardID)
		tokenStateStorer.collection[tokenState.ShardID] =
			tokenStateStorer.mongo.client.Database(DataBaseName).Collection(collectionName)
	}
	_, err := tokenStateStorer.collection[tokenState.ShardID].InsertOne(ctx, tokenState)
	return err
}
