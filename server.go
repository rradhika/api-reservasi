package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rradhika/api-reservasi/handler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
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

	e := echo.New()
	e.GET("/", handler.Welcome())
	_ = e.Start(":3001")

	//Start the bot in a separate thread
	go StartBot()

	//To stop the program to finish and close
	select {}

	//You can also use https://golang.org/pkg/sync/#WaitGroup instead.
}

/*
func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		data := M{"Message": "Hello", "Counter": 2}
		return c.JSON(http.StatusOK, data)
	})

	e.GET("/test", func(c echo.Context) error {
		name := c.QueryParam("name")
		data := fmt.Sprintf("Hello %s", name)

		return c.String(http.StatusOK, data)
	})
	e.Logger.Fatal(e.Start(":3001"))
}*/

//StartBot to initiate bot /start
func StartBot() {
	bot, err := tgbotapi.NewBotAPI(goDotEnvVariable("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		fnm, lnm := update.Message.From.FirstName, update.Message.From.LastName
		msg := fmt.Sprintf("Welcome %s %s, are you gonna book meeting room?", fnm, lnm)
		gpMsg := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
		//gpMsg := tgbotapi.NewKeyboardButtonContact("Welcome Jack, are you gonna book meeting room?")
		bot.Send(gpMsg)
	}
}
