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

// list metric
const (
	BeaconBlock = "BeaconBlock"
	ShardBlock  = "ShardBlock"
)

func AnalyzeTimeSeriesBeaconBlockMetric(paymentAddress string, value float64) {
	sendMetricDataToGrafana(paymentAddress, value, BeaconBlock)
}

func AnalyzeTimeSeriesShardBlockMetric(paymentAddress string, value float64) {
	go sendMetricDataToGrafana(paymentAddress, value, ShardBlock)
}

func sendMetricDataToGrafana(id string, value float64, metric string) {

	grafanaURL := os.Getenv("GrafanaURL")
	if grafanaURL == "" {
		return
	}

	nodeName := os.Getenv("NodeName")
	if nodeName == "" {
		nodeName = id
	}
	if nodeName == "" || value == 0 || metric == "" {
		return
	}

	dataBinary := fmt.Sprintf("%s,node=%s value=%f %d000000000", metric, nodeName, value, time.Now().Unix())
	req, err := http.NewRequest(http.MethodPost, grafanaURL, bytes.NewBuffer([]byte(dataBinary)))
	if err != nil {
		log.Println("Create Request failed with err: ", err)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Println("Push to Grafana error:", err)
		return
	}
}
