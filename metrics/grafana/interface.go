package grafana

type MetricTool interface {
	SendTimeSeriesMetricData(params map[string]interface{})
	SendTimeSeriesMetricDataWithTime(params map[string]interface{})
	GetExternalAddress() string
}

var metricTool MetricTool

func InitMetricTool(tool MetricTool) {
	metricTool = tool
}

func AnalyzeTimeSeriesMetricData(params map[string]interface{}) {
	if metricTool == nil {
		return
	}
	metricTool.SendTimeSeriesMetricData(params)
}

func AnalyzeTimeSeriesMetricDataWithTime(params map[string]interface{}) {
	if metricTool == nil {
		return
	}
	metricTool.SendTimeSeriesMetricDataWithTime(params)
}
