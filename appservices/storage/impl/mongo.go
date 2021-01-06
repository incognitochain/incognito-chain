package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/appservices/data"
	"github.com/incognitochain/incognito-chain/appservices/storage"
	"github.com/incognitochain/incognito-chain/appservices/storage/model"
	"github.com/incognitochain/incognito-chain/appservices/storage/repository"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	NumOfShard = 32

	DataBaseName = "Incognito-Testnet"

	//Beacon
	BeaconState = "BeaconState"

	//Shard
	ShardState = "ShardState"

	//Transaction
	Transaction = "Transaction"

	PublicKeyToTransactionHash = "PublicKeyToTransactionHash"

	//InputCoin
	InputCoin = "InputCoin"

	//Shard OutputCoin
	ShardOutputCoin = "ShardOutputCoin"

	//Commitment
	ShardCommitmentIndex = "ShardCommitmentIndex"

	//Cross Shard Output Coin
	CrossShardOutputCoin= "CrossShardOutputCoin"

	//TokenState
	TokenState = "TokenState"

	//BridgeTokenState
	BridgeTokenState = "BridgeTokenState"

	//RewardState
	RewardState = "RewardState"

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

func IsMongoDupKey(err error) bool {
/*	wce, ok := err.(mongo.WriteError)
	if !ok {
		log.Printf("%v", err)
		return false
	}
	log.Printf("message: %s  code: %d", wce.Message, wce.Code)
	return wce.Code == 11000 || wce.Code == 11001 || wce.Code == 12582 || wce.Code == 16460 && strings.Contains(wce.Message, " E11000 ")*/
	log.Printf("message: %s", err.Error())
	return strings.Contains(err.Error(), "E11000 duplicate key error")
}

func LoadMongoDBDriver(dbConnectionString string) error {
	log.Printf("Init mongodb to server %s", dbConnectionString)
	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(dbConnectionString)
	client, err := mongo.Connect( ctx, clientOptions)
	if err != nil {
		return err
	}

	mongoDBDriver := &mongoDBDriver{client: client}
	err = mongoDBDriver.createIndex(ctx)
	if err != nil {
		return err
	}

	err = storage.AddDBDriver(storage.MONGODB, mongoDBDriver)

	if err != nil {
		return err
	}
	return nil
}

type mongoDBDriver struct {
	client *mongo.Client

	//Beacon
	beaconCollection *mongo.Collection

	//Shard
	shardCollection [256]*mongo.Collection

	//Transaction
	transactionCollection [256]*mongo.Collection

	//PublicKey To Transaction Hash
	publicKeyToTransactionHashCollection *mongo.Collection

	//InputCoin
	inputCoinCollection [256]*mongo.Collection

	//Shard OutputCoin
	shardOutputCoinCollection [256]*mongo.Collection

	//Cross Shard OutputCoin
	crossShardOutputCoinCollection [256]*mongo.Collection

	//Commitment
	shardCommitmentIndexCollection [256]*mongo.Collection

	//Token State
	tokenStateCollection [256]*mongo.Collection

	//Bridge Token State
	bridgeTokenStateCollection *mongo.Collection

	//Reward State
	rewardStateCollection [256]*mongo.Collection

	//PDE
	pdeShareCollection *mongo.Collection
	pdePoolForPairCollection *mongo.Collection
	pdeTradingFeeCollection *mongo.Collection
	waitingPDEContributionCollection *mongo.Collection

	//Portal
	custodianCollection *mongo.Collection
	waitingPortingRequestCollection *mongo.Collection
	finalExchangeRatesCollection *mongo.Collection
	waitingRedeemRequestCollection *mongo.Collection
	matchedRedeemRequestCollection *mongo.Collection
	lockedCollateralCollection *mongo.Collection

}

func (m *mongoDBDriver) GetBeaconStateRepository () repository.BeaconStateRepository {
	return m
}

func (m *mongoDBDriver) GetShardStateRepository () repository.ShardStateRepository {
	return m
}

func (m *mongoDBDriver)  createIndex(ctx context.Context) error {
	log.Printf("Init Index")

	//Beacon
	if err := m.createIndexForBeaconCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Beacon Index")

	//Shard
	if err := m.createIndexForShardCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Shard Index")

	//Transaction
	if err := m.createIndexForTransactionCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Transaction Index")

	//PublicKey To Transaction Hash
	if err := m.createIndexForPublicKeyToTransactionHashCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish PublicKey Index")

	//InputCoin
	if err := m.createIndexForInputCoinCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Input Coin Index")

	//Shard OutputCoin
	if err := m.createIndexForShardOutputCoinCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Shard Output Coin Index")

	//Cross Shard OutputCoin
	if err := m.createIndexForCrossShardOutputCoinCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Cross Shard Output Coin Index")


	//Commitment
	if err := m.createIndexForShardCommitmentIndexCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Coin Comitment Index")

	//Token State
	if err := m.createIndexForTokenState(ctx); err != nil {
		return err
	}
	log.Printf("Finish Token State Index")

	//Bridge Token State
	if err := m.createIndexForBridgeTokenCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish Bridge Token State Index")

	//Reward State
	if err := m.createIndexForRewardState(ctx); err != nil {
		return err
	}
	log.Printf("Finish Reward State Index")

	//PDE
	if err := m.createIndexForPDEPoolForPairCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForPDEShareCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForPDETradingFeeCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForWaitingPDEContributionCollection(ctx); err != nil {
		return err
	}
	log.Printf("Finish PDE Index")

	//Custodian
	if err := m.createIndexForCustodianCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForWaitingPortingRequestCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForFinalExchangeRatesCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForWaitingRedeemRequestCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForMatchedRedeemRequestCollection(ctx); err != nil {
		return err
	}

	if err := m.createIndexForLockedCollateralCollection(ctx); err != nil {
		return err
	}

	log.Printf("Finish Custodian Index")
	log.Printf("Finish Init Index")

	return nil
}

func (m *mongoDBDriver) createIndexForBeaconCollection(ctx context.Context) error {
	m.beaconCollection = m.client.Database(DataBaseName).Collection(BeaconState)
	indexView := m.beaconCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "blockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "height" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForShardCollection(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", ShardState, i)
		m.shardCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.shardCollection[i].Indexes()
		blockHashIndex := mongo.IndexModel{
			Keys:    bson.D{bson.E{Key: "blockhash", Value: 1}},
			Options: options.Index().SetUnique(true),
		}
		blockHeightIndex := mongo.IndexModel{
			Keys: bson.D{bson.E{Key: "height", Value: 1}},
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex}); err != nil {
			return err
		}
	}
	return nil

}

