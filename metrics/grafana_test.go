package metrics

import (
	"os"
	"testing"
)

func TestPoolDataSendTimeSeriesMetricDataGrafana(T *testing.T) {
	os.Setenv("GRAFANAURL", GrafanaURL)
	data := map[string]interface{}{
		Measurement:      PoolSize,
		MeasurementValue: float64(10),
	}
	SendTimeSeriesMetricDataGrafana(data)
}

func TestBlockDataSendTimeSeriesMetricDataGrafana(T *testing.T) {
	os.Setenv("GRAFANAURL", GrafanaURL)
	data := map[string]interface{}{
		Measurement:      TxInOneBlock,
		MeasurementValue: float64(10000),
		Tag:              BlockHeightTag,
		TagValue:         "1000",
	}
	SendTimeSeriesMetricDataGrafana(data)
}
