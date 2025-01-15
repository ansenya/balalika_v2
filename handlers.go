package main

import (
	"dolphin_bot/db"
	"dolphin_bot/ollama"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

const MaxMessageLength = 4096

func handleCommands(bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	command := update.Message.Command()
	switch command {
	case "start":
		msg.Text = "Hey! I'm bot created by @hipahopa.\n\n" +
			"I can say a lot of stupid things. That's why I was created.\n\n" +
			"Don't expect me to be helpful."
	case "clear":
		err := db.ClearHistory(update.Message.Chat.ID)
		if err != nil {
			return err
		}
		msg.Text = "Successfully cleared history"
	case "prompt":
		prompt := update.Message.CommandArguments()
		if prompt == "" {
			prompt, err := db.GetPrompt(update.Message.Chat.ID)
			if err != nil {
				return err
			}
			msg.Text = fmt.Sprintf("Right now your prompt is:\n%s\n\nIf you want to change it, type /prompt <your prompt>", prompt)
		} else {
			err := db.ClearHistory(update.Message.Chat.ID)
			if err != nil {
				return err
			}
			if err := db.SavePrompt(update.Message.Chat.ID, prompt); err != nil {
				return err
			}
			msg.Text = "Updated prompt (and cleared history)."
		}
	case "hey":
		msg.Text = "I'm here"
		msg.ReplyToMessageID = update.Message.MessageID
	default:
		msg.Text = "I don't know that command"
	}

	_, err := bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func handleTextMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
	if update.Message.Chat.Type != "private" &&
		(update.Message.ReplyToMessage == nil ||
			update.Message.ReplyToMessage.From.UserName != bot.Self.UserName) {
		return nil
	}
	done := make(chan bool)
	defer close(done)

	go func() {
		sendActionTyping(bot, update)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				sendActionTyping(bot, update)
			}
		}
	}()
	messages, err := db.GetHistory(update.Message.Chat.ID)
	if err != nil {
		return err
	}
	messages = append(messages, ollama.Message{
		Role:    "user",
		Content: fmt.Sprintf("%s: %s", update.Message.From.FirstName, update.Message.Text),
	})

	response, err := ollama.RetrieveAnswer(&messages)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, truncateMessage(response.Message.Content))

	if update.Message.Chat.Type != "private" {
		msg.ReplyToMessageID = update.Message.MessageID
	}

	_, err = bot.Send(msg)
	if err != nil {
		return err
	}

	err = db.SaveMessage(update.Message.Chat.ID, &messages[len(messages)-2])
	if err != nil {
		return err
	}
	err = db.SaveMessage(update.Message.Chat.ID, &messages[len(messages)-1])
	if err != nil {
		return err
	}

	return nil
}

func truncateMessage(text string) string {
	if len(text) > MaxMessageLength {
		return text[:MaxMessageLength-3] + "..."
	}
	return text
}
