package controllers

import (
	"time"
)

func ConvertTimeYYYYMMDD(input string) time.Time {
	time, _ := time.Parse("2006-01-02", input)
	return time
}
