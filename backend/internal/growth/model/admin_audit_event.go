package model

type AdminAuditEvent struct {
	ID          string         `json:"id"`
	EventDomain string         `json:"event_domain"`
	EventType   string         `json:"event_type"`
	Level       string         `json:"level"`
	Module      string         `json:"module,omitempty"`
	ObjectType  string         `json:"object_type,omitempty"`
	ObjectID    string         `json:"object_id,omitempty"`
	ActorUserID string         `json:"actor_user_id,omitempty"`
	Title       string         `json:"title"`
	Summary     string         `json:"summary,omitempty"`
	Detail      string         `json:"detail,omitempty"`
	Status      string         `json:"status,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	DedupeKey   string         `json:"dedupe_key,omitempty"`
	CreatedAt   string         `json:"created_at,omitempty"`
}

type AdminAuditEventFilter struct {
	EventDomain string `json:"event_domain,omitempty"`
	EventType   string `json:"event_type,omitempty"`
	Level       string `json:"level,omitempty"`
	Module      string `json:"module,omitempty"`
	ObjectType  string `json:"object_type,omitempty"`
	ObjectID    string `json:"object_id,omitempty"`
	ActorUserID string `json:"actor_user_id,omitempty"`
	Status      string `json:"status,omitempty"`
}

type AdminAuditEventSummary struct {
	TotalCount   int            `json:"total_count"`
	OpenCount    int            `json:"open_count"`
	InfoCount    int            `json:"info_count"`
	WarningCount int            `json:"warning_count"`
	CriticalCount int           `json:"critical_count"`
	ByDomain     map[string]int `json:"by_domain"`
}
