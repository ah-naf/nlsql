package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"nlsql/config"
	"nlsql/internal/models"
)

const (
	ApiURL    = "https://api.together.xyz/v1/chat/completions"
	ModelName = "meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8"
)

// Connect sends a chat completion request to the LLM and returns its reply.
func Connect(messages []models.Message) (string, error) {
	config.LoadEnv()
	token := os.Getenv("LLM_API_KEY")
	if token == "" {
		return "", fmt.Errorf("LLM_API_KEY not set")
	}

	payload := struct {
		Model    string           `json:"model"`
		Messages []models.Message `json:"messages"`
	}{
		Model:    ModelName,
		Messages: messages,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, _ := http.NewRequest("POST", ApiURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", b)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}
	return out.Choices[0].Message.Content, nil
}
