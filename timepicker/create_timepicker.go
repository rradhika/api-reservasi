package timepicker

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//Timepicker struct
type Timepicker struct {
	bot *tgbotapi.BotAPI
}

// CreateTimepicker will generate hour & minutes
func CreateTimepicker(callbackSet string, setMsg string, jam string, menit string) tgbotapi.InlineKeyboardMarkup {
	var keyboard tgbotapi.InlineKeyboardMarkup
	var row []tgbotapi.InlineKeyboardButton

	//Callback Button
	upJamCallbackData := CreateCallbackData("UPDATE-JAM-"+callbackSet, CreateNextJam(jam), menit)
	upMenitCallbackData := CreateCallbackData("UPDATE-JAM-"+callbackSet, jam, CreateNextMenit(menit))
	downMenitCallbackData := CreateCallbackData("UPDATE-JAM-"+callbackSet, jam, CreatePrevMenit(menit))
	downJamCallbackData := CreateCallbackData("UPDATE-JAM-"+callbackSet, CreatePrevJam(jam), menit)
	setJamMenit := CreateCallbackData("SET-JAM-"+callbackSet, jam, menit)

	row = []tgbotapi.InlineKeyboardButton{}

	row = append(row, tgbotapi.InlineKeyboardButton{Text: "🔼", CallbackData: &upJamCallbackData})
	row = append(row, tgbotapi.InlineKeyboardButton{Text: "🔼", CallbackData: &upMenitCallbackData})

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	row = []tgbotapi.InlineKeyboardButton{}
	row = append(row, tgbotapi.InlineKeyboardButton{Text: jam, CallbackData: &upMenitCallbackData})
	row = append(row, tgbotapi.InlineKeyboardButton{Text: menit, CallbackData: &upMenitCallbackData})

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	row = []tgbotapi.InlineKeyboardButton{}

	row = append(row, tgbotapi.InlineKeyboardButton{Text: "🔽", CallbackData: &downJamCallbackData})
	row = append(row, tgbotapi.InlineKeyboardButton{Text: "🔽", CallbackData: &downMenitCallbackData})

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	row = []tgbotapi.InlineKeyboardButton{}

	row = append(row, tgbotapi.InlineKeyboardButton{Text: "✓ " + setMsg, CallbackData: &setJamMenit})

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)

	return keyboard
}
