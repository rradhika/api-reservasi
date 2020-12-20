package telegram_calendar

import (
	"strconv"
	"strings"
	"time"
)

func checkYearAndMonth(year *int, month *time.Month) {

	if *year == 0 {
		*year = time.Now().Year()
	}
	if *month == 0 {
		*month = time.Now().Month()
	}

}

func createCallbackData(action string, year int, month time.Month, day int) string {
	return strings.Join([]string{action, strconv.Itoa(year), strconv.Itoa(int(month)), strconv.Itoa(day)}, ";")
}

func separateCallbackData(data string) []string {
	return strings.Split(data, ";")
}
