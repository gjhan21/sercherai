package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func newID(prefix string) string {
	seq := repoIDSequence.Add(1)
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), seq)
}

func nullableString(value string) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func formatNullTime(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format(time.RFC3339)
	}
	return ""
}

type marketRhythmTaskTemplate struct {
	Slot    string
	TaskKey string
}

func defaultMarketRhythmTaskTemplates() []marketRhythmTaskTemplate {
	return []marketRhythmTaskTemplate{
		{Slot: "08:30", TaskKey: "morning_stock_publish"},
		{Slot: "11:30", TaskKey: "midday_news_publish"},
		{Slot: "15:30", TaskKey: "close_tracking_review"},
		{Slot: "周末", TaskKey: "weekend_review_digest"},
	}
}

func normalizeMarketRhythmDate(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Now().Format("2006-01-02"), nil
	}
	parsed, err := time.Parse("2006-01-02", text)
	if err != nil {
		return "", err
	}
	return parsed.Format("2006-01-02"), nil
}

func normalizeMarketRhythmStatus(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "TODO":
		return "TODO"
	case "IN_PROGRESS":
		return "IN_PROGRESS"
	case "DONE":
		return "DONE"
	case "BLOCKED":
		return "BLOCKED"
	default:
		return ""
	}
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		if _, exists := seen[text]; exists {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func marshalStringList(items []string) (string, error) {
	if len(items) == 0 {
		return "", nil
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func parseStringList(raw sql.NullString) []string {
	text := strings.TrimSpace(raw.String)
	if !raw.Valid || text == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(text), &items); err == nil {
		return normalizeStringList(items)
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	return normalizeStringList(parts)
}

func scanMarketRhythmTask(scanner interface {
	Scan(dest ...interface{}) error
}) (model.MarketRhythmTask, error) {
	var item model.MarketRhythmTask
	var taskDate time.Time
	var sourceLinksJSON sql.NullString
	var completedAt sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&taskDate,
		&item.Slot,
		&item.TaskKey,
		&item.Status,
		&item.Owner,
		&item.Notes,
		&sourceLinksJSON,
		&completedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.MarketRhythmTask{}, err
	}
	item.Date = taskDate.Format("2006-01-02")
	item.SourceLinks = parseStringList(sourceLinksJSON)
	if completedAt.Valid {
		item.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}
	if createdAt.Valid {
		item.CreatedAt = createdAt.Time.Format(time.RFC3339)
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
	}
	return item, nil
}

func getMarketRhythmTaskByIDTx(tx *sql.Tx, id string) (model.MarketRhythmTask, error) {
	row := tx.QueryRow(`
SELECT id, task_date, slot, task_key, status, COALESCE(owner, ''), COALESCE(notes, ''), source_links_json, completed_at, created_at, updated_at
FROM market_rhythm_tasks
WHERE id = ?
LIMIT 1`, id)
	return scanMarketRhythmTask(row)
}

func (r *MySQLGrowthRepo) adminGetMarketRhythmTaskByID(id string) (model.MarketRhythmTask, error) {
	row := r.db.QueryRow(`
SELECT id, task_date, slot, task_key, status, COALESCE(owner, ''), COALESCE(notes, ''), source_links_json, completed_at, created_at, updated_at
FROM market_rhythm_tasks
WHERE id = ?
LIMIT 1`, id)
	return scanMarketRhythmTask(row)
}

func marketRhythmCompletedAtValue(existing string, status string) interface{} {
	if normalizeMarketRhythmStatus(status) != "DONE" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(existing)); err == nil {
		return parsed
	}
	return time.Now()
}

func vipRemainingDays(expireAt sql.NullTime, now time.Time) int {
	if !expireAt.Valid {
		return 0
	}
	if !expireAt.Time.After(now) {
		return 0
	}
	hours := expireAt.Time.Sub(now).Hours()
	if hours <= 0 {
		return 0
	}
	days := int(math.Ceil(hours / 24))
	if days < 0 {
		return 0
	}
	return days
}

func vipStatusFromLevelAndExpire(memberLevel string, expireAt sql.NullTime, now time.Time) string {
	level := strings.ToUpper(strings.TrimSpace(memberLevel))
	if !strings.HasPrefix(level, "VIP") {
		return "FREE"
	}
	if !expireAt.Valid {
		return "ACTIVE"
	}
	if expireAt.Time.After(now) {
		return "ACTIVE"
	}
	return "EXPIRED"
}

func resolveVIPStatusAndDays(memberLevel string, expireAt sql.NullTime) (string, int) {
	now := time.Now()
	return vipStatusFromLevelAndExpire(memberLevel, expireAt, now), vipRemainingDays(expireAt, now)
}

func isVerifiedKYCStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "APPROVED", "VERIFIED":
		return true
	default:
		return false
	}
}

func resolveMembershipActivationState(memberLevel string, vipExpireAt sql.NullTime) string {
	if vipStatusFromLevelAndExpire(memberLevel, vipExpireAt, time.Now()) != "ACTIVE" {
		return "NON_MEMBER"
	}
	return "ACTIVE"
}

func parseRepoConfigBool(raw string, fallback bool) bool {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return fallback
	}
	switch text {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func parseRepoConfigInt(raw string, fallback int) int {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fallback
	}
	return value
}
