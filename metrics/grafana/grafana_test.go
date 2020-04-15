package grafana

import (
	"os"
	"testing"
)

func TestPoolDataSendTimeSeriesMetricDataGrafana(T *testing.T) {
	data := map[string]interface{}{
		Measurement:      PoolSize,
		MeasurementValue: float64(10),
	}
	grafanaEmptyUrl := NewGrafana("", "")
	grafanaEmptyUrl.SendTimeSeriesMetricData(data)
	os.Setenv("GRAFANAURL", GrafanaURL)
	grafana := NewGrafana(os.Getenv("GRAFANAURL"), "")
	grafana.SendTimeSeriesMetricData(data)
}

func TestBlockDataSendTimeSeriesMetricDataGrafana(T *testing.T) {
	os.Setenv("GRAFANAURL", GrafanaURL)
	data := map[string]interface{}{
		Measurement:      TxInOneBlock,
		MeasurementValue: float64(10000),
		Tag:              BlockHeightTag,
		TagValue:         "1000",
	}
	grafana := NewGrafana(os.Getenv("GRAFANAURL"), "")
	grafana.SendTimeSeriesMetricData(data)
}
