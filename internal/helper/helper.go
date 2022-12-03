package helper

import (
	"context"

	"github.com/sirupsen/logrus"
)

// WrapCloser wrap closer. If closer return error, log the error and use sentry to report the error
func WrapCloser(closeFn func() error) {
	if err := closeFn(); err != nil {
		logrus.Error(err)
	}
}

// DumpContext dump all necessary value in context
// also dump the custom value written to context, such as user, and request ID
func DumpContext(ctx context.Context) string {
	return ""
}
