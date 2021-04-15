package aggregatelog

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/olivere/elastic"
)

const LOG_AGGREGATION_INDEX = "log_aggregation"

type MessageData struct {
	Time     time.Time `json:"time"`
	LogLevel string    `json:"level"`
	Message  string    `json:"message"`
	NodeID   string    `json:"nodeid,omitempty"`
	TestName string    `json:"testname,omitempty"`
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

func SendElasticMessage(message string, params ...interface{}) error {
	return SendMessageToElastic(message, INFO_LEVEL)
}

func SendElasticError(err error, params ...interface{}) error {
	var message string
	if err != nil {
		message = err.Error()
	} else {
		message = "Something wrong :|"
	}
	return SendMessageToElastic(message, ERROR_LEVEL)
}

func SendElasticDebug(message string, params ...interface{}) error {
	return SendMessageToElastic(message, DEBUG_LEVEL)
}
func SendElasticWarning(message string, params ...interface{}) error {
	return SendMessageToElastic(message, WARNING_LEVEL)
}
func SendElasticFatal(message string, params ...interface{}) error {
	return SendMessageToElastic(message, FATAL_LEVEL)
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
	validErr := ValidateElasticClient()
	if validErr != nil {
		return validErr
	}

	messageObject := MessageData{
		Time:     time.Now(),
		Message:  message,
		LogLevel: level,
		NodeID:   os.Getenv("NodeID"),
		TestName: os.Getenv("TestName"),
	}
	// putResult, err := elasticClient.Index().
	_, err := elasticClient.Index().
		Index(LOG_AGGREGATION_INDEX).
		Type("log").
		BodyJson(messageObject).
		Do(ctx)
	if err != nil {
		return err
	}
	// log.Printf("Indexed tweet %s to index %s, type %s\n", putResult.Id, putResult.Index, putResult.Type)
	// log.Printf("Indexed tweet %s to index %s, type %s\n", putResult.Id, putResult.Index, putResult.Type)
	return nil
}
