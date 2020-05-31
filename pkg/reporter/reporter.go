package reporter

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// Initialize makes sure sentry is ready to use
func Initialize(dsn string) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		return errors.Wrap(err, "error initializing Sentry")
	}
	return nil
}

// Report uploads error info to sentry
func Report(err error) {
	if err == nil {
		return
	}
	defer sentry.Flush(2 * time.Second)
	log.Printf("[ERROR] %s", err.Error())
	sentry.CaptureException(err)
}

// ErrorReportingRoutine accepts error chan and reports any errors on it.
func ErrorReportingRoutine(errors <-chan error) {
	for err := range errors {
		Report(err)
	}
}
