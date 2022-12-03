package helper

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/model"
	"github.com/sirupsen/logrus"
)

// WrapCloser wrap closer. If closer return error, log the error and use sentry to report the error
func WrapCloser(closeFn func() error) {
	if err := closeFn(); err != nil {
		logrus.Error(err)
	}
}

func GetRequestIDFromCtx(ctx context.Context) string {
	id, ok := ctx.Value(model.ReqIDCtxKey).(string)
	if !ok {
		logrus.WithField("ctx", utils.DumpIncomingContext).Warn("error getting request id")
		return ""
	}

	return id
}

// DumpContext dump all necessary value in context
// also dump the custom value written to context, such as user, and request ID
func DumpContext(ctx context.Context) string {
	reqID := GetRequestIDFromCtx(ctx)

	data := map[string]any{
		"request_id": reqID,
		"any":        utils.DumpIncomingContext(ctx),
	}

	v, err := json.Marshal(data)
	if err != nil {
		logrus.Error(err)
		return ""
	}

	return string(v)
}

// GenerateRequestID generate request ID using UUID v4 and stripping the '-'
func GenerateID() string {
	id := uuid.New()
	return strings.ReplaceAll(id.String(), "-", "")
}
