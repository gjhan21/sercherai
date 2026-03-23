package handler

import (
	"encoding/json"
	"strings"
)

type strategyPublishConflictPayload struct {
	Detail string `json:"detail"`
}

func extractStrategyPublishConflictDetail(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "", false
	}
	lower := strings.ToLower(message)
	if !strings.Contains(lower, "returned 409 when publishing job") && !strings.Contains(message, "发布策略拦截") {
		return "", false
	}
	if start := strings.Index(message, "{"); start >= 0 {
		var payload strategyPublishConflictPayload
		if json.Unmarshal([]byte(message[start:]), &payload) == nil && strings.TrimSpace(payload.Detail) != "" {
			return strings.TrimSpace(payload.Detail), true
		}
	}
	if marker := strings.LastIndex(message, ":"); marker >= 0 && marker+1 < len(message) {
		return strings.TrimSpace(message[marker+1:]), true
	}
	return message, true
}
