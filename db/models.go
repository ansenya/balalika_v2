package db

import (
	"dolphin_bot/llm"
)

type ChatHistory struct {
	ChatId   int64         `bson:"chat_id"`
	Messages []llm.Message `bson:"messages"`
}

type Prompt struct {
	Prompt string `bson:"prompt"`
}

type Model struct {
	Model string `bson:"model"`
}
