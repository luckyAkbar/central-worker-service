package helper

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kumparan/go-utils"
	"github.com/luckyAkbar/central-worker-service/internal/config"
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
		"user":       utils.Dump(model.GetUserFromCtx(ctx)),
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

// HTMLContentForReportSecretMessagingService return html content for report secret messaging email for admin
func HTMLContentForReportSecretMessagingService(msgNode *model.SecretMessageNode) string {
	return fmt.Sprintf(`
	<html>
		<h1>User Report on Secret Messaging Service</h1>
		<p>Hi, Telegram Bot Admin! There is a user report on Secret Messaging Service. Please check the report below.</p>

		<p>Message Node ID: %d</p>
		<p>Secret Message Session ID: %s</p>
		<p>Text: %s</p>

		<p>Thanks in advance!</p>
		<p>this is an auto-generated email, please do not reply to this email.</p>
	</html>
	`, msgNode.ID, msgNode.SessionID, msgNode.Text)
}

// BaseLetter source to generate the random string
type BaseLetter string

// BaseLetter constants
const (
	AlphabetCaps  BaseLetter = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphabetLower BaseLetter = "abcdefghijklmnopqrstuvwxyz"
	Numeric       BaseLetter = "0123456789"
	AlphaNumeric  BaseLetter = AlphabetLower + AlphabetCaps + Numeric
)

const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

// GenerateToken generate random alphanumeric character adapted from
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func GenerateToken(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(AlphaNumeric) {
			sb.WriteByte(AlphaNumeric[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

// FilterImageMimetype check mimetype and validate if allowed
func FilterImageMimetype(mimetype string) error {
	logger := logrus.WithField("func", "helper_image_mimetype")
	err := errors.New("image mimetype is not supported")

	if mimetype == "" {
		logger.Info("mimetype is empty")
		return err
	}

	for _, allowedType := range config.ImageMediaAllowedTypes() {
		if mimetype == allowedType {
			return nil
		}
	}

	logger.Info("mimetype is not allowed: ", mimetype)

	return err
}

// SaveMediaImageToLocalStorage save the file media to local storage
func SaveMediaImageToLocalStorage(file *multipart.FileHeader, storagePath string, fullname string) error {
	logger := logrus.WithFields(logrus.Fields{
		"file":        utils.Dump(file),
		"storagePath": storagePath,
		"fullname":    fullname,
	})

	src, err := file.Open()
	if err != nil {
		logger.Error("failed to open file: ", err)
		return err
	}

	defer WrapCloser(src.Close)

	dst, err := os.Create(fmt.Sprintf("%s/%s", storagePath, fullname))
	if err != nil {
		logger.Error("failed to create file: ", err)
		return err
	}

	defer WrapCloser(dst.Close)

	total, err := io.Copy(dst, src)
	if err != nil {
		logger.Error("failed to copy file: ", err)
		return err
	}

	logger.Info("success save image with total size: ", total)

	return nil
}

// GenerateStartAndEndDate will create time instance based on date
// return (start, end, error)
// start is the date with 00:00:00
// end is the date with 23:59:59
func GenerateStartAndEndDate(date string) (time.Time, time.Time, error) {
	switch strings.ToLower(date) {
	default:
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			logrus.WithError(err).Info("failed to parse date: ", date)
			return t, t, err
		}

		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()), nil
	case "today":
		// TODO: maybe should confirm user timezone?
		t := time.Now()
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()), nil

	case "yesterday":
		t := time.Now().AddDate(0, 0, -1)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location()), nil
	}
}
