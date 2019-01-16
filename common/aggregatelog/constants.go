package aggregatelog

import "os"

const (
	SENTRY_LOG_SERVICENAME  = "sentry"
	ELASTIC_LOG_SERVICENAME = "elastic"
)

var SENTRY_DSN = os.Getenv("SENTRY_DSN")
var ELASTIC_URL = os.Getenv("ELASTIC_URL")
