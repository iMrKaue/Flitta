package utils

import "time"

func GetTodayDate() string {
	return time.Now().Format("2006-01-02")
}
