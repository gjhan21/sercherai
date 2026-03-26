package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

var allowedMarketBackfillAssetTypes = []string{"STOCK", "FUTURES", "INDEX", "ETF", "LOF", "CBOND"}

const marketBackfillLongHistoryChunkDays = 180

type marketBackfillBadRequestError struct {
	message string
}

func (e *marketBackfillBadRequestError) Error() string {
	return e.message
}

func (e *marketBackfillBadRequestError) BadRequest() bool {
	return true
}

type marketBackfillDateRange struct {
	FromText string
	ToText   string
	FromTime time.Time
	ToTime   time.Time
	Days     int
}

type marketBackfillLongHistoryOptions struct {
	Enabled   bool
	DateRange marketBackfillDateRange
	ChunkDays int
}

type marketBackfillDateChunk struct {
	FromText string
	ToText   string
	FromTime time.Time
	ToTime   time.Time
	Index    int
	Total    int
}

func newMarketBackfillBadRequestError(message string) error {
	return &marketBackfillBadRequestError{message: strings.TrimSpace(message)}
}

func normalizeMarketBackfillAssetType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "STOCK":
		return "STOCK"
	case "FUTURES":
		return "FUTURES"
	case "INDEX":
		return "INDEX"
	case "ETF":
		return "ETF"
	case "LOF":
		return "LOF"
	case "CBOND":
		return "CBOND"
	default:
		return ""
	}
}

func normalizeMarketBackfillAssetScope(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		normalized := normalizeMarketBackfillAssetType(value)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func normalizeMarketBackfillRunType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "INCREMENTAL":
		return "INCREMENTAL"
	case "REBUILD_ONLY":
		return "REBUILD_ONLY"
	default:
		return "FULL"
	}
}

func normalizeMarketBackfillStage(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "UNIVERSE":
		return "UNIVERSE"
	case "MASTER":
		return "MASTER"
	case "QUOTES":
		return "QUOTES"
	case "DAILY_BASIC":
		return "DAILY_BASIC"
	case "MONEYFLOW":
		return "MONEYFLOW"
	case "TRUTH":
		return "TRUTH"
	case "COVERAGE_SUMMARY":
		return "COVERAGE_SUMMARY"
	default:
		return ""
	}
}

func normalizeMarketBackfillStages(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		normalized := normalizeMarketBackfillStage(value)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func normalizeMarketBackfillRetryMode(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "FROM_STAGE":
		return "FROM_STAGE"
	default:
		return "FAILED_ONLY"
	}
}

func normalizeMarketBackfillRunStatus(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "PENDING":
		return "PENDING"
	case "PARTIAL_SUCCESS":
		return "PARTIAL_SUCCESS"
	case "SUCCESS":
		return "SUCCESS"
	case "FAILED":
		return "FAILED"
	case "CANCELLED":
		return "CANCELLED"
	default:
		return "RUNNING"
	}
}

func normalizeMarketBackfillDetailStatus(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "PENDING":
		return "PENDING"
	case "RUNNING":
		return "RUNNING"
	case "SUCCESS":
		return "SUCCESS"
	case "FAILED":
		return "FAILED"
	case "SKIPPED":
		return "SKIPPED"
	default:
		return "PENDING"
	}
}

func marshalJSONText(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(payload)
}

func unmarshalStringSlice(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func unmarshalStageProgress(raw string) []model.MarketBackfillStageProgress {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []model.MarketBackfillStageProgress
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func unmarshalSummaryMap(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var item map[string]any
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil
	}
	return item
}

func parseNullableTimeRFC3339(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.RFC3339)
}

func parseNullableString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func parseMarketBackfillDateInput(value string, fieldName string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, newMarketBackfillBadRequestError(fmt.Sprintf("%s 必须是 YYYY-MM-DD 格式", fieldName))
	}
	return parsed, nil
}

func parseMarketBackfillDateRange(tradeDateFrom string, tradeDateTo string) (marketBackfillDateRange, error) {
	fromText := strings.TrimSpace(tradeDateFrom)
	toText := strings.TrimSpace(tradeDateTo)
	fromTime, err := parseMarketBackfillDateInput(fromText, "trade_date_from")
	if err != nil {
		return marketBackfillDateRange{}, err
	}
	toTime, err := parseMarketBackfillDateInput(toText, "trade_date_to")
	if err != nil {
		return marketBackfillDateRange{}, err
	}
	if fromText == "" || toText == "" {
		return marketBackfillDateRange{
			FromText: fromText,
			ToText:   toText,
			FromTime: fromTime,
			ToTime:   toTime,
		}, nil
	}
	if toTime.Before(fromTime) {
		return marketBackfillDateRange{}, newMarketBackfillBadRequestError("trade_date_from 不能晚于 trade_date_to")
	}
	return marketBackfillDateRange{
		FromText: fromText,
		ToText:   toText,
		FromTime: fromTime,
		ToTime:   toTime,
		Days:     int(toTime.Sub(fromTime)/(24*time.Hour)) + 1,
	}, nil
}