func (m *mongoDBDriver) createIndexForTransactionCollection(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", Transaction, i)
		m.transactionCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.transactionCollection[i].Indexes()
		transactionHashIndex := mongo.IndexModel{
			Keys:    bson.D{bson.E{Key: "hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{transactionHashIndex}); err != nil {
			return err
		}
	}
	return nil

}

func (m *mongoDBDriver) createIndexForShardOutputCoinCollection(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", ShardOutputCoin, i)
		m.shardOutputCoinCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.shardOutputCoinCollection[i].Indexes()
		shardOutputCoinSNDerivatorIndex := mongo.IndexModel{
			Keys:    bson.D{
				bson.E{Key: "tokenid", Value: 1},
				bson.E{Key: "snderivator", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		}
		shardOutputCoinTokenIdPublicKeyLockTimeIndex := mongo.IndexModel{
			Keys: bson.D{bson.E{Key: "tokenid", Value: 1},
				bson.E{ Key:   "publickey", Value: 1 },
				bson.E{ Key:   "locktime", Value: -1 },
			},
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{shardOutputCoinSNDerivatorIndex, shardOutputCoinTokenIdPublicKeyLockTimeIndex}); err != nil {
			return err
		}
	}
	return nil

}

func (m *mongoDBDriver) createIndexForCrossShardOutputCoinCollection(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", CrossShardOutputCoin, i)
		m.crossShardOutputCoinCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.crossShardOutputCoinCollection[i].Indexes()
		crossShardOutputCoinSNDerivatorIndex := mongo.IndexModel{
			Keys:    bson.D{
				bson.E{Key: "tokenid", Value: 1},
				bson.E{Key: "snderivator", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		}
		crossShardOutputCoinTokenIdPublicKeyLockTimeIndex := mongo.IndexModel{
			Keys: bson.D{bson.E{Key: "tokenid", Value: 1},
				bson.E{ Key:   "publickey", Value: 1 },
				bson.E{ Key:   "locktime", Value: -1 },
			},
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{crossShardOutputCoinSNDerivatorIndex, crossShardOutputCoinTokenIdPublicKeyLockTimeIndex}); err != nil {
			return err
		}
	}
	return nil

}

func (m *mongoDBDriver) createIndexForInputCoinCollection(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", InputCoin, i)
		m.inputCoinCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.inputCoinCollection[i].Indexes()
		inputCoinSNIndex := mongo.IndexModel{
			Keys:    bson.D{
				bson.E{Key: "tokenid", Value: 1},
				bson.E{Key: "serialnumber", Value: 1},
				bson.E{Key: "shardheight", Value: -1},
			},
			Options: options.Index().SetUnique(true),
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{inputCoinSNIndex}); err != nil {
			return err
		}
	}
	return nil

}

func (m *mongoDBDriver) createIndexForShardCommitmentIndexCollection(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", ShardCommitmentIndex, i)
		m.shardCommitmentIndexCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.shardCommitmentIndexCollection[i].Indexes()
		shardCommitmentIndexTokenIdIndexIndex := mongo.IndexModel{
			Keys:    bson.D{
				bson.E{Key: "tokenid", Value: 1},
				bson.E{Key: "index", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{shardCommitmentIndexTokenIdIndexIndex}); err != nil {
			return err
		}
	}
	return nil

}

func (m *mongoDBDriver) createIndexForPublicKeyToTransactionHashCollection(ctx context.Context) error {
	m.publicKeyToTransactionHashCollection = m.client.Database(DataBaseName).Collection(PublicKeyToTransactionHash)
	indexView := m.publicKeyToTransactionHashCollection.Indexes()
	publicKeyShardHashTransactionHashIndex := mongo.IndexModel{
		Keys:    bson.D{
			bson.E{Key: "publickey", Value: 1},
			bson.E{Key: "shardhash", Value: 1},
			bson.E{Key: "transactionhash", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{publicKeyShardHashTransactionHashIndex}); err != nil {
		return err
	}

	return nil

}

func (m *mongoDBDriver) createIndexForTokenState(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", TokenState, i)
		m.tokenStateCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.tokenStateCollection[i].Indexes()
		blockHashIndex := mongo.IndexModel{
			Keys:    bson.D{bson.E{Key: "shardhash", Value: 1}},
			Options: options.Index().SetUnique(true),
		}
		blockHeightIndex := mongo.IndexModel{
			Keys: bson.D{bson.E{Key: "shardheight", Value: 1}},
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex}); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDBDriver) createIndexForRewardState(ctx context.Context) error {
	for i := 0; i < NumOfShard; i++ {
		collectionName := fmt.Sprintf("%s-%d", RewardState, i)
		m.rewardStateCollection[i] = m.client.Database(DataBaseName).Collection(collectionName)
		indexView := m.rewardStateCollection[i].Indexes()
		blockHashIndex := mongo.IndexModel{
			Keys:    bson.D{bson.E{Key: "shardhash", Value: 1}},
			Options: options.Index().SetUnique(true),
		}
		blockHeightIndex := mongo.IndexModel{
			Keys: bson.D{bson.E{Key: "shardheight", Value: 1}},
		}
		if _, err := indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex}); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDBDriver) createIndexForBridgeTokenCollection(ctx context.Context) error {
	m.bridgeTokenStateCollection = m.client.Database(DataBaseName).Collection(BridgeTokenState)
	indexView := m.bridgeTokenStateCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForCustodianCollection(ctx context.Context) error {
	m.custodianCollection = m.client.Database(DataBaseName).Collection(Custodian)
	indexView := m.custodianCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForFinalExchangeRatesCollection(ctx context.Context) error {
	m.finalExchangeRatesCollection = m.client.Database(DataBaseName).Collection(FinalExchangeRates)
	indexView := m.finalExchangeRatesCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForLockedCollateralCollection(ctx context.Context) error {
	m.lockedCollateralCollection = m.client.Database(DataBaseName).Collection(LockedCollateral)
	indexView := m.lockedCollateralCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForMatchedRedeemRequestCollection(ctx context.Context) error {
	m.matchedRedeemRequestCollection = m.client.Database(DataBaseName).Collection(MatchedRedeemRequest)
	indexView := m.matchedRedeemRequestCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForPDEPoolForPairCollection(ctx context.Context) error {
	m.pdePoolForPairCollection = m.client.Database(DataBaseName).Collection(PDEPoolForPair)
	indexView := m.pdePoolForPairCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForPDEShareCollection(ctx context.Context) error {
	m.pdeShareCollection = m.client.Database(DataBaseName).Collection(PDEShare)
	indexView := m.pdeShareCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForPDETradingFeeCollection(ctx context.Context) error {
	m.pdeTradingFeeCollection = m.client.Database(DataBaseName).Collection(PDETradingFee)
	indexView := m.pdeTradingFeeCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForWaitingPDEContributionCollection(ctx context.Context) error {
	m.waitingPDEContributionCollection = m.client.Database(DataBaseName).Collection(WaitingPDEContribution)
	indexView := m.waitingPDEContributionCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForWaitingPortingRequestCollection(ctx context.Context) error {
	m.waitingPortingRequestCollection = m.client.Database(DataBaseName).Collection(WaitingPortingRequest)
	indexView := m.waitingPortingRequestCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) createIndexForWaitingRedeemRequestCollection(ctx context.Context) error {
	m.waitingRedeemRequestCollection = m.client.Database(DataBaseName).Collection(WaitingRedeemRequest)
	indexView := m.waitingRedeemRequestCollection.Indexes()
	blockHashIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconblockhash" , Value: 1 }},
		Options: options.Index().SetUnique(true),
	}
	blockHeightIndex := mongo.IndexModel{
		Keys:    bson.D{ bson.E{ Key: "beaconheight" , Value: 1 }},
	}
	_, err:= indexView.CreateMany(ctx, []mongo.IndexModel{blockHashIndex, blockHeightIndex})
	return err
}

func (m *mongoDBDriver) StoreLatestBeaconState(ctx context.Context ,beacon *data.Beacon) error {
	//Logger.log.Infof("Store beacon with block hash %v and block height %d", beacon.BlockHash, beacon.Height)
	beaconState := getBeaconFromBeaconState(beacon)

	//Logger.log.Infof("This beacon contain %d PDE Share ", len(beacon.PDEShare))

	//PDE
	pdeShares := getPDEShareFromBeaconState(beacon)
	pdePoolPairs := getPDEPoolForPairStateFromBeaconState(beacon)
	pdeTradingFees := getPDETradingFeeFromBeaconState(beacon)
	waitingPDEContributionStates := getWaitingPDEContributionStateFromBeaconState(beacon)

	//Portal v2
	custodians := getCustodianFromBeaconState(beacon)
	waitingPortingRequests := getWaitingPortingRequestFromBeaconState(beacon)
	matchedRedeemRequests := getMatchedRedeemRequestFromBeaconState(beacon)
	waitingRedeemRequests := getWaitingRedeemRequestFromBeaconState(beacon)
	finalExchangeRates := getFinalExchangeRatesFromBeaconState(beacon)
	lockedCollaterals := getLockedCollateralFromBeaconState(beacon)

	//Bridge
	bridgeTokenState := getBrideTokenFromBeaconState(beacon)

	return m.storeAllBeaconStateDataWithTransaction(ctx, beaconState, pdeShares, pdePoolPairs, pdeTradingFees, waitingPDEContributionStates, custodians, waitingPortingRequests, matchedRedeemRequests, waitingRedeemRequests, finalExchangeRates, lockedCollaterals, bridgeTokenState)
}

func (m *mongoDBDriver) storeAllBeaconStateDataWithTransaction(ctx context.Context, beaconState model.BeaconState, pdeShares model.PDEShare, pdePoolPairs model.PDEPoolForPair, pdeTradingFees model.PDETradingFee, waitingPDEContributionStates model.WaitingPDEContribution, custodians model.Custodian, waitingPortingRequests model.WaitingPortingRequest, matchedRedeemRequests model.RedeemRequest, waitingRedeemRequests model.RedeemRequest, finalExchangeRates model.FinalExchangeRate, lockedCollaterals model.LockedCollateral, bridgeTokenState model.BridgeTokenState) error  {
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	session, err := m.client.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) error {
		if err := session.StartTransaction(txnOpts); err != nil {
			return err
		}

		if err := m.storeAllBeaconStateData(sessionContext, beaconState, pdeShares, pdePoolPairs, pdeTradingFees,
			waitingPDEContributionStates, custodians, waitingPortingRequests, matchedRedeemRequests, waitingRedeemRequests,
			finalExchangeRates, lockedCollaterals, bridgeTokenState); err != nil {
			return err
		}

		if err := session.CommitTransaction(sessionContext); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if abortErr := session.AbortTransaction(ctx); abortErr != nil {
			return abortErr
		}
	}
	return err
}

func (m *mongoDBDriver) storeAllBeaconStateData(ctx context.Context, beaconState model.BeaconState, pdeShares model.PDEShare, pdePoolPairs model.PDEPoolForPair, pdeTradingFees model.PDETradingFee, waitingPDEContributionStates model.WaitingPDEContribution, custodians model.Custodian, waitingPortingRequests model.WaitingPortingRequest, matchedRedeemRequests model.RedeemRequest, waitingRedeemRequests model.RedeemRequest, finalExchangeRates model.FinalExchangeRate, lockedCollaterals model.LockedCollateral, bridgeTokenState model.BridgeTokenState) error {
	//Beacon
	_, err := m.beaconCollection.InsertOne(ctx, beaconState)
	if err != nil {
		return err
	}

	//Update Next Hash
	filter := bson.M{"blockhash": beaconState.PreviousBlockHash}

	update := bson.M{
		"$set": bson.M{"nextblockhash": beaconState.BlockHash},
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	result := m.beaconCollection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		return result.Err()
	}


	//PDE
	_, err = m.pdeShareCollection.InsertOne(ctx, pdeShares)
	if err != nil {
		return err
	}
	_, err = m.pdePoolForPairCollection.InsertOne(ctx, pdePoolPairs)
	if err != nil {
		return err
	}
	_, err = m.pdeTradingFeeCollection.InsertOne(ctx, pdeTradingFees)
	if err != nil {
		return err
	}
	_, err = m.waitingPDEContributionCollection.InsertOne(ctx, waitingPDEContributionStates)
	if err != nil {
		return err
	}

	//Portal v2
	_, err = m.custodianCollection.InsertOne(ctx, custodians)
	if err != nil {
		return err
	}
	_, err = m.waitingPortingRequestCollection.InsertOne(ctx, waitingPortingRequests)
	if err != nil {
		return err
	}
	_, err = m.matchedRedeemRequestCollection.InsertOne(ctx, matchedRedeemRequests)
	if err != nil {
		return err
	}
	_, err = m.waitingRedeemRequestCollection.InsertOne(ctx, waitingRedeemRequests)
	if err != nil {
		return err
	}
	_, err = m.finalExchangeRatesCollection.InsertOne(ctx, finalExchangeRates)
	if err != nil {
		return err
	}
	_, err = m.lockedCollateralCollection.InsertOne(ctx, lockedCollaterals)
	if err != nil {
		return err
	}

	//Bridge Token
	_, err = m.bridgeTokenStateCollection.InsertOne(ctx, bridgeTokenState)
	if err != nil {
		return err
	}
	return nil
}

func (m *mongoDBDriver) StoreLatestShardState(ctx context.Context ,shard *data.Shard) error {

	shardId := shard.ShardID
	//Logger.log.Infof("Store shard with block hash %v and block height %d of Shard ID %d", shard.BlockHash, shard.Height, shard.ShardID)
	shardState := getShardFromShardState(shard)

	transactions := getTransactionFromShardState(shard)
	inputCoins := getInputCoinFromShardState(shard)
	outputCoins := getOutputCoinForThisShardFromShardState(shard)
	crossOutputCoins := getCrossShardOutputCoinFromShardState(shard)
	publicKeyToHashes := getPublicKeyToTransactionHash(shard)
	commitments := getCommitmentFromShardState(shard)

	tokenState := GetTokenStateFromShardState(shard)
	rewardState := GetRewardStateFromShardState(shard)

	return m.storeAllShardStateDataWithTransaction(ctx, shardId, shardState, transactions, inputCoins, outputCoins, crossOutputCoins, commitments, publicKeyToHashes, tokenState, rewardState)

}

func (m *mongoDBDriver) storeAllShardStateDataWithTransaction(ctx context.Context, shardId byte, shardState model.ShardState, transactions []model.Transaction, inputCoins []model.InputCoin, outputCoins []model.OutputCoin, crossOutputCoins []model.OutputCoin, commitments []model.Commitment, publicKeyToHashes []model.PublicKeyToTransactionHash, tokenState model.TokenState, rewardState model.CommitteeRewardState) error {
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	session, err := m.client.StartSession()
	if err != nil {
		return err
	}

	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) error {
		log.Printf("Shard Enter session")
		if err = session.StartTransaction(txnOpts); err != nil {
			return err
		}

		if err := m.storeAllShardStateData(sessionContext, shardId, shardState, transactions, inputCoins, outputCoins,
			crossOutputCoins, commitments, publicKeyToHashes, tokenState, rewardState); err != nil {
			return err
		}

		if err = session.CommitTransaction(ctx); err != nil {
			log.Printf("Commit Transaction un successfully")
			return err
		}
		return nil
	})

	if err != nil {
		if abortErr := session.AbortTransaction(ctx); abortErr != nil {
			return abortErr
		}
	}
	return err
}

func (m *mongoDBDriver) storeAllShardStateData(ctx context.Context, shardId byte, shardState model.ShardState, transactions []model.Transaction, inputCoins []model.InputCoin, outputCoins []model.OutputCoin, crossOutputCoins []model.OutputCoin, commitments []model.Commitment, publicKeyToHashes []model.PublicKeyToTransactionHash, tokenState model.TokenState, rewardState model.CommitteeRewardState) error {
	_, err := m.shardCollection[shardId].InsertOne(ctx, shardState)
	if err != nil {
		return err
	}

	//Update Next Hash
	filter := bson.M{"blockhash": shardState.PreviousBlockHash}

	update := bson.M{
		"$set": bson.M{"nextblockhash": shardState.BlockHash},
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	result := m.shardCollection[shardId].FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		return result.Err()
	}

	for _, value := range transactions {
		_, err = m.transactionCollection[shardId].InsertOne(ctx, value)
		if err != nil {
			return err
		}
	}

	for _, value := range inputCoins {
		_, err = m.inputCoinCollection[shardId].InsertOne(ctx, value)
		if err != nil {
			return err
		}
	}

	for _, value := range outputCoins {
		_, err = m.shardOutputCoinCollection[shardId].InsertOne(ctx, value)
		if err != nil {
			return err
		}
	}

	for _, value := range crossOutputCoins {
		_, err = m.crossShardOutputCoinCollection[shardId].InsertOne(ctx, value)
		if err != nil {
			return err
		}
	}

	for _, value := range commitments {
		_, err = m.shardCommitmentIndexCollection[shardId].InsertOne(ctx, value)
		if err != nil {
			return err
		}
	}

	for _, value := range publicKeyToHashes {
		_, err = m.publicKeyToTransactionHashCollection.InsertOne(ctx, value)
		if err != nil {
			return err
		}
	}

	_, err = m.tokenStateCollection[shardId].InsertOne(ctx, tokenState)
	if err != nil {
		return err
	}

	_, err = m.rewardStateCollection[shardId].InsertOne(ctx, rewardState)
	if err != nil {
		return err
	}

	log.Printf("Commit Transaction sucessfully")
	return nil
}


func getBrideTokenFromBeaconState(beacon *data.Beacon) model.BridgeTokenState {
	brideTokenInfos := make([]model.BridgeTokenInfo, 0, len(beacon.BridgeToken))
	for _, token := range beacon.BridgeToken {
		brideTokenInfos = append(brideTokenInfos, model.BridgeTokenInfo{
			TokenID:         token.TokenID.String(),
			Amount:          strconv.FormatUint(token.Amount, 10) ,
			ExternalTokenID: token.ExternalTokenID,
			Network:         token.Network,
			IsCentralized:   token.IsCentralized,
		})
	}
	return model.BridgeTokenState{
		BeaconBlockHash:    beacon.BlockHash,
		BeaconEpoch:        beacon.Epoch,
		BeaconHeight:       beacon.Height,
		BeaconTime:         beacon.Time,
		BridgeTokenInfo: brideTokenInfos,
	}
}

func getBeaconFromBeaconState(beacon *data.Beacon) model.BeaconState {
	return model.BeaconState{
		ShardID:                                beacon.ShardID,
		BlockHash:                              beacon.BlockHash,
		PreviousBlockHash:                      beacon.PreviousBlockHash,
		NextBlockHash:                          "",
		BestShardHash:                          beacon.BestShardHash,
		BestShardHeight:                        beacon.BestShardHeight,
		Epoch:                                  beacon.Epoch,
		Height:                                 beacon.Height,
		ProposerIndex:                          beacon.ProposerIndex,
		BeaconCommittee:                        beacon.BeaconCommittee,
		BeaconPendingValidator:                 beacon.BeaconPendingValidator,
		CandidateBeaconWaitingForCurrentRandom: beacon.CandidateBeaconWaitingForNextRandom,
		CandidateShardWaitingForCurrentRandom:  beacon.CandidateShardWaitingForCurrentRandom,
		CandidateBeaconWaitingForNextRandom:    beacon.CandidateBeaconWaitingForNextRandom,
		CandidateShardWaitingForNextRandom:     beacon.CandidateShardWaitingForNextRandom,
		ShardCommittee:                         beacon.ShardCommittee,
		ShardPendingValidator:                  beacon.ShardPendingValidator,
		AutoStaking:                            beacon.AutoStaking,
		CurrentRandomNumber:                    beacon.CurrentRandomNumber,
		CurrentRandomTimeStamp:                 beacon.CurrentRandomTimeStamp,
		MaxBeaconCommitteeSize:                 beacon.MaxBeaconCommitteeSize,
		MinBeaconCommitteeSize:                 beacon.MinBeaconCommitteeSize,
		MaxShardCommitteeSize:                  beacon.MaxShardCommitteeSize,
		MinShardCommitteeSize:                  beacon.MinShardCommitteeSize,
		ActiveShards:                           beacon.ActiveShards,
		LastCrossShardState:                    beacon.LastCrossShardState,
		Time:                                   beacon.Time,
		ConsensusAlgorithm:                     beacon.ConsensusAlgorithm,
		ShardConsensusAlgorithm:                beacon.ShardConsensusAlgorithm,
		Instruction:                            beacon.Instruction,
		BlockProducer:                          beacon.BlockProducer,
		BlockProducerPublicKey:                 beacon.BlockProducerPublicKey,
		BlockProposer:                          beacon.BlockProposer,
		ValidationData:                         beacon.ValidationData,
		Version:                                beacon.Version,
		Round:                                  beacon.Round,
		Size:                                   beacon.Size,
		ShardState:                             beacon.ShardState,
		RewardReceiver:                         beacon.RewardReceiver,
		IsGetRandomNumber:                      beacon.IsGetRandomNumber,
	}
}

func getPDEShareFromBeaconState(beacon *data.Beacon) model.PDEShare {
	pdeShareInfos := make([]model.PDEShareInfo, 0, len(beacon.PDEShare))
	for _, share := range beacon.PDEShare {
		pdeShareInfos = append(pdeShareInfos, model.PDEShareInfo{
			Token1ID:           share.Token1ID,
			Token2ID:           share.Token2ID,
			ContributorAddress: share.ContributorAddress,
			Amount:             strconv.FormatUint(share.Amount, 10),
		})
	}
	return model.PDEShare{
		BeaconBlockHash:    beacon.BlockHash,
		BeaconEpoch:        beacon.Epoch,
		BeaconHeight:       beacon.Height,
		BeaconTime:         beacon.Time,
		PDEShareInfo:       pdeShareInfos,
	}
}
func getWaitingPDEContributionStateFromBeaconState(beacon *data.Beacon) model.WaitingPDEContribution {
	waitingPDEContributionInfos := make([]model.WaitingPDEContributionInfo, 0, len(beacon.WaitingPDEContributionState))
	for _, waiting := range beacon.WaitingPDEContributionState {
		waitingPDEContributionInfos = append(waitingPDEContributionInfos, model.WaitingPDEContributionInfo{
			PairID:             waiting.PairID,
			ContributorAddress: waiting.ContributorAddress,
			TokenID:            waiting.TokenID,
			Amount:             strconv.FormatUint(waiting.Amount, 10),
			TXReqID:            waiting.TXReqID,
		})
	}
	return model.WaitingPDEContribution{
		BeaconBlockHash:    beacon.BlockHash,
		BeaconEpoch:        beacon.Epoch,
		BeaconHeight:       beacon.Height,
		BeaconTime:         beacon.Time,
		WaitingPDEContributionInfo: waitingPDEContributionInfos,
	}
}

func getPDETradingFeeFromBeaconState(beacon *data.Beacon) model.PDETradingFee {
	pdeTradingFeeInfos := make([]model.PDETradingFeeInfo, 0, len(beacon.PDETradingFee))
	for _, pdeTradingFee := range beacon.PDETradingFee {
		pdeTradingFeeInfos = append(pdeTradingFeeInfos, model.PDETradingFeeInfo{
			Token1ID:           pdeTradingFee.Token1ID,
			Token2ID:           pdeTradingFee.Token2ID,
			ContributorAddress: pdeTradingFee.ContributorAddress,
			Amount:             strconv.FormatUint(pdeTradingFee.Amount, 10),
		})
	}
	return model.PDETradingFee{
		BeaconBlockHash:    beacon.BlockHash,
		BeaconEpoch:        beacon.Epoch,
		BeaconHeight:       beacon.Height,
		BeaconTime:         beacon.Time,
		PDETradingFeeInfo:  pdeTradingFeeInfos,
	}
}

func getPDEPoolForPairStateFromBeaconState(beacon *data.Beacon) model.PDEPoolForPair {
	pdeFoolForPairInfos := make([]model.PDEPoolForPairInfo, 0, len(beacon.PDEPoolPair))
	for _, pdeFoolForPair := range beacon.PDEPoolPair {
		pdeFoolForPairInfos = append(pdeFoolForPairInfos, model.PDEPoolForPairInfo{
			Token1ID:        pdeFoolForPair.Token1ID,
			Token1PoolValue: strconv.FormatUint(pdeFoolForPair.Token1PoolValue, 10),
			Token2ID:        pdeFoolForPair.Token2ID,
			Token2PoolValue: strconv.FormatUint(pdeFoolForPair.Token2PoolValue, 10),
		})
	}
	return model.PDEPoolForPair{
		BeaconBlockHash: beacon.BlockHash,
		BeaconEpoch:     beacon.Epoch,
		BeaconHeight:    beacon.Height,
		BeaconTime:      beacon.Time,
		PDEPoolForPairInfo: pdeFoolForPairInfos,
	}
}

func getCustodianFromBeaconState(beacon *data.Beacon) model.Custodian {
	custodianInfos := make([]model.CustodianInfo, 0, len(beacon.Custodian))
	for _, custodian := range beacon.Custodian {
		info := model.CustodianInfo{
			IncognitoAddress:       custodian.IncognitoAddress,
			TotalCollateral:        custodian.TotalCollateral,
			FreeCollateral:         custodian.FreeCollateral,
			HoldingPubTokens:       make(map[string]string),
			LockedAmountCollateral: make(map[string]string),
			RemoteAddresses:        custodian.RemoteAddresses,
			RewardAmount:           make(map[string]string),
		}
		for key, val := range custodian.HoldingPubTokens {
			info.HoldingPubTokens[key] = strconv.FormatUint(val, 10)
		}

		for key, val := range custodian.LockedAmountCollateral {
			info.LockedAmountCollateral[key] = strconv.FormatUint(val, 10)
		}

		for key, val := range custodian.RewardAmount {
			info.RewardAmount[key] = strconv.FormatUint(val, 10)
		}

		custodianInfos = append(custodianInfos,info)
	}
	return model.Custodian{
		BeaconBlockHash:        beacon.BlockHash,
		BeaconEpoch:            beacon.Epoch,
		BeaconHeight:           beacon.Height,
		BeaconTime:             beacon.Time,
		CustodianInfo:          custodianInfos,
	}
}

func getWaitingPortingRequestFromBeaconState(beacon *data.Beacon) model.WaitingPortingRequest {
	waitingPortingRequestInfos := make([]model.WaitingPortingRequestInfo, 0, len(beacon.WaitingPortingRequest))
	for _, w := range beacon.WaitingPortingRequest {
		waitingPortingRequestInfos = append(waitingPortingRequestInfos, model.WaitingPortingRequestInfo{
			UniquePortingID:     w.UniquePortingID,
			TokenID:             w.TokenID,
			PorterAddress:       w.PorterAddress,
			Amount:              strconv.FormatUint(w.Amount, 10),
			Custodians:          getMatchingPortingCustodianDetailFromWaitingPortingRequest(w),
			PortingFee:          w.PortingFee,
			WaitingBeaconHeight: w.BeaconHeight,
			TXReqID:             w.TXReqID,
		})
	}
	return model.WaitingPortingRequest{
		BeaconBlockHash:     beacon.BlockHash,
		BeaconEpoch:         beacon.Epoch,
		BeaconHeight:        beacon.Height,
		BeaconTime:          beacon.Time,
		WaitingPortingRequestInfo: waitingPortingRequestInfos,
	}
}
func getMatchingPortingCustodianDetailFromWaitingPortingRequest(request data.WaitingPortingRequest) []model.MatchingPortingCustodianDetail{
	result := make([]model.MatchingPortingCustodianDetail, 0)
	for _, custodian := range request.Custodians {
		result = append(result, model.MatchingPortingCustodianDetail{
			IncAddress:             custodian.IncAddress,
			RemoteAddress:          custodian.RemoteAddress,
			Amount:                 strconv.FormatUint(custodian.Amount, 10),
			LockedAmountCollateral: strconv.FormatUint(custodian.LockedAmountCollateral, 10),
		})
	}
	return result

}

func getFinalExchangeRatesFromBeaconState(beacon *data.Beacon) model.FinalExchangeRate {
	finalExchangeRateInfos := make([]model.FinalExchangeRateInfo, 0, len(beacon.FinalExchangeRates.Rates))
	for key, amount := range beacon.FinalExchangeRates.Rates {
		finalExchangeRateInfos = append(finalExchangeRateInfos, model.FinalExchangeRateInfo{
			Amount:          strconv.FormatUint(amount.Amount, 10),
			TokenID:         key,
		})
	}
	return model.FinalExchangeRate{
		BeaconBlockHash: beacon.BlockHash,
		BeaconEpoch:     beacon.Epoch,
		BeaconHeight:    beacon.Height,
		BeaconTime:      beacon.Time,
		FinalExchangeRateInfo: finalExchangeRateInfos,
	}
}

func getMatchedRedeemRequestFromBeaconState(beacon *data.Beacon) model.RedeemRequest {
	redeemRequestInfos := make([]model.RedeemRequestInfo, 0, len(beacon.MatchedRedeemRequest))
	for _, matchedRedeem := range beacon.MatchedRedeemRequest {
		redeemRequestInfos = append(redeemRequestInfos, model.RedeemRequestInfo{
			UniqueRedeemID:        matchedRedeem.UniqueRedeemID,
			TokenID:               matchedRedeem.TokenID,
			RedeemerAddress:       matchedRedeem.RedeemerAddress,
			RedeemerRemoteAddress: matchedRedeem.RedeemerRemoteAddress,
			RedeemAmount:          strconv.FormatUint(matchedRedeem.RedeemAmount, 10),
			Custodians:            getMatchingRedeemCustodianDetail(matchedRedeem),
			RedeemFee:             matchedRedeem.RedeemFee,
			RedeemBeaconHeight:    matchedRedeem.BeaconHeight,
			TXReqID:               matchedRedeem.TXReqID,
		})
	}
	return model.RedeemRequest{
		BeaconBlockHash:       beacon.BlockHash,
		BeaconEpoch:           beacon.Epoch,
		BeaconHeight:          beacon.Height,
		BeaconTime:            beacon.Time,
		RedeemRequestInfo: redeemRequestInfos,
	}
}

func getMatchingRedeemCustodianDetail(request data.RedeemRequest) []model.MatchingRedeemCustodianDetail {
	result:=make( []model.MatchingRedeemCustodianDetail, 0)
	for _, custodian := range request.Custodians {
		result = append(result, model.MatchingRedeemCustodianDetail{
			IncAddress:    custodian.IncAddress,
			RemoteAddress: custodian.RemoteAddress,
			Amount:        strconv.FormatUint(custodian.Amount, 10),
		})
	}

	return result
}

func getWaitingRedeemRequestFromBeaconState(beacon *data.Beacon) model.RedeemRequest {
	redeemRequestInfos := make([]model.RedeemRequestInfo, 0, len(beacon.WaitingRedeemRequest))
	for _, waitingRedeem := range beacon.WaitingRedeemRequest {
		redeemRequestInfos = append(redeemRequestInfos, model.RedeemRequestInfo{
			UniqueRedeemID:        waitingRedeem.UniqueRedeemID,
			TokenID:               waitingRedeem.TokenID,
			RedeemerAddress:       waitingRedeem.RedeemerAddress,
			RedeemerRemoteAddress: waitingRedeem.RedeemerRemoteAddress,
			RedeemAmount:          strconv.FormatUint(waitingRedeem.RedeemAmount, 10),
			Custodians:            getMatchingRedeemCustodianDetail(waitingRedeem),
			RedeemFee:             waitingRedeem.RedeemFee,
			RedeemBeaconHeight:    waitingRedeem.BeaconHeight,
			TXReqID:               waitingRedeem.TXReqID,
		})
	}
	return model.RedeemRequest{
		BeaconBlockHash:       beacon.BlockHash,
		BeaconEpoch:           beacon.Epoch,
		BeaconHeight:          beacon.Height,
		BeaconTime:            beacon.Time,
		RedeemRequestInfo: redeemRequestInfos,
	}
}

func getLockedCollateralFromBeaconState(beacon *data.Beacon) model.LockedCollateral {
	lockedCollateralInfos := make([]model.LockedCollateralInfo, 0, len(beacon.LockedCollateralState.LockedCollateralDetail))
	for key, lockedDetail := range beacon.LockedCollateralState.LockedCollateralDetail {
		lockedCollateralInfos = append(lockedCollateralInfos, model.LockedCollateralInfo{
			TotalLockedCollateralForRewards: beacon.LockedCollateralState.TotalLockedCollateralForRewards,
			CustodianAddress:                key,
			Amount:                          strconv.FormatUint(lockedDetail, 10),
		})
	}
	return model.LockedCollateral{
		BeaconBlockHash:                 beacon.BlockHash,
		BeaconEpoch:                     beacon.Epoch,
		BeaconHeight:                    beacon.Height,
		BeaconTime:                      beacon.Time,
		LockedCollateralInfo: lockedCollateralInfos,
	}
}

//Store Beacon
func getShardFromShardState(shard *data.Shard) model.ShardState {
	return model.ShardState{
		ShardID:                shard.ShardID,
		BlockHash:              shard.BlockHash,
		PreviousBlockHash:      shard.PreviousBlockHash,
		NextBlockHash:          "",
		Height:                 shard.Height,
		Version:                shard.Version,
		TxRoot:                 shard.TxRoot,
		ShardTxRoot:            shard.ShardTxRoot,
		CrossTransactionRoot:   shard.CrossTransactionRoot,
		InstructionsRoot:       shard.InstructionsRoot,
		CommitteeRoot:          shard.CommitteeRoot,
		PendingValidatorRoot:   shard.PendingValidatorRoot,
		StakingTxRoot:          shard.StakingTxRoot,
		InstructionMerkleRoot:  shard.InstructionMerkleRoot,
		TotalTxsFee:            shard.TotalTxsFee,
		Time:                   shard.Time,
		TxHashes:               shard.TxHashes,
		Txs:                    shard.Txs,
		BlockProducer:          shard.BlockProducer,
		BlockProducerPubKeyStr: shard.BlockProducerPubKeyStr,
		Proposer:               shard.Proposer,
		ProposeTime:            shard.ProposeTime,
		ValidationData:         shard.ValidationData,
		ConsensusType:          shard.ConsensusType,
		Data:                   shard.Data,
		BeaconHeight:           shard.BeaconHeight,
		BeaconBlockHash:        shard.BeaconBlockHash,
		Round:                  shard.Round,
		Epoch:                  shard.Epoch,
		Reward:                 shard.Reward,
		RewardBeacon:           shard.RewardBeacon,
		Fee:                    shard.Fee,
		Size:                   shard.Size,
		Instruction:            shard.Instruction,
		CrossShardBitMap:       shard.CrossShardBitMap,
		NumTxns:                shard.NumTxns,
		TotalTxns:              shard.TotalTxns,
		NumTxnsExcludeSalary:   shard.NumTxnsExcludeSalary,
		TotalTxnsExcludeSalary: shard.TotalTxnsExcludeSalary,
		ActiveShards:           shard.ActiveShards,
		ConsensusAlgorithm:     shard.ConsensusType,
		NumOfBlocksByProducers: shard.NumOfBlocksByProducers,
		MaxShardCommitteeSize:  shard.MaxShardCommitteeSize,
		MinShardCommitteeSize:  shard.MinShardCommitteeSize,
		ShardProposerIdx:       shard.ShardProposerIdx,
		MetricBlockHeight:      shard.MetricBlockHeight,
		BestCrossShard:         shard.BestCrossShard,
		ShardCommittee:         shard.ShardCommittee,
		ShardPendingValidator:  shard.ShardPendingValidator,
		StakingTx: shard.StakingTx,
	}
}

func getTransactionFromShardState(shard *data.Shard) []model.Transaction {
	transactions := make([]model.Transaction, 0, len(shard.Transactions))
	for _, transaction := range shard.Transactions {
		newTransaction := model.Transaction{
			ShardId:              shard.ShardID,
			ShardHash:            shard.BlockHash,
			ShardHeight:          shard.BeaconHeight,
			Image:                 "",
			IsPrivacy:             transaction.IsPrivacy,
			TxSize:				  transaction.TxSize,
			Index:                transaction.Index,
			Hash:                 transaction.Hash,
			Version:              transaction.Version,
			Type:                 transaction.Type,
			LockTime:             time.Unix(transaction.LockTime, 0).Format(common.DateOutputFormat),
			Fee:                  transaction.Fee,
			Info:                 string(transaction.Info),
			SigPubKey:            base58.Base58Check{}.Encode(transaction.SigPubKey, 0x0),
			Sig:                  base58.Base58Check{}.Encode(transaction.Sig, 0x0),
			PubKeyLastByteSender: transaction.PubKeyLastByteSender,
			InputCoinPubKey: transaction.InputCoinPubKey,
			IsInBlock: true,
			IsInMempool: false,
		}
		newTransaction.ProofDetail, newTransaction.Proof = 	getProofDetail(transaction)
		newTransaction.CustomTokenData =  ""
		if transaction.Metadata != nil {
			metaData, _ := json.MarshalIndent(transaction.Metadata, "", "\t")
			newTransaction.Metadata = string(metaData)
		}
		if transaction.TxPrivacy != nil {
			newTransaction.PrivacyCustomTokenID = transaction.TxPrivacy.PropertyID
			newTransaction.PrivacyCustomTokenName = transaction.TxPrivacy.PropertyName
			newTransaction.PrivacyCustomTokenSymbol = transaction.TxPrivacy.PropertySymbol
			newTransaction.PrivacyCustomTokenData = transaction.PrivacyCustomTokenData
			newTransaction.PrivacyCustomTokenIsPrivacy = transaction.TxPrivacy.Tx.IsPrivacy
			newTransaction.PrivacyCustomTokenFee = transaction.TxPrivacy.Tx.Fee
			newTransaction.PrivacyCustomTokenProofDetail, newTransaction.PrivacyCustomTokenProof = getProofDetail(transaction.TxPrivacy.Tx)
		}
		transactions = append(transactions, newTransaction)
	}
	return transactions
}

func getProofDetail (normalTx *data.Transaction) (jsonresult.ProofDetail, *string) {
	if normalTx.Proof == nil {
		return jsonresult.ProofDetail{}, nil
	}
	b, _:= normalTx.Proof.MarshalJSON()
	proof := string(b)
	return jsonresult.ProofDetail{
		InputCoins:  getProofDetailInputCoin(normalTx.Proof),
		OutputCoins: getProofDetailOutputCoin(normalTx.Proof),
	}, &proof
}

func getProofDetailInputCoin(proof *zkp.PaymentProof) []*jsonresult.CoinDetail {
	inputCoins := make([]*jsonresult.CoinDetail, 0)
	for _, input := range proof.GetInputCoins() {
		in := jsonresult.CoinDetail{
			CoinDetails: jsonresult.Coin{},
		}
		if input.CoinDetails != nil {
			in.CoinDetails.Value = input.CoinDetails.GetValue()
			in.CoinDetails.Info = base58.Base58Check{}.Encode(input.CoinDetails.GetInfo(), 0x0)
			if input.CoinDetails.GetCoinCommitment() != nil {
				in.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(input.CoinDetails.GetCoinCommitment().ToBytesS(), 0x0)
			}
			if input.CoinDetails.GetRandomness() != nil {
				in.CoinDetails.Randomness = *input.CoinDetails.GetRandomness()
			}
			if input.CoinDetails.GetSNDerivator() != nil {
				in.CoinDetails.SNDerivator = *input.CoinDetails.GetSNDerivator()
			}
			if input.CoinDetails.GetSerialNumber() != nil {
				in.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(input.CoinDetails.GetSerialNumber().ToBytesS(), 0x0)
			}
			if input.CoinDetails.GetPublicKey() != nil {
				in.CoinDetails.PublicKey = base58.Base58Check{}.Encode(input.CoinDetails.GetPublicKey().ToBytesS(), 0x0)
			}
		}
		inputCoins = append(inputCoins, &in)
	}
	return inputCoins
}

func getProofDetailOutputCoin(proof *zkp.PaymentProof) []*jsonresult.CoinDetail {
	outputCoins := make([]*jsonresult.CoinDetail, 0)
	for _, output := range proof.GetOutputCoins() {
		out := jsonresult.CoinDetail{
			CoinDetails: jsonresult.Coin{},
		}
		if output.CoinDetails != nil {
			out.CoinDetails.Value = output.CoinDetails.GetValue()
			out.CoinDetails.Info = base58.Base58Check{}.Encode(output.CoinDetails.GetInfo(), 0x0)
			if output.CoinDetails.GetCoinCommitment() != nil {
				out.CoinDetails.CoinCommitment = base58.Base58Check{}.Encode(output.CoinDetails.GetCoinCommitment().ToBytesS(), 0x0)
			}
			if output.CoinDetails.GetRandomness() != nil {
				out.CoinDetails.Randomness = *output.CoinDetails.GetRandomness()
			}
			if output.CoinDetails.GetSNDerivator() != nil {
				out.CoinDetails.SNDerivator = *output.CoinDetails.GetSNDerivator()
			}
			if output.CoinDetails.GetSerialNumber() != nil {
				out.CoinDetails.SerialNumber = base58.Base58Check{}.Encode(output.CoinDetails.GetSerialNumber().ToBytesS(), 0x0)
			}
			if output.CoinDetails.GetPublicKey() != nil {
				out.CoinDetails.PublicKey = base58.Base58Check{}.Encode(output.CoinDetails.GetPublicKey().ToBytesS(), 0x0)
			}
			if output.CoinDetailsEncrypted != nil {
				out.CoinDetailsEncrypted = base58.Base58Check{}.Encode(output.CoinDetailsEncrypted.Bytes(), 0x0)
			}
		}
		outputCoins = append(outputCoins , &out)
	}
	return outputCoins
}


func getInputCoinFromShardState(shard *data.Shard) []model.InputCoin {
	inputCoins := make([]model.InputCoin, 0, len(shard.Transactions))
	for _, transaction := range shard.Transactions {
		for _, input := range transaction.InputCoins {
			inputCoin := model.InputCoin{
				ShardId:         shard.ShardID,
				ShardHash:       shard.BlockHash,
				ShardHeight:     shard.BeaconHeight,
				TransactionHash: transaction.Hash,
				Value:           strconv.FormatUint(input.Value, 10),
				Info:            string(input.Info),
				TokenID:         input.TokenID,
			}
			if input.PublicKey != nil {
				inputCoin.PublicKey =   base58.Base58Check{}.Encode(input.PublicKey.ToBytesS(), common.ZeroByte)
			}
			if input.CoinCommitment != nil {
				inputCoin.CoinCommitment = base58.Base58Check{}.Encode(input.CoinCommitment.ToBytesS(), common.ZeroByte)
			}
			if input.SNDerivator != nil {
				inputCoin.SNDerivator = base58.Base58Check{}.Encode(input.SNDerivator.ToBytesS(), common.ZeroByte)
			}
			if input.SerialNumber != nil {
				inputCoin.SerialNumber = base58.Base58Check{}.Encode(input.SerialNumber.ToBytesS(), common.ZeroByte)
			}
			if input.Randomness != nil {
				inputCoin.Randomness = base58.Base58Check{}.Encode(input.Randomness.ToBytesS(), common.ZeroByte)
			}
			inputCoins = append(inputCoins, inputCoin)
		}

	}
	return inputCoins
}
func getCrossShardOutputCoinFromShardState(shard *data.Shard) []model.OutputCoin {
	outputCoins := make([]model.OutputCoin, 0, len(shard.OutputCoins))
	for _, output := range shard.OutputCoins {
		if output.ToShardID == shard.ShardID {
			continue
		}
		outputCoin := model.OutputCoin{
			ShardId:          shard.ShardID,
			ShardHash:        shard.BlockHash,
			ShardHeight:      shard.BeaconHeight,
			TransactionHash:  output.TransactionHash,
			Value:            strconv.FormatUint(output.Value, 10),
			Info:             string(output.Info),
			TokenID:          output.TokenID,
			FromShardID:      output.FromShardID,
			ToShardID:        output.ToShardID,
			FromCrossShard:   output.FromCrossShard,
			CrossBlockHash:   output.CrossBlockHash,
			CrossBlockHeight: output.CrossBlockHeight,
			PropertyName:     output.PropertyName,
			PropertySymbol:   output.PropertySymbol,
			Type:             output.Type,
			Mintable:         output.Mintable,
			Amount:           strconv.FormatUint(output.Amount, 10),
			LockTime:		  output.LockTime,
			TransactionMemo: string(output.TransactionMemo),

		}
		if output.PublicKey != nil {
			outputCoin.PublicKey = base58.Base58Check{}.Encode(output.PublicKey.ToBytesS(), common.ZeroByte)
		}
		if output.CoinCommitment != nil {
			outputCoin.CoinCommitment = base58.Base58Check{}.Encode(output.CoinCommitment.ToBytesS(), common.ZeroByte)
		}
		if output.SNDerivator != nil {
			outputCoin.SNDerivator = base58.Base58Check{}.Encode(output.SNDerivator.ToBytesS(), common.ZeroByte)
		}
		if output.SerialNumber != nil {
			outputCoin.SerialNumber = base58.Base58Check{}.Encode(output.SerialNumber.ToBytesS(), common.ZeroByte)
		}
		if output.Randomness != nil {
			outputCoin.Randomness = base58.Base58Check{}.Encode(output.Randomness.ToBytesS(), common.ZeroByte)
		}
		if output.CoinDetailsEncrypted != nil {
			outputCoin.CoinDetailsEncrypted = base58.Base58Check{}.Encode(output.CoinDetailsEncrypted.Bytes(), common.ZeroByte)
		}
		outputCoins = append(outputCoins, outputCoin)
	}
	return outputCoins
}
func getOutputCoinForThisShardFromShardState(shard *data.Shard) []model.OutputCoin {
	outputCoins := make([]model.OutputCoin, 0, len(shard.OutputCoins))
	for _, output := range shard.OutputCoins {
		if output.ToShardID != shard.ShardID {
			continue
		}
		outputCoin := model.OutputCoin{
			ShardId:          shard.ShardID,
			ShardHash:        shard.BlockHash,
			ShardHeight:      shard.BeaconHeight,
			TransactionHash:  output.TransactionHash,
			Value:            strconv.FormatUint(output.Value, 10),
			Info:             string(output.Info),
			TokenID:          output.TokenID,
			FromShardID:      output.FromShardID,
			ToShardID:        output.ToShardID,
			FromCrossShard:   output.FromCrossShard,
			CrossBlockHash:   output.CrossBlockHash,
			CrossBlockHeight: output.CrossBlockHeight,
			PropertyName:     output.PropertyName,
			PropertySymbol:   output.PropertySymbol,
			Type:             output.Type,
			Mintable:         output.Mintable,
			Amount:           strconv.FormatUint(output.Amount, 10),
			LockTime:		  output.LockTime,
			TransactionMemo: string(output.TransactionMemo),

		}
		if output.PublicKey != nil {
			outputCoin.PublicKey = base58.Base58Check{}.Encode(output.PublicKey.ToBytesS(), common.ZeroByte)
		}
		if output.CoinCommitment != nil {
			outputCoin.CoinCommitment = base58.Base58Check{}.Encode(output.CoinCommitment.ToBytesS(), common.ZeroByte)
		}
		if output.SNDerivator != nil {
			outputCoin.SNDerivator = base58.Base58Check{}.Encode(output.SNDerivator.ToBytesS(), common.ZeroByte)
		}
		if output.SerialNumber != nil {
			outputCoin.SerialNumber = base58.Base58Check{}.Encode(output.SerialNumber.ToBytesS(), common.ZeroByte)
		}
		if output.Randomness != nil {
			outputCoin.Randomness = base58.Base58Check{}.Encode(output.Randomness.ToBytesS(), common.ZeroByte)
		}
		if output.CoinDetailsEncrypted != nil {
			outputCoin.CoinDetailsEncrypted = base58.Base58Check{}.Encode(output.CoinDetailsEncrypted.Bytes(), common.ZeroByte)
		}

		outputCoins = append(outputCoins, outputCoin)
	}
	return outputCoins
}

func getCommitmentFromShardState(shard *data.Shard) []model.Commitment {
	commitments := make([]model.Commitment, 0, len(shard.Commitments))

	for _, commitment := range shard.Commitments {
		commitments = append(commitments, model.Commitment{
			ShardHash:       shard.BlockHash,
			ShardHeight:     shard.Height,
			TransactionHash: commitment.TransactionHash,
			TokenID:         commitment.TokenID,
			ShardId:         commitment.ShardID,
			Commitment:      base58.Base58Check{}.Encode(commitment.Commitment,common.ZeroByte),
			Index:           commitment.Index,
		})
	}
	return commitments
}

func GetTokenStateFromShardState(shard *data.Shard) model.TokenState {
	tokenState := model.TokenState{
		ShardID:     shard.ShardID,
		ShardHash:   shard.BlockHash,
		ShardHeight: shard.Height,
	}
	tokenInformations := make([]model.TokenInformation, 0, len(shard.TokenState))

	for _, token := range shard.TokenState {
		tokenInformations = append(tokenInformations, model.TokenInformation{
			TokenID:        token.TokenID,
			PropertyName:   token.PropertyName,
			PropertySymbol: token.PropertySymbol,
			TokenType:      token.TokenType,
			Mintable:       token.Mintable,
			Amount:         strconv.FormatUint(token.Amount, 10),
			Info:           token.Info,
			InitTx:         token.InitTx,
			Txs:            token.Txs,
		})
	}
	tokenState.Token = tokenInformations
	return tokenState
}

func GetRewardStateFromShardState(shard *data.Shard) model.CommitteeRewardState {
	rewardsState := model.CommitteeRewardState{
		ShardID:     shard.ShardID,
		ShardHash:   shard.BlockHash,
		ShardHeight: shard.Height,
	}
	rewards := make([]model.CommitteeReward, 0, 2000)

	for address, token := range shard.CommitteeRewardState {
		for token, amount := range token {
			rewards = append(rewards, model.CommitteeReward{
				Address: address,
				TokenId: token,
				Amount:  strconv.FormatUint(amount, 10),
			})
		}

	}
	rewardsState.CommitteeReward = rewards
	return rewardsState
}

func getPublicKeyToTransactionHash(shard *data.Shard) []model.PublicKeyToTransactionHash {
	result := make([]model.PublicKeyToTransactionHash, 0, len(shard.OutputCoins))
	publicKeyMap := make(map[string]bool)
	for _, output := range shard.OutputCoins {
		if len(output.TransactionHash) == 0 {
			continue
		}

		public := model.PublicKeyToTransactionHash{
			ShardId:         shard.ShardID,
			ShardHash:       shard.BlockHash,
			ShardHeight:     shard.Height,
			TransactionHash: output.TransactionHash,
		}
		if output.PublicKey != nil {
			public.PublicKey = base58.Base58Check{}.Encode(output.PublicKey.ToBytesS(), common.ZeroByte)
		}
		if _ , ok := publicKeyMap[public.PublicKey]; ok {
			continue
		} else {
			publicKeyMap[public.PublicKey] = true
		}
		result = append(result, public)
	}
	return result
}