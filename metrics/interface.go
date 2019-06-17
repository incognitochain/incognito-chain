package metrics

type MetricTool interface {
	SendTimeSeriesMetricData(params map[string]interface{})
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
