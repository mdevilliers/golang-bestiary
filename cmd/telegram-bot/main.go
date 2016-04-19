package main

import (
	"log"
	"os"

	"gopkg.in/telegram-bot-api.v4"
)

func main() {

	telegram_key := os.Getenv("TELEGRAM_KEY")

	if len(telegram_key) == 0 {
		panic("TELEGRAM_KEY env variable not set!")
	}

	bot, err := tgbotapi.NewBotAPI(telegram_key)
	if err != nil {
		log.Panic(err)

	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)

	}

}
