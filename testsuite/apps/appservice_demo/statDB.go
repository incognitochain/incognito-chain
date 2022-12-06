package main

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

type statDB struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}
type DetailInfo struct {
	MetaName       string
	PreviousLinkTx string
	TokenID        string
	Amount         uint64
}

type StatInfo struct {
	ShardID      int
	BlockHeight  uint64
	BlockTime    time.Time
	BlockHash    string
	Tx           string
	Outcoin      []coin.Coin
	Incoin       [][]string
	MetadataType int
	Metadata     string
	Detail       DetailInfo
	Epoch        uint64
}

func NewStatDB(endpoint string, dbName, collectionName string) (*statDB, error) {

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
	res := &statDB{client, collection}
	//set indexing
	res.indexing("ShardID", "BlockHeight")
	res.indexing("ShardID", "BlockTime")
	res.indexing("Tx")
	res.indexing("PreviousLinkTx")
	res.indexing("BlockTime")

	res.indexing("TokenName")
	res.indexing("MetadataType")
	res.indexing("Metadata.NftID")
	res.indexing("BlockTime", "MetadataType")
	res.indexing("Epoch", "MetadataType")
	res.indexing("MetadataType.TokenID")
	res.indexing("MetadataType", "MetadataType.TokenID")

	return res, nil
}

func (s *statDB) indexing(fields ...string) {
	opts := options.CreateIndexes().SetMaxTime(10 * time.Minute)
	keys := bson.D{}
	for _, f := range fields {
		keys = append(keys, bson.E{f, -1})
	}
	index := mongo.IndexModel{}
	index.Keys = keys
	s.Collection.Indexes().CreateOne(context.Background(), index, opts)
}

func (s *statDB) lastBlock(cid int) uint64 {
	opts := options.Find().SetSort(bson.D{{"BlockHeight", -1}}).SetLimit(1)

	cur, err := s.Collection.Find(context.TODO(), map[string]interface{}{
		"ShardID": cid,
	}, opts)

	if err != nil {
		fmt.Println(err)
		panic(1)
	}

	if !cur.Next(context.TODO()) {
		return 1
	}

	var tmp bson.M
	err = cur.Decode(&tmp)
	if err != nil {
		panic(err)
	}
	return uint64(tmp["BlockHeight"].(int64))
}

func (s *statDB) set(info StatInfo) error {
	filter := map[string]interface{}{
		"Tx": info.Tx,
	}
	var metadata interface{}
	if len(info.Metadata) > 0 {
		err := bson.UnmarshalExtJSON([]byte(info.Metadata), true, &metadata)
		if err != nil {
			panic(err)
		}
	}
	outcoins := []string{}
	for _, coin := range info.Outcoin {
		publicKey := base58.Base58Check{}.Encode(coin.(*privacy.CoinV2).GetPublicKey().ToBytesS(), common.ZeroByte)
		outcoins = append(outcoins, publicKey)
	}
	doc := bson.M{
		"$set": bson.M{
			"ShardID":        info.ShardID,
			"BlockHeight":    info.BlockHeight,
			"BlockTime":      info.BlockTime,
			"BlockHash":      info.BlockHash,
			"Epoch":          info.Epoch,
			"Tx":             info.Tx,
			"PreviousLinkTx": info.Detail.PreviousLinkTx,
			"Outcoin":        outcoins,
			"Incoin":         info.Incoin,
			"MetadataType":   info.MetadataType,
			"Metadata":       metadata,
			"TokenID":        info.Detail.TokenID,
			"TokenName":      mapToken(info.Detail.TokenID),
			"Amount":         info.Detail.Amount,
			"MetadataName":   info.Detail.MetaName,
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

func mapToken(s string) string {
	n := strings.Replace(s, "0000000000000000000000000000000000000000000000000000000000000004", "PRV", -1)
	n = strings.Replace(n, "076a4423fa20922526bd50b0d7b0dc1c593ce16e15ba141ede5fb5a28aa3f229", "USDT(Unify)", -1)
	n = strings.Replace(n, "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696", "BTC", -1)
	n = strings.Replace(n, "3ee31eba6376fc16cadb52c8765f20b6ebff92c0b1c5ab5fc78c8c25703bb19e", "ETH(Unify)", -1)
	n = strings.Replace(n, "7450ad98cb8c967afb76503944ab30b4ce3560ed8f3acc3155f687641ae34135", "LTC", -1)
	n = strings.Replace(n, "26df4d1bca9fd1a8871a24b9b84fc97f3dd62ca8809975c6d971d1b79d1d9f31", "MATIC", -1)
	n = strings.Replace(n, "447b088f1c2a8e08bff622ef43a477e98af22b64ea34f99278f4b550d285fbff", "DASH", -1)
	n = strings.Replace(n, "be02b225bcd26eeae00d3a51e554ac0adcdcc09de77ad03202904666d427a7e4", "BUSD", -1)
	n = strings.Replace(n, "e5032c083f0da67ca141331b6005e4a3740c50218f151a5e829e9d03227e33e2", "BNB", -1)
	n = strings.Replace(n, "545ef6e26d4d428b16117523935b6be85ec0a63e8c2afeb0162315eb0ce3d151", "USDC(Unify)", -1)
	n = strings.Replace(n, "c01e7dc1d1aba995c19b257412340b057f8ad1482ccb6a9bb0adce61afbf05d4", "XMR", -1)
	n = strings.Replace(n, "a609150120c0247407e6d7725f2a9701dcbb7bab5337a70b9cef801f34bc2b5c", "Zcash", -1)
	n = strings.Replace(n, "6eed691cb14d11066f939630ff647f5f1c843a8f964d9a4d295fa9cd1111c474", "Fantom", -1)
	n = strings.Replace(n, "3f89c75324b46f13c7b036871060e641d996a24c09b3065835cb1d38b799d6c1", "Dai", -1)
	n = strings.Replace(n, "0d953a47a7a488cee562e64c80c25d3dbe29d3b477ccd2b54408c0553a93f126", "Dai(Unify)", -1)

	return n
}
