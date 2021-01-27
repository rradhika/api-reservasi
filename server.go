package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	. "github.com/rradhika/api-reservasi/telegram_calendar"
	tp "github.com/rradhika/api-reservasi/timepicker"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	. "github.com/rradhika/api-reservasi/datastore"
	"github.com/rradhika/api-reservasi/handler"
	"github.com/rradhika/api-reservasi/models"
)

var bot *tgbotapi.BotAPI
var updChannel tgbotapi.UpdatesChannel

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

	AuthBot()

	//Start initiate Cron every 9 AM
	c := cron.New()
	c.AddFunc("0 9 * * ?", RunPenggunaRuanganByGroup)

	//For notif every employees
	// c.AddFunc("@every 0h0m10s", RunPenggunaRuanganEmployee)

	//For notif in group
	// c.AddFunc("@every 0h0m10s", RunPenggunaRuanganByGroup)
	c.Start()
	///End Cron

	go StartBot()

	e := echo.New()
	e.GET("/", handler.Welcome())
	e.GET("/employees", handler.GetEmployees(db))
	_ = e.Start(":3001")

	//To stop the program to finish and close
	select {}

	//You can also use https://golang.org/pkg/sync/#WaitGroup instead.
}

//AuthBot authenticate the BOT
func AuthBot() {

	var err error
	bot, err = tgbotapi.NewBotAPI(goDotEnvVariable("BOT_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updChannel, err = bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatalln(err)
	}

}

//StartBot where the command and text processed
func StartBot() {

	tc := TelegramCalendar{}

	for update := range updChannel {
		//Declare tanggal sekarang
		yNow, mNow, dNow := time.Now().Date()

		nowRaw := time.Now()
		timeStampString := nowRaw.Format(LayoutFullSQL)
		timeStamp, _ := time.Parse(LayoutFullSQL, timeStampString)
		hr, min := timeStamp.Format(LayoutHour), timeStamp.Format(LayoutMinute)

		model := models.Employee{}
		roomsMod := models.Room{}
		revMod := models.Reservation{}
		revData := models.ReservationData{}

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

			now := time.Now().Format(LayoutTanggal)

			employees := model.GetEmployee(update.CallbackQuery.From.UserName)

			//Contains:
			//check_schedule: to check schedule
			//reserve_place: to reserve new meeting room
			//pilih_ruangan: to set meeting room
			//cancel_reservation: to cancel existing reservation
			switch {
			case strings.Contains(data, "check_schedule"):
				// msg := fmt.Sprintf("You choose %s", update.CallbackQuery.Message.Chat.ID)
				// rspn = "Check Penggunaan Ruangan"

				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)

				checkSchedule := model.CheckToday()
				if !checkSchedule {
					list += "Pemakaian ruang meeting dan fun room " + now + " masih kosong"
				} else {
					room = roomsMod.MeetingRoom(employees.Location)
					list += "<b>Pemakaian ruang meeting dan fun room " + now + ":</b> \n\n"
					for _, rm := range room {
						list += "<b>" + rm.Name + "</b> :  \n"
						// fmt.Println("INI LHO: " + strconv.Itoa(int(rm.ID)))
						empl = model.CheckPenggunaRuangan(rm.ID)

						for _, emp := range empl {
							sd, _ := time.Parse(LayoutISO, emp.StartDate)
							ed, _ := time.Parse(LayoutISO, emp.EndDate)

							sdSq, _ := time.Parse(LayoutSQL, sd.Format(LayoutSQL))
							edSq, _ := time.Parse(LayoutSQL, ed.Format(LayoutSQL))
							if !sameDate(edSq, sdSq) {
								sdNow, _ := time.Parse(LayoutSQL, nowRaw.Format(LayoutSQL))
								if !sameDate(edSq, sdNow) {
									ed, _ = time.Parse(LayoutTime, tp.JamTerakhir+":"+tp.MenitTerakhir)
								}
							}
							fmt.Printf("Tanggal sd: %s", sdSq)
							fmt.Printf("Tanggal ed: %s", edSq)

							list += emp.Telegram + " :  " + sd.Format(LayoutTime) + " - " + ed.Format(LayoutTime) + "\n"
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
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)
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
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)
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

				//VALIDASI - Tanggal minimal hari ini
				//REFORMAT BULAN & TANGGAL
				mo, _ := strconv.Atoi(month)
				if mo < 10 {
					month = "0" + month
				}
				da, _ := strconv.Atoi(day)
				if da < 10 {
					day = "0" + day
				}
				//REFORMAT BULAN & TANGGAL
				// ns, _ := time.Parse(LayoutFullSQL, strconv.Itoa(yNow)+"-"+strconv.Itoa(int(mNow))+"-"+strconv.Itoa(dNow)+" "+tp.JamTerakhir+":"+tp.MenitTerakhir+":00")
				// srt, _ := time.Parse(LayoutFullSQL, year+"-"+month+"-"+day+" "+hr+":"+min+":00")
				ns, _ := time.Parse(LayoutSQL, nowRaw.Format(LayoutSQL))
				srt, _ := time.Parse(LayoutSQL, year+"-"+month+"-"+day)

				fmt.Printf("Data: %s", data)
				fmt.Printf("yNow: %s", strconv.Itoa(yNow)+"-"+strconv.Itoa(int(mNow))+"-"+strconv.Itoa(dNow))
				fmt.Printf("Tanggal Hari Ini: %s", ns.Format(LayoutSQL))
				fmt.Printf("Tanggal Dipilih: %s", srt.Format(LayoutSQL))

				if !validDate(srt, ns) {
					loader = "Mohon Pilih Tanggal minimal hari ini"
					bot.AnswerCallbackQuery(
						tgbotapi.CallbackConfig{
							CallbackQueryID: update.CallbackQuery.ID,
							Text:            loader,
						},
					)
					continue
				}
				if sameDate(srt, ns) {
					ns, _ = time.Parse(LayoutTime, hr+":"+min)
					srt, _ = time.Parse(LayoutTime, tp.JamTerakhir+":"+tp.MenitTerakhir)
					fmt.Printf("sameDate: Jam Max: %s", ns)
					fmt.Printf("sameDate: Jam Dipilih: %s", srt)

					//JIKA JAM SEKARANG(NS) LEBIH DARI JAM KANTOR (SRT)
					if !validHour(ns, srt) {
						loader = "Mohon pilih tanggal besok ya, hari ini office hour sudah selesai. "
						bot.AnswerCallbackQuery(
							tgbotapi.CallbackConfig{
								CallbackQueryID: update.CallbackQuery.ID,
								Text:            loader,
							},
						)
						continue
					}

				}

				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)

				//END VALIDASI

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
			case strings.Contains(data, "UPDATE-CALENDAR"):
				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)

				resData := tp.SeparateCallbackData(data)
				selectorUpdate := resData[1]
				tipeUpdate := resData[2]
				yearUpdate, _ := strconv.ParseInt(resData[3], 10, 64)
				monthUpdate, _ := strconv.ParseInt(resData[4], 10, 64)
				selectedMonth := time.Date(int(yearUpdate), time.Month(monthUpdate), 1, 0, 0, 0, 0, time.Local)
				selectorMonth := time.Time{}
				switch selectorUpdate {
				case "PREV-MONTH":
					selectorMonth = selectedMonth.AddDate(0, -1, 0)

				case "NEXT-MONTH":
					selectorMonth = selectedMonth.AddDate(0, 1, 0)

				}

				var msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					update.CallbackQuery.Message.Text,
				)
				var msgMarkup = tgbotapi.NewEditMessageReplyMarkup(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					tc.CreateCalendar(tipeUpdate, selectorMonth.Year(), selectorMonth.Month()),
				)
				msg.ReplyMarkup = msgMarkup.ReplyMarkup
				bot.Send(msg)
				continue
			case strings.Contains(data, "UPDATE-JAM-MULAI"):
				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)
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
				sd, _ := time.Parse(LayoutISO, tempData.DateStart)
				startDate := sd.Format(LayoutSQL) + " " + resData[1] + ":" + resData[2] + ":00"

				//VALIDASI - Jam minimal adalah jam hari ini
				nsNow, _ := time.Parse(LayoutSQL, sd.Format(LayoutSQL))
				srtNow, _ := time.Parse(LayoutSQL, nowRaw.Format(LayoutSQL))
				if sameDate(srtNow, nsNow) {
					ns, _ := time.Parse(LayoutTime, nowRaw.Format(LayoutTime))
					srt, _ := time.Parse(LayoutTime, resData[1]+":"+resData[2])

					fmt.Printf("CheckJamMulai: Jam Minimum: %s", ns)
					fmt.Printf("CheckJamMulai: Jam Dipilih: %s", srt)

					//JIKA JAM DIPILIH(NS) LEBIH KECIL JAM SEKARANG (SRT)
					start := ns.UTC()
					check := srt.UTC()
					if check.Before(start) {
						loader = fmt.Sprintf("Mohon Pilih Jam minimal : %s", ns.Format(LayoutTime))
						bot.AnswerCallbackQuery(
							tgbotapi.CallbackConfig{
								CallbackQueryID: update.CallbackQuery.ID,
								Text:            loader,
							},
						)
						continue
					}
				}

				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)

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

				//REFORMAT BULAN & TANGGAL
				mo, _ := strconv.Atoi(month)
				if mo < 10 {
					month = "0" + month
				}
				da, _ := strconv.Atoi(day)
				if da < 10 {
					day = "0" + day
				}
				//REFORMAT BULAN & TANGGAL

				//VALIDASI - Tanggal minimal sama dengan tanggal mulai
				tempData := revMod.GetTemp(update.CallbackQuery.Message.Chat.ID)
				dateTempMulai := strings.Split(tempData.DateStart, "T")
				ns, _ := time.Parse(LayoutSQL, dateTempMulai[0])
				srt, _ := time.Parse(LayoutSQL, year+"-"+month+"-"+day)

				fmt.Printf("Tanggal Mulai Asli: %s \n", tempData.DateStart)
				fmt.Printf("Tanggal Mulai Tanggal: %s \n", dateTempMulai[0])
				fmt.Printf("Tanggal Mulai: %s \n", ns)
				fmt.Printf("Tanggal Dipilih: %s \n", srt)

				if !validDate(srt, ns) {
					loader = fmt.Sprintf("Mohon pilih tanggal minimal tanggal mulai: %s", ns.Format(LayoutTanggal))
					bot.AnswerCallbackQuery(
						tgbotapi.CallbackConfig{
							CallbackQueryID: update.CallbackQuery.ID,
							Text:            loader,
						},
					)
					continue
				}
				// fmt.Printf("Tanggal Mulai Asli: %s", tempData.DateStart)
				// fmt.Printf("Tanggal Mulai Tanggal: %s", dateTempMulai[0])
				// fmt.Printf("Tanggal Mulai: %s", ns)
				// fmt.Printf("Tanggal Dipilih: %s", srt)
				//END VALIDASI

				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)

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
				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)
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
				sd, _ := time.Parse(LayoutISO, tempData.DateStart)
				de, _ := time.Parse(LayoutISO, tempData.DateEnd)
				endDate := de.Format(LayoutSQL) + " " + resData[1] + ":" + resData[2] + ":00"
				toBeQuery := models.Reservation{ChatID: update.CallbackQuery.Message.Chat.ID, DateEnd: endDate}
				// UpdateEndDate temporary
				_, err := revMod.UpdateEndDate(&toBeQuery)

				//VALIDASI
				nsNow, _ := time.Parse(LayoutSQL, de.Format(LayoutSQL))
				srtNow, _ := time.Parse(LayoutSQL, nowRaw.Format(LayoutSQL))
				sdDate, _ := time.Parse(LayoutSQL, sd.Format(LayoutSQL))
				deDate, _ := time.Parse(LayoutSQL, de.Format(LayoutSQL))

				if sameDate(srtNow, nsNow) || sameDate(sdDate, deDate) {
					ns, _ := time.Parse(LayoutTime, sd.Format(LayoutTime))
					srt, _ := time.Parse(LayoutTime, resData[1]+":"+resData[2])

					fmt.Printf("CheckJamMulai: Jam Minimum: %s", ns)
					fmt.Printf("CheckJamMulai: Jam Dipilih: %s", srt)

					//JIKA JAM DIPILIH(NS) LEBIH KECIL JAM SEKARANG (SRT)
					start := ns.UTC()
					check := srt.UTC()
					if check.Before(start) || check.Equal(start) {
						loader = fmt.Sprintf("Mohon pilih Lebih dari jam : %s", ns.Format(LayoutTime))
						bot.AnswerCallbackQuery(
							tgbotapi.CallbackConfig{
								CallbackQueryID: update.CallbackQuery.ID,
								Text:            loader,
							},
						)
						continue
					}
				}

				tempData = revMod.GetTemp(update.CallbackQuery.Message.Chat.ID)

				//VALIDASI PENGGUNAAN RUANGAN
				toBeSearch := models.ReservationData{
					EmployeeID: employees.Esid,
					RoomID:     tempData.RoomID,
					StartDate:  tempData.DateStart,
					EndDate:    tempData.DateEnd,
				}
				//Validate jika employee memesan pada waktu yang sama
				validateEmployee := revData.CheckWaktuEmployee(&toBeSearch)
				if !validateEmployee {
					loader = "Anda sudah ada jadwal di waktu yang sama."
					bot.AnswerCallbackQuery(
						tgbotapi.CallbackConfig{
							CallbackQueryID: update.CallbackQuery.ID,
							Text:            loader,
						},
					)

					list = fmt.Sprintln("Anda sudah ada jadwal pemakaian ruangan meeting/fun pada jam yang sama. Silahkan pilih waktu lainnya atau batalkan reservasi yang sudah ada.")

					var msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						list,
					)
					bot.Send(msg)

					//DeleteTemporary data
					revMod.DeleteTemp(update.CallbackQuery.Message.Chat.ID)
					continue
				}
				//End Validate

				//Validate jika ruangan dipesan pada waktu yang sama
				validateRoom := revData.CheckWaktuRuangan(&toBeSearch)
				if !validateRoom {
					loader = "Ruangan tidak tersedia di waktu yang dipilih."
					bot.AnswerCallbackQuery(
						tgbotapi.CallbackConfig{
							CallbackQueryID: update.CallbackQuery.ID,
							Text:            loader,
						},
					)

					list = fmt.Sprintln("Ruangan yang Anda pesan tidak tersedia di waktu yang Anda pilih.")

					var msg = tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						list,
					)
					bot.Send(msg)

					//DeleteTemporary data
					revMod.DeleteTemp(update.CallbackQuery.Message.Chat.ID)
					continue
				}
				//End Validate

				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)

				if err != nil {
					log.Fatal(err)
				}

				toBeInserted := models.ReservationData{
					EmployeeID: employees.Esid,
					RoomID:     tempData.RoomID,
					StartDate:  tempData.DateStart,
					EndDate:    tempData.DateEnd,
				}

				_, err = revData.CreateData(&toBeInserted)
				//DeleteTemporary data
				revMod.DeleteTemp(update.CallbackQuery.Message.Chat.ID)

				if err != nil {
					log.Fatal(err)
				}

				startD, _ := time.Parse(LayoutISO, toBeInserted.StartDate)
				endD, _ := time.Parse(LayoutISO, toBeInserted.EndDate)

				list = fmt.Sprintf("Jadwal pemakaian ruangan meeting/fun room untuk tanggal <b>%s</b> sampai <b>%s</b> sudah berhasil diset. Terima kasih", startD.Format(LayoutFullUser), endD.Format(LayoutFullUser))

				var msg = tgbotapi.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					list,
				)
				msg.ParseMode = "HTML"
				bot.Send(msg)
				continue
			case strings.Contains(data, "cancel_reservation"):
				loader = "Processing..."
				bot.AnswerCallbackQuery(
					tgbotapi.CallbackConfig{
						CallbackQueryID: update.CallbackQuery.ID,
						Text:            loader,
					},
				)
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

			validateEmp := model.CheckEmployee(update.Message.From.UserName)

			if !validateEmp {
				fnm, lnm, usn := update.Message.From.FirstName, update.Message.From.LastName, update.Message.From.UserName
				msg := fmt.Sprintf("Halo %s %s, username Anda: @%s belum terdaftar di sistem karyawan SPE. Silahkan hubungi tim HC untuk mendaftarkan. Terima kasih", fnm, lnm, usn)
				reply := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
				bot.Send(reply)
				continue
			}

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

