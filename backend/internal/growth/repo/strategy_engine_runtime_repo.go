package repo

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

const (
	strategyScenarioTemplateConfigPrefix = "strategy_engine.scenario_template."
	strategyPublishPolicyConfigPrefix    = "strategy_engine.publish_policy."
	strategyBuiltInPolicyOperator        = "system-bootstrap"
)

func (r *MySQLGrowthRepo) AdminListStrategyScenarioTemplates(targetType string, status string, page int, pageSize int) ([]model.StrategyScenarioTemplate, int, error) {
	if err := r.ensureStrategyScenarioTemplatesMaterialized(targetType); err != nil {
		return nil, 0, err
	}
	items, err := r.listStrategyScenarioTemplatesAll()
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]model.StrategyScenarioTemplate, 0, len(items))
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
	return paginateStrategyScenarioTemplates(filtered, page, pageSize), len(filtered), nil
}

func (r *MySQLGrowthRepo) AdminCreateStrategyScenarioTemplate(item model.StrategyScenarioTemplate) (string, error) {
	item = normalizeStrategyScenarioTemplate(item)
	item.ID = newStrategyConfigID("sc")
	if err := r.upsertStrategyScenarioTemplate(item, item.UpdatedBy); err != nil {
		return "", err
	}
	return item.ID, nil
}

func (r *MySQLGrowthRepo) AdminUpdateStrategyScenarioTemplate(id string, item model.StrategyScenarioTemplate) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return sql.ErrNoRows
	}
	current, err := r.getStrategyScenarioTemplate(id)
	if err != nil {
		return err
	}
	item = normalizeStrategyScenarioTemplate(item)
	item.ID = id
	if item.UpdatedBy == "" {
		item.UpdatedBy = current.UpdatedBy
	}
	return r.upsertStrategyScenarioTemplate(item, item.UpdatedBy)
}

func (r *MySQLGrowthRepo) AdminListStrategyPublishPolicies(targetType string, status string, page int, pageSize int) ([]model.StrategyPublishPolicy, int, error) {
	if err := r.ensureStrategyPublishPoliciesMaterialized(targetType); err != nil {
		return nil, 0, err
	}
	items, err := r.listStrategyPublishPoliciesAll()
	if err != nil {
		return nil, 0, err
	}
	filtered := make([]model.StrategyPublishPolicy, 0, len(items))
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
	return paginateStrategyPublishPolicies(filtered, page, pageSize), len(filtered), nil
}

func (r *MySQLGrowthRepo) AdminCreateStrategyPublishPolicy(item model.StrategyPublishPolicy) (string, error) {
	item = normalizeStrategyPublishPolicy(item)
	item.ID = newStrategyConfigID("po")
	if err := r.upsertStrategyPublishPolicy(item, item.UpdatedBy); err != nil {
		return "", err
	}
	return item.ID, nil
}

func (r *MySQLGrowthRepo) AdminUpdateStrategyPublishPolicy(id string, item model.StrategyPublishPolicy) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return sql.ErrNoRows
	}
	current, err := r.getStrategyPublishPolicy(id)
	if err != nil {
		return err
	}
	item = normalizeStrategyPublishPolicy(item)
	item.ID = id
	if item.UpdatedBy == "" {
		item.UpdatedBy = current.UpdatedBy
	}
	return r.upsertStrategyPublishPolicy(item, item.UpdatedBy)
}

func (r *MySQLGrowthRepo) ResolveActiveStrategyScenarioTemplate(targetType string) (*model.StrategyScenarioTemplate, error) {
	if err := r.ensureStrategyScenarioTemplatesMaterialized(targetType); err != nil {
		return nil, err
	}
	items, err := r.listStrategyScenarioTemplatesAll()
	if err != nil {
		return nil, err
	}
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	for _, item := range items {
		if item.Status == "ACTIVE" && item.IsDefault && item.TargetType == targetType {
			cloned := item
			return &cloned, nil
		}
	}
	fallback := builtInStrategyScenarioTemplate(targetType)
	if fallback == nil {
		return nil, nil
	}
	cloned := *fallback
	return &cloned, nil
}