func splitMarketBackfillDateChunks(dateRange marketBackfillDateRange, chunkDays int) []marketBackfillDateChunk {
	if dateRange.Days <= 0 {
		return nil
	}
	if chunkDays <= 0 {
		chunkDays = marketBackfillLongHistoryChunkDays
	}
	chunks := make([]marketBackfillDateChunk, 0, (dateRange.Days+chunkDays-1)/chunkDays)
	current := dateRange.FromTime
	for !current.After(dateRange.ToTime) {
		end := current.AddDate(0, 0, chunkDays-1)
		if end.After(dateRange.ToTime) {
			end = dateRange.ToTime
		}
		chunks = append(chunks, marketBackfillDateChunk{
			FromText: current.Format("2006-01-02"),
			ToText:   end.Format("2006-01-02"),
			FromTime: current,
			ToTime:   end,
		})
		current = end.AddDate(0, 0, 1)
	}
	total := len(chunks)
	for idx := range chunks {
		chunks[idx].Index = idx + 1
		chunks[idx].Total = total
	}
	return chunks
}

func resolveMarketBackfillLongHistoryOptions(runType string, assetScope []string, tradeDateFrom string, tradeDateTo string, stages []string) (marketBackfillLongHistoryOptions, error) {
	dateRange, err := parseMarketBackfillDateRange(tradeDateFrom, tradeDateTo)
	if err != nil {
		return marketBackfillLongHistoryOptions{}, err
	}
	options := marketBackfillLongHistoryOptions{
		DateRange: dateRange,
		ChunkDays: marketBackfillLongHistoryChunkDays,
	}
	if dateRange.Days <= 365 {
		return options, nil
	}
	if len(assetScope) != 1 || assetScope[0] != "STOCK" {
		return marketBackfillLongHistoryOptions{}, newMarketBackfillBadRequestError("超过 365 天的长历史回补目前只支持单一 STOCK 资产范围")
	}
	if normalizeMarketBackfillRunType(runType) != "FULL" {
		return marketBackfillLongHistoryOptions{}, newMarketBackfillBadRequestError("超过 365 天的股票长历史回补当前只允许 FULL 运行类型")
	}
	normalizedStages := normalizeMarketBackfillStages(stages)
	if len(normalizedStages) > 0 {
		hasQuotes := false
		for _, stage := range normalizedStages {
			if stage == "QUOTES" {
				hasQuotes = true
				break
			}
		}
		if !hasQuotes {
			return marketBackfillLongHistoryOptions{}, newMarketBackfillBadRequestError("长历史股票回补的阶段范围必须包含 QUOTES")
		}
	}
	options.Enabled = true
	return options, nil
}

func resolveMarketBackfillLongHistoryOptionsFromRun(run model.MarketBackfillRun) marketBackfillLongHistoryOptions {
	options, _ := resolveMarketBackfillLongHistoryOptions(run.RunType, normalizeMarketBackfillAssetScope(run.AssetScope), run.TradeDateFrom, run.TradeDateTo, nil)
	if summaryChunkDays := marketBackfillSummaryInt(run.Summary, "long_history_chunk_days", 0); summaryChunkDays > 0 {
		options.ChunkDays = summaryChunkDays
	}
	if run.Summary != nil {
		if enabled, ok := run.Summary["long_history_mode"].(bool); ok {
			options.Enabled = enabled
		}
	}
	return options
}

func marketBackfillLongHistorySourceProviderError() error {
	return newMarketBackfillBadRequestError("超过 365 天的股票长历史回补当前仅支持 provider 为 TUSHARE 的数据源")
}

func isMarketBackfillTushareProvider(item model.DataSource) bool {
	provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
	if provider == "" {
		provider = strings.ToUpper(strings.TrimSpace(item.SourceKey))
	}
	return provider == "TUSHARE"
}

func (r *MySQLGrowthRepo) validateMarketBackfillLongHistorySourceKey(sourceKey string, longHistory marketBackfillLongHistoryOptions) error {
	if !longHistory.Enabled {
		return nil
	}
	resolvedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if resolvedSourceKey == "" {
		resolvedSourceKey = "TUSHARE"
	}
	sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey)
	if err != nil {
		return marketBackfillLongHistorySourceProviderError()
	}
	if !isMarketBackfillTushareProvider(sourceItem) {
		return marketBackfillLongHistorySourceProviderError()
	}
	return nil
}

func (r *InMemoryGrowthRepo) validateMarketBackfillLongHistorySourceKey(sourceKey string, longHistory marketBackfillLongHistoryOptions) error {
	if !longHistory.Enabled {
		return nil
	}
	resolvedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if resolvedSourceKey == "" {
		resolvedSourceKey = "TUSHARE"
	}
	items, _, err := r.AdminListDataSources(1, 100)
	if err != nil {
		return err
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.SourceKey), resolvedSourceKey) {
			if isMarketBackfillTushareProvider(item) {
				return nil
			}
			return marketBackfillLongHistorySourceProviderError()
		}
	}
	return marketBackfillLongHistorySourceProviderError()
}

func marketBackfillSummaryInt(summary map[string]any, key string, fallback int) int {
	if summary == nil {
		return fallback
	}
	switch value := summary[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return fallback
	}
}

