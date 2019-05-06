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
	TxPoolValidated  = "TxPoolValidated"
	TxPoolValidatedWithType  = "TxPoolValidatedWithType"
	TxPoolEntered  = "TxPoolEntered"
	TxPoolEnteredWithType  = "TxPoolEnteredWithType"
	TxPoolAddedAfterValidation  = "TxPoolAddedAfterValidation"
	TxPoolRemoveAfterInBlock = "TxPoolRemoveAfterInBlock"
	TxPoolRemoveAfterInBlockWithType = "TxPoolRemoveAfterInBlockWithType"
	TxPoolRemoveAfterLifeTime = "TxPoolRemoveAfterLifeTime"
	TxAddedIntoPoolType = "TxAddedIntoPoolType"
	TxPoolPrivacyOrNot = "TxAddedIntoPoolType"
	PoolSize = "PoolSize"
	TxValidateByItSelfInPoolType = "TxValidateByItSelfInPoolType"
	TxInOneBlock = "TxInOneBlock"
)
// tag
const (
	BeaconBlock = "BeaconBlock"
	ShardBlock  = "ShardBlock"
	
	TxSizeMetric = "txsize"
	TxSizeWithTypeMetric = "txsizewithtype"
	PoolSizeMetric = "poolsize"
	TxTypeMetic = "txtype"
	VTBITxTypeMetic = "vtbitxtype"
	TxPrivacyOrNotMetric = "txprivacyornot"
	BlockHeight = "blockheight"
	
)
//Tag value
const (
	TxPrivacy = "privacy"
	TxNormalPrivacy = "normaltxprivacy"
	TxNoPrivacy = "noprivacy"
	TxNormalNoPrivacy = "normaltxnoprivacy"
)
func AnalyzeTimeSeriesTxSizeMetric(txSize string, metric string, value float64){
	sendTimeSeriesTransactionMetricDataInfluxDB(TxSizeMetric, txSize, metric, value)
}
func AnalyzeTimeSeriesTxSizeWithTypeMetric(txSizeWithType string, metric string, value float64){
	sendTimeSeriesTransactionMetricDataInfluxDB(TxSizeWithTypeMetric, txSizeWithType, metric, value)
}
func AnalyzeTimeSeriesTxsInOneBlockMetric(blockHeight string, value float64){
	sendTimeSeriesTransactionMetricDataInfluxDB(BlockHeight, blockHeight, TxInOneBlock, value)
}
func AnalyzeTimeSeriesTxTypeMetric(txType string, value float64){
	sendTimeSeriesTransactionMetricDataInfluxDB(TxTypeMetic, txType, TxAddedIntoPoolType, value)
}
func AnalyzeTimeSeriesTxPrivacyOrNotMetric(txType string, value float64){
	sendTimeSeriesTransactionMetricDataInfluxDB(TxPrivacyOrNotMetric, txType, TxPoolPrivacyOrNot, value)
}
func AnalyzeTimeSeriesVTBITxTypeMetric(txType string, value float64){
	sendTimeSeriesTransactionMetricDataInfluxDB(VTBITxTypeMetic, txType, TxValidateByItSelfInPoolType, value)
}
func AnalyzeTimeSeriesPoolSizeMetric(numOfTxs string, value float64){
	sendTimeSeriesTransactionMetricDataInfluxDB(PoolSizeMetric, numOfTxs, PoolSize, value)
}
func sendTimeSeriesTransactionMetricDataInfluxDB(metricTag string, tagValue string, metric string, value ...float64) {
	//os.Setenv("GrafanaURL", "http://128.199.96.206:8086/write?db=mydb")
	databaseUrl := os.Getenv("GrafanaURL")
	if databaseUrl == "" {
		return
	}
	dataBinary := fmt.Sprintf("%s,%+v=%s value=%f %d000000000", metric, metricTag, tagValue, value[0], time.Now().Unix())
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
	_,err = client.Do(req)
	if err != nil {
		log.Println("Push to Grafana error:", err)
		return
	}
	//fmt.Println("Grafana Response: ", res)
}

func AnalyzeTimeSeriesBeaconBlockMetric(paymentAddress string, value float64) {
	sendTimeSeriesMetricDataInfluxDB(paymentAddress, BeaconBlock, value)
}

func AnalyzeTimeSeriesShardBlockMetric(paymentAddress string, value float64) {
	go sendTimeSeriesMetricDataInfluxDB(paymentAddress, ShardBlock, value)
}

func sendTimeSeriesMetricDataInfluxDB(id string, metric string, value ...float64) {
	
	os.Setenv("GrafanaURL", "http://128.199.96.206:8086/write?db=mydb")
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
	_,err = client.Do(req)
	if err != nil {
		log.Println("Push to Grafana error:", err)
		return
	}
	//fmt.Println("Grafana Response: ", res)
}
