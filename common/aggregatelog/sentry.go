package aggregatelog

import (
	"errors"

	raven "github.com/getsentry/raven-go"
)

var sentryClient *raven.Client

func ValidateClient() error {
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

func CaptureSentryMessage(message string) error {
	err := ValidateClient()
	if err != nil {
		return err
	}
	sentryClient.CaptureMessage(message, nil)
	return nil
}

func CaptureSentryError(err error) error {
	clientErr := ValidateClient()
	if clientErr != nil {
		return clientErr
	}
	sentryClient.CaptureError(err, nil)
	return nil
}
