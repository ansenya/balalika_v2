package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

var url = os.Getenv("OLLAMA_URL")

func RetrieveAnswer(messages *[]Message, model string) (*Response, error) {
	request := Request{
		Model:    "deepseek-llm:7b",
		Messages: *messages,
		Stream:   false,
		Options:  Options{Temperature: 0.6},
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url+"/chat", "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	*messages = append(*messages, Message{
		Role:    "assistant",
		Content: response.Message.Content,
	})
	return &response, nil
}
