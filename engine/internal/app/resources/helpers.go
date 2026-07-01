package resources

import (
	"encoding/json"
	"regexp"
	"time"
)

var secretPattern = regexp.MustCompile(`(?i)(sk-[a-z0-9._+=:/-]+|bearer\s+[a-z0-9._+=:/-]+|api[_-]?key\s*[:=]\s*[\S]+|token\s*[:=]\s*[\S]+)`)

func latencyMS(startedAt time.Time) int {
	return int(time.Since(startedAt).Milliseconds())
}

func LatencyMS(startedAt time.Time) int {
	return latencyMS(startedAt)
}

func redactSecrets(value string) string {
	return secretPattern.ReplaceAllString(value, "[REDACTED]")
}

func RedactSecrets(value string) string {
	return redactSecrets(value)
}

func decodeToolArguments(arguments string) map[string]any {
	if arguments == "" {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(arguments), &decoded); err != nil {
		return map[string]any{"raw": arguments}
	}
	return decoded
}

func DecodeToolArguments(arguments string) map[string]any {
	return decodeToolArguments(arguments)
}
