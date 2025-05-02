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

// Connect sends messages to Gemini API and returns the response.
func Connect(messages []models.Message) (string, error) {
	config.LoadEnv()

	apiURL := os.Getenv("GEMINI_API_URL")
	if apiURL == "" {
		return "", fmt.Errorf("GEMINI_API_URL not set")
	}

	// Convert messages into Gemini-compatible parts
	var parts []map[string]string
	for _, msg := range messages {
		parts = append(parts, map[string]string{"text": msg.Content})
	}

	// Create the payload
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": parts,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Send request
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error: %s", data)
	}

	// Decode the response
	var gemResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gemResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content returned from Gemini")
	}

	return gemResp.Candidates[0].Content.Parts[0].Text, nil
}

// Connect sends a chat completion request to the LLM and returns its reply.
// Connects to together ai api
// func Connect(messages []models.Message) (string, error) {
// 	config.LoadEnv()
// 	token := os.Getenv("LLM_API_KEY")
// 	if token == "" {
// 		return "", fmt.Errorf("LLM_API_KEY not set")
// 	}
// 	ApiURL := os.Getenv("LLM_API_URL")
// 	if ApiURL == "" {
// 		return "", fmt.Errorf("LLM_API_URL not set")
// 	}
// 	ModelName := os.Getenv("LLM_API_MODEL_NAME")
// 	if ModelName == "" {
// 		return "", fmt.Errorf("LLM_API_MODEL_NAME not set")
// 	}

// 	payload := struct {
// 		Model    string           `json:"model"`
// 		Messages []models.Message `json:"messages"`
// 	}{
// 		Model:    ModelName,
// 		Messages: messages,
// 	}

// 	data, err := json.Marshal(payload)
// 	if err != nil {
// 		return "", err
// 	}
// 	req, _ := http.NewRequest("POST", ApiURL, bytes.NewBuffer(data))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+token)

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		b, _ := io.ReadAll(resp.Body)
// 		return "", fmt.Errorf("API error: %s", b)
// 	}

// 	var out struct {
// 		Choices []struct {
// 			Message struct {
// 				Content string `json:"content"`
// 			} `json:"message"`
// 		} `json:"choices"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
// 		return "", err
// 	}
// 	if len(out.Choices) == 0 {
// 		return "", fmt.Errorf("no response from LLM")
// 	}
// 	return out.Choices[0].Message.Content, nil
// }
