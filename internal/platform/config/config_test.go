package config

import "testing"

func TestLoadRejectsDevelopmentUserOutsideDevelopment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_DEV_USER_ID", "11111111-1111-4111-8111-111111111111")

	if _, err := Load(); err == nil {
		t.Fatal("expected production development identity to be rejected")
	}
}
