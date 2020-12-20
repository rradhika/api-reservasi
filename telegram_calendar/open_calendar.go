package telegram_calendar

import (
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// You can add some bot_example message text before calendar.
// Otherwise bot_example message will contain month and year. Example: June , 2019 .
func (tc *TelegramCalendar) OpenCalendar(tipe string, update int64, message string) (tgbotapi.Message, error) {

	if len(message) == 0 {
		message = time.Now().Month().String() + " , " + strconv.Itoa(time.Now().Year())
	}

	var msg = tgbotapi.NewMessage(update, message)

	msg.ReplyMarkup = tc.CreateCalendar(tipe, 0, 0)
	calendarMessage, err := tc.bot.Send(msg)
	if err != nil {
		return tgbotapi.Message{}, err
	}
	msg.Text = ""
	msg.ReplyMarkup = nil

	return calendarMessage, nil

}