func buildMarketBackfillSummary(input model.MarketBackfillCreateInput, assetScope []string, snapshotItems []model.MarketUniverseSnapshotItem, longHistory marketBackfillLongHistoryOptions) map[string]any {
	summary := map[string]any{
		"requested_stages":         normalizeMarketBackfillStages(input.Stages),
		"asset_scope":              assetScope,
		"universe_item_count":      len(snapshotItems),
		"rebuild_truth_after_sync": input.RebuildTruthAfterSync,
		"long_history_mode":        longHistory.Enabled,
	}
	if longHistory.DateRange.FromText != "" {
		summary["requested_trade_date_from"] = longHistory.DateRange.FromText
	}
	if longHistory.DateRange.ToText != "" {
		summary["requested_trade_date_to"] = longHistory.DateRange.ToText
	}
	if longHistory.Enabled {
		chunks := splitMarketBackfillDateChunks(longHistory.DateRange, longHistory.ChunkDays)
		summary["long_history_chunk_days"] = longHistory.ChunkDays
		summary["long_history_chunk_count"] = len(chunks)
	}
	return summary
}

func shouldSkipTruthRebuildForLongHistory(run model.MarketBackfillRun) bool {
	longHistory := resolveMarketBackfillLongHistoryOptionsFromRun(run)
	if !longHistory.Enabled || run.Summary == nil {
		return false
	}
	value, ok := run.Summary["rebuild_truth_after_sync"]
	if !ok {
		return false
	}
	flag, ok := value.(bool)
	return ok && !flag
}

func defaultMarketBackfillStageProgress() []model.MarketBackfillStageProgress {
	return []model.MarketBackfillStageProgress{
		{Stage: "UNIVERSE", Status: "RUNNING", TotalBatches: 1},
		{Stage: "MASTER", Status: "PENDING"},
		{Stage: "QUOTES", Status: "PENDING"},
		{Stage: "DAILY_BASIC", Status: "PENDING"},
		{Stage: "MONEYFLOW", Status: "PENDING"},
		{Stage: "TRUTH", Status: "PENDING"},
		{Stage: "COVERAGE_SUMMARY", Status: "PENDING"},
	}
}

func buildMarketBackfillStageProgressAfterUniverse(assetScope []string) []model.MarketBackfillStageProgress {
	progress := defaultMarketBackfillStageProgress()
	if len(progress) == 0 {
		return progress
	}
	totalBatches := len(normalizeMarketBackfillAssetScope(assetScope))
	if totalBatches <= 0 {
		totalBatches = 1
	}
	progress[0].Status = "SUCCESS"
	progress[0].TotalBatches = totalBatches
	progress[0].CompletedBatches = totalBatches
	return progress
}

func buildMarketUniverseRunDetails(runID string, schedulerRunID string, sourceKey string, items []model.MarketUniverseSnapshotItem, assetScope []string, now time.Time) []model.MarketBackfillRunDetail {
	orderedScope := normalizeMarketBackfillAssetScope(assetScope)
	byAsset := make(map[string][]model.MarketUniverseSnapshotItem, len(orderedScope))
	for _, item := range items {
		byAsset[item.AssetType] = append(byAsset[item.AssetType], item)
	}
	details := make([]model.MarketBackfillRunDetail, 0, len(orderedScope))
	startedAt := now.Format(time.RFC3339)
	for _, assetType := range orderedScope {
		assetItems := byAsset[assetType]
		sample := make([]string, 0, minInt(len(assetItems), 3))
		for idx := 0; idx < len(assetItems) && idx < 3; idx++ {
			sample = append(sample, assetItems[idx].InstrumentKey)
		}
		details = append(details, model.MarketBackfillRunDetail{
			ID:             newID("mbd"),
			RunID:          runID,
			SchedulerRunID: schedulerRunID,
			Stage:          "UNIVERSE",
			AssetType:      assetType,
			BatchKey:       fmt.Sprintf("UNIVERSE-%s-001", assetType),
			SourceKey:      strings.ToUpper(strings.TrimSpace(sourceKey)),
			SymbolCount:    len(assetItems),
			SymbolSample:   sample,
			Status:         "SUCCESS",
			FetchedCount:   len(assetItems),
			UpsertedCount:  len(assetItems),
			StartedAt:      startedAt,
			FinishedAt:     startedAt,
			CreatedAt:      startedAt,
			UpdatedAt:      startedAt,
		})
	}
	return details
}

