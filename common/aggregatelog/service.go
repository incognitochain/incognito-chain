package aggregatelog

import "errors"

type LogService struct {
	InitService    InitService
	CaptureMessage CaptureMessage
	CaptureError   CaptureError
}

type CaptureError func(err error) error
type CaptureMessage func(message string) error
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
	RegisterService(SENTRY_LOG_SERVICENAME, &LogService{
		InitSentry,
		CaptureSentryMessage,
		CaptureSentryError,
	})

	RegisterService(ELASTIC_LOG_SERVICENAME, &LogService{
		InitElastic,
		SendElasticMessage,
		SendElasticError,
	})
}
