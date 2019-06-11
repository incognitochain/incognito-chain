package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Measurement
const (
	TxPoolValidated                  = "TxPoolValidated"
	TxPoolValidatedDetails          = "TxPoolValidatedDetails"
	TxPoolValidatedWithType          = "TxPoolValidatedWithType"
	TxPoolEntered                    = "TxPoolEntered"
	TxPoolEnteredWithType            = "TxPoolEnteredWithType"
	TxPoolAddedAfterValidation       = "TxPoolAddedAfterValidation"
	TxPoolRemoveAfterInBlock         = "TxPoolRemoveAfterInBlock"
	TxPoolRemoveAfterInBlockWithType = "TxPoolRemoveAfterInBlockWithType"
	TxPoolRemoveAfterLifeTime         = "TxPoolRemoveAfterLifeTime"
	TxAddedIntoPoolType               = "TxAddedIntoPoolType"
	TxPoolPrivacyOrNot                = "TxPoolPrivacyOrNot"
	PoolSize                          = "PoolSize"
	TxValidateByItSelfInPoolType      = "TxValidateByItSelfInPoolType"
	TxInOneBlock                      = "TxInOneBlock"
	DuplicateTxs                      = "DuplicateTxs"
	CreateAndSaveTxViewPointFromBlock = "CreateAndSaveTxViewPointFromBlock"
	NumOfBlockInsertToChain           = "NumOfBlockInsertToChain"
	TxPoolRemovedNumber               = "TxPoolRemovedNumber"
	TxPoolRemovedTime                 = "TxPoolRemovedTime"
	TxPoolRemovedTimeDetails          = "TxPoolRemovedTimeDetails"
	TxPoolTxBeginEnter                = "TxPoolTxBeginEnter"
	TxPoolTxEntered                   = "TxPoolTxEntered"
 )

// tag
const (
	BeaconBlock = "BeaconBlock"
	ShardBlock  = "ShardBlock"
	
	TxSizeMetric         = "txsize"
	TxSizeWithTypeMetric = "txsizewithtype"
	PoolSizeMetric       = "poolsize"
	TxTypeMetic          = "txtype"
	ValidateCondition          = "validatecond"
	VTBITxTypeMetic      = "vtbitxtype"
	TxPrivacyOrNotMetric = "txprivacyornot"
	BlockHeight          = "blockheight"
	TxHash               = "txhash"
	ShardID              = "shardid"
	Func = "func"

)

//Tag value
const (
	TxPrivacy                             = "privacy"
	TxNormalPrivacy                       = "normaltxprivacy"
	TxNoPrivacy                           = "noprivacy"
	TxNormalNoPrivacy                     = "normaltxnoprivacy"
	FuncCreateAndSaveTxViewPointFromBlock = "func-CreateAndSaveTxViewPointFromBlock"
	Beacon                                = "beacon"
	Condition1                      = "condition1"
	Condition2                      = "condition2"
	Condition3                      = "condition3"
	Condition4                      = "condition4"
	Condition5                      = "condition5"
	Condition6                      = "condition6"
	Condition7                      = "condition7"
	Condition8                      = "condition8"
	Condition9                      = "condition9"
	Condition10                      = "condition10"
	Condition11                      = "condition11"
)
// test value
var (
	blockpersecond int
)
func AnalyzeTimeSeriesTxSizeMetric(txSize string, metric string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxSizeMetric, txSize, metric, value)
}
func AnalyzeTimeSeriesTxSizeWithTypeMetric(txSizeWithType string, metric string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxSizeWithTypeMetric, txSizeWithType, metric, value)
}
func AnalyzeTimeSeriesTxsInOneBlockMetric(blockHeight string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(BlockHeight, blockHeight, TxInOneBlock, value)
}
func AnalyzeTimeSeriesTxTypeMetric(txType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxTypeMetic, txType, TxAddedIntoPoolType, value)
}
func AnalyzeTimeSeriesTxRemovedMetric(txType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxTypeMetic, txType, TxPoolRemovedNumber, value)
}
func AnalyzeTimeSeriesTxBeginEnterMetric(txType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxTypeMetic, txType, TxPoolTxBeginEnter, value)
}
func AnalyzeTimeSeriesTxEnteredMetric(txType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxTypeMetic, txType, TxPoolTxEntered, value)
}
func AnalyzeTimeSeriesTxPrivacyOrNotMetric(txType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxPrivacyOrNotMetric, txType, TxPoolPrivacyOrNot, value)
}
func AnalyzeTimeSeriesVTBITxTypeMetric(txType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(VTBITxTypeMetic, txType, TxValidateByItSelfInPoolType, value)
}
func AnalyzeTimeSeriesPoolSizeMetric(numOfTxs string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(PoolSizeMetric, numOfTxs, PoolSize, value)
}
func AnalyzeTimeSeriesTxDuplicateTimesMetric(txHash string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxHash, txHash, DuplicateTxs, value)
}
func AnalyzeTimeSeriesBlockPerSecondTimesMetric(shardID string, value float64, blockHeight uint64) {
	sendTimeSeriesMetricDataInfluxDBV2(ShardID, shardID, NumOfBlockInsertToChain, value)
}
func AnalyzeTimeSeriesTxRemovedTimeMetric(txType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(TxTypeMetic, txType, TxPoolRemovedTime, value)
}
func AnalyzeTimeSeriesTxValidationTimeDetailsMetric(conditionType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(ValidateCondition, conditionType, TxPoolValidatedDetails, value)
}
func AnalyzeTimeSeriesTxRemovedTimeDetailsTimeMetric(conditionType string, value float64) {
	sendTimeSeriesMetricDataInfluxDBV2(ValidateCondition, conditionType, TxPoolRemovedTimeDetails, value)
}
func AnalyzeFuncCreateAndSaveTxViewPointFromBlock(time float64) {
	sendTimeSeriesMetricDataInfluxDBV2(Func, FuncCreateAndSaveTxViewPointFromBlock, CreateAndSaveTxViewPointFromBlock, time)
}

