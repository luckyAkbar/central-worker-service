// Package helper contains all generic helper functionality to support DRY principle
package helper

import (
	"fmt"

	"github.com/luckyAkbar/central-worker-service/internal/model"
)

// FlattenAndFormatDiaries will format diaries using this format
// fmt.Sprintf("<strong>%s</strong> - %s\n", diary.CreatedAt.Format("02 Jan 2006"), diary.Note)
func FlattenAndFormatDiaries(diaries []model.Diary) string {
	var result string
	for _, diary := range diaries {
		result += fmt.Sprintf("<strong>%s</strong> - %s\n", diary.CreatedAt.Format("02 Jan 2006"), diary.Note)
	}

	return result
}