func (r *MySQLGrowthRepo) ensureStrategyScenarioTemplatesMaterialized(targetType string) error {
	items, err := r.listStrategyScenarioTemplatesAll()
	if err != nil {
		return err
	}
	activeDefaults := make(map[string]bool, len(items))
	for _, item := range items {
		if item.Status == "ACTIVE" && item.IsDefault {
			activeDefaults[item.TargetType] = true
		}
	}
	for _, requiredTarget := range strategyScenarioTemplateTargetsForBootstrap(targetType) {
		if activeDefaults[requiredTarget] {
			continue
		}
		builtIn := builtInStrategyScenarioTemplate(requiredTarget)
		if builtIn == nil {
			continue
		}
		builtIn.UpdatedBy = strategyBuiltInPolicyOperator
		if err := r.persistStrategyConfig(strategyScenarioTemplateConfigPrefix+builtIn.ID, builtIn, builtIn.Description, strategyBuiltInPolicyOperator); err != nil {
			return err
		}
		activeDefaults[requiredTarget] = true
	}
	return nil
}

func strategyScenarioTemplateTargetsForBootstrap(targetType string) []string {
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	switch targetType {
	case "STOCK":
		return []string{"STOCK"}
	case "FUTURES":
		return []string{"FUTURES"}
	case "", "ALL":
		return []string{"STOCK", "FUTURES"}
	default:
		return []string{"STOCK", "FUTURES"}
	}
}

func builtInStrategyScenarioTemplate(targetType string) *model.StrategyScenarioTemplate {
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	switch targetType {
	case "STOCK":
		return &model.StrategyScenarioTemplate{
			ID:         "scenario_default_stock",
			Name:       "股票四象限模板",
			TargetType: "STOCK",
			Status:     "ACTIVE",
			IsDefault:  true,
			Items: []model.StrategyScenarioTemplateItem{
				{Scenario: "bull", Label: "进攻", ThesisTemplate: "景气扩散与资金跟随，趋势继续强化。", Action: "加仓", RiskSignal: "低", ScoreBias: 0},
				{Scenario: "base", Label: "常态", ThesisTemplate: "维持当前节奏，等待下一轮验证。", Action: "持有", RiskSignal: "中", ScoreBias: 0},
				{Scenario: "bear", Label: "收缩", ThesisTemplate: "市场回撤导致估值与情绪压缩。", Action: "减仓", RiskSignal: "中高", ScoreBias: 0},
				{Scenario: "shock", Label: "防守", ThesisTemplate: "黑天鹅或流动性冲击下先保命。", Action: "回避", RiskSignal: "高", ScoreBias: 0},
			},
			Description: "系统自动落库的股票默认场景模板，可在后台继续编辑与审计。",
			UpdatedBy:   "system",
		}
	case "FUTURES":
		return &model.StrategyScenarioTemplate{
			ID:         "scenario_default_futures",
			Name:       "期货六世界模板",
			TargetType: "FUTURES",
			Status:     "ACTIVE",
			IsDefault:  true,
			Items: []model.StrategyScenarioTemplateItem{
				{Scenario: "base", Label: "常态", ThesisTemplate: "主线逻辑延续但尚未出现强加速，继续跟踪基差、库存与持仓确认。", Action: "观察", RiskSignal: "中", ScoreBias: 0},
				{Scenario: "trend_continue", Label: "趋势延续", ThesisTemplate: "趋势、持仓与成交同步确认，顺势策略仍可延展。", Action: "顺势跟进", RiskSignal: "中低", ScoreBias: 0},
				{Scenario: "policy_positive", Label: "政策利多", ThesisTemplate: "政策边际改善强化需求或风险偏好，优先验证资金是否跟随。", Action: "择机加仓", RiskSignal: "中", ScoreBias: 0},
				{Scenario: "policy_negative", Label: "政策压制", ThesisTemplate: "监管或政策扰动压制预期，需要收缩仓位并重估胜率。", Action: "降杠杆", RiskSignal: "中高", ScoreBias: 0},
				{Scenario: "supply_shock", Label: "供给冲击", ThesisTemplate: "供给侧扰动抬升波动，关注库存、升贴水和跨品种传导。", Action: "事件驱动", RiskSignal: "高", ScoreBias: 0},
				{Scenario: "liquidity_shock", Label: "流动性冲击", ThesisTemplate: "流动性恶化或风险偏好骤降，优先控制回撤与滑点风险。", Action: "防守回避", RiskSignal: "高", ScoreBias: 0},
			},
			Description: "系统自动落库的期货默认场景模板，可在后台继续编辑与审计。",
			UpdatedBy:   "system",
		}
	default:
		return nil
	}
}

