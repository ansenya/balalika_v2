package db

import (
	"dolphin_bot/ollama"
)

type ChatHistory struct {
	ChatId   int64            `bson:"chat_id"`
	Messages []ollama.Message `bson:"messages"`
}

type Prompt struct {
	Prompt string `bson:"prompt"`
}
