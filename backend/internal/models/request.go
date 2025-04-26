package models

// DBRequest holds connection parameters for a Postgres database.
type DBRequest struct {
	Host             string `json:"host"`
	Port             string `json:"port"`
	User             string `json:"user"`
	Pass             string `json:"pass"`
	DBName           string `json:"dbname"`
	Provider         string `json:"provider"`
	SSLMode          string `json:"sslmode"`
	ConnectionString string `json:"connectionString"`
}

// NLQueryRequest is the JSON body for a natural-language→SQL request.
type NLQueryRequest struct {
	Config       DBRequest `json:"config"`
	Prompt       string    `json:"prompt"`
	Confirmed    bool      `json:"confirmed"`
	SQLToConfirm string    `json:"sqlToConfirm"`
	SessionID    string    `json:"sessionId"`
}
