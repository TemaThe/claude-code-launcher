package redact

import "testing"

func TestEnvRedactionByPattern(t *testing.T) {
	got := Env(map[string]string{
		"OPENAI_API_KEY": "secret",
		"HOME":           "/tmp/home",
	}, []string{"key", "token"})
	if got["OPENAI_API_KEY"] != "[REDACTED]" {
		t.Fatalf("expected API key redaction")
	}
	if got["HOME"] != "/tmp/home" {
		t.Fatalf("HOME should not be redacted")
	}
}
