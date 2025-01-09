package main

import (
	"dolphin_bot/db"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"sync"
)

var chatLocks sync.Map

func main() {
	// "mongodb://localhost:27017"
	err := db.ConnectMongoDB(os.Getenv("MONGO_URL"))
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go func() {
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

				//chatID := update.Message.Chat.ID
				//lock := getChatLock(chatID)
				//if !lock.TryLock() {
				//	if update.Message.Chat.Type != "private" &&
				//		(update.Message.ReplyToMessage == nil ||
				//			update.Message.ReplyToMessage.From.UserName != bot.Self.UserName) {
				//		return
				//	}
				//	msg := tgbotapi.NewMessage(chatID, "**Only one request for a chat at a time is allowed.**")
				//	msg.ParseMode = "markdown"
				//	sent, err := bot.Send(msg)
				//	if err != nil {
				//		log.Printf("Error sending 'wait' message: %v", err)
				//	}
				//	sendActionTyping(bot, &update)
				//	go func() {
				//		time.Sleep(5 * time.Second)
				//		messagesToDelete := []tgbotapi.DeleteMessageConfig{
				//			tgbotapi.NewDeleteMessage(chatID, sent.MessageID),
				//			tgbotapi.NewDeleteMessage(chatID, update.Message.MessageID),
				//		}
				//		for _, message := range messagesToDelete {
				//			go func() {
				//				_, err := bot.Send(message)
				//				if err != nil {
				//					log.Printf("Error deleting 'wait' message: %v", err)
				//				}
				//			}()
				//		}
				//	}()
				//	return
				//}
				//
				//defer lock.Unlock()

				if update.Message.IsCommand() {
					err := handleCommands(bot, &update)
					if err != nil {
						log.Printf("Error while handling command: %v", err)
						_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("I cannot perform this command(\n%v", err)))
						if err != nil {
							log.Printf("Error while sending message: %v", err)
						}
					}
				} else {
					err := handleTextMessage(bot, &update)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
						if errors.Is(err, http.ErrAbortHandler) {
							msg.Text = "Backend is not responsive, cannot get answer("
						} else {
							msg.Text = fmt.Sprintf("Some error occured( %v", err)
						}
						_, _ = bot.Send(msg)
						log.Printf("Error while retrieving answer: %v", err)
					}
				}
			}()
		}
	}
}

func getChatLock(chatID int64) *sync.Mutex {
	lock, ok := chatLocks.Load(chatID)
	if !ok {
		lock = &sync.Mutex{}
		chatLocks.Store(chatID, lock)
	}
	return lock.(*sync.Mutex)
}

func sendActionTyping(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	_, _ = bot.Send(tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping))
}
