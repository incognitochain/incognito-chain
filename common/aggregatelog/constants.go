package aggregatelog

import "os"

const (
	SENTRY_LOG_SERVICENAME  = "sentry"
	ELASTIC_LOG_SERVICENAME = "elastic"

	INFO_LEVEL    = "info"
	ERROR_LEVEL   = "error"
	WARNING_LEVEL = "warning"
	DEBUG_LEVEL   = "debug"
	FATAL_LEVEL   = "fatal"
)

var SENTRY_DSN = os.Getenv("SENTRY_DSN")
var ELASTIC_URL = os.Getenv("ELASTIC_URL")
