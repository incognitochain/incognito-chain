package aggregatelog

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/olivere/elastic"
)

const LOG_AGGREGATION_INDEX = "log_aggregation"

type Message struct {
	Time     time.Time `json:"time"`
	LogLevel string    `json:"level"`
	Message  string    `json:"message"`
}

var elasticClient *elastic.Client
var ctx = context.Background()

func ValidateElasticClient() error {
	if elasticClient == nil {
		return errors.New("Elastic client not initialized")
	}
	return nil
}

func InitElastic(params map[string]interface{}) error {
	urlValue, ok := params["elastic_url"]
	if !ok || urlValue == "" {
		return errors.New("Elastic url config empty")
	}
	url, ok := urlValue.(string)
	if !ok {
		return errors.New("Elastic urlconfig invalid")
	}
	err := CreateElasticClient(url)
	if err != nil {
		return err
	}
	indexExisted := CheckLogIndexExisted(elasticClient, LOG_AGGREGATION_INDEX)
	if !indexExisted {
		CreateLogIndex(elasticClient, LOG_AGGREGATION_INDEX)
	}
	return nil
}

func CreateElasticClient(url string) error {

	if url == "" {
		return errors.New("Elastic URL setting invalid")
	}
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))
	if err != nil {
		return err
	}
	elasticClient = client
	return nil
}

func SendElasticMessage(message string) error {
	err := ValidateElasticClient()
	if err != nil {
		return err
	}
	err = SendMessageToElastic(message, "INFO")
	if err != nil {
		return err
	}
	return nil
}

func SendElasticError(err error) error {
	validErr := ValidateElasticClient()
	if validErr != nil {
		return validErr
	}
	sendError := SendMessageToElastic(err.Error(), "ERROR")
	if sendError != nil {
		return sendError
	}
	return nil
}

// CHECK ELASTIC TABLE LOG INDEX EXIST
func CheckLogIndexExisted(client *elastic.Client, index string) bool {
	existed, err := client.IndexExists(index).Do(ctx)
	if err != nil {
		log.Println("check index error", err)
		return false
	}
	if !existed {
		return false
	}
	return true
}

func CreateLogIndex(client *elastic.Client, index string) error {
	createIndex, err := client.CreateIndex(index).Do(ctx)
	if err != nil {
		log.Println("create elastic index log error", err)
		return err
	}
	if !createIndex.Acknowledged {
		// Not acknowledged
		log.Println("not ack")
	} else {
		log.Println("ack")
	}
	return nil
}

func SendMessageToElastic(message, level string) error {
	messageObject := Message{
		Time:     time.Now(),
		Message:  message,
		LogLevel: level,
	}
	putResult, err := elasticClient.Index().
		Index(LOG_AGGREGATION_INDEX).
		Type("log").
		BodyJson(messageObject).
		Do(ctx)
	if err != nil {
		return err
	}
	log.Printf("Indexed tweet %s to index %s, type %s\n", putResult.Id, putResult.Index, putResult.Type)
	return nil
}
