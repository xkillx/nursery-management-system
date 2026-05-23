package application

import (
	"fmt"
	"time"
)

var londonLocation *time.Location

func init() {
	var err error
	londonLocation, err = time.LoadLocation("Europe/London")
	if err != nil {
		panic(fmt.Sprintf("load Europe/London timezone: %v", err))
	}
}

func LondonNow() (utc time.Time, localDate time.Time) {
	utc = time.Now().UTC()
	localDate = utc.In(londonLocation)
	return utc, localDate
}

func LondonLocalDate(utc time.Time) time.Time {
	return utc.In(londonLocation)
}