func (r *MySQLGrowthRepo) AdminCreateMarketDataBackfillRun(input model.MarketBackfillCreateInput, operator string) (model.MarketBackfillRun, error) {
	assetScope := normalizeMarketBackfillAssetScope(input.AssetScope)
	if len(assetScope) == 0 {
		return model.MarketBackfillRun{}, fmt.Errorf("asset_scope is required")
	}
	runType := normalizeMarketBackfillRunType(input.RunType)
	batchSize := input.BatchSize
	if batchSize <= 0 {
		batchSize = 200
	}
	longHistory, err := resolveMarketBackfillLongHistoryOptions(runType, assetScope, input.TradeDateFrom, input.TradeDateTo, input.Stages)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}
	if err := r.validateMarketBackfillLongHistorySourceKey(input.SourceKey, longHistory); err != nil {
		return model.MarketBackfillRun{}, err
	}
	now := time.Now()
	snapshot, snapshotItems, err := r.AdminBuildMarketUniverseSnapshot(input.SourceKey, assetScope, operator)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}
	schedulerRunID, err := r.AdminCreateSchedulerJobRun("market_data_full_backfill", "MANUAL", "RUNNING", "market data backfill created", "", operator)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}

	runID := newID("mbr")
	stageProgress := buildMarketBackfillStageProgressAfterUniverse(assetScope)
	summary := buildMarketBackfillSummary(input, assetScope, snapshotItems, longHistory)
	if _, err := r.db.Exec(`
INSERT INTO market_backfill_runs (
	id, scheduler_run_id, run_type, asset_scope, trade_date_from, trade_date_to, source_key, batch_size,
	universe_snapshot_id, status, current_stage, stage_progress_json, summary_json, error_message,
	created_by, created_at, updated_at, finished_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		schedulerRunID,
		runType,
		marshalJSONText(assetScope),
		nullableString(strings.TrimSpace(input.TradeDateFrom)),
		nullableString(strings.TrimSpace(input.TradeDateTo)),
		truncateByRunes(snapshot.SourceKey, 64),
		batchSize,
		snapshot.ID,
		"RUNNING",
		"MASTER",
		marshalJSONText(stageProgress),
		marshalJSONText(summary),
		nil,
		truncateByRunes(normalizeUTF8Text(operator), 64),
		now,
		now,
		nil,
	); err != nil {
		return model.MarketBackfillRun{}, err
	}

	details := buildMarketUniverseRunDetails(runID, schedulerRunID, snapshot.SourceKey, snapshotItems, assetScope, now)
	for _, detail := range details {
		if _, err := r.db.Exec(`
INSERT INTO market_backfill_run_details (
	id, run_id, scheduler_run_id, stage, asset_type, batch_key, source_key, symbol_count, symbol_sample,
	trade_date_from, trade_date_to, status, fetched_count, upserted_count, truth_count, warning_text,
	error_text, started_at, finished_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			detail.ID,
			detail.RunID,
			nullableString(detail.SchedulerRunID),
			detail.Stage,
			nullableString(detail.AssetType),
			nullableString(detail.BatchKey),
			nullableString(detail.SourceKey),
			detail.SymbolCount,
			nullableString(marshalJSONText(detail.SymbolSample)),
			nullableString(strings.TrimSpace(detail.TradeDateFrom)),
			nullableString(strings.TrimSpace(detail.TradeDateTo)),
			detail.Status,
			detail.FetchedCount,
			detail.UpsertedCount,
			detail.TruthCount,
			nullableString(detail.WarningText),
			nullableString(detail.ErrorText),
			now,
			now,
			now,
			now,
		); err != nil {
			return model.MarketBackfillRun{}, err
		}
	}
	executed, execErr := r.executeMarketDataBackfillRun(runID)
	if execErr != nil {
		if latest, getErr := r.AdminGetMarketDataBackfillRun(runID); getErr == nil {
			return latest, execErr
		}
		return model.MarketBackfillRun{}, execErr
	}
	return executed, nil
}