func (r *MySQLGrowthRepo) ResolveActiveStrategyPublishPolicy(targetType string) (*model.StrategyPublishPolicy, error) {
	if err := r.ensureStrategyPublishPoliciesMaterialized(targetType); err != nil {
		return nil, err
	}
	items, err := r.listStrategyPublishPoliciesAll()
	if err != nil {
		return nil, err
	}
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	for _, item := range items {
		if item.Status == "ACTIVE" && item.IsDefault && item.TargetType == targetType {
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
	fallback := builtInStrategyPublishPolicy(targetType)
	if fallback == nil {
		return nil, nil
	}
	cloned := *fallback
	return &cloned, nil
}

func (r *MySQLGrowthRepo) ensureStrategyPublishPoliciesMaterialized(targetType string) error {
	items, err := r.listStrategyPublishPoliciesAll()
	if err != nil {
		return err
	}
	activeDefaults := make(map[string]bool, len(items))
	for _, item := range items {
		if item.Status == "ACTIVE" && item.IsDefault {
			activeDefaults[item.TargetType] = true
		}
	}
	for _, requiredTarget := range strategyPublishPolicyTargetsForBootstrap(targetType) {
		if activeDefaults[requiredTarget] {
			continue
		}
		builtIn := builtInStrategyPublishPolicy(requiredTarget)
		if builtIn == nil {
			continue
		}
		builtIn.UpdatedBy = strategyBuiltInPolicyOperator
		if err := r.persistStrategyConfig(strategyPublishPolicyConfigPrefix+builtIn.ID, builtIn, builtIn.Description, strategyBuiltInPolicyOperator); err != nil {
			return err
		}
		activeDefaults[requiredTarget] = true
	}
	return nil
}

func strategyPublishPolicyTargetsForBootstrap(targetType string) []string {
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

func builtInStrategyPublishPolicy(targetType string) *model.StrategyPublishPolicy {
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	switch targetType {
	case "FUTURES":
		return &model.StrategyPublishPolicy{
			ID:                   "policy_default_futures",
			Name:                 "期货默认发布门槛",
			TargetType:           "FUTURES",
			Status:               "ACTIVE",
			IsDefault:            true,
			MaxRiskLevel:         "HIGH",
			MaxWarningCount:      5,
			AllowVetoedPublish:   true,
			DefaultPublisher:     "strategy-engine",
			OverrideNoteTemplate: "期货策略默认允许带 veto 与 warning 发布，但必须保留完整风控说明与复盘链路。",
			Description:          "系统自动落库的期货默认发布策略，可在后台继续编辑与审计。",
			UpdatedBy:            "system",
		}
	case "", "ALL", "STOCK":
		return &model.StrategyPublishPolicy{
			ID:                   "policy_default_all",
			Name:                 "默认发布门槛",
			TargetType:           "ALL",
			Status:               "ACTIVE",
			IsDefault:            true,
			MaxRiskLevel:         "MEDIUM",
			MaxWarningCount:      3,
			AllowVetoedPublish:   false,
			DefaultPublisher:     "strategy-engine",
			OverrideNoteTemplate: "人工覆盖发布，需记录原因与复盘结论。",
			Description:          "系统自动落库的默认发布策略，可在后台继续编辑与审计。",
			UpdatedBy:            "system",
		}
	default:
		return nil
	}
}

func (r *MySQLGrowthRepo) listStrategyScenarioTemplatesAll() ([]model.StrategyScenarioTemplate, error) {
	rows, err := r.db.Query(`
SELECT config_key, config_value, description, updated_by, updated_at
FROM system_configs
WHERE config_key LIKE ?
ORDER BY updated_at DESC, config_key DESC`, strategyScenarioTemplateConfigPrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyScenarioTemplate, 0)
	for rows.Next() {
		var configKey string
		var configValue string
		var description sql.NullString
		var updatedBy sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&configKey, &configValue, &description, &updatedBy, &updatedAt); err != nil {
			return nil, err
		}
		var item model.StrategyScenarioTemplate
		if err := json.Unmarshal([]byte(configValue), &item); err != nil {
			continue
		}
		if item.ID == "" {
			item.ID = strings.TrimPrefix(configKey, strategyScenarioTemplateConfigPrefix)
		}
		item = normalizeStrategyScenarioTemplate(item)
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

func (r *MySQLGrowthRepo) listStrategyPublishPoliciesAll() ([]model.StrategyPublishPolicy, error) {
	rows, err := r.db.Query(`
SELECT config_key, config_value, description, updated_by, updated_at
FROM system_configs
WHERE config_key LIKE ?
ORDER BY updated_at DESC, config_key DESC`, strategyPublishPolicyConfigPrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyPublishPolicy, 0)
	for rows.Next() {
		var configKey string
		var configValue string
		var description sql.NullString
		var updatedBy sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&configKey, &configValue, &description, &updatedBy, &updatedAt); err != nil {
			return nil, err
		}
		var item model.StrategyPublishPolicy
		if err := json.Unmarshal([]byte(configValue), &item); err != nil {
			continue
		}
		if item.ID == "" {
			item.ID = strings.TrimPrefix(configKey, strategyPublishPolicyConfigPrefix)
		}
		item = normalizeStrategyPublishPolicy(item)
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

func (r *MySQLGrowthRepo) getStrategyScenarioTemplate(id string) (model.StrategyScenarioTemplate, error) {
	items, err := r.listStrategyScenarioTemplatesAll()
	if err != nil {
		return model.StrategyScenarioTemplate{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return model.StrategyScenarioTemplate{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) getStrategyPublishPolicy(id string) (model.StrategyPublishPolicy, error) {
	items, err := r.listStrategyPublishPoliciesAll()
	if err != nil {
		return model.StrategyPublishPolicy{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return model.StrategyPublishPolicy{}, sql.ErrNoRows
}

func (r *MySQLGrowthRepo) upsertStrategyScenarioTemplate(item model.StrategyScenarioTemplate, operator string) error {
	item = normalizeStrategyScenarioTemplate(item)
	if item.IsDefault {
		items, err := r.listStrategyScenarioTemplatesAll()
		if err != nil {
			return err
		}
		for _, existing := range items {
			if existing.ID == item.ID || existing.TargetType != item.TargetType || !existing.IsDefault {
				continue
			}
			existing.IsDefault = false
			if err := r.persistStrategyConfig(strategyScenarioTemplateConfigPrefix+existing.ID, existing, existing.Description, operator); err != nil {
				return err
			}
		}
	}
	return r.persistStrategyConfig(strategyScenarioTemplateConfigPrefix+item.ID, item, item.Description, operator)
}

func (r *MySQLGrowthRepo) upsertStrategyPublishPolicy(item model.StrategyPublishPolicy, operator string) error {
	item = normalizeStrategyPublishPolicy(item)
	if item.IsDefault {
		items, err := r.listStrategyPublishPoliciesAll()
		if err != nil {
			return err
		}
		for _, existing := range items {
			if existing.ID == item.ID || existing.TargetType != item.TargetType || !existing.IsDefault {
				continue
			}
			existing.IsDefault = false
			if err := r.persistStrategyConfig(strategyPublishPolicyConfigPrefix+existing.ID, existing, existing.Description, operator); err != nil {
				return err
			}
		}
	}
	return r.persistStrategyConfig(strategyPublishPolicyConfigPrefix+item.ID, item, item.Description, operator)
}

func normalizeStrategyScenarioTemplate(item model.StrategyScenarioTemplate) model.StrategyScenarioTemplate {
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
	items := make([]model.StrategyScenarioTemplateItem, 0, len(item.Items))
	for _, raw := range item.Items {
		raw.Scenario = strings.ToLower(strings.TrimSpace(raw.Scenario))
		if raw.Scenario == "" {
			continue
		}
		if _, ok := seen[raw.Scenario]; ok {
			continue
		}
		seen[raw.Scenario] = struct{}{}
		raw.Label = strings.TrimSpace(raw.Label)
		if raw.Label == "" {
			raw.Label = strings.ToUpper(raw.Scenario)
		}
		raw.ThesisTemplate = strings.TrimSpace(raw.ThesisTemplate)
		raw.Action = strings.TrimSpace(raw.Action)
		raw.RiskSignal = strings.TrimSpace(raw.RiskSignal)
		items = append(items, raw)
	}
	item.Items = items
	return item
}

func normalizeStrategyPublishPolicy(item model.StrategyPublishPolicy) model.StrategyPublishPolicy {
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
	item.MaxRiskLevel = strings.ToUpper(strings.TrimSpace(item.MaxRiskLevel))
	if item.MaxRiskLevel == "" {
		item.MaxRiskLevel = "MEDIUM"
	}
	if item.MaxWarningCount < 0 {
		item.MaxWarningCount = 0
	}
	item.DefaultPublisher = strings.TrimSpace(item.DefaultPublisher)
	if item.DefaultPublisher == "" {
		item.DefaultPublisher = "strategy-engine"
	}
	item.OverrideNoteTemplate = strings.TrimSpace(item.OverrideNoteTemplate)
	item.Description = strings.TrimSpace(item.Description)
	return item
}

func paginateStrategyScenarioTemplates(items []model.StrategyScenarioTemplate, page int, pageSize int) []model.StrategyScenarioTemplate {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []model.StrategyScenarioTemplate{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func paginateStrategyPublishPolicies(items []model.StrategyPublishPolicy, page int, pageSize int) []model.StrategyPublishPolicy {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []model.StrategyPublishPolicy{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}
