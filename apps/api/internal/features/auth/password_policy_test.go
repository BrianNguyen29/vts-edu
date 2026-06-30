package auth

import "testing"

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid mixed", "StrongPass1", false},
		{"valid with symbol", "AnotherP@ssw0rd", false},
		{"too short", "Short1", true},
		{"no uppercase", "strongpass1", true},
		{"no lowercase", "STRONGPASS1", true},
		{"no digit", "StrongPass", true},
		{"blocked exact", "Password123!", true},
		{"blocked lowercase", "password", true},
		{"blocked 12345678", "12345678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if tt.wantErr && err == nil {
				t.Errorf("ValidatePasswordStrength(%q) expected error", tt.password)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePasswordStrength(%q) unexpected error: %v", tt.password, err)
			}
		})
	}
}