func (r *MySQLGrowthRepo) AdminListMarketDataBackfillRuns(status string, runType string, assetType string, sourceKey string, page int, pageSize int) ([]model.MarketBackfillRun, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := (page - 1) * pageSize
	filters := []string{"1=1"}
	args := make([]any, 0, 8)
	if normalized := normalizeMarketBackfillRunStatus(status); strings.TrimSpace(status) != "" {
		filters = append(filters, "status = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeMarketBackfillRunType(runType); strings.TrimSpace(runType) != "" {
		filters = append(filters, "run_type = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeMarketBackfillAssetType(assetType); normalized != "" {
		filters = append(filters, "CAST(asset_scope AS CHAR) LIKE ?")
		args = append(args, "%"+normalized+"%")
	}
	if trimmed := strings.ToUpper(strings.TrimSpace(sourceKey)); trimmed != "" {
		filters = append(filters, "source_key = ?")
		args = append(args, trimmed)
	}
	where := " WHERE " + strings.Join(filters, " AND ")

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM market_backfill_runs"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT id, scheduler_run_id, run_type, asset_scope, COALESCE(DATE_FORMAT(trade_date_from, '%Y-%m-%d'), ''),
       COALESCE(DATE_FORMAT(trade_date_to, '%Y-%m-%d'), ''), COALESCE(source_key, ''), batch_size,
       universe_snapshot_id, status, current_stage, COALESCE(CAST(stage_progress_json AS CHAR), ''),
       COALESCE(CAST(summary_json AS CHAR), ''), COALESCE(error_message, ''), COALESCE(created_by, ''),
       created_at, updated_at, finished_at
FROM market_backfill_runs`+where+`
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.MarketBackfillRun, 0, pageSize)
	for rows.Next() {
		var item model.MarketBackfillRun
		var assetScopeJSON, stageProgressJSON, summaryJSON string
		var finishedAt sql.NullTime
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.SchedulerRunID,
			&item.RunType,
			&assetScopeJSON,
			&item.TradeDateFrom,
			&item.TradeDateTo,
			&item.SourceKey,
			&item.BatchSize,
			&item.UniverseSnapshotID,
			&item.Status,
			&item.CurrentStage,
			&stageProgressJSON,
			&summaryJSON,
			&item.ErrorMessage,
			&item.CreatedBy,
			&createdAt,
			&updatedAt,
			&finishedAt,
		); err != nil {
			return nil, 0, err
		}
		item.AssetScope = unmarshalStringSlice(assetScopeJSON)
		item.StageProgress = unmarshalStageProgress(stageProgressJSON)
		item.Summary = unmarshalSummaryMap(summaryJSON)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		item.FinishedAt = parseNullableTimeRFC3339(finishedAt)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminGetMarketDataBackfillRun(id string) (model.MarketBackfillRun, error) {
	items, _, err := r.AdminListMarketDataBackfillRuns("", "", "", "", 1, 200)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}
	for _, item := range items {
		if item.ID == strings.TrimSpace(id) {
			return item, nil
		}
	}
	return model.MarketBackfillRun{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) AdminListMarketDataBackfillRunDetails(runID string, stage string, assetType string, status string, page int, pageSize int) ([]model.MarketBackfillRunDetail, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := (page - 1) * pageSize
	filters := []string{"run_id = ?"}
	args := []any{strings.TrimSpace(runID)}
	if normalized := normalizeMarketBackfillStage(stage); normalized != "" {
		filters = append(filters, "stage = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeMarketBackfillAssetType(assetType); normalized != "" {
		filters = append(filters, "asset_type = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeMarketBackfillDetailStatus(status); strings.TrimSpace(status) != "" {
		filters = append(filters, "status = ?")
		args = append(args, normalized)
	}
	where := " WHERE " + strings.Join(filters, " AND ")

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM market_backfill_run_details"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, run_id, COALESCE(scheduler_run_id, ''), stage, COALESCE(asset_type, ''), COALESCE(batch_key, ''),
       COALESCE(source_key, ''), symbol_count, COALESCE(CAST(symbol_sample AS CHAR), ''),
       COALESCE(DATE_FORMAT(trade_date_from, '%Y-%m-%d'), ''), COALESCE(DATE_FORMAT(trade_date_to, '%Y-%m-%d'), ''),
       status, fetched_count, upserted_count, truth_count, COALESCE(warning_text, ''), COALESCE(error_text, ''),
       started_at, finished_at, created_at, updated_at
FROM market_backfill_run_details`+where+`
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.MarketBackfillRunDetail, 0, pageSize)
	for rows.Next() {
		var item model.MarketBackfillRunDetail
		var symbolSampleJSON string
		var startedAt, createdAt, updatedAt time.Time
		var finishedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.SchedulerRunID,
			&item.Stage,
			&item.AssetType,
			&item.BatchKey,
			&item.SourceKey,
			&item.SymbolCount,
			&symbolSampleJSON,
			&item.TradeDateFrom,
			&item.TradeDateTo,
			&item.Status,
			&item.FetchedCount,
			&item.UpsertedCount,
			&item.TruthCount,
			&item.WarningText,
			&item.ErrorText,
			&startedAt,
			&finishedAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, err
		}
		item.SymbolSample = unmarshalStringSlice(symbolSampleJSON)
		item.StartedAt = startedAt.Format(time.RFC3339)
		item.FinishedAt = parseNullableTimeRFC3339(finishedAt)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminRetryMarketDataBackfillRun(runID string, input model.MarketBackfillRetryInput, operator string) (model.MarketBackfillRun, error) {
	base, err := r.AdminGetMarketDataBackfillRun(runID)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}
	newSchedulerRunID, err := r.AdminRetrySchedulerJobRun(base.SchedulerRunID, "MANUAL", "RUNNING", "retry market data backfill run", "", operator)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}
	newRunID := newID("mbr")
	now := time.Now()
	currentStage := base.CurrentStage
	if normalized := normalizeMarketBackfillStage(input.Stage); normalized != "" && normalizeMarketBackfillRetryMode(input.RetryMode) == "FROM_STAGE" {
		currentStage = normalized
	}
	retried := base
	retried.ID = newRunID
	retried.SchedulerRunID = newSchedulerRunID
	retried.Status = "RUNNING"
	retried.CurrentStage = currentStage
	retried.ErrorMessage = ""
	retried.CreatedBy = operator
	retried.CreatedAt = now.Format(time.RFC3339)
	retried.UpdatedAt = now.Format(time.RFC3339)
	retried.FinishedAt = ""
	if _, err := r.db.Exec(`
INSERT INTO market_backfill_runs (
	id, scheduler_run_id, run_type, asset_scope, trade_date_from, trade_date_to, source_key, batch_size,
	universe_snapshot_id, status, current_stage, stage_progress_json, summary_json, error_message,
	created_by, created_at, updated_at, finished_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		retried.ID,
		retried.SchedulerRunID,
		retried.RunType,
		marshalJSONText(retried.AssetScope),
		nullableString(strings.TrimSpace(retried.TradeDateFrom)),
		nullableString(strings.TrimSpace(retried.TradeDateTo)),
		nullableString(strings.TrimSpace(retried.SourceKey)),
		retried.BatchSize,
		retried.UniverseSnapshotID,
		retried.Status,
		retried.CurrentStage,
		marshalJSONText(retried.StageProgress),
		marshalJSONText(retried.Summary),
		nil,
		nullableString(strings.TrimSpace(retried.CreatedBy)),
		now,
		now,
		nil,
	); err != nil {
		return model.MarketBackfillRun{}, err
	}
	executed, execErr := r.executeMarketDataBackfillRun(retried.ID)
	if execErr != nil {
		if latest, getErr := r.AdminGetMarketDataBackfillRun(retried.ID); getErr == nil {
			return latest, execErr
		}
		return model.MarketBackfillRun{}, execErr
	}
	return executed, nil
}

func (r *MySQLGrowthRepo) AdminListMarketUniverseSnapshots(page int, pageSize int) ([]model.MarketUniverseSnapshot, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM market_universe_snapshots").Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, COALESCE(CAST(scope AS CHAR), ''), COALESCE(source_key, ''), snapshot_date,
       COALESCE(CAST(summary_json AS CHAR), ''), COALESCE(created_by, ''), created_at
FROM market_universe_snapshots
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.MarketUniverseSnapshot, 0, pageSize)
	for rows.Next() {
		var item model.MarketUniverseSnapshot
		var scopeJSON, summaryJSON string
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &scopeJSON, &item.SourceKey, &item.SnapshotDate, &summaryJSON, &item.CreatedBy, &createdAt); err != nil {
			return nil, 0, err
		}
		item.Scope = unmarshalStringSlice(scopeJSON)
		item.AssetSummaries = buildMarketUniverseAssetSummariesFromJSON(summaryJSON)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminGetMarketUniverseSnapshot(id string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error) {
	var snapshot model.MarketUniverseSnapshot
	var scopeJSON, summaryJSON string
	var createdAt time.Time
	err := r.db.QueryRow(`
SELECT id, COALESCE(CAST(scope AS CHAR), ''), COALESCE(source_key, ''), snapshot_date,
       COALESCE(CAST(summary_json AS CHAR), ''), COALESCE(created_by, ''), created_at
FROM market_universe_snapshots
WHERE id = ?`, strings.TrimSpace(id)).
		Scan(&snapshot.ID, &scopeJSON, &snapshot.SourceKey, &snapshot.SnapshotDate, &summaryJSON, &snapshot.CreatedBy, &createdAt)
	if err != nil {
		return model.MarketUniverseSnapshot{}, nil, err
	}
	snapshot.Scope = unmarshalStringSlice(scopeJSON)
	snapshot.AssetSummaries = buildMarketUniverseAssetSummariesFromJSON(summaryJSON)
	snapshot.CreatedAt = createdAt.Format(time.RFC3339)

	rows, err := r.db.Query(`
SELECT id, snapshot_id, asset_type, instrument_key, COALESCE(external_symbol, ''), COALESCE(display_name, ''),
       COALESCE(exchange_code, ''), COALESCE(status, ''), COALESCE(DATE_FORMAT(list_date, '%Y-%m-%d'), ''),
       COALESCE(DATE_FORMAT(delist_date, '%Y-%m-%d'), ''), COALESCE(CAST(raw_metadata_json AS CHAR), ''), created_at
FROM market_universe_snapshot_items
WHERE snapshot_id = ?
ORDER BY asset_type ASC, instrument_key ASC`, snapshot.ID)
	if err != nil {
		return model.MarketUniverseSnapshot{}, nil, err
	}
	defer rows.Close()
	items := make([]model.MarketUniverseSnapshotItem, 0)
	for rows.Next() {
		var item model.MarketUniverseSnapshotItem
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.SnapshotID, &item.AssetType, &item.InstrumentKey, &item.ExternalSymbol, &item.DisplayName, &item.ExchangeCode, &item.Status, &item.ListDate, &item.DelistDate, &item.MetadataJSON, &createdAt); err != nil {
			return model.MarketUniverseSnapshot{}, nil, err
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return snapshot, items, rows.Err()
}

func buildMarketUniverseAssetSummariesFromJSON(raw string) []model.MarketUniverseSnapshotAssetItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var envelope struct {
		AssetSummaries []model.MarketUniverseSnapshotAssetItem `json:"asset_summaries"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err == nil && len(envelope.AssetSummaries) > 0 {
		return envelope.AssetSummaries
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminGetMarketCoverageSummary() (model.MarketCoverageSummary, error) {
	summary := model.MarketCoverageSummary{
		AssetItems: make([]model.MarketCoverageSummaryAssetItem, 0, len(allowedMarketBackfillAssetTypes)),
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM market_instruments WHERE asset_class = 'STOCK'`).Scan(&summary.TotalUniverseCount); err != nil {
		return summary, err
	}
	summary.MasterCoverageCount = summary.TotalUniverseCount
	if err := r.db.QueryRow(`SELECT COUNT(DISTINCT instrument_key) FROM market_daily_bar_truth WHERE asset_class = 'STOCK'`).Scan(&summary.QuotesCoverageCount); err != nil {
		return summary, err
	}
	if err := r.db.QueryRow(`SELECT COUNT(DISTINCT symbol) FROM stock_daily_basic`).Scan(&summary.DailyBasicCoverageCount); err != nil {
		return summary, err
	}
	if err := r.db.QueryRow(`SELECT COUNT(DISTINCT symbol) FROM stock_moneyflow_daily`).Scan(&summary.MoneyflowCoverageCount); err != nil {
		return summary, err
	}
	var latestTradeDate sql.NullTime
	if err := r.db.QueryRow(`SELECT MAX(trade_date) FROM market_daily_bar_truth WHERE asset_class = 'STOCK'`).Scan(&latestTradeDate); err != nil {
		return summary, err
	}
	summary.LatestTradeDate = parseNullableTimeRFC3339(latestTradeDate)
	if summary.LatestTradeDate != "" {
		summary.LatestTradeDate = latestTradeDate.Time.Format("2006-01-02")
	}

	rows, err := r.db.Query(`
SELECT COALESCE(selected_source_key, '') AS source_key, COUNT(*)
FROM market_daily_bar_truth
WHERE asset_class = 'STOCK'
GROUP BY COALESCE(selected_source_key, '')
ORDER BY COUNT(*) DESC, source_key ASC
LIMIT 5`)
	if err != nil && isMarketStatusSchemaCompatError(err) {
		// Schema compatibility fallback for historical deployments that used `source_key`.
		rows, err = r.db.Query(`
SELECT COALESCE(source_key, '') AS source_key, COUNT(*)
FROM market_daily_bar_truth
WHERE asset_class = 'STOCK'
GROUP BY COALESCE(source_key, '')
ORDER BY COUNT(*) DESC, source_key ASC
LIMIT 5`)
	}
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	for rows.Next() {
		var item model.MarketCoverageSourceSummaryItem
		if err := rows.Scan(&item.SourceKey, &item.Count); err != nil {
			return summary, err
		}
		summary.FallbackSourceSummary = append(summary.FallbackSourceSummary, item)
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}
	summary.AssetItems = append(summary.AssetItems,
		model.MarketCoverageSummaryAssetItem{
			AssetType:               "STOCK",
			UniverseCount:           summary.TotalUniverseCount,
			MasterCoverageCount:     summary.MasterCoverageCount,
			QuotesCoverageCount:     summary.QuotesCoverageCount,
			DailyBasicCoverageCount: summary.DailyBasicCoverageCount,
			MoneyflowCoverageCount:  summary.MoneyflowCoverageCount,
			LatestTradeDate:         summary.LatestTradeDate,
		},
		model.MarketCoverageSummaryAssetItem{AssetType: "INDEX"},
		model.MarketCoverageSummaryAssetItem{AssetType: "ETF"},
		model.MarketCoverageSummaryAssetItem{AssetType: "LOF"},
		model.MarketCoverageSummaryAssetItem{AssetType: "CBOND"},
	)
	return summary, nil
}

func (r *InMemoryGrowthRepo) AdminCreateMarketDataBackfillRun(input model.MarketBackfillCreateInput, operator string) (model.MarketBackfillRun, error) {
	assetScope := normalizeMarketBackfillAssetScope(input.AssetScope)
	if len(assetScope) == 0 {
		return model.MarketBackfillRun{}, fmt.Errorf("asset_scope is required")
	}
	longHistory, err := resolveMarketBackfillLongHistoryOptions(input.RunType, assetScope, input.TradeDateFrom, input.TradeDateTo, input.Stages)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}
	if err := r.validateMarketBackfillLongHistorySourceKey(input.SourceKey, longHistory); err != nil {
		return model.MarketBackfillRun{}, err
	}
	if operator == "" {
		operator = "system"
	}
	r.mu.Lock()

	now := time.Now()
	nowText := now.Format(time.RFC3339)
	snapshot, snapshotItems, err := r.buildMarketUniverseSnapshotLocked(input.SourceKey, assetScope, operator)
	if err != nil {
		r.mu.Unlock()
		return model.MarketBackfillRun{}, err
	}

	run := model.MarketBackfillRun{
		ID:                 "mbr_" + strings.ToLower(strings.ReplaceAll(newID("run"), "_", "")),
		SchedulerRunID:     "jr_" + strings.ToLower(strings.ReplaceAll(newID("job"), "_", "")),
		RunType:            normalizeMarketBackfillRunType(input.RunType),
		AssetScope:         assetScope,
		TradeDateFrom:      strings.TrimSpace(input.TradeDateFrom),
		TradeDateTo:        strings.TrimSpace(input.TradeDateTo),
		SourceKey:          snapshot.SourceKey,
		BatchSize:          input.BatchSize,
		UniverseSnapshotID: snapshot.ID,
		Status:             "RUNNING",
		CurrentStage:       "MASTER",
		StageProgress:      buildMarketBackfillStageProgressAfterUniverse(assetScope),
		Summary:            buildMarketBackfillSummary(input, assetScope, snapshotItems, longHistory),
		CreatedBy:          operator,
		CreatedAt:          nowText,
		UpdatedAt:          nowText,
		FinishedAt:         "",
	}
	if run.BatchSize <= 0 {
		run.BatchSize = 200
	}
	r.marketBackfillRuns[run.ID] = run
	r.marketBackfillRunDetails[run.ID] = buildMarketUniverseRunDetails(run.ID, run.SchedulerRunID, run.SourceKey, snapshotItems, assetScope, now)
	r.mu.Unlock()

	executed, execErr := r.executeMarketDataBackfillRun(run.ID)
	if execErr != nil {
		return model.MarketBackfillRun{}, execErr
	}
	return executed, nil
}

func (r *InMemoryGrowthRepo) AdminListMarketDataBackfillRuns(status string, runType string, assetType string, sourceKey string, page int, pageSize int) ([]model.MarketBackfillRun, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]model.MarketBackfillRun, 0, len(r.marketBackfillRuns))
	for _, item := range r.marketBackfillRuns {
		if normalized := strings.TrimSpace(status); normalized != "" && item.Status != normalizeMarketBackfillRunStatus(normalized) {
			continue
		}
		if normalized := strings.TrimSpace(runType); normalized != "" && item.RunType != normalizeMarketBackfillRunType(normalized) {
			continue
		}
		if normalized := normalizeMarketBackfillAssetType(assetType); normalized != "" {
			found := false
			for _, current := range item.AssetScope {
				if current == normalized {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if trimmed := strings.ToUpper(strings.TrimSpace(sourceKey)); trimmed != "" && item.SourceKey != trimmed {
			continue
		}
		items = append(items, item)
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminGetMarketDataBackfillRun(id string) (model.MarketBackfillRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.marketBackfillRuns[strings.TrimSpace(id)]
	if !ok {
		return model.MarketBackfillRun{}, sql.ErrNoRows
	}
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminListMarketDataBackfillRunDetails(runID string, stage string, assetType string, status string, page int, pageSize int) ([]model.MarketBackfillRunDetail, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := r.marketBackfillRunDetails[strings.TrimSpace(runID)]
	filtered := make([]model.MarketBackfillRunDetail, 0, len(items))
	for _, item := range items {
		if normalized := normalizeMarketBackfillStage(stage); normalized != "" && item.Stage != normalized {
			continue
		}
		if normalized := normalizeMarketBackfillAssetType(assetType); normalized != "" && item.AssetType != normalized {
			continue
		}
		if normalized := strings.TrimSpace(status); normalized != "" && item.Status != normalizeMarketBackfillDetailStatus(normalized) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, len(filtered), nil
}

func (r *InMemoryGrowthRepo) AdminRetryMarketDataBackfillRun(runID string, input model.MarketBackfillRetryInput, operator string) (model.MarketBackfillRun, error) {
	r.mu.Lock()
	base, ok := r.marketBackfillRuns[strings.TrimSpace(runID)]
	if !ok {
		r.mu.Unlock()
		return model.MarketBackfillRun{}, sql.ErrNoRows
	}
	now := time.Now().Format(time.RFC3339)
	retried := base
	retried.ID = "mbr_" + strings.ToLower(strings.ReplaceAll(newID("retry"), "_", ""))
	retried.SchedulerRunID = "jr_" + strings.ToLower(strings.ReplaceAll(newID("retry"), "_", ""))
	retried.Status = "RUNNING"
	retried.ErrorMessage = ""
	retried.CreatedBy = operator
	retried.CreatedAt = now
	retried.UpdatedAt = now
	retried.FinishedAt = ""
	if normalized := normalizeMarketBackfillStage(input.Stage); normalized != "" && normalizeMarketBackfillRetryMode(input.RetryMode) == "FROM_STAGE" {
		retried.CurrentStage = normalized
	}
	r.marketBackfillRuns[retried.ID] = retried
	r.marketBackfillRunDetails[retried.ID] = r.marketBackfillRunDetails[base.ID]
	r.mu.Unlock()

	executed, execErr := r.executeMarketDataBackfillRun(retried.ID)
	if execErr != nil {
		return model.MarketBackfillRun{}, execErr
	}
	return executed, nil
}

func (r *InMemoryGrowthRepo) AdminListMarketUniverseSnapshots(page int, pageSize int) ([]model.MarketUniverseSnapshot, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]model.MarketUniverseSnapshot, 0, len(r.marketUniverseSnapshots))
	for _, item := range r.marketUniverseSnapshots {
		items = append(items, item)
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminGetMarketUniverseSnapshot(id string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.marketUniverseSnapshots[strings.TrimSpace(id)]
	if !ok {
		return model.MarketUniverseSnapshot{}, nil, sql.ErrNoRows
	}
	return item, r.marketUniverseItems[item.ID], nil
}

func (r *InMemoryGrowthRepo) AdminGetMarketCoverageSummary() (model.MarketCoverageSummary, error) {
	return model.MarketCoverageSummary{
		TotalUniverseCount:      5,
		MasterCoverageCount:     4,
		QuotesCoverageCount:     3,
		DailyBasicCoverageCount: 2,
		MoneyflowCoverageCount:  2,
		LatestTradeDate:         "2026-03-24",
		FallbackSourceSummary: []model.MarketCoverageSourceSummaryItem{
			{SourceKey: "TUSHARE", Count: 3},
			{SourceKey: "AKSHARE", Count: 1},
		},
		AssetItems: []model.MarketCoverageSummaryAssetItem{
			{AssetType: "STOCK", UniverseCount: 2, MasterCoverageCount: 2, QuotesCoverageCount: 2, DailyBasicCoverageCount: 2, MoneyflowCoverageCount: 2, LatestTradeDate: "2026-03-24"},
			{AssetType: "INDEX", UniverseCount: 1, MasterCoverageCount: 1, QuotesCoverageCount: 1, LatestTradeDate: "2026-03-24"},
			{AssetType: "ETF", UniverseCount: 1, MasterCoverageCount: 1},
			{AssetType: "LOF", UniverseCount: 1},
			{AssetType: "CBOND", UniverseCount: 0},
		},
		CanonicalKeyGapCount: 0,
		DisplayNameGapCount:  1,
		ListDateGapCount:     1,
	}, nil
}
