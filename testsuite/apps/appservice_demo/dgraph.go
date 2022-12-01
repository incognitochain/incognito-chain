package main

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/dgraph-io/dgo/v200"
//	"github.com/dgraph-io/dgo/v200/protos/api"
//	"google.golang.org/grpc"
//	"log"
//)
//
//type DgraphClient struct {
//	client *dgo.Dgraph
//}
//
//func NewDgraphClient() *DgraphClient {
//	conn, err := grpc.Dial("51.222.153.212:9080", grpc.WithInsecure())
//	if err != nil {
//		panic(err)
//	}
//	dc := api.NewDgraphClient(conn)
//	dg := dgo.NewDgraphClient(dc)
//	return &DgraphClient{dg}
//}
//
//func (dg *DgraphClient) SetSchema() {
//	op := &api.Operation{}
//	op.Schema = `
//		type Transaction {
//			hash
//			md
//			mdName
//			tokenID
//			tokenName
//			amount
//			incoins
//			outcoins
//		}
//		type Coin {
//			hash
//			tokenID
//			tokenName
//			amount
//			inCoinOfTxs
//		}
//
//		mdName: string .
//		md: int .
//		hash : string @index(exact) .
//		tokenID: string @index(exact) .
//		tokenName: string @index(exact) .
//		amount: int .
//		incoins: [uid] .
//		inCoinOfTxs: [uid] .
//		outcoins: [uid] @reverse .
//	`
//
//	ctx := context.Background()
//	if err := dg.client.Alter(ctx, op); err != nil {
//		log.Fatal(err)
//	}
//}
//
//type Coin struct {
//	Uid         string        `json:"uid,omitempty"`
//	Hash        string        `json:"hash,omitempty"`
//	TokenID     string        `json:"tokenID,omitempty"`
//	TokenName   string        `json:"tokenName,omitempty"`
//	Amount      int           `json:"amount,omitempty"`
//	InCoinOfTxs []Transaction `json:"inCoinOfTx,omitempty"`
//	DType       []string      `json:"dgraph.type,omitempty"`
//}
//
//type Transaction struct {
//	Uid          string        `json:"uid,omitempty"`
//	Hash         string        `json:"hash,omitempty"`
//	LinkTx       []Transaction `json:"linkTx,omitempty"`
//	Metadata     int           `json:"md,omitempty"`
//	MetadataName string        `json:"mdName,omitempty"`
//	TokenID      string        `json:"tokenID,omitempty"`
//	TokenName    string        `json:"tokenName,omitempty"`
//	Amount       int           `json:"amount,omitempty"`
//	InCoins      [][]Coin      `json:"incoins,omitempty"`
//	OutCoins     []Coin        `json:"outCoins,omitempty"`
//	DType        []string      `json:"dgraph.type,omitempty"`
//}
//
//func (dg *DgraphClient) SetTransaction(hash string, linkTx string, md int, mdName, tokenID, tokenName string, amount int, inputCoins [][]string, outCoins []string) {
//	_outcoins := []Coin{}
//	for _, out := range outCoins {
//		p := Coin{
//			Uid:       "_:" + out,
//			Hash:      out,
//			TokenID:   tokenID,
//			TokenName: tokenName,
//			Amount:    amount,
//			DType:     []string{"Coin"},
//		}
//		_outcoins = append(_outcoins, p)
//	}
//	p := Transaction{
//		Uid:          "_:" + hash,
//		Hash:         hash,
//		Metadata:     md,
//		MetadataName: mdName,
//		TokenID:      tokenID,
//		TokenName:    tokenName,
//		Amount:       amount,
//		OutCoins:     _outcoins,
//		DType:        []string{"Transaction"},
//	}
//	pb, err := json.MarshalIndent(p, "", "\t")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(string(pb))
//	//TODO: link input coin
//	//TODO: link request-response transaction
//
//	dg.client.NewTxn().
//		dg.SetJson(pb)
//}
//
//func (dg *DgraphClient) SetOutCoin(hash string, tokenID, tokenName string, amount int) Coin {
//	p := Coin{
//		Uid:       "_:" + hash,
//		Hash:      hash,
//		TokenID:   tokenID,
//		TokenName: tokenName,
//		Amount:    amount,
//		DType:     []string{"Coin"},
//	}
//
//	pb, err := json.Marshal(p)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	dg.SetJson(pb)
//	return p
//}
//
//func (dg *DgraphClient) SetJson(data []byte) {
//	mu := &api.Mutation{
//		CommitNow: true,
//	}
//	mu.SetJson = data
//	ctx := context.Background()
//	response, err := dg.client.NewTxn().Mutate(ctx, mu)
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(response)
//}