//validDate untuk mengecek tanggalnya sekarang atau masa depan
func validDate(start, check time.Time) bool {
	start = start.UTC()
	check = check.UTC()
	// if check.Equal(start) {
	// 	ns, _ := time.Parse(LayoutTime, start.Format(LayoutTime))
	// 	srt, _ := time.Parse(LayoutTime, check.Format(LayoutTime))
	// 	validDate(ns, srt)
	// }
	return check.Before(start) || check.Equal(start)
}

func sameDate(start, check time.Time) bool {
	start = start.UTC()
	check = check.UTC()
	// if check.Equal(start) {
	// 	ns, _ := time.Parse(LayoutTime, start.Format(LayoutTime))
	// 	srt, _ := time.Parse(LayoutTime, check.Format(LayoutTime))
	// 	validDate(ns, srt)
	// }
	return check.Equal(start)
}

//validDate untuk mengecek tanggalnya sekarang atau masa depan
func validHour(start, check time.Time) bool {
	start = start.UTC()
	check = check.UTC()
	// if check.Equal(start) {
	// 	ns, _ := time.Parse(LayoutTime, start.Format(LayoutTime))
	// 	srt, _ := time.Parse(LayoutTime, check.Format(LayoutTime))
	// 	validDate(ns, srt)
	// }
	return check.After(start)
}

