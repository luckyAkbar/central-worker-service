package model

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// DeleteDiaryCallbackDataSeparator separator for delete diary callback data
	DeleteDiaryCallbackDataSeparator = ";"
)

// TelegramCallbackQuery is a query indicating for callback handler. All value that will be used
// for callback query in callback data, should use type TelegramCallbackQuery
type TelegramCallbackQuery string

// list of the callback query data (all type e.g static, prefix, suffix)
var (
	RegisterSecretMessagingService TelegramCallbackQuery = "register_secret_telegram_messaging"
	DeleteDiaryPrefix              TelegramCallbackQuery = "delete_diary"
)

// GenerateDeleteDiaryCallbackQuery will return callback data for delete diary query
func GenerateDeleteDiaryCallbackQuery(diaryID string) string {
	return fmt.Sprintf("%s%s%s", DeleteDiaryPrefix, DeleteDiaryCallbackDataSeparator, diaryID)
}

// GetDiaryIDFromDiaryCallbackQuery will extract diary ID from delete diary callback data
// error returned when the supposed string format is wrong
func GetDiaryIDFromDiaryCallbackQuery(data string) (string, error) {
	res := strings.Split(data, DeleteDiaryCallbackDataSeparator)

	if len(res) != 2 {
		return "", errors.New("invalid data on get diary ID for delete ID callback query")
	}

	return res[1], nil
}

// TelegramBot interface for telegram bot
type TelegramBot interface {
	// RegisterHandlers register all handlers
	RegisterHandlers()
}
