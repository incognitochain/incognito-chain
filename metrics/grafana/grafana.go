package grafana

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type Grafana struct {
	url             string
	externalAddress string
}

func NewGrafana(url, externalAddress string) Grafana {
	return Grafana{
		url:             url,
		externalAddress: externalAddress,
	}
}
func StartSystemMetrics() {
	if metricTool == nil {
		return
	}
	var externalAddress string
	if metricTool.GetExternalAddress() != "" {
		externalAddress = metricTool.GetExternalAddress()
	}
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		go metricTool.SendTimeSeriesMetricData(map[string]interface{}{
			Measurement:      NumberOfGoRoutine,
			MeasurementValue: float64(runtime.NumGoroutine()),
			Tag:              ExternalAddressTag,
			TagValue:         externalAddress,
		})
	}
}
func (grafana *Grafana) GetExternalAddress() string {
	return grafana.externalAddress
}

//Influxdb write query
//<measurement>[,<tag-key>=<tag-value>...] <field-key>=<field-value>[,<field2-key>=<field2-value>...] [unix-nano-timestamp]
func (grafana *Grafana) SendTimeSeriesMetricData(params map[string]interface{}) {
	if grafana.url == "" {
		return
	}
	var (
		measurement string
		tag         string
		tagValue    string
		value       float64
		dataBinary  string
	)
	switch len(params) {
	case 2:
		measurement = params[Measurement].(string)
		value = params[MeasurementValue].(float64)
		dataBinary = fmt.Sprintf("%s value=%f %d", measurement, value, time.Now().UnixNano())
	case 4:
		measurement = params[Measurement].(string)
		tag = params[Tag].(string)
		tagValue = params[TagValue].(string)
		value = params[MeasurementValue].(float64)
		dataBinary = fmt.Sprintf("%s,%+v=%s value=%f %d", measurement, tag, tagValue, value, time.Now().UnixNano())
	default:
		return
	}
	req, err := http.NewRequest(http.MethodPost, grafana.url, bytes.NewBuffer([]byte(dataBinary)))
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	client := &http.Client{}
	client.Do(req)
	return
}
func (grafana *Grafana) SendTimeSeriesMetricDataWithTime(params map[string]interface{}) {
	if grafana.url == "" {
		return
	}
	var (
		measurement string
		tag         string
		tagValue    string
		value       float64
		dataBinary  string
		writeTime   int64
	)
	switch len(params) {
	case 3:
		measurement = params[Measurement].(string)
		value = params[MeasurementValue].(float64)
		writeTime = params[Time].(int64)
		dataBinary = fmt.Sprintf("%s value=%f %d", measurement, value, writeTime*1000000000)
	case 5:
		measurement = params[Measurement].(string)
		tag = params[Tag].(string)
		tagValue = params[TagValue].(string)
		value = params[MeasurementValue].(float64)
		writeTime = params[Time].(int64)
		dataBinary = fmt.Sprintf("%s,%+v=%s value=%f %d", measurement, tag, tagValue, value, writeTime*1000000000)
	default:
		return
	}
	req, err := http.NewRequest(http.MethodPost, grafana.url, bytes.NewBuffer([]byte(dataBinary)))
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	client := &http.Client{}
	client.Do(req)
	return
}
