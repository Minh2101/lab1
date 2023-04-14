package controllers

import (
	"time"
)

func ConvertTimeYYYYMMDD(input string) time.Time {
	time, _ := time.Parse("2006-01-02", input)
	return time
}
func Now() time.Time {
	return time.Now().Add(time.Duration(+7) * time.Hour)
}
