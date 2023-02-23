package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// CommonCallbackDataSeparator common separator for callback query data
	CommonCallbackDataSeparator = ";"
)

// TelegramCallbackQuery is a query indicating for callback handler. All value that will be used
// for callback query in callback data, should use type TelegramCallbackQuery
type TelegramCallbackQuery string

// list of the callback query data (all type e.g static, prefix, suffix)
var (
	RegisterSecretMessagingService TelegramCallbackQuery = "register_secret_telegram_messaging"
	DeleteDiaryPrefix              TelegramCallbackQuery = "delete_diary"
	ReportSecretMessagePrefix      TelegramCallbackQuery = "report_secret_message"
	BlockSecretMessagingUserPrefix TelegramCallbackQuery = "block_secret_messaging_user"
	GagMemeServiceSubscription     TelegramCallbackQuery = "gag_meme_service_subscription"
)

// GenerateDeleteDiaryCallbackQuery will return callback data for delete diary query
func GenerateDeleteDiaryCallbackQuery(diaryID string) string {
	return fmt.Sprintf("%s%s%s", DeleteDiaryPrefix, CommonCallbackDataSeparator, diaryID)
}

// GetDiaryIDFromDiaryCallbackQuery will extract diary ID from delete diary callback data
// error returned when the supposed string format is wrong
func GetDiaryIDFromDiaryCallbackQuery(data string) (string, error) {
	res := strings.Split(data, CommonCallbackDataSeparator)

	if len(res) != 2 {
		return "", errors.New("invalid data on get diary ID for delete ID callback query")
	}

	return res[1], nil
}

// GenerateReportSecretMessageCallbackQuery generate data for callback query in report secret messaging
func GenerateReportSecretMessageCallbackQuery(secretMessageID int64) string {
	return fmt.Sprintf("%s%s%d", ReportSecretMessagePrefix, CommonCallbackDataSeparator, secretMessageID)
}

// GetSecretMessageIDFromReportSecretMessageCallbackQuery will extract secret message ID from report secret message callback data
func GetSecretMessageIDFromReportSecretMessageCallbackQuery(data string) (int64, error) {
	res := strings.Split(data, CommonCallbackDataSeparator)

	if len(res) != 2 {
		return 0, errors.New("invalid data on get secret message ID for report secret message callback query")
	}

	return strconv.ParseInt(res[1], 10, 64)
}

// GenerateBlockSecretMessagingUserCallbackQuery generate data for callback query in block secret messaging user
func GenerateBlockSecretMessagingUserCallbackQuery(userID int64) string {
	return fmt.Sprintf("%s%s%d", BlockSecretMessagingUserPrefix, CommonCallbackDataSeparator, userID)
}

// GetUserIDFromSecretMessagingUserCallbackQuery will extract user ID from block secret messaging user callback data
func GetUserIDFromSecretMessagingUserCallbackQuery(data string) (int64, error) {
	res := strings.Split(data, CommonCallbackDataSeparator)

	if len(res) != 2 {
		return 0, errors.New("invalid data on get user ID for block secret messaging user callback query")
	}

	return strconv.ParseInt(res[1], 10, 64)
}

// TelegramBot interface for telegram bot
type TelegramBot interface {
	// RegisterHandlers register all handlers
	RegisterHandlers()
}
