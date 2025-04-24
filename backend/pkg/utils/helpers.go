package utils

import (
	"regexp"
	"strings"
)

var destructiveRE = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE)`)
var dbOperationRE = regexp.MustCompile(`(?i)\b(table|database|column|row|insert|update|delete|drop|alter|create|select|from|where|join|schema)\b`)

// NeedsConfirmation returns true if the SQL is a potentially destructive statement.
func NeedsConfirmation(sqlQ string) bool {
	return destructiveRE.MatchString(strings.TrimSpace(sqlQ))
}

// IsDBOperation returns true if the prompt likely refers to a DB operation.
func IsDBOperation(prompt string) bool {
	return dbOperationRE.MatchString(prompt)
}

// ParseCSV splits a comma-separated list into trimmed strings.
func ParseCSV(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".\n")
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// OperationType returns a human-friendly name for the SQL operation.
func OperationType(sqlQ string) string {
	up := strings.ToUpper(strings.TrimSpace(sqlQ))
	switch {
	case strings.HasPrefix(up, "INSERT"):
		return "INSERT operation"
	case strings.HasPrefix(up, "UPDATE"):
		return "UPDATE operation"
	case strings.HasPrefix(up, "DELETE"):
		return "DELETE operation"
	case strings.HasPrefix(up, "DROP"):
		return "DROP operation"
	case strings.HasPrefix(up, "ALTER"):
		return "ALTER operation"
	case strings.HasPrefix(up, "CREATE"):
		return "CREATE operation"
	}
	return "SQL operation"
}
