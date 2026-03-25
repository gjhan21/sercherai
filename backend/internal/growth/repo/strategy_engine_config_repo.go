package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

const (
	strategySeedSetConfigPrefix      = "strategy_engine.seed_set."
	strategyAgentProfileConfigPrefix = "strategy_engine.agent_profile."
)

func (r *MySQLGrowthRepo) AdminListStrategySeedSets(targetType string, status string, page int, pageSize int) ([]model.StrategySeedSet, int, error) {
	items, err := r.listStrategySeedSetsAll()
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]model.StrategySeedSet, 0, len(items))
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	status = strings.ToUpper(strings.TrimSpace(status))
	for _, item := range items {
		if targetType != "" && item.TargetType != targetType {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		filtered = append(filtered, item)
	}
	return paginateStrategySeedSets(filtered, page, pageSize), len(filtered), nil
}

func (r *MySQLGrowthRepo) AdminCreateStrategySeedSet(item model.StrategySeedSet) (string, error) {
	item = normalizeStrategySeedSet(item)
	item.ID = newStrategyConfigID("sd")
	if err := r.upsertStrategySeedSet(item, item.UpdatedBy); err != nil {
		return "", err
	}
	return item.ID, nil
}

func (r *MySQLGrowthRepo) AdminUpdateStrategySeedSet(id string, item model.StrategySeedSet) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return sql.ErrNoRows
	}
	current, err := r.getStrategySeedSet(id)
	if err != nil {
		return err
	}
	item = normalizeStrategySeedSet(item)
	item.ID = id
	if item.UpdatedBy == "" {
		item.UpdatedBy = current.UpdatedBy
	}
	return r.upsertStrategySeedSet(item, item.UpdatedBy)
}

func (r *MySQLGrowthRepo) AdminListStrategyAgentProfiles(targetType string, status string, page int, pageSize int) ([]model.StrategyAgentProfile, int, error) {
	if err := r.ensureStrategyAgentProfilesMaterialized(targetType); err != nil {
		return nil, 0, err
	}
	items, err := r.listStrategyAgentProfilesAll()
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]model.StrategyAgentProfile, 0, len(items))
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	status = strings.ToUpper(strings.TrimSpace(status))
	for _, item := range items {
		if targetType != "" && item.TargetType != targetType {
			continue
		}
		if status != "" && item.Status != status {
			continue
		}
		filtered = append(filtered, item)
	}
	return paginateStrategyAgentProfiles(filtered, page, pageSize), len(filtered), nil
}

func (r *MySQLGrowthRepo) AdminCreateStrategyAgentProfile(item model.StrategyAgentProfile) (string, error) {
	item = normalizeStrategyAgentProfile(item)
	item.ID = newStrategyConfigID("ag")
	if err := r.upsertStrategyAgentProfile(item, item.UpdatedBy); err != nil {
		return "", err
	}
	return item.ID, nil
}

func (r *MySQLGrowthRepo) AdminUpdateStrategyAgentProfile(id string, item model.StrategyAgentProfile) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return sql.ErrNoRows
	}
	current, err := r.getStrategyAgentProfile(id)
	if err != nil {
		return err
	}
	item = normalizeStrategyAgentProfile(item)
	item.ID = id
	if item.UpdatedBy == "" {
		item.UpdatedBy = current.UpdatedBy
	}
	return r.upsertStrategyAgentProfile(item, item.UpdatedBy)
}

