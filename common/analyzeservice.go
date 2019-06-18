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
	PoolSize                          = "PoolSize"
 )

// tag
const (
	BeaconBlock = "BeaconBlock"
	ShardBlock  = "ShardBlock"
	BlockHeight          = "blockheight"
	TxHash               = "txhash"
	ShardID              = "shardid"
	Func = "func"

)

//Tag value
const (
	Beacon                                = "beacon"
)

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
