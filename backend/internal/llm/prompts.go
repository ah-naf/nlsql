package llm

import (
	"fmt"
	"strings"

	"nlsql/internal/models"
)

// BuildTableDetectionPrompt asks the LLM to pick relevant tables.
func BuildTableDetectionPrompt(tables []string, query string) string {
	return fmt.Sprintf(`
You are a database schema assistant.

You are given a list of table names:

%s

Based on the user's request, return only the **relevant table names** from the list above.
If the user wants to create a new table or perform operations not related to specific existing tables, return "!!".
Do not include any descriptions or explanations. Only output the table names as a comma-separated list.

### Request
%s

### Output (comma-separated table names only)
`, strings.Join(tables, ", "), query)
}

// BuildSQLPromptWithHistory generates the SQL prompt including schema & history.
func BuildSQLPromptWithHistory(schema map[string]models.TableInfo, userQuery, historyContext string) string {
	var sb strings.Builder
	for tbl, info := range schema {
		sb.WriteString(fmt.Sprintf("Table: %s\n", tbl))
		for _, col := range info.Columns {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", col.Name, col.DataType))
		}
		sb.WriteString("\n")
	}
	return fmt.Sprintf(`
You are a SQL expert. Given the schema, conversation history, and user request, generate a valid SQL query.
Only return the SQL. Do not explain anything. Do not format the sql. Give it in raw text.

For CREATE TABLE operations, use appropriate data types and constraints.

### Schema
%s

### Conversation History
%s

### Request
%s

### SQL
`, sb.String(), historyContext, userQuery)
}
