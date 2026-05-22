package repo

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminListStockSelectionProfiles(status string, page int, pageSize int) ([]model.StockSelectionProfile, int, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	offset := (page - 1) * pageSize

	templateItems, _, err := r.AdminListStockSelectionProfileTemplates("", 1, 200)
	if err != nil {
		return nil, 0, err
	}
	templateMap := make(map[string]model.StockSelectionProfileTemplate, len(templateItems))
	var defaultTemplate *model.StockSelectionProfileTemplate
	for index := range templateItems {
		template := templateItems[index]
		templateMap[template.ID] = template
		if template.IsDefault && defaultTemplate == nil {
			cloned := template
			defaultTemplate = &cloned
		}
	}

	args := make([]any, 0, 4)
	filter := ""
	if status != "" {
		filter = " WHERE p.status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_selection_profiles p"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

query := `
SELECT
  p.id,
  p.name,
  COALESCE(p.template_id, ''),
  COALESCE(t.name, ''),
  p.status,
  p.is_default,
  COALESCE(v.current_version, 0) AS current_version,
  p.selection_mode_default,
  p.universe_scope,
  COALESCE(CAST(p.universe_config AS CHAR), ''),
  COALESCE(CAST(p.seed_mining_config AS CHAR), ''),
  COALESCE(CAST(p.factor_config AS CHAR), ''),
  COALESCE(CAST(p.portfolio_config AS CHAR), ''),
  COALESCE(CAST(p.publish_config AS CHAR), ''),
  COALESCE(CAST(p.market_analysis_config AS CHAR), ''),
  COALESCE(CAST(p.candidate_pool_config AS CHAR), ''),
  COALESCE(CAST(p.short_term_head_config AS CHAR), ''),
  COALESCE(CAST(p.swing_head_config AS CHAR), ''),
  COALESCE(p.description, ''),
  COALESCE(p.updated_by, ''),
  p.updated_at,
  p.created_at
FROM stock_selection_profiles p
LEFT JOIN stock_selection_profile_templates t ON t.id = p.template_id
LEFT JOIN (
  SELECT profile_id, MAX(version_no) AS current_version
  FROM stock_selection_profile_versions
  GROUP BY profile_id
) v ON v.profile_id = p.id` + filter + `
ORDER BY p.is_default DESC, p.updated_at DESC, p.id ASC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StockSelectionProfile, 0)
	profileIDs := make([]string, 0)
	for rows.Next() {
		var item model.StockSelectionProfile
		var isDefault bool
		var universeConfig, seedConfig, factorConfig, portfolioConfig, publishConfig string
		var marketAnalysisConfig, candidatePoolConfig, shortTermHeadConfig, swingHeadConfig string
		var updatedAt, createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.TemplateID,
			&item.TemplateName,
			&item.Status,
			&isDefault,
			&item.CurrentVersion,
			&item.SelectionModeDefault,
			&item.UniverseScope,
			&universeConfig,
			&seedConfig,
			&factorConfig,
			&portfolioConfig,
			&publishConfig,
			&marketAnalysisConfig,
			&candidatePoolConfig,
			&shortTermHeadConfig,
			&swingHeadConfig,
			&item.Description,
			&item.UpdatedBy,
			&updatedAt,
			&createdAt,
		); err != nil {
			return nil, 0, err
		}
		item.IsDefault = isDefault
		item.UniverseConfig = parseJSONMap(universeConfig)
		item.SeedMiningConfig = parseJSONMap(seedConfig)
		item.FactorConfig = parseJSONMap(factorConfig)
		item.PortfolioConfig = parseJSONMap(portfolioConfig)
		item.PublishConfig = parseJSONMap(publishConfig)
		item.MarketAnalysisConfig = parseJSONMap(marketAnalysisConfig)
		item.CandidatePoolConfig = parseJSONMap(candidatePoolConfig)
		item.ShortTermHeadConfig = parseJSONMap(shortTermHeadConfig)
		item.SwingHeadConfig = parseJSONMap(swingHeadConfig)
		template, hasTemplate := templateMap[item.TemplateID]
		if !hasTemplate && defaultTemplate != nil {
			template = *defaultTemplate
			hasTemplate = true
		}
		if len(item.MarketAnalysisConfig) == 0 {
			item.MarketAnalysisConfig = deriveStockSelectionMarketAnalysisConfig(item)
		}
		if len(item.CandidatePoolConfig) == 0 {
			item.CandidatePoolConfig = deriveStockSelectionCandidatePoolConfig(item)
		}
		if len(item.ShortTermHeadConfig) == 0 {
			item.ShortTermHeadConfig = deriveStockSelectionShortTermHeadConfig(item)
		}
		if len(item.SwingHeadConfig) == 0 {
			item.SwingHeadConfig = deriveStockSelectionSwingHeadConfig(item)
		}
		if hasTemplate {
			item.MarketAnalysisConfig = mergeStockSelectionConfigMaps(template.MarketAnalysisDefaults, item.MarketAnalysisConfig)
			item.CandidatePoolConfig = mergeStockSelectionConfigMaps(template.CandidatePoolDefaults, item.CandidatePoolConfig)
			item.ShortTermHeadConfig = mergeStockSelectionConfigMaps(template.ShortTermHeadDefaults, item.ShortTermHeadConfig)
			item.SwingHeadConfig = mergeStockSelectionConfigMaps(template.SwingHeadDefaults, item.SwingHeadConfig)
		}
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
		profileIDs = append(profileIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	versionMap, err := r.listStockSelectionProfileVersions(profileIDs)
	if err != nil {
		return nil, 0, err
	}
	for index := range items {
		items[index].Versions = versionMap[items[index].ID]
		if items[index].CurrentVersion == 0 && len(items[index].Versions) > 0 {
			items[index].CurrentVersion = items[index].Versions[0].VersionNo
		}
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateStockSelectionProfile(item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error) {
	item = normalizeStockSelectionProfile(item)
	item.ID = newID("ssp")
	item.CurrentVersion = 1
	if strings.TrimSpace(changeNote) == "" {
		changeNote = "创建 profile"
	}
	if err := r.upsertStockSelectionProfile(item, 1, changeNote, item.UpdatedBy, true); err != nil {
		return model.StockSelectionProfile{}, err
	}
	return r.getStockSelectionProfile(item.ID)
}

func (r *MySQLGrowthRepo) AdminUpdateStockSelectionProfile(id string, item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error) {
	current, err := r.getStockSelectionProfile(id)
	if err != nil {
		return model.StockSelectionProfile{}, err
	}
	item = normalizeStockSelectionProfile(item)
	item.ID = current.ID
	item.CurrentVersion = current.CurrentVersion + 1
	if item.UpdatedBy == "" {
		item.UpdatedBy = current.UpdatedBy
	}
	if strings.TrimSpace(changeNote) == "" {
		changeNote = fmt.Sprintf("更新 profile 到 v%d", item.CurrentVersion)
	}
	if err := r.upsertStockSelectionProfile(item, item.CurrentVersion, changeNote, item.UpdatedBy, false); err != nil {
		return model.StockSelectionProfile{}, err
	}
	return r.getStockSelectionProfile(id)
}

func (r *MySQLGrowthRepo) AdminPublishStockSelectionProfile(id string, operator string) (model.StockSelectionProfile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.StockSelectionProfile{}, sql.ErrNoRows
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.StockSelectionProfile{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.Exec("UPDATE stock_selection_profiles SET is_default = 0, updated_at = NOW() WHERE id <> ?", id); err != nil {
		return model.StockSelectionProfile{}, err
	}
	if _, err = tx.Exec("UPDATE stock_selection_profiles SET is_default = 1, updated_by = ?, updated_at = NOW() WHERE id = ?", operator, id); err != nil {
		return model.StockSelectionProfile{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.StockSelectionProfile{}, err
	}
	return r.getStockSelectionProfile(id)
}

func (r *MySQLGrowthRepo) AdminRollbackStockSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.StockSelectionProfile, error) {
	id = strings.TrimSpace(id)
	if id == "" || versionNo <= 0 {
		return model.StockSelectionProfile{}, sql.ErrNoRows
	}
	current, err := r.getStockSelectionProfile(id)
	if err != nil {
		return model.StockSelectionProfile{}, err
	}
	version, err := r.getStockSelectionProfileVersion(id, versionNo)
	if err != nil {
		return model.StockSelectionProfile{}, err
	}
	restored := normalizeStockSelectionProfile(profileFromSnapshot(version.Snapshot))
	restored.ID = id
	restored.CurrentVersion = current.CurrentVersion + 1
	restored.UpdatedBy = operator
	if strings.TrimSpace(changeNote) == "" {
		changeNote = fmt.Sprintf("从版本 v%d 回滚", versionNo)
	}
	if err := r.upsertStockSelectionProfile(restored, restored.CurrentVersion, changeNote, operator, false); err != nil {
		return model.StockSelectionProfile{}, err
	}
	return r.getStockSelectionProfile(id)
}

func (r *MySQLGrowthRepo) getDefaultStockSelectionProfile() (*model.StockSelectionProfile, error) {
	items, _, err := r.AdminListStockSelectionProfiles("ACTIVE", 1, 50)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.IsDefault {
			cloned := item
			return &cloned, nil
		}
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	cloned := items[0]
	return &cloned, nil
}

func (r *MySQLGrowthRepo) getStockSelectionProfile(id string) (model.StockSelectionProfile, error) {
	items, _, err := r.AdminListStockSelectionProfiles("", 1, 200)
	if err != nil {
		return model.StockSelectionProfile{}, err
	}
	for _, item := range items {
		if item.ID == strings.TrimSpace(id) {
			return item, nil
		}
	}
	return model.StockSelectionProfile{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) getStockSelectionProfileVersion(profileID string, versionNo int) (model.StockSelectionProfileVersion, error) {
	rows, err := r.listStockSelectionProfileVersions([]string{profileID})
	if err != nil {
		return model.StockSelectionProfileVersion{}, err
	}
	for _, item := range rows[profileID] {
		if item.VersionNo == versionNo {
			return item, nil
		}
	}
	return model.StockSelectionProfileVersion{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) AdminListStockSelectionProfileVersions(profileID string) ([]model.StockSelectionProfileVersion, error) {
	rows, err := r.listStockSelectionProfileVersions([]string{strings.TrimSpace(profileID)})
	if err != nil {
		return nil, err
	}
	return rows[strings.TrimSpace(profileID)], nil
}

func (r *MySQLGrowthRepo) listStockSelectionProfileVersions(profileIDs []string) (map[string][]model.StockSelectionProfileVersion, error) {
	if len(profileIDs) == 0 {
		return map[string][]model.StockSelectionProfileVersion{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(profileIDs)), ",")
	args := make([]any, 0, len(profileIDs))
	for _, id := range profileIDs {
		args = append(args, id)
	}
	query := fmt.Sprintf(`
SELECT id, profile_id, version_no, COALESCE(CAST(snapshot_json AS CHAR), ''), COALESCE(change_note, ''), COALESCE(created_by, ''), created_at
FROM stock_selection_profile_versions
WHERE profile_id IN (%s)
ORDER BY profile_id ASC, version_no DESC`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]model.StockSelectionProfileVersion, len(profileIDs))
	for rows.Next() {
		var item model.StockSelectionProfileVersion
		var snapshotJSON string
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.ProfileID, &item.VersionNo, &snapshotJSON, &item.ChangeNote, &item.CreatedBy, &createdAt); err != nil {
			return nil, err
		}
		item.Snapshot = parseJSONMap(snapshotJSON)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		result[item.ProfileID] = append(result[item.ProfileID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for profileID := range result {
		sort.SliceStable(result[profileID], func(i, j int) bool {
			return result[profileID][i].VersionNo > result[profileID][j].VersionNo
		})
	}
	return result, nil
}

func (r *MySQLGrowthRepo) upsertStockSelectionProfile(item model.StockSelectionProfile, versionNo int, changeNote string, operator string, createOnly bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if item.IsDefault {
		if _, err = tx.Exec("UPDATE stock_selection_profiles SET is_default = 0, updated_at = NOW() WHERE id <> ?", item.ID); err != nil {
			return err
		}
	}

	universeJSON := stockSelectionMustJSON(item.UniverseConfig)
	seedJSON := stockSelectionMustJSON(item.SeedMiningConfig)
	factorJSON := stockSelectionMustJSON(item.FactorConfig)
	portfolioJSON := stockSelectionMustJSON(item.PortfolioConfig)
	publishJSON := stockSelectionMustJSON(item.PublishConfig)
	marketAnalysisJSON := stockSelectionMustJSON(item.MarketAnalysisConfig)
	candidatePoolJSON := stockSelectionMustJSON(item.CandidatePoolConfig)
	shortTermHeadJSON := stockSelectionMustJSON(item.ShortTermHeadConfig)
	swingHeadJSON := stockSelectionMustJSON(item.SwingHeadConfig)

	if createOnly {
		_, err = tx.Exec(`
INSERT INTO stock_selection_profiles (
  id, name, template_id, status, is_default, selection_mode_default, universe_scope,
  universe_config, seed_mining_config, factor_config, portfolio_config, publish_config,
  market_analysis_config, candidate_pool_config, short_term_head_config, swing_head_config,
  description, updated_by, updated_at, created_at
) VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			item.ID,
			item.Name,
			item.TemplateID,
			item.Status,
			item.IsDefault,
			item.SelectionModeDefault,
			item.UniverseScope,
			universeJSON,
			seedJSON,
			factorJSON,
			portfolioJSON,
			publishJSON,
			marketAnalysisJSON,
			candidatePoolJSON,
			shortTermHeadJSON,
			swingHeadJSON,
			item.Description,
			operator,
		)
	} else {
		_, err = tx.Exec(`
UPDATE stock_selection_profiles
SET name = ?,
    template_id = NULLIF(?, ''),
    status = ?,
    is_default = ?,
    selection_mode_default = ?,
    universe_scope = ?,
    universe_config = ?,
    seed_mining_config = ?,
    factor_config = ?,
    portfolio_config = ?,
    publish_config = ?,
    market_analysis_config = ?,
    candidate_pool_config = ?,
    short_term_head_config = ?,
    swing_head_config = ?,
    description = ?,
    updated_by = ?,
    updated_at = NOW()
WHERE id = ?`,
			item.Name,
			item.TemplateID,
			item.Status,
			item.IsDefault,
			item.SelectionModeDefault,
			item.UniverseScope,
			universeJSON,
			seedJSON,
			factorJSON,
			portfolioJSON,
			publishJSON,
			marketAnalysisJSON,
			candidatePoolJSON,
			shortTermHeadJSON,
			swingHeadJSON,
			item.Description,
			operator,
			item.ID,
		)
	}
	if err != nil {
		return err
	}

	versionID := fmt.Sprintf("%s_v%d", item.ID, versionNo)
	_, err = tx.Exec(`
INSERT INTO stock_selection_profile_versions (id, profile_id, version_no, snapshot_json, change_note, created_by, created_at)
VALUES (?, ?, ?, ?, ?, ?, NOW())`,
		versionID,
		item.ID,
		versionNo,
		stockSelectionMustJSON(buildStockSelectionProfileSnapshot(item)),
		changeNote,
		operator,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func normalizeStockSelectionProfile(item model.StockSelectionProfile) model.StockSelectionProfile {
	item.Name = strings.TrimSpace(item.Name)
	item.TemplateID = strings.TrimSpace(item.TemplateID)
	item.Status = strings.ToUpper(strings.TrimSpace(item.Status))
	item.SelectionModeDefault = strings.ToUpper(strings.TrimSpace(item.SelectionModeDefault))
	item.UniverseScope = strings.TrimSpace(item.UniverseScope)
	item.Description = strings.TrimSpace(item.Description)
	item.UpdatedBy = strings.TrimSpace(item.UpdatedBy)
	if item.Status == "" {
		item.Status = "ACTIVE"
	}
	if item.SelectionModeDefault == "" {
		item.SelectionModeDefault = model.StrategyEngineStockSelectionModeAuto
	}
	if item.UniverseScope == "" {
		item.UniverseScope = model.StrategyEngineDefaultStockUniverseScope
	}
	if item.UniverseConfig == nil {
		item.UniverseConfig = map[string]any{}
	}
	if item.SeedMiningConfig == nil {
		item.SeedMiningConfig = map[string]any{}
	}
	if item.FactorConfig == nil {
		item.FactorConfig = map[string]any{}
	}
	if item.PortfolioConfig == nil {
		item.PortfolioConfig = map[string]any{}
	}
	if item.PublishConfig == nil {
		item.PublishConfig = map[string]any{}
	}
	if item.MarketAnalysisConfig == nil {
		item.MarketAnalysisConfig = map[string]any{}
	}
	if item.CandidatePoolConfig == nil {
		item.CandidatePoolConfig = map[string]any{}
	}
	if item.ShortTermHeadConfig == nil {
		item.ShortTermHeadConfig = map[string]any{}
	}
	if item.SwingHeadConfig == nil {
		item.SwingHeadConfig = map[string]any{}
	}
	return item
}

func buildStockSelectionProfileSnapshot(item model.StockSelectionProfile) map[string]any {
	return map[string]any{
		"id":                     item.ID,
		"name":                   item.Name,
		"template_id":            item.TemplateID,
		"status":                 item.Status,
		"is_default":             item.IsDefault,
		"selection_mode_default": item.SelectionModeDefault,
		"universe_scope":         item.UniverseScope,
		"universe_config":        item.UniverseConfig,
		"seed_mining_config":     item.SeedMiningConfig,
		"factor_config":          item.FactorConfig,
		"portfolio_config":       item.PortfolioConfig,
		"publish_config":         item.PublishConfig,
		"market_analysis_config": item.MarketAnalysisConfig,
		"candidate_pool_config":  item.CandidatePoolConfig,
		"short_term_head_config": item.ShortTermHeadConfig,
		"swing_head_config":      item.SwingHeadConfig,
		"description":            item.Description,
	}
}

func profileFromSnapshot(snapshot map[string]any) model.StockSelectionProfile {
	item := model.StockSelectionProfile{
		ID:                   stringValue(snapshot["id"]),
		Name:                 stringValue(snapshot["name"]),
		TemplateID:           stringValue(snapshot["template_id"]),
		Status:               strings.ToUpper(stringValue(snapshot["status"])),
		IsDefault:            boolValue(snapshot["is_default"]),
		SelectionModeDefault: strings.ToUpper(stringValue(snapshot["selection_mode_default"])),
		UniverseScope:        stringValue(snapshot["universe_scope"]),
		UniverseConfig:       mapValue(snapshot["universe_config"]),
		SeedMiningConfig:     mapValue(snapshot["seed_mining_config"]),
		FactorConfig:         mapValue(snapshot["factor_config"]),
		PortfolioConfig:      mapValue(snapshot["portfolio_config"]),
		PublishConfig:        mapValue(snapshot["publish_config"]),
		MarketAnalysisConfig: mapValue(snapshot["market_analysis_config"]),
		CandidatePoolConfig:  mapValue(snapshot["candidate_pool_config"]),
		ShortTermHeadConfig:  mapValue(snapshot["short_term_head_config"]),
		SwingHeadConfig:      mapValue(snapshot["swing_head_config"]),
		Description:          stringValue(snapshot["description"]),
	}
	return normalizeStockSelectionProfile(item)
}

func deriveStockSelectionMarketAnalysisConfig(item model.StockSelectionProfile) map[string]any {
	result := mergeStockSelectionConfigMaps(nil, item.MarketAnalysisConfig)
	if len(result) > 0 {
		return result
	}
	result = map[string]any{}
	copyStockSelectionPayloadFields(result, item.FactorConfig, []string{"lookback_days", "trend_bias", "resonance_bias"})
	copyStockSelectionPayloadFields(result, item.PublishConfig, []string{"review_required"})
	return result
}

func deriveStockSelectionCandidatePoolConfig(item model.StockSelectionProfile) map[string]any {
	result := mergeStockSelectionConfigMaps(nil, item.CandidatePoolConfig)
	if len(result) > 0 {
		return result
	}
	result = map[string]any{}
	copyStockSelectionPayloadFields(result, item.UniverseConfig, []string{"min_listing_days", "min_avg_turnover", "price_min", "price_max", "industry_whitelist", "industry_blacklist", "theme_whitelist", "theme_blacklist"})
	copyStockSelectionPayloadFields(result, item.SeedMiningConfig, []string{"bucket_limit", "seed_pool_cap", "candidate_pool_limit"})
	return result
}

func deriveStockSelectionShortTermHeadConfig(item model.StockSelectionProfile) map[string]any {
	result := mergeStockSelectionConfigMaps(nil, item.ShortTermHeadConfig)
	if len(result) > 0 {
		return result
	}
	result = map[string]any{}
	copyStockSelectionPayloadFields(result, item.PortfolioConfig, []string{"limit", "min_score", "max_risk_level", "watchlist_limit"})
	copyStockSelectionPayloadFields(result, item.FactorConfig, []string{"quant_weight", "event_weight", "resonance_weight", "liquidity_risk_weight"})
	return result
}

func deriveStockSelectionSwingHeadConfig(item model.StockSelectionProfile) map[string]any {
	result := mergeStockSelectionConfigMaps(nil, item.SwingHeadConfig)
	if len(result) > 0 {
		return result
	}
	result = map[string]any{}
	copyStockSelectionPayloadFields(result, item.PortfolioConfig, []string{"watchlist_limit", "max_symbol_per_bucket", "max_symbols_per_sector"})
	copyStockSelectionPayloadFields(result, item.FactorConfig, []string{"quant_weight", "resonance_weight"})
	return result
}
