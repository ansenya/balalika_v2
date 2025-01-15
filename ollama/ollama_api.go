package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

var url = os.Getenv("OLLAMA_URL")

func RetrieveAnswer(messages *[]Message) (*Response, error) {
	request := Request{
		Model:    "deepseek-llm:7b",
		Messages: *messages,
		Stream:   false,
		Options:  Options{Temperature: 1.0},
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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
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