func sendTimeSeriesMetricDataInfluxDBV2(metricTag string, tagValue string, metric string, value ...float64) error {
	//os.Setenv("GrafanaURL", "http://128.199.96.206:8086/write?db=mydb")
	databaseUrl := os.Getenv("GRAFANAURL")
	if databaseUrl == "" {
		return nil
	}
	dataBinary := fmt.Sprintf("%s,%+v=%s value=%f %d", metric, metricTag, tagValue, value[0], time.Now().UnixNano())
	req, err := http.NewRequest(http.MethodPost, databaseUrl, bytes.NewBuffer([]byte(dataBinary)))
	if err != nil {
		log.Println("Create Request failed with err: ", err)
		return err
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	
	client := &http.Client{}
	//res, err := client.Do(req)
	_, err = client.Do(req)
	if err != nil {
		log.Println("Push to Grafana error:", err)
		return err
	}
	return nil
	//fmt.Println("Grafana Response: ", res)
}

func AnalyzeTimeSeriesBeaconBlockMetric(paymentAddress string, value float64) {
	sendTimeSeriesMetricDataInfluxDB(paymentAddress, BeaconBlock, value)
}

func AnalyzeTimeSeriesShardBlockMetric(paymentAddress string, value float64) {
	go sendTimeSeriesMetricDataInfluxDB(paymentAddress, ShardBlock, value)
}

func sendTimeSeriesMetricDataInfluxDB(id string, metric string, value ...float64) {
	
	//os.Setenv("GrafanaURL", "http://128.199.96.206:8086/write?db=mydb")
	databaseUrl := os.Getenv("GrafanaURL")
	if databaseUrl == "" {
		return
	}
	
	nodeName := os.Getenv("NodeName")
	if nodeName == "" {
		nodeName = id
	}
	if nodeName == "" || len(value) == 0 || value[0] == 0 || metric == "" {
		return
	}
	
	dataBinary := ""
	if len(value) == 1 {
		dataBinary = fmt.Sprintf("%s,node=%s value=%f %d000000000", metric, nodeName, value[0], time.Now().Unix())
	} else {
		dataBinary = fmt.Sprintf("%s,node=%s ", metric, nodeName)
		for i, value := range value {
			dataBinary += fmt.Sprintf("value%d=%f", i, value)
		}
		dataBinary += fmt.Sprintf(" %d000000000", time.Now().Unix())
	}
	req, err := http.NewRequest(http.MethodPost, databaseUrl, bytes.NewBuffer([]byte(dataBinary)))
	if err != nil {
		log.Println("Create Request failed with err: ", err)
		return
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	
	client := &http.Client{}
	//res, err := client.Do(req)
	_, err = client.Do(req)
	if err != nil {
		log.Println("Push to Grafana error:", err)
		return
	}
	//fmt.Println("Grafana Response: ", res)
}