//RunPenggunaRuanganByGroup untuk mengirimkan notifikasi pengguna ruangan
func RunPenggunaRuanganByGroup() {
	model := models.Employee{}
	roomsMod := models.Room{}

	checkSchedule := model.CheckToday()
	now := time.Now().Format(LayoutTanggal)
	nowRaw := time.Now()

	var list string
	if !checkSchedule {
		list += "Pemakaian ruang meeting dan fun room " + now + " masih kosong"
	} else {
		room := roomsMod.MeetingRoom("yogyakarta")
		list += "<b>Pemakaian ruang meeting dan fun room " + now + ":</b> \n\n"
		for _, rm := range room {
			list += "<b>" + rm.Name + "</b> :  \n"
			// fmt.Println("INI LHO: " + strconv.Itoa(int(rm.ID)))
			empl := model.CheckPenggunaRuangan(rm.ID)

			for _, emp := range empl {
				sd, _ := time.Parse(LayoutISO, emp.StartDate)
				ed, _ := time.Parse(LayoutISO, emp.EndDate)

				sdSq, _ := time.Parse(LayoutSQL, sd.Format(LayoutSQL))
				edSq, _ := time.Parse(LayoutSQL, ed.Format(LayoutSQL))
				if !sameDate(edSq, sdSq) {
					sdNow, _ := time.Parse(LayoutSQL, nowRaw.Format(LayoutSQL))
					if !sameDate(edSq, sdNow) {
						ed, _ = time.Parse(LayoutTime, tp.JamTerakhir+":"+tp.MenitTerakhir)
					}
				}
				fmt.Printf("Tanggal sd: %s", sdSq)
				fmt.Printf("Tanggal ed: %s", edSq)

				list += emp.Telegram + " :  " + sd.Format(LayoutTime) + " - " + ed.Format(LayoutTime) + "\n"
			}
			list += "\n"

		}
		list += "Bagi SPEcial team yang akan ijin menggunakan ruang meeting dan fun room, silahkan di infokan dgn tim HC-GA ya. \n\n"
		list += "Thank you"

	}
	chatID, _ := strconv.ParseInt(goDotEnvVariable("BOT_CHATID"), 10, 64)
	msg := tgbotapi.NewMessage(chatID, list)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

//RunPenggunaRuanganEmployee untuk mengirimkan notifikasi pengguna ruangan ke semua user
func RunPenggunaRuanganEmployee() {
	model := models.Employee{}
	roomsMod := models.Room{}

	checkSchedule := model.CheckToday()
	now := time.Now().Format(LayoutTanggal)
	nowRaw := time.Now()
	employees := model.GetEmployees()

	for _, employee := range employees {
		var list string
		if !checkSchedule {
			list += "Pemakaian ruang meeting dan fun room " + now + " masih kosong"
		} else {
			room := roomsMod.MeetingRoom("yogyakarta")
			list += "<b>Pemakaian ruang meeting dan fun room " + now + ":</b> \n\n"
			for _, rm := range room {
				list += "<b>" + rm.Name + "</b> :  \n"
				// fmt.Println("INI LHO: " + strconv.Itoa(int(rm.ID)))
				empl := model.CheckPenggunaRuangan(rm.ID)

				for _, emp := range empl {
					sd, _ := time.Parse(LayoutISO, emp.StartDate)
					ed, _ := time.Parse(LayoutISO, emp.EndDate)

					sdSq, _ := time.Parse(LayoutSQL, sd.Format(LayoutSQL))
					edSq, _ := time.Parse(LayoutSQL, ed.Format(LayoutSQL))
					if !sameDate(edSq, sdSq) {
						sdNow, _ := time.Parse(LayoutSQL, nowRaw.Format(LayoutSQL))
						if !sameDate(edSq, sdNow) {
							ed, _ = time.Parse(LayoutTime, tp.JamTerakhir+":"+tp.MenitTerakhir)
						}
					}
					fmt.Printf("Tanggal sd: %s", sdSq)
					fmt.Printf("Tanggal ed: %s", edSq)

					list += emp.Telegram + " :  " + sd.Format(LayoutTime) + " - " + ed.Format(LayoutTime) + "\n"
				}
				list += "\n"

			}
			list += "Bagi SPEcial team yang akan ijin menggunakan ruang meeting dan fun room, silahkan di infokan dgn tim HC-GA ya. \n\n"
			list += "Thank you"

		}
		chatID, _ := strconv.ParseInt(employee.TelegramChatID, 10, 64)
		msg := tgbotapi.NewMessage(chatID, list)
		msg.ParseMode = "HTML"
		bot.Send(msg)
	}
}
