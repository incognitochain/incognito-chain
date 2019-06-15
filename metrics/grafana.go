package metrics

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

//Influxdb write query
//<measurement>[,<tag-key>=<tag-value>...] <field-key>=<field-value>[,<field2-key>=<field2-value>...] [unix-nano-timestamp]
func SendTimeSeriesMetricDataGrafana(params map[string]interface{}) {
	databaseUrl := os.Getenv("GRAFANAURL")
	if databaseUrl == "" {
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
	req, err := http.NewRequest(http.MethodPost, databaseUrl, bytes.NewBuffer([]byte(dataBinary)))
	if err != nil {
		Logger.log.Debug("Create Request failed with err: ", err)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		Logger.log.Debug("Push to Grafana error:", err)
		return
	}
}
