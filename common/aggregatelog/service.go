package aggregatelog

import (
	"errors"
)

type LogService struct {
	InitService    InitService
	CaptureMessage CaptureMessage
	CaptureError   CaptureError
	CaptureDebug   CaptureDebug
	CaptureWarning CaptureWarning
	CaptureFatal   CaptureFatal
}

type CaptureMessage func(message string, params ...interface{}) error
type CaptureError func(err error, params ...interface{}) error
type CaptureDebug func(message string, params ...interface{}) error
type CaptureWarning func(message string, params ...interface{}) error
type CaptureFatal func(message string, params ...interface{}) error

type InitService func(params map[string]interface{}) error

var LogServices = make(map[string]*LogService)

func RegisterService(serviceName string, service *LogService) {
	LogServices[serviceName] = service
}

func GetService(serviceName string) (*LogService, error) {
	service, ok := LogServices[serviceName]
	if !ok {
		return nil, errors.New("Service not exist")
	}
	return service, nil
}

func init() {
	/*RegisterService(SENTRY_LOG_SERVICENAME, &LogService{
		InitSentry,
		CaptureSentryMessage,
		CaptureSentryError,
		CaptureSentryDebug,
		CaptureSentryWarning,
		CaptureSentryFatal,
	})*/

	RegisterService(ELASTIC_LOG_SERVICENAME, &LogService{
		InitElastic,
		SendElasticMessage,
		SendElasticError,
		SendElasticDebug,
		SendElasticWarning,
		SendElasticFatal,
	})
}
