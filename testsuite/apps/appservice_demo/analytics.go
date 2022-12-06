package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	time2 "time"
)

func FindScore(stat *statDB, coinDB *CoinDB) {
	tx := "5bccf424d85025c2ff2e655f06404f35c670bd3e88eea79e02824f0e1c2f7b87"
	var result = bson.M{}
	res := stat.Collection.FindOne(context.Background(), bson.D{{"Tx", tx}}).Decode(&result)
	fmt.Println(res)
	fmt.Println(result["Incoin"])

	//
	for _, coins := range result["Incoin"].(primitive.A) {
		for _, coin := range coins.(primitive.A) {
			var result = bson.M{}
			res := coinDB.Collection.FindOne(context.Background(), bson.D{{"Pubkey", coin.(string)}}).Decode(&result)
			fmt.Println(res)
			var t time2.Time
			if result["CreateTime"] != nil {
				t = result["CreateTime"].(primitive.DateTime).Time().Local()
			}
			fmt.Println(coin, t, len(result["UsedTx"].(primitive.A)), result["TokenName"])
		}

	}
}

func (s *statDB) MetadataPerEpoch() {
	matchState := bson.D{
		{"$match", bson.D{
			{"MetadataType", 287},
		},
		}}

	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", bson.D{
				{"Epoch", "$Epoch"},
			}},
			{
				"time", bson.D{{"$last", "$BlockTime"}},
			},
			{
				"total", bson.D{{"$sum", 1}},
			},
		},
		},
	}
	sortStage := bson.D{
		{"$sort", bson.D{{"_id.Epoch", 1}}}}

	ctx := context.TODO()
	cursor, err := s.Collection.Aggregate(ctx, mongo.Pipeline{matchState, groupStage, sortStage})
	if err != nil {
		panic(err)
	}
	// display the results
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		panic(err)
	}

	for _, result := range results {
		if result["total"].(int32) > 10 {
			fmt.Printf(" %v : %v %v \n", result["_id"].(bson.M)["Epoch"], result["total"], result["time"].(primitive.DateTime).Time().Local())
		}

	}

}
