package aggregatelog

import "os"

const (
	SENTRY_LOG_SERVICENAME = "sentry"
)

var SENTRY_DSN = os.Getenv("SENTRY_DSN")
