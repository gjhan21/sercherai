package repo

import (
	"database/sql"
	"strings"

	"sercherai/backend/internal/growth/model"
)

var inMemoryStrategySeedSets = []model.StrategySeedSet{
	{
		ID:          "seed_default_stock",
		Name:        "核心蓝筹观察池",
		TargetType:  "STOCK",
		Status:      "ACTIVE",
		IsDefault:   true,
		Items:       []string{"600519.SH", "601318.SH", "300750.SZ", "600036.SH"},
		Description: "默认股票种子池",
		UpdatedBy:   "system",
		UpdatedAt:   "2026-03-17T09:00:00+08:00",
	},
	{
		ID:          "seed_default_futures",
		Name:        "核心期货观察池",
		TargetType:  "FUTURES",
		Status:      "ACTIVE",
		IsDefault:   true,
		Items:       []string{"IF2606", "IH2606", "IC2606"},
		Description: "默认期货种子池",
		UpdatedBy:   "system",
		UpdatedAt:   "2026-03-17T09:00:00+08:00",
	},
}

var inMemoryStrategyAgentProfiles = []model.StrategyAgentProfile{
	{
		ID:                "agent_default_all",
		Name:              "默认五角评审",
		TargetType:        "ALL",
		Status:            "ACTIVE",
		IsDefault:         true,
		EnabledAgents:     []string{"trend", "event", "liquidity", "risk", "basis"},
		PositiveThreshold: 3,
		NegativeThreshold: 2,
		AllowVeto:         true,
		Description:       "默认多角色评审配置",
		UpdatedBy:         "system",
		UpdatedAt:         "2026-03-17T09:00:00+08:00",
	},
	{
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
		Description:                     "期货默认角色配置，允许短历史时受控回退 MOCK",
		UpdatedBy:                       "system",
		UpdatedAt:                       "2026-03-20T09:00:00+08:00",
	},
}

var inMemoryStrategyScenarioTemplates = []model.StrategyScenarioTemplate{
	{
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
		Description: "默认股票场景模板",
		UpdatedBy:   "system",
		UpdatedAt:   "2026-03-18T09:00:00+08:00",
	},
	{
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
		Description: "默认期货场景模板",
		UpdatedBy:   "system",
		UpdatedAt:   "2026-03-22T09:00:00+08:00",
	},
}

var inMemoryStrategyPublishPolicies = []model.StrategyPublishPolicy{
	{
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
		Description:          "默认发布策略",
		UpdatedBy:            "system",
		UpdatedAt:            "2026-03-18T09:00:00+08:00",
	},
	{
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
		Description:          "期货默认发布策略",
		UpdatedBy:            "system",
		UpdatedAt:            "2026-03-19T09:00:00+08:00",
	},
}

