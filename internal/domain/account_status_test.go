package domain

import "testing"

func TestAccountStatusCompatibility(t *testing.T) {
	for _, status := range []string{"active", "approved", " ACTIVE ", "APPROVED"} {
		if !IsActiveAccount(status) {
			t.Errorf("active account %q was rejected", status)
		}
	}
	for _, status := range []string{"suspended", "disabled", "inactive", "pending", "rejected", "", "unknown"} {
		if IsActiveAccount(status) {
			t.Errorf("inactive/unverified account %q was activated", status)
		}
	}
}
