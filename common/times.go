package common

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"time"
)

const Layout = "2006-01-02"

/**
Get the date of this week's Monday
*/
func GetFirstDateOfWeek() (weekMonday string) {
	now := time.Now()
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	weekMonday = weekStartDate.Format(Layout)
	logs.Info("curWeekMonday: ", weekMonday)
	return
}

/**
Get the Monday date of the previous week
*/
func GetLastWeekFirstDate() (weekMonday string) {
	thisWeekMonday := GetFirstDateOfWeek()
	TimeMonday, _ := time.Parse(Layout, thisWeekMonday)
	lastWeekMonday := TimeMonday.AddDate(0, 0, -7)
	weekMonday = lastWeekMonday.Format(Layout)
	logs.Info("lastWeekMonday: ", weekMonday)
	return
}

// Get the start and end dates of the previous month
func GetLastMonthDate() (string, string) {
	year, month, _ := time.Now().Date()
	thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	start := thisMonth.AddDate(0, -1, 0).Format(DATE_FORMAT)
	// days = -1 last data || days = 0 first data
	end := thisMonth.AddDate(0, 0, 0).Format(DATE_FORMAT)
	timeRange := fmt.Sprintf("startTime~endTime: %s~%s", start, end)
	logs.Info(timeRange)
	return start, end
}
