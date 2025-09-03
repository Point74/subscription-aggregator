package utils

import (
	"fmt"
	"time"
)

func parseDate(date string) (time.Time, error) {
	if date == "" {
		return time.Time{}, fmt.Errorf("date is empty")
	}

	resultTime, err := time.Parse("01-2006", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("date is invalid: %w", err)
	}

	return resultTime, nil
}
