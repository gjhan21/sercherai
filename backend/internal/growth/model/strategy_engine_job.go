package model

type StrategyEngineJobResult struct {
	Summary     string         `json:"summary"`
	PayloadEcho map[string]any `json:"payload_echo"`
	Artifacts   map[string]any `json:"artifacts"`
	Warnings    []string       `json:"warnings"`
}

type StrategyEngineJobRecord struct {
	JobID                string                        `json:"job_id"`
	JobType              string                        `json:"job_type"`
	Status               string                        `json:"status"`
	RequestedBy          string                        `json:"requested_by"`
	TraceID              string                        `json:"trace_id"`
	TradeDate            string                        `json:"trade_date,omitempty"`
	Payload              map[string]any                `json:"payload"`
	Result               *StrategyEngineJobResult      `json:"result,omitempty"`
	ResultSummary        string                        `json:"result_summary,omitempty"`
	SelectedCount        int                           `json:"selected_count"`
	PayloadCount         int                           `json:"payload_count"`
	WarningCount         int                           `json:"warning_count"`
	PublishCount         int                           `json:"publish_count"`
	LatestPublishID      string                        `json:"latest_publish_id,omitempty"`
	LatestPublishVersion int                           `json:"latest_publish_version,omitempty"`
	LatestPublishAt      string                        `json:"latest_publish_at,omitempty"`
	LatestPublishMode    string                        `json:"latest_publish_mode,omitempty"`
	LatestPublishSource  string                        `json:"latest_publish_source,omitempty"`
	Replays              []StrategyEnginePublishReplay `json:"replays,omitempty"`
	ErrorMessage         string                        `json:"error_message,omitempty"`
	CreatedAt            string                        `json:"created_at"`
	StartedAt            string                        `json:"started_at,omitempty"`
	FinishedAt           string                        `json:"finished_at,omitempty"`
	StorageSource        string                        `json:"storage_source,omitempty"`
	SyncedAt             string                        `json:"synced_at,omitempty"`
}

func (r StrategyEngineJobRecord) ResultPayloadEcho() map[string]any {
	if r.Result == nil {
		return nil
	}
	return r.Result.PayloadEcho
}