func (r *InMemoryGrowthRepo) AdminListStrategySeedSets(targetType string, status string, page int, pageSize int) ([]model.StrategySeedSet, int, error) {
	filtered := make([]model.StrategySeedSet, 0, len(inMemoryStrategySeedSets))
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	status = strings.ToUpper(strings.TrimSpace(status))
	for _, item := range inMemoryStrategySeedSets {
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

func (r *InMemoryGrowthRepo) AdminCreateStrategySeedSet(item model.StrategySeedSet) (string, error) {
	item = normalizeStrategySeedSet(item)
	item.ID = "seed_preview"
	item.UpdatedBy = "inmemory"
	item.UpdatedAt = "2026-03-17T09:00:00+08:00"
	return item.ID, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateStrategySeedSet(id string, item model.StrategySeedSet) error {
	if strings.TrimSpace(id) == "" {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminListStrategyAgentProfiles(targetType string, status string, page int, pageSize int) ([]model.StrategyAgentProfile, int, error) {
	filtered := make([]model.StrategyAgentProfile, 0, len(inMemoryStrategyAgentProfiles))
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	status = strings.ToUpper(strings.TrimSpace(status))
	for _, item := range inMemoryStrategyAgentProfiles {
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

func (r *InMemoryGrowthRepo) AdminCreateStrategyAgentProfile(item model.StrategyAgentProfile) (string, error) {
	item = normalizeStrategyAgentProfile(item)
	item.ID = "agent_preview"
	item.UpdatedBy = "inmemory"
	item.UpdatedAt = "2026-03-17T09:00:00+08:00"
	return item.ID, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateStrategyAgentProfile(id string, item model.StrategyAgentProfile) error {
	if strings.TrimSpace(id) == "" {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminListStrategyScenarioTemplates(targetType string, status string, page int, pageSize int) ([]model.StrategyScenarioTemplate, int, error) {
	filtered := make([]model.StrategyScenarioTemplate, 0, len(inMemoryStrategyScenarioTemplates))
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	status = strings.ToUpper(strings.TrimSpace(status))
	for _, item := range inMemoryStrategyScenarioTemplates {
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

func (r *InMemoryGrowthRepo) AdminCreateStrategyScenarioTemplate(item model.StrategyScenarioTemplate) (string, error) {
	item = normalizeStrategyScenarioTemplate(item)
	item.ID = "scenario_preview"
	item.UpdatedBy = "inmemory"
	item.UpdatedAt = "2026-03-18T09:00:00+08:00"
	return item.ID, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateStrategyScenarioTemplate(id string, item model.StrategyScenarioTemplate) error {
	if strings.TrimSpace(id) == "" {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminListStrategyPublishPolicies(targetType string, status string, page int, pageSize int) ([]model.StrategyPublishPolicy, int, error) {
	filtered := make([]model.StrategyPublishPolicy, 0, len(inMemoryStrategyPublishPolicies))
	targetType = strings.ToUpper(strings.TrimSpace(targetType))
	status = strings.ToUpper(strings.TrimSpace(status))
	for _, item := range inMemoryStrategyPublishPolicies {
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

func (r *InMemoryGrowthRepo) AdminCreateStrategyPublishPolicy(item model.StrategyPublishPolicy) (string, error) {
	item = normalizeStrategyPublishPolicy(item)
	item.ID = "publish_policy_preview"
	item.UpdatedBy = "inmemory"
	item.UpdatedAt = "2026-03-18T09:00:00+08:00"
	return item.ID, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateStrategyPublishPolicy(id string, item model.StrategyPublishPolicy) error {
	if strings.TrimSpace(id) == "" {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminListStrategyEngineJobs(jobType string, status string, page int, pageSize int) ([]model.StrategyEngineJobRecord, int, error) {
	return []model.StrategyEngineJobRecord{}, 0, nil
}

func (r *InMemoryGrowthRepo) AdminGetStrategyEngineJob(jobID string) (model.StrategyEngineJobRecord, error) {
	if strings.TrimSpace(jobID) == "" {
		return model.StrategyEngineJobRecord{}, sql.ErrNoRows
	}
	return model.StrategyEngineJobRecord{}, sql.ErrNoRows
}

func (r *InMemoryGrowthRepo) AdminPublishStrategyEngineJob(jobID string, operator string, force bool, overrideReason string) (model.StrategyEnginePublishRecord, error) {
	if strings.TrimSpace(jobID) == "" {
		return model.StrategyEnginePublishRecord{}, sql.ErrNoRows
	}
	return model.StrategyEnginePublishRecord{
		PublishID:     "pub_demo_001",
		JobID:         strings.TrimSpace(jobID),
		JobType:       "stock-selection",
		Version:       1,
		CreatedAt:     "2026-03-24T09:30:00+08:00",
		TradeDate:     "2026-03-24",
		ReportSummary: "demo publish",
		SelectedCount: 3,
		PayloadCount:  3,
		Replay: model.StrategyEnginePublishReplay{
			PublishID:       "pub_demo_001",
			JobID:           strings.TrimSpace(jobID),
			PublishVersion:  1,
			Operator:        strings.TrimSpace(operator),
			ForcePublish:    force,
			OverrideReason:  strings.TrimSpace(overrideReason),
			WarningCount:    1,
			StorageSource:   "LOCAL_ONLY",
			WarningMessages: []string{"demo warning"},
		},
	}, nil
}
