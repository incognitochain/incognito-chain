package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type CoinDB struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

type CoinInfo struct {
	Pubkey     string
	CreateTime time.Time
	CreatedTx  string
	UsedTx     []string
	TokenID    string
	TokenName  string
	Amount     uint64
}

func NewCoinDB(endpoint string, dbName, collectionName string) (*CoinDB, error) {

	client, err := mongo.NewClient(options.Client().ApplyURI(endpoint))
	if err != nil {
		panic("Cannot new client")
		return nil, err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		panic("Cannot connect")
		return nil, err
	}
	collection := client.Database(dbName).Collection(collectionName)
	res := &CoinDB{client, collection}
	//set indexing
	res.indexing("Pubkey")
	res.indexing("CreateTime")
	res.indexing("UsedTx")
	return res, nil
}

func (s *CoinDB) indexing(fields ...string) {
	opts := options.CreateIndexes().SetMaxTime(10 * time.Minute)
	keys := bson.D{}
	for _, f := range fields {
		keys = append(keys, bson.E{f, -1})
	}
	index := mongo.IndexModel{}
	index.Keys = keys
	_, e := s.Collection.Indexes().CreateOne(context.Background(), index, opts)
	fmt.Println(e)
}

func (s *CoinDB) set(info CoinInfo) error {
	filter := map[string]interface{}{
		"Pubkey": info.Pubkey,
	}
	doc := bson.M{
		"$set": bson.M{
			"Pubkey":     info.Pubkey,
			"CreateTime": info.CreateTime,
			"TokenID":    info.TokenID,
			"TokenName":  mapToken(info.TokenID),
			"CreatedTx":  info.CreatedTx,
			"Amount":     info.Amount,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	result := s.Collection.FindOneAndUpdate(context.TODO(), filter, doc, opts)
	err := result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}

func (s *CoinDB) UpdateCoinLink(pk string, tx string) error {
	filter := map[string]interface{}{
		"Pubkey": pk,
	}
	doc := bson.M{
		"$addToSet": bson.D{{"UsedTx", tx}},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	result := s.Collection.FindOneAndUpdate(context.TODO(), filter, doc, opts)
	err := result.Err()
	if err != nil && err.Error() == "mongo: no documents in result" {
		return nil
	}
	return err
}
