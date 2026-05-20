package redact

import "strings"

func Env(env map[string]string, patterns []string) map[string]string {
	out := make(map[string]string, len(env))
	for key, value := range env {
		redacted := false
		lower := strings.ToLower(key)
		for _, pattern := range patterns {
			if strings.Contains(lower, strings.ToLower(pattern)) {
				out[key] = "[REDACTED]"
				redacted = true
				break
			}
		}
		if !redacted {
			out[key] = value
		}
	}
	return out
}
