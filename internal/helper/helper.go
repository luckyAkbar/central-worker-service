package helper

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// GetRequestIDFromCtx self explained
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

// GenerateID generate ID using UUID v4 and stripping the '-'
func GenerateID() string {
	id := uuid.New()
	return strings.ReplaceAll(id.String(), "-", "")
}

// CreateHashSHA512 generate sha512 hash and return result in base64 encoded string
func CreateHashSHA512(data []byte) string {
	sha := sha512.New()
	sha.Write(data)

	return EncodeBase64(sha.Sum(nil))
}

// EncodeBase64 encode data to base64 string
func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeBase64 decode data to []byte and return err if any
func DecodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// HTMLContentForUserRegistrationEmail create html content for user registration email
func HTMLContentForUserRegistrationEmail(username, url string) string {
	return fmt.Sprintf(`
	<html>
		<h1>Thanks for registering on our platform, Central Service!</h1>
		<p>Hello, %s!, recently you register your account on our platform. To activate the account, please click link below.</p>
		<p><a href="%s">activate</a></p>
		<br><br>
		<p>But, if you never register on our platform, just ignore this message.</p>
		<p>Thanks in advance!</p>
	</html>
	
	
	`, username, url)
}
