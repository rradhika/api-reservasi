package telegram_calendar

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (tc *TelegramCalendar) CloseCalendar(message tgbotapi.Message) error {

	var editMsg = tgbotapi.NewEditMessageText(
		message.Chat.ID,
		message.MessageID,
		message.Text,
	)
	var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
		message.Chat.ID,
		message.MessageID,
		tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}},
	)
	editMsg.ReplyMarkup = msgMarkup.ReplyMarkup
	if _, err := tc.bot.Send(editMsg); err != nil {
		return err
	}

	return nil

}
