package models

import "time"

// HistoryItem is one exchange in an NL→SQL conversation.
type HistoryItem struct {
	Prompt   string    `json:"prompt"`
	SQL      string    `json:"sql"`
	Response string    `json:"response"`
	Time     time.Time `json:"time"`
}

// ConversationHistory holds the last MAX items and a timestamp.
type ConversationHistory struct {
	Items    []HistoryItem `json:"items"`
	ClientIP string        `json:"client_ip"`
	LastUsed time.Time     `json:"last_used"`
}

// Message is one chat message for the LLM.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}