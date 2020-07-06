package aggregatelog

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/olivere/elastic"
)

type MessageDataTest struct {
	Time     time.Time `json:"time"`
	LogLevel string    `json:"level"`
	Message  string    `json:"message"`
}

func ElasticTest() {
	// Starting with elastic.v5, you must pass a context to execute each service
	ctx := context.Background()

	// Obtain a client and connect to the default Elasticsearch installation
	// on 127.0.0.1:9200. Of course you can configure your client to connect
	// to other hosts and configure it in various other ways.
	elastic.SetHealthcheckTimeoutStartup(time.Second * 10)
	client, err := elastic.NewClient(
		elastic.SetURL("http://35.198.243.55:9200"), elastic.SetSniff(false))

	if err != nil {
		// Handle error
		panic(err)
	}

	// Ping the Elasticsearch server to get e.g. the version number
	info, code, err := client.Ping("http://35.198.243.55:9200").Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}

	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("log_aggregation").Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}

	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex("log_aggregation").Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
			log.Println("not ack")
		} else {
			log.Println("ack")
		}
	} else {
		log.Println("GetIndex Exists")
	}

	message1 := MessageDataTest{Time: time.Now(), Message: "Take Five", LogLevel: "WARN"}
	put1, err := client.Index().
		Index("log_aggregation").
		Type("log").
		BodyJson(message1).
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)
}
