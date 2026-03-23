package repo

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminListFuturesSelectionProfiles(status string, page int, pageSize int) ([]model.FuturesSelectionProfile, int, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	offset := (page - 1) * pageSize

	args := make([]any, 0, 4)
	filter := ""
	if status != "" {
		filter = " WHERE p.status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_selection_profiles p"+filter, args...).Scan(&total); err != nil {
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
  COALESCE(p.style_default, 'balanced'),
  COALESCE(p.contract_scope, ''),
  COALESCE(CAST(p.universe_config AS CHAR), ''),
  COALESCE(CAST(p.factor_config AS CHAR), ''),
  COALESCE(CAST(p.portfolio_config AS CHAR), ''),
  COALESCE(CAST(p.publish_config AS CHAR), ''),
  COALESCE(p.description, ''),
  COALESCE(p.updated_by, ''),
  p.updated_at,
  p.created_at
FROM futures_selection_profiles p
LEFT JOIN futures_selection_profile_templates t ON t.id = p.template_id
LEFT JOIN (
  SELECT profile_id, MAX(version_no) AS current_version
  FROM futures_selection_profile_versions
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

	items := make([]model.FuturesSelectionProfile, 0)
	profileIDs := make([]string, 0)
	for rows.Next() {
		var item model.FuturesSelectionProfile
		var isDefault bool
		var universeConfig, factorConfig, portfolioConfig, publishConfig string
		var updatedAt, createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.TemplateID,
			&item.TemplateName,
			&item.Status,
			&isDefault,
			&item.CurrentVersion,
			&item.StyleDefault,
			&item.ContractScope,
			&universeConfig,
			&factorConfig,
			&portfolioConfig,
			&publishConfig,
			&item.Description,
			&item.UpdatedBy,
			&updatedAt,
			&createdAt,
		); err != nil {
			return nil, 0, err
		}
		item.IsDefault = isDefault
		item.UniverseConfig = parseJSONMap(universeConfig)
		item.FactorConfig = parseJSONMap(factorConfig)
		item.PortfolioConfig = parseJSONMap(portfolioConfig)
		item.PublishConfig = parseJSONMap(publishConfig)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
		profileIDs = append(profileIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	versionMap, err := r.listFuturesSelectionProfileVersions(profileIDs)
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

func (r *MySQLGrowthRepo) AdminListFuturesSelectionProfileVersions(profileID string) ([]model.FuturesSelectionProfileVersion, error) {
	rows, err := r.listFuturesSelectionProfileVersions([]string{strings.TrimSpace(profileID)})
	if err != nil {
		return nil, err
	}
	return rows[strings.TrimSpace(profileID)], nil
}

func (r *MySQLGrowthRepo) AdminCreateFuturesSelectionProfile(item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error) {
	item = normalizeFuturesSelectionProfile(item)
	item.ID = newID("fspf")
	item.CurrentVersion = 1
	if strings.TrimSpace(changeNote) == "" {
		changeNote = "创建期货配置"
	}
	if err := r.upsertFuturesSelectionProfile(item, 1, changeNote, item.UpdatedBy, true); err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	return r.getFuturesSelectionProfile(item.ID)
}

func (r *MySQLGrowthRepo) AdminUpdateFuturesSelectionProfile(id string, item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error) {
	current, err := r.getFuturesSelectionProfile(id)
	if err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	item = normalizeFuturesSelectionProfile(item)
	item.ID = current.ID
	item.CurrentVersion = current.CurrentVersion + 1
	if item.UpdatedBy == "" {
		item.UpdatedBy = current.UpdatedBy
	}
	if strings.TrimSpace(changeNote) == "" {
		changeNote = fmt.Sprintf("更新期货配置到 v%d", item.CurrentVersion)
	}
	if err := r.upsertFuturesSelectionProfile(item, item.CurrentVersion, changeNote, item.UpdatedBy, false); err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	return r.getFuturesSelectionProfile(id)
}

func (r *MySQLGrowthRepo) AdminPublishFuturesSelectionProfile(id string, operator string) (model.FuturesSelectionProfile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.FuturesSelectionProfile{}, sql.ErrNoRows
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.Exec("UPDATE futures_selection_profiles SET is_default = 0, updated_at = NOW() WHERE id <> ?", id); err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	if _, err = tx.Exec("UPDATE futures_selection_profiles SET is_default = 1, updated_by = ?, updated_at = NOW() WHERE id = ?", strings.TrimSpace(operator), id); err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	return r.getFuturesSelectionProfile(id)
}

func (r *MySQLGrowthRepo) AdminRollbackFuturesSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.FuturesSelectionProfile, error) {
	id = strings.TrimSpace(id)
	if id == "" || versionNo <= 0 {
		return model.FuturesSelectionProfile{}, sql.ErrNoRows
	}
	current, err := r.getFuturesSelectionProfile(id)
	if err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	version, err := r.getFuturesSelectionProfileVersion(id, versionNo)
	if err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	restored := normalizeFuturesSelectionProfile(futuresSelectionProfileFromSnapshot(version.Snapshot))
	restored.ID = id
	restored.CurrentVersion = current.CurrentVersion + 1
	restored.UpdatedBy = strings.TrimSpace(operator)
	if strings.TrimSpace(changeNote) == "" {
		changeNote = fmt.Sprintf("从版本 v%d 回滚", versionNo)
	}
	if err := r.upsertFuturesSelectionProfile(restored, restored.CurrentVersion, changeNote, restored.UpdatedBy, false); err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	return r.getFuturesSelectionProfile(id)
}

func (r *MySQLGrowthRepo) AdminListFuturesSelectionProfileTemplates(status string, page int, pageSize int) ([]model.FuturesSelectionProfileTemplate, int, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	offset := (page - 1) * pageSize

	args := make([]any, 0, 4)
	filter := ""
	if status != "" {
		filter = " WHERE status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_selection_profile_templates"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
  id,
  template_key,
  name,
  COALESCE(description, ''),
  COALESCE(market_regime_bias, ''),
  is_default,
  status,
  COALESCE(CAST(universe_defaults_json AS CHAR), ''),
  COALESCE(CAST(factor_defaults_json AS CHAR), ''),
  COALESCE(CAST(portfolio_defaults_json AS CHAR), ''),
  COALESCE(CAST(publish_defaults_json AS CHAR), ''),
  COALESCE(updated_by, ''),
  DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_profile_templates`+filter+`
ORDER BY is_default DESC, updated_at DESC, id ASC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.FuturesSelectionProfileTemplate, 0)
	for rows.Next() {
		var item model.FuturesSelectionProfileTemplate
		var isDefault bool
		var universeJSON, factorJSON, portfolioJSON, publishJSON string
		if err := rows.Scan(
			&item.ID,
			&item.TemplateKey,
			&item.Name,
			&item.Description,
			&item.MarketRegimeBias,
			&isDefault,
			&item.Status,
			&universeJSON,
			&factorJSON,
			&portfolioJSON,
			&publishJSON,
			&item.UpdatedBy,
			&item.UpdatedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		item.IsDefault = isDefault
		item.UniverseDefaults = parseJSONMap(universeJSON)
		item.FactorDefaults = parseJSONMap(factorJSON)
		item.PortfolioDefaults = parseJSONMap(portfolioJSON)
		item.PublishDefaults = parseJSONMap(publishJSON)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminCreateFuturesSelectionProfileTemplate(item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error) {
	item = normalizeFuturesSelectionProfileTemplate(item)
	item.ID = newID("fstpl")
	tx, err := r.db.Begin()
	if err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if item.IsDefault {
		if _, err = tx.Exec("UPDATE futures_selection_profile_templates SET is_default = 0, updated_at = NOW() WHERE id <> ?", item.ID); err != nil {
			return model.FuturesSelectionProfileTemplate{}, err
		}
	}
	_, err = tx.Exec(`
INSERT INTO futures_selection_profile_templates (
  id, template_key, name, description, market_regime_bias, is_default, status,
  universe_defaults_json, factor_defaults_json, portfolio_defaults_json, publish_defaults_json,
  updated_by, updated_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		item.ID,
		item.TemplateKey,
		item.Name,
		item.Description,
		item.MarketRegimeBias,
		item.IsDefault,
		item.Status,
		stockSelectionMustJSON(item.UniverseDefaults),
		stockSelectionMustJSON(item.FactorDefaults),
		stockSelectionMustJSON(item.PortfolioDefaults),
		stockSelectionMustJSON(item.PublishDefaults),
		item.UpdatedBy,
	)
	if err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	return r.getFuturesSelectionProfileTemplate(item.ID)
}

func (r *MySQLGrowthRepo) AdminUpdateFuturesSelectionProfileTemplate(id string, item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error) {
	current, err := r.getFuturesSelectionProfileTemplate(id)
	if err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	item = normalizeFuturesSelectionProfileTemplate(item)
	item.ID = current.ID
	if item.TemplateKey == "" {
		item.TemplateKey = current.TemplateKey
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if item.IsDefault {
		if _, err = tx.Exec("UPDATE futures_selection_profile_templates SET is_default = 0, updated_at = NOW() WHERE id <> ?", item.ID); err != nil {
			return model.FuturesSelectionProfileTemplate{}, err
		}
	}
	_, err = tx.Exec(`
UPDATE futures_selection_profile_templates
SET template_key = ?,
    name = ?,
    description = ?,
    market_regime_bias = ?,
    is_default = ?,
    status = ?,
    universe_defaults_json = ?,
    factor_defaults_json = ?,
    portfolio_defaults_json = ?,
    publish_defaults_json = ?,
    updated_by = ?,
    updated_at = NOW()
WHERE id = ?`,
		item.TemplateKey,
		item.Name,
		item.Description,
		item.MarketRegimeBias,
		item.IsDefault,
		item.Status,
		stockSelectionMustJSON(item.UniverseDefaults),
		stockSelectionMustJSON(item.FactorDefaults),
		stockSelectionMustJSON(item.PortfolioDefaults),
		stockSelectionMustJSON(item.PublishDefaults),
		firstNonEmpty(item.UpdatedBy, current.UpdatedBy),
		item.ID,
	)
	if err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	return r.getFuturesSelectionProfileTemplate(item.ID)
}

func (r *MySQLGrowthRepo) AdminSetDefaultFuturesSelectionProfileTemplate(id string, operator string) (model.FuturesSelectionProfileTemplate, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.FuturesSelectionProfileTemplate{}, sql.ErrNoRows
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.Exec("UPDATE futures_selection_profile_templates SET is_default = 0, updated_at = NOW() WHERE id <> ?", id); err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	if _, err = tx.Exec("UPDATE futures_selection_profile_templates SET is_default = 1, updated_by = ?, updated_at = NOW() WHERE id = ?", strings.TrimSpace(operator), id); err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	return r.getFuturesSelectionProfileTemplate(id)
}

func (r *MySQLGrowthRepo) AdminListFuturesSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.FuturesSelectionEvaluationLeaderboardItem, error) {
	args := make([]any, 0, 4)
	conditions := []string{"e.evaluation_scope = 'PORTFOLIO'"}
	if strings.TrimSpace(templateID) != "" {
		conditions = append(conditions, "r.template_id = ?")
		args = append(args, strings.TrimSpace(templateID))
	}
	if strings.TrimSpace(profileID) != "" {
		conditions = append(conditions, "r.profile_id = ?")
		args = append(args, strings.TrimSpace(profileID))
	}
	if strings.TrimSpace(marketRegime) != "" {
		conditions = append(conditions, "r.market_regime = ?")
		args = append(args, strings.ToUpper(strings.TrimSpace(marketRegime)))
	}

	rows, err := r.db.Query(`
SELECT
  COALESCE(r.template_id, ''),
  COALESCE(t.name, ''),
  r.profile_id,
  COALESCE(p.name, ''),
  COALESCE(r.market_regime, ''),
  r.run_id,
  e.contract,
  e.horizon_day,
  e.return_pct,
  e.hit_flag,
  e.max_drawdown_pct
FROM futures_selection_run_evaluations e
JOIN futures_selection_runs r ON r.run_id = e.run_id
LEFT JOIN futures_selection_profiles p ON p.id = r.profile_id
LEFT JOIN futures_selection_profile_templates t ON t.id = r.template_id
WHERE `+strings.Join(conditions, " AND ")+`
ORDER BY COALESCE(r.template_id, ''), r.profile_id, COALESCE(r.market_regime, ''), e.contract, e.horizon_day`, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return []model.FuturesSelectionEvaluationLeaderboardItem{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	type horizonAggregate struct {
		returnSum float64
		hitSum    float64
		count     int
	}
	type leaderboardKey struct {
		templateID   string
		profileID    string
		marketRegime string
	}
	type leaderboardAggregate struct {
		model.FuturesSelectionEvaluationLeaderboardItem
		horizon    map[int]*horizonAggregate
		sampleKeys map[string]struct{}
	}

	aggregateMap := map[leaderboardKey]*leaderboardAggregate{}
	order := make([]leaderboardKey, 0)
	for rows.Next() {
		var templateIDValue, templateName, profileIDValue, profileName, regime string
		var runID, contract string
		var horizonDay int
		var returnPct, maxDrawdown float64
		var hitFlag bool
		if err := rows.Scan(&templateIDValue, &templateName, &profileIDValue, &profileName, &regime, &runID, &contract, &horizonDay, &returnPct, &hitFlag, &maxDrawdown); err != nil {
			return nil, err
		}
		key := leaderboardKey{templateID: templateIDValue, profileID: profileIDValue, marketRegime: regime}
		entry, ok := aggregateMap[key]
		if !ok {
			entry = &leaderboardAggregate{
				FuturesSelectionEvaluationLeaderboardItem: model.FuturesSelectionEvaluationLeaderboardItem{
					TemplateID:       templateIDValue,
					TemplateName:     templateName,
					ProfileID:        profileIDValue,
					ProfileName:      profileName,
					MarketRegime:     regime,
					ReturnByHorizon:  map[string]float64{},
					HitRateByHorizon: map[string]float64{},
				},
				horizon:    map[int]*horizonAggregate{},
				sampleKeys: map[string]struct{}{},
			}
			aggregateMap[key] = entry
			order = append(order, key)
		}
		if _, exists := entry.horizon[horizonDay]; !exists {
			entry.horizon[horizonDay] = &horizonAggregate{}
		}
		entry.horizon[horizonDay].returnSum += returnPct
		if hitFlag {
			entry.horizon[horizonDay].hitSum += 1
		}
		entry.horizon[horizonDay].count++
		entry.sampleKeys[strings.ToUpper(strings.TrimSpace(runID))+"::"+strings.ToUpper(strings.TrimSpace(contract))] = struct{}{}
		if len(entry.sampleKeys) == 1 || maxDrawdown < entry.MaxDrawdownPct {
			entry.MaxDrawdownPct = maxDrawdown
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]model.FuturesSelectionEvaluationLeaderboardItem, 0, len(order))
	for _, key := range order {
		entry := aggregateMap[key]
		entry.SampleCount = len(entry.sampleKeys)
		for horizonDay, aggregate := range entry.horizon {
			if aggregate.count == 0 {
				continue
			}
			horizonKey := fmt.Sprintf("%d", horizonDay)
			entry.ReturnByHorizon[horizonKey] = aggregate.returnSum / float64(aggregate.count)
			entry.HitRateByHorizon[horizonKey] = aggregate.hitSum / float64(aggregate.count)
		}
		items = append(items, entry.FuturesSelectionEvaluationLeaderboardItem)
	}

	sort.SliceStable(items, func(i, j int) bool {
		left := items[i].ReturnByHorizon["5"]
		right := items[j].ReturnByHorizon["5"]
		if left != right {
			return left > right
		}
		if items[i].TemplateName != items[j].TemplateName {
			return items[i].TemplateName < items[j].TemplateName
		}
		return items[i].ProfileName < items[j].ProfileName
	})
	return items, nil
}

func (r *MySQLGrowthRepo) getFuturesSelectionProfileVersion(profileID string, versionNo int) (model.FuturesSelectionProfileVersion, error) {
	rows, err := r.listFuturesSelectionProfileVersions([]string{strings.TrimSpace(profileID)})
	if err != nil {
		return model.FuturesSelectionProfileVersion{}, err
	}
	for _, item := range rows[strings.TrimSpace(profileID)] {
		if item.VersionNo == versionNo {
			return item, nil
		}
	}
	return model.FuturesSelectionProfileVersion{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) listFuturesSelectionProfileVersions(profileIDs []string) (map[string][]model.FuturesSelectionProfileVersion, error) {
	if len(profileIDs) == 0 {
		return map[string][]model.FuturesSelectionProfileVersion{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(profileIDs)), ",")
	args := make([]any, 0, len(profileIDs))
	for _, id := range profileIDs {
		args = append(args, id)
	}
	query := fmt.Sprintf(`
SELECT id, profile_id, version_no, COALESCE(CAST(snapshot_json AS CHAR), ''), COALESCE(change_note, ''), COALESCE(created_by, ''), created_at
FROM futures_selection_profile_versions
WHERE profile_id IN (%s)
ORDER BY profile_id ASC, version_no DESC`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]model.FuturesSelectionProfileVersion, len(profileIDs))
	for rows.Next() {
		var item model.FuturesSelectionProfileVersion
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

func (r *MySQLGrowthRepo) upsertFuturesSelectionProfile(item model.FuturesSelectionProfile, versionNo int, changeNote string, operator string, createOnly bool) error {
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
		if _, err = tx.Exec("UPDATE futures_selection_profiles SET is_default = 0, updated_at = NOW() WHERE id <> ?", item.ID); err != nil {
			return err
		}
	}

	universeJSON := stockSelectionMustJSON(item.UniverseConfig)
	factorJSON := stockSelectionMustJSON(item.FactorConfig)
	portfolioJSON := stockSelectionMustJSON(item.PortfolioConfig)
	publishJSON := stockSelectionMustJSON(item.PublishConfig)

	if createOnly {
		_, err = tx.Exec(`
INSERT INTO futures_selection_profiles (
  id, name, template_id, status, is_default, style_default, contract_scope,
  universe_config, factor_config, portfolio_config, publish_config,
  description, updated_by, updated_at, created_at
) VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			item.ID,
			item.Name,
			item.TemplateID,
			item.Status,
			item.IsDefault,
			item.StyleDefault,
			item.ContractScope,
			universeJSON,
			factorJSON,
			portfolioJSON,
			publishJSON,
			item.Description,
			operator,
		)
	} else {
		_, err = tx.Exec(`
UPDATE futures_selection_profiles
SET name = ?,
    template_id = NULLIF(?, ''),
    status = ?,
    is_default = ?,
    style_default = ?,
    contract_scope = ?,
    universe_config = ?,
    factor_config = ?,
    portfolio_config = ?,
    publish_config = ?,
    description = ?,
    updated_by = ?,
    updated_at = NOW()
WHERE id = ?`,
			item.Name,
			item.TemplateID,
			item.Status,
			item.IsDefault,
			item.StyleDefault,
			item.ContractScope,
			universeJSON,
			factorJSON,
			portfolioJSON,
			publishJSON,
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
INSERT INTO futures_selection_profile_versions (id, profile_id, version_no, snapshot_json, change_note, created_by, created_at)
VALUES (?, ?, ?, ?, ?, ?, NOW())`,
		versionID,
		item.ID,
		versionNo,
		stockSelectionMustJSON(buildFuturesSelectionProfileSnapshot(item)),
		changeNote,
		operator,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func normalizeFuturesSelectionProfile(item model.FuturesSelectionProfile) model.FuturesSelectionProfile {
	item.Name = strings.TrimSpace(item.Name)
	item.TemplateID = strings.TrimSpace(item.TemplateID)
	item.Status = strings.ToUpper(strings.TrimSpace(item.Status))
	item.StyleDefault = strings.ToLower(strings.TrimSpace(item.StyleDefault))
	item.ContractScope = strings.ToUpper(strings.TrimSpace(item.ContractScope))
	item.Description = strings.TrimSpace(item.Description)
	item.UpdatedBy = strings.TrimSpace(item.UpdatedBy)
	if item.Status == "" {
		item.Status = "ACTIVE"
	}
	if item.StyleDefault == "" {
		item.StyleDefault = "balanced"
	}
	if item.ContractScope == "" {
		item.ContractScope = "DOMINANT_ALL"
	}
	if item.UniverseConfig == nil {
		item.UniverseConfig = map[string]any{}
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
	return item
}

func buildFuturesSelectionProfileSnapshot(item model.FuturesSelectionProfile) map[string]any {
	return map[string]any{
		"id":               item.ID,
		"name":             item.Name,
		"template_id":      item.TemplateID,
		"status":           item.Status,
		"is_default":       item.IsDefault,
		"style_default":    item.StyleDefault,
		"contract_scope":   item.ContractScope,
		"universe_config":  item.UniverseConfig,
		"factor_config":    item.FactorConfig,
		"portfolio_config": item.PortfolioConfig,
		"publish_config":   item.PublishConfig,
		"description":      item.Description,
	}
}

func futuresSelectionProfileFromSnapshot(snapshot map[string]any) model.FuturesSelectionProfile {
	item := model.FuturesSelectionProfile{
		ID:              stringValue(snapshot["id"]),
		Name:            stringValue(snapshot["name"]),
		TemplateID:      stringValue(snapshot["template_id"]),
		Status:          strings.ToUpper(stringValue(snapshot["status"])),
		IsDefault:       boolValue(snapshot["is_default"]),
		StyleDefault:    strings.ToLower(stringValue(snapshot["style_default"])),
		ContractScope:   strings.ToUpper(stringValue(snapshot["contract_scope"])),
		UniverseConfig:  mapValue(snapshot["universe_config"]),
		FactorConfig:    mapValue(snapshot["factor_config"]),
		PortfolioConfig: mapValue(snapshot["portfolio_config"]),
		PublishConfig:   mapValue(snapshot["publish_config"]),
		Description:     stringValue(snapshot["description"]),
	}
	return normalizeFuturesSelectionProfile(item)
}

func normalizeFuturesSelectionProfileTemplate(item model.FuturesSelectionProfileTemplate) model.FuturesSelectionProfileTemplate {
	item.TemplateKey = strings.ToUpper(strings.TrimSpace(item.TemplateKey))
	item.Name = strings.TrimSpace(item.Name)
	item.Description = strings.TrimSpace(item.Description)
	item.MarketRegimeBias = strings.ToUpper(strings.TrimSpace(item.MarketRegimeBias))
	item.Status = strings.ToUpper(strings.TrimSpace(item.Status))
	item.UpdatedBy = strings.TrimSpace(item.UpdatedBy)
	if item.Status == "" {
		item.Status = "ACTIVE"
	}
	if item.UniverseDefaults == nil {
		item.UniverseDefaults = map[string]any{}
	}
	if item.FactorDefaults == nil {
		item.FactorDefaults = map[string]any{}
	}
	if item.PortfolioDefaults == nil {
		item.PortfolioDefaults = map[string]any{}
	}
	if item.PublishDefaults == nil {
		item.PublishDefaults = map[string]any{}
	}
	return item
}
