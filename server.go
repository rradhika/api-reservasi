package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/rradhika/api-reservasi/telegram_calendar"
	tp "github.com/rradhika/api-reservasi/timepicker"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	. "github.com/rradhika/api-reservasi/datastore"
	"github.com/rradhika/api-reservasi/handler"
	"github.com/rradhika/api-reservasi/models"
)

const (
	layoutISO     = "2006-01-02T15:04:05+07:00"
	layoutTime    = "15:04"
	layoutTanggal = "2 January 2006"
)

//M is exported
type M map[string]interface{}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func main() {

	NewDB()
	db := Db
	// model := models.Employee{}

	// empl := model.CheckPenggunaRuangan()
	// fmt.Println(empl)

	go StartBot()

	e := echo.New()
	e.GET("/", handler.Welcome())
	e.GET("/employees", handler.GetEmployees(db))
	_ = e.Start(":3001")

	//To stop the program to finish and close
	select {}

	//You can also use https://golang.org/pkg/sync/#WaitGroup instead.
}

func StartBot() {
	// db, err := datastore.NewDB()
	bot, err := tgbotapi.NewBotAPI(goDotEnvVariable("BOT_TOKEN"))
	model := models.Employee{}
	roomsMod := models.Room{}
	revMod := models.Reservation{}
	revData := models.ReservationData{}

	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	tc := TelegramCalendar{}

	for update := range updates {

		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		var list string

		if update.CallbackQuery != nil {
			var msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")

			var loader string
			// var rspn string
			var empl []models.Employee
			var room []models.Room

			// var em string
			data := update.CallbackQuery.Data

			loader = "Please wait.."

			now := time.Now().Format(layoutTanggal)

			bot.AnswerCallbackQuery(
				tgbotapi.CallbackConfig{
					CallbackQueryID: update.CallbackQuery.ID,
					Text:            loader,
				},
			)

			//Contains:
			//check_schedule: to check schedule
			//reserve_place: to reserve new meeting room
			//pilih_ruangan: to set meeting room
			//cancel_reservation: to cancel existing reservation
			switch {
			case strings.Contains(data, "check_schedule"):
				// msg := fmt.Sprintf("You choose %s", update.CallbackQuery.Message.Chat.ID)
				// rspn = "Check Penggunaan Ruangan"
				checkSchedule := model.CheckToday()
				if !checkSchedule {
					list += "Pemakaian ruang meeting dan fun room " + now + " masih kosong"
				} else {
					room = roomsMod.MeetingRoom("yogyakarta")
					list += "<b>Pemakaian ruang meeting dan fun room " + now + ":</b> \n\n"
					for _, rm := range room {
						list += "<b>" + rm.Name + "</b> :  \n"
						// fmt.Println("INI LHO: " + strconv.Itoa(int(rm.ID)))
						empl = model.CheckPenggunaRuangan(rm.ID)

						for _, emp := range empl {
							sd, _ := time.Parse(layoutISO, emp.StartDate)
							ed, _ := time.Parse(layoutISO, emp.EndDate)
							list += emp.Telegram + " :  " + sd.Format(layoutTime) + " - " + ed.Format(layoutTime) + "\n"
						}
						list += "\n"

					}
					list += "Bagi SPEcial team yang akan ijin menggunakan ruang meeting dan fun room, silahkan di infokan dgn tim HC-GA ya. \n\n"
					list += "Thank you"

				}
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, list)
				msg.ParseMode = "HTML"
				bot.Send(msg)
				continue
			case strings.Contains(data, "reserve_place"):
				// rspn = "Reservasi Ruangan"

				room = roomsMod.MeetingRoom("yogyakarta")

				btns := []tgbotapi.InlineKeyboardButton{}
				for _, rm := range room {
					cbrm := "pilih_ruangan;" + strconv.Itoa(int(rm.ID))
					btns = append(btns, tgbotapi.InlineKeyboardButton{Text: rm.Name, CallbackData: &cbrm})
				}

				roomsButton := tgbotapi.NewInlineKeyboardMarkup(btns)
				list = "Silahkan pilih ruangan:"

				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, list)
				msg.ReplyMarkup = &roomsButton
				bot.Send(msg)
				continue
			case strings.Contains(data, "pilih_ruangan"):
				cData := tp.SeparateCallbackData(data)
				i, _ := strconv.ParseInt(cData[1], 10, 64)
				checkChatID := revMod.CheckChatID(update.CallbackQuery.Message.Chat.ID)
				reservation := models.Reservation{}

				toBeQuery := models.Reservation{ChatID: update.CallbackQuery.Message.Chat.ID, RoomID: i}

				//Cek apakah chat ID ini ada di table temp? Jika tidak maka create
				if !checkChatID {
					_, err := reservation.Create(&toBeQuery)

					if err != nil {
						log.Fatal(err)
					}
					loader = "Ruangan Disimpan"

					//Else maka update pilihan ruangan
				} else {
					_, err := reservation.UpdateRoom(&toBeQuery)

					if err != nil {
						log.Fatal(err)
					}
					loader = "Ruangan Diupdate"

				}
				list = fmt.Sprintln("Silahkan Pilih Tanggal Mulai")

				var msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					list,
				)
				var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					tc.CreateCalendar("MULAI", 0, 0),
				)
				msg.ReplyMarkup = msgMarkup.ReplyMarkup
				bot.Send(msg)

				continue
				// list = fmt.Sprintln("Pilih Jam Mulai")
				// setMsg := "Set Jam Mulai"
				// keyboard := tp.CreateTimepicker("MULAI", setMsg, tp.JamPertama, tp.MenitPertama)

				// msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, list)
				// msg.ReplyMarkup = &keyboard
			case strings.Contains(data, "CALENDAR-MULAI"):
				//fmt.Sprintln(strings.Contains(data, "NEXT-MONTH"))
				resData := tp.SeparateCallbackData(data)
				action := resData[1]
				year := resData[2]
				month := resData[3]
				day := resData[4]

				switch action {

				case "DAY":
					reservation := models.Reservation{}
					startDate := year + "-" + month + "-" + day + " 00:00:00"
					toBeQuery := models.Reservation{ChatID: update.CallbackQuery.Message.Chat.ID, DateStart: startDate}

					//Update StartDate
					_, err := reservation.UpdateStartDate(&toBeQuery)

					if err != nil {
						log.Fatal(err)
					}
					list = fmt.Sprintln("Pilih Jam Mulai")
					setMsg := "Set Jam Mulai"

					var msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						list,
					)
					var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						tp.CreateTimepicker("MULAI", setMsg, tp.JamPertama, tp.MenitPertama),
					)
					msg.ReplyMarkup = msgMarkup.ReplyMarkup
					bot.Send(msg)
					continue
				}
				continue
			case strings.Contains(data, "UPDATE-JAM-MULAI"):
				list = fmt.Sprintln("Pilih Jam Mulai")
				setMsg := "Set Jam Mulai"
				resData := tp.SeparateCallbackData(data)

				var msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					update.CallbackQuery.Message.Text,
				)
				var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					tp.CreateTimepicker("MULAI", setMsg, resData[1], resData[2]),
				)
				msg.ReplyMarkup = msgMarkup.ReplyMarkup
				bot.Send(msg)
				continue
			case strings.Contains(data, "SET-JAM-MULAI"):
				// now := time.Now()
				resData := tp.SeparateCallbackData(data)
				tempData := revMod.GetTemp(update.CallbackQuery.Message.Chat.ID)
				sd, _ := time.Parse(layoutISO, tempData.DateStart)
				startDate := sd.Format(LayoutSQL) + " " + resData[1] + ":" + resData[2] + ":00"
				toBeQuery := models.Reservation{ChatID: update.CallbackQuery.Message.Chat.ID, DateStart: startDate}

				// UpdateStartDate
				_, err := revMod.UpdateStartDate(&toBeQuery)

				if err != nil {
					log.Fatal(err)
				}

				list = fmt.Sprintln("Silahkan Pilih Tanggal Berakhir")

				var msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					list,
				)
				var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					tc.CreateCalendar("AKHIR", 0, 0),
				)
				msg.ReplyMarkup = msgMarkup.ReplyMarkup
				bot.Send(msg)
				continue
			case strings.Contains(data, "CALENDAR-AKHIR"):
				//fmt.Sprintln(strings.Contains(data, "NEXT-MONTH"))
				resData := tp.SeparateCallbackData(data)
				action := resData[1]
				year := resData[2]
				month := resData[3]
				day := resData[4]

				switch action {

				case "DAY":
					reservation := models.Reservation{}
					endDate := year + "-" + month + "-" + day + " 00:00:00"
					toBeQuery := models.Reservation{ChatID: update.CallbackQuery.Message.Chat.ID, DateEnd: endDate}

					//Update StartDate
					_, err := reservation.UpdateEndDate(&toBeQuery)

					if err != nil {
						log.Fatal(err)
					}
					list = fmt.Sprintln("Pilih Jam Berakhir")
					setMsg := "Set Jam Berakhir"

					var msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						list,
					)
					var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						tp.CreateTimepicker("AKHIR", setMsg, tp.JamPertama, tp.MenitPertama),
					)
					msg.ReplyMarkup = msgMarkup.ReplyMarkup
					bot.Send(msg)
					continue
				}
				continue
			case strings.Contains(data, "UPDATE-JAM-AKHIR"):
				list = fmt.Sprintln("Pilih Jam Berakhir")
				setMsg := "Set Jam Berakhir"
				resData := tp.SeparateCallbackData(data)

				var msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					update.CallbackQuery.Message.Text,
				)
				var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					tp.CreateTimepicker("AKHIR", setMsg, resData[1], resData[2]),
				)
				msg.ReplyMarkup = msgMarkup.ReplyMarkup
				bot.Send(msg)
				continue
			case strings.Contains(data, "SET-JAM-AKHIR"):
				// now := time.Now()
				resData := tp.SeparateCallbackData(data)
				tempData := revMod.GetTemp(update.CallbackQuery.Message.Chat.ID)
				de, _ := time.Parse(layoutISO, tempData.DateEnd)
				endDate := de.Format(LayoutSQL) + " " + resData[1] + ":" + resData[2] + ":00"
				toBeQuery := models.Reservation{ChatID: update.CallbackQuery.Message.Chat.ID, DateEnd: endDate}

				// UpdateStartDate
				_, err := revMod.UpdateEndDate(&toBeQuery)

				if err != nil {
					log.Fatal(err)
				}

				tempData = revMod.GetTemp(update.CallbackQuery.Message.Chat.ID)
				emp := model.GetEmployee(update.CallbackQuery.From.UserName)
				toBeInserted := models.ReservationData{
					EmployeeID: emp.Esid,
					RoomID:     tempData.RoomID,
					StartDate:  tempData.DateStart,
					EndDate:    tempData.DateEnd,
				}

				_, err = revData.CreateData(&toBeInserted)
				revMod.DeleteTemp(update.CallbackQuery.Message.Chat.ID)

				if err != nil {
					log.Fatal(err)
				}

				list = fmt.Sprintln("Jadwal pemakaian ruangan meeting/fun room sudah berhasil diset. Terima kasih")

				var msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					list,
				)
				bot.Send(msg)
				continue
			case strings.Contains(data, "cancel_reservation"):
				emp := model.GetEmployee(update.CallbackQuery.From.UserName)
				revData.DeleteData(emp.Esid)
				revMod.DeleteTemp(update.CallbackQuery.Message.Chat.ID)

				list = fmt.Sprintln("Reservation has been cancelled")

				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, list)
				msg.ParseMode = "HTML"
				bot.Send(msg)
				continue

			}

		}

		msgText := update.Message.Text

		//Start Command
		if strings.Contains(msgText, "/start") {
			fnm, lnm := update.Message.From.FirstName, update.Message.From.LastName
			msg := fmt.Sprintf("Welcome %s %s, how are you doing? Please choose this option below.", fnm, lnm)

			cs := "check_schedule"
			rp := "reserve_place"
			cl := "cancel_reservation"

			markup := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.InlineKeyboardButton{
						Text:         "Check Schedule",
						CallbackData: &cs,
					},
					tgbotapi.InlineKeyboardButton{
						Text:         "Reserve Place",
						CallbackData: &rp,
					},
					tgbotapi.InlineKeyboardButton{
						Text:         "Cancel Reservation",
						CallbackData: &cl,
					},
				),
			)

			reply := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
			reply.ReplyMarkup = &markup
			bot.Send(reply)
		}

	}

}
