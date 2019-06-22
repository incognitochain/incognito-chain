package aggregatelog

/*import (
	"errors"

	raven "github.com/getsentry/raven-go"
)

var sentryClient *raven.Client

func ValidateSentryClient() error {
	if sentryClient == nil {
		return errors.New("Sentry client not initialized")
	}
	return nil
}

func InitSentry(params map[string]interface{}) error {
	DSNValue, ok := params["DSN"]
	if !ok || DSNValue == "" {
		return errors.New("Sentry DNS config empty")
	}
	DSN, ok := DSNValue.(string)
	if !ok {
		return errors.New("Sentry DNS config invalid")
	}
	err := CreateSentryService(DSN)
	if err != nil {
		return err
	}
	return nil
}

func CreateSentryService(DNS string) error {

	if DNS == "" {
		return errors.New("Sentry DSN setting invalid")
	}
	client, err := raven.NewClient(DNS, nil)
	if err != nil {
		return err
	}
	sentryClient = client
	return nil
}

func CaptureSentryMessage(message string, params ...interface{}) error {
	tags := map[string]string{
		"level": INFO_LEVEL,
		"type":  INFO_LEVEL,
	}
	return sendMessageToSentry(message, tags)
}

func CaptureSentryError(err error, params ...interface{}) error {
	clientErr := ValidateSentryClient()
	if clientErr != nil {
		return clientErr
	}
	tags := map[string]string{
		"level": ERROR_LEVEL,
		"type":  ERROR_LEVEL,
	}
	sentryClient.CaptureError(err, tags, nil)
	return nil
}

func CaptureSentryDebug(message string, params ...interface{}) error {
	tags := map[string]string{
		"level": DEBUG_LEVEL,
		"type":  DEBUG_LEVEL,
	}
	return sendMessageToSentry(message, tags)
}

func CaptureSentryWarning(message string, params ...interface{}) error {
	tags := map[string]string{
		"level": WARNING_LEVEL,
		"type":  WARNING_LEVEL,
	}
	return sendMessageToSentry(message, tags)
}

func CaptureSentryFatal(message string, params ...interface{}) error {
	tags := map[string]string{
		"level": FATAL_LEVEL,
		"type":  FATAL_LEVEL,
	}
	return sendMessageToSentry(message, tags)
}

func sendMessageToSentry(message string, tags map[string]string) error {
	clientErr := ValidateSentryClient()
	if clientErr != nil {
		return clientErr
	}
	sentryClient.CaptureMessage(message, tags, nil)
	return nil
}*/