func (r *MySQLGrowthRepo) ResolveActiveStrategySeedSet(targetType string) (*model.StrategySeedSet, error) {
	items, err := r.listStrategySeedSetsAll()
	if err != nil {
		return nil, err
	}
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	for _, item := range items {
		if item.Status != "ACTIVE" || !item.IsDefault {
			continue
		}
		if item.TargetType == targetType {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *MySQLGrowthRepo) ResolveActiveStrategyAgentProfile(targetType string) (*model.StrategyAgentProfile, error) {
	if err := r.ensureStrategyAgentProfilesMaterialized(targetType); err != nil {
		return nil, err
	}
	items, err := r.listStrategyAgentProfilesAll()
	if err != nil {
		return nil, err
	}
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	for _, item := range items {
		if item.Status != "ACTIVE" || !item.IsDefault {
			continue
		}
		if item.TargetType == targetType {
			cloned := item
			return &cloned, nil
		}
	}
	for _, item := range items {
		if item.Status == "ACTIVE" && item.IsDefault && item.TargetType == "ALL" {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *MySQLGrowthRepo) ensureStrategyAgentProfilesMaterialized(targetType string) error {
	items, err := r.listStrategyAgentProfilesAll()
	if err != nil {
		return err
	}
	activeDefaults := make(map[string]bool, len(items))
	for _, item := range items {
		if item.Status == "ACTIVE" && item.IsDefault {
			activeDefaults[item.TargetType] = true
		}
	}
	for _, requiredTarget := range strategyAgentProfileTargetsForBootstrap(targetType) {
		if activeDefaults[requiredTarget] {
			continue
		}
		builtIn := builtInStrategyAgentProfile(requiredTarget)
		if builtIn == nil {
			continue
		}
		builtIn.UpdatedBy = strategyBuiltInPolicyOperator
		if err := r.persistStrategyConfig(strategyAgentProfileConfigPrefix+builtIn.ID, builtIn, builtIn.Description, strategyBuiltInPolicyOperator); err != nil {
			return err
		}
		activeDefaults[requiredTarget] = true
	}
	return nil
}

func strategyAgentProfileTargetsForBootstrap(targetType string) []string {
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	switch targetType {
	case "FUTURES":
		return []string{"FUTURES", "ALL"}
	case "STOCK", "ALL":
		return []string{"ALL"}
	default:
		return []string{"ALL", "FUTURES"}
	}
}

func builtInStrategyAgentProfile(targetType string) *model.StrategyAgentProfile {
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	switch targetType {
	case "FUTURES":
		return &model.StrategyAgentProfile{
			ID:                              "agent_default_futures",
			Name:                            "期货默认五角评审",
			TargetType:                      "FUTURES",
			Status:                          "ACTIVE",
			IsDefault:                       true,
			EnabledAgents:                   []string{"trend", "event", "liquidity", "risk", "basis"},
			PositiveThreshold:               3,
			NegativeThreshold:               2,
			AllowVeto:                       true,
			AllowMockFallbackOnShortHistory: true,
			Description:                     "系统自动落库的期货默认角色配置，允许短历史时受控回退 MOCK。",
			UpdatedBy:                       "system",
		}
	case "", "ALL", "STOCK":
		return &model.StrategyAgentProfile{
			ID:                "agent_default_all",
			Name:              "默认五角评审",
			TargetType:        "ALL",
			Status:            "ACTIVE",
			IsDefault:         true,
			EnabledAgents:     []string{"trend", "event", "liquidity", "risk", "basis"},
			PositiveThreshold: 3,
			NegativeThreshold: 2,
			AllowVeto:         true,
			Description:       "系统自动落库的默认多角色评审配置。",
			UpdatedBy:         "system",
		}
	default:
		return nil
	}
}

func (r *MySQLGrowthRepo) listStrategySeedSetsAll() ([]model.StrategySeedSet, error) {
	rows, err := r.db.Query(`
SELECT config_key, config_value, description, updated_by, updated_at
FROM system_configs
WHERE config_key LIKE ?
ORDER BY updated_at DESC, config_key DESC`, strategySeedSetConfigPrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategySeedSet, 0)
	for rows.Next() {
		var configKey string
		var configValue string
		var description sql.NullString
		var updatedBy sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&configKey, &configValue, &description, &updatedBy, &updatedAt); err != nil {
			return nil, err
		}
		var item model.StrategySeedSet
		if err := json.Unmarshal([]byte(configValue), &item); err != nil {
			continue
		}
		if item.ID == "" {
			item.ID = strings.TrimPrefix(configKey, strategySeedSetConfigPrefix)
		}
		item = normalizeStrategySeedSet(item)
		if description.Valid && strings.TrimSpace(item.Description) == "" {
			item.Description = strings.TrimSpace(description.String)
		}
		if updatedBy.Valid {
			item.UpdatedBy = strings.TrimSpace(updatedBy.String)
		}
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) listStrategyAgentProfilesAll() ([]model.StrategyAgentProfile, error) {
	rows, err := r.db.Query(`
SELECT config_key, config_value, description, updated_by, updated_at
FROM system_configs
WHERE config_key LIKE ?
ORDER BY updated_at DESC, config_key DESC`, strategyAgentProfileConfigPrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyAgentProfile, 0)
	for rows.Next() {
		var configKey string
		var configValue string
		var description sql.NullString
		var updatedBy sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&configKey, &configValue, &description, &updatedBy, &updatedAt); err != nil {
			return nil, err
		}
		var item model.StrategyAgentProfile
		if err := json.Unmarshal([]byte(configValue), &item); err != nil {
			continue
		}
		if item.ID == "" {
			item.ID = strings.TrimPrefix(configKey, strategyAgentProfileConfigPrefix)
		}
		item = normalizeStrategyAgentProfile(item)
		if description.Valid && strings.TrimSpace(item.Description) == "" {
			item.Description = strings.TrimSpace(description.String)
		}
		if updatedBy.Valid {
			item.UpdatedBy = strings.TrimSpace(updatedBy.String)
		}
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) getStrategySeedSet(id string) (model.StrategySeedSet, error) {
	items, err := r.listStrategySeedSetsAll()
	if err != nil {
		return model.StrategySeedSet{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return model.StrategySeedSet{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) getStrategyAgentProfile(id string) (model.StrategyAgentProfile, error) {
	items, err := r.listStrategyAgentProfilesAll()
	if err != nil {
		return model.StrategyAgentProfile{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return model.StrategyAgentProfile{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) upsertStrategySeedSet(item model.StrategySeedSet, operator string) error {
	item = normalizeStrategySeedSet(item)
	if item.IsDefault {
		items, err := r.listStrategySeedSetsAll()
		if err != nil {
			return err
		}
		for _, existing := range items {
			if existing.ID == item.ID || existing.TargetType != item.TargetType || !existing.IsDefault {
				continue
			}
			existing.IsDefault = false
			if err := r.persistStrategyConfig(strategySeedSetConfigPrefix+existing.ID, existing, existing.Description, operator); err != nil {
				return err
			}
		}
	}
	return r.persistStrategyConfig(strategySeedSetConfigPrefix+item.ID, item, item.Description, operator)
}

func (r *MySQLGrowthRepo) upsertStrategyAgentProfile(item model.StrategyAgentProfile, operator string) error {
	item = normalizeStrategyAgentProfile(item)
	if item.IsDefault {
		items, err := r.listStrategyAgentProfilesAll()
		if err != nil {
			return err
		}
		for _, existing := range items {
			if existing.ID == item.ID || existing.TargetType != item.TargetType || !existing.IsDefault {
				continue
			}
			existing.IsDefault = false
			if err := r.persistStrategyConfig(strategyAgentProfileConfigPrefix+existing.ID, existing, existing.Description, operator); err != nil {
				return err
			}
		}
	}
	return r.persistStrategyConfig(strategyAgentProfileConfigPrefix+item.ID, item, item.Description, operator)
}

func (r *MySQLGrowthRepo) persistStrategyConfig(configKey string, payload any, description string, operator string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), description = VALUES(description), updated_by = VALUES(updated_by), updated_at = VALUES(updated_at)`,
		newID("cfg"), configKey, string(body), strings.TrimSpace(description), strings.TrimSpace(operator), time.Now())
	if err == nil && strings.TrimSpace(operator) == strategyBuiltInPolicyOperator {
		r.recordStrategyConfigBootstrapAuditEvent(configKey, description, operator)
	}
	return err
}

func newStrategyConfigID(prefix string) string {
	seq := repoIDSequence.Add(1)
	return fmt.Sprintf(
		"%s_%s_%s",
		strings.TrimSpace(prefix),
		strconv.FormatInt(time.Now().UnixNano(), 36),
		strconv.FormatUint(uint64(seq), 36),
	)
}

func normalizeStrategySeedSet(item model.StrategySeedSet) model.StrategySeedSet {
	item.Name = strings.TrimSpace(item.Name)
	item.TargetType = strings.ToUpper(strings.TrimSpace(item.TargetType))
	if item.TargetType == "" {
		item.TargetType = "STOCK"
	}
	item.Status = strings.ToUpper(strings.TrimSpace(item.Status))
	if item.Status == "" {
		item.Status = "ACTIVE"
	}
	if item.Status != "ACTIVE" {
		item.IsDefault = false
	}
	item.Description = strings.TrimSpace(item.Description)
	seen := map[string]struct{}{}
	items := make([]string, 0, len(item.Items))
	for _, raw := range item.Items {
		normalized := strings.ToUpper(strings.TrimSpace(raw))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	item.Items = items
	return item
}

func normalizeStrategyAgentProfile(item model.StrategyAgentProfile) model.StrategyAgentProfile {
	item.Name = strings.TrimSpace(item.Name)
	item.TargetType = strings.ToUpper(strings.TrimSpace(item.TargetType))
	if item.TargetType == "" {
		item.TargetType = "ALL"
	}
	item.Status = strings.ToUpper(strings.TrimSpace(item.Status))
	if item.Status == "" {
		item.Status = "ACTIVE"
	}
	if item.Status != "ACTIVE" {
		item.IsDefault = false
	}
	item.Description = strings.TrimSpace(item.Description)
	if item.PositiveThreshold <= 0 {
		item.PositiveThreshold = 3
	}
	if item.NegativeThreshold <= 0 {
		item.NegativeThreshold = 2
	}
	seen := map[string]struct{}{}
	agents := make([]string, 0, len(item.EnabledAgents))
	for _, raw := range item.EnabledAgents {
		normalized := strings.ToLower(strings.TrimSpace(raw))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		agents = append(agents, normalized)
	}
	if len(agents) == 0 {
		agents = []string{"trend", "event", "liquidity", "risk", "basis"}
	}
	item.EnabledAgents = agents
	return item
}

func paginateStrategySeedSets(items []model.StrategySeedSet, page int, pageSize int) []model.StrategySeedSet {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []model.StrategySeedSet{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func paginateStrategyAgentProfiles(items []model.StrategyAgentProfile, page int, pageSize int) []model.StrategyAgentProfile {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []model.StrategyAgentProfile{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}
