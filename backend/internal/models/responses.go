package models

// DatabaseListResponse is returned by GET /databases.
type DatabaseListResponse struct {
	Databases []string `json:"databases"`
}

// NLQueryResponse is returned by POST /nlq.
type NLQueryResponse struct {
	SQL               string                   `json:"sql"`
	ResultTable       []map[string]interface{} `json:"result_table,omitempty"`
	SessionID         string                   `json:"session_id"`
	NeedsConfirmation bool                     `json:"needs_confirmation,omitempty"`
	SQLPreview        string                   `json:"sql_preview,omitempty"`
	Message           string                   `json:"message,omitempty"`
	SQLType           string                   `json:"sql_type,omitempty"`
	Affected          int64                    `json:"affected,omitempty"`
}

// BriefSchemaItem is one table with its row count.
type BriefSchemaItem struct {
	Name     string `json:"name"`
	RowCount int    `json:"row_count"`
}

// BriefSchemaResponse is returned by GET /schema?brief=true.
type BriefSchemaResponse struct {
	Tables []BriefSchemaItem `json:"tables"`
}

// SchemaResponse is returned by GET /schema.
type SchemaResponse struct {
	Schema map[string]TableInfo `json:"schema"`
}

// TableSchemaResponse is returned by GET /schema/:tableName.
type TableSchemaResponse struct {
	Table TableInfo `json:"table"`
}
