package domain

import "strings"

func NormalizeAccountStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active", "approved":
		return "active"
	case "suspended", "inactive", "disabled":
		return "suspended"
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func IsActiveAccount(status string) bool { return NormalizeAccountStatus(status) == "active" }
