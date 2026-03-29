package repo

import (
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

type strategyForecastL3ResearchPack struct {
	TargetType        string
	TargetKey         string
	TargetLabel       string
	CoreThesis        string
	HistoricalNotes   []string
	RelatedHighlights []string
	RiskBoundary      string
	Invalidations     []string
	ActionHints       []string
	L2PrimaryScenario string
	L2ConsensusAction string
	L2Vetoed          bool
	L2VetoReason      string
	EvaluationSummary string
}

type strategyForecastL3RoleResult struct {
	Role       string
	Stance     string
	Confidence float64
	Summary    string
	Veto       bool
}

func buildStrategyForecastL3Report(
	run model.StrategyForecastL3Run,
	pack strategyForecastL3ResearchPack,
	roles []strategyForecastL3RoleResult,
	now time.Time,
) model.StrategyForecastL3Report {
	reportID := newID("l3report")
	targetLabel := firstNonEmpty(pack.TargetLabel, run.TargetLabel, run.TargetKey)
	primaryScenario := resolveStrategyForecastL3PrimaryScenario(run.TargetType, roles, pack)
	executiveSummary := buildStrategyForecastL3ExecutiveSummary(targetLabel, pack, primaryScenario)
	actionGuidance := buildStrategyForecastL3ActionGuidance(pack, primaryScenario)
	triggerChecklist := buildStrategyForecastL3TriggerChecklist(pack)
	roleDisagreements := buildStrategyForecastL3RoleDisagreements(roles)
	alternativeScenarios := buildStrategyForecastL3AlternativeScenarios(run.TargetType, pack)
	invalidationSignals := buildStrategyForecastL3InvalidationSignals(pack)

	report := model.StrategyForecastL3Report{
		ID:                   reportID,
		RunID:                run.ID,
		Version:              1,
		ExecutiveSummary:     executiveSummary,
		PrimaryScenario:      primaryScenario,
		AlternativeScenarios: alternativeScenarios,
		TriggerChecklist:     triggerChecklist,
		InvalidationSignals:  invalidationSignals,
		RoleDisagreements:    roleDisagreements,
		ActionGuidance:       actionGuidance,
		CreatedAt:            now.UTC().Format(time.RFC3339),
		UpdatedAt:            now.UTC().Format(time.RFC3339),
	}
	report.Summary = model.StrategyForecastL3Summary{
		RunID:            run.ID,
		Status:           model.StrategyForecastL3StatusSucceeded,
		EngineKey:        firstNonEmpty(run.EngineKey, model.StrategyForecastL3EngineLocalSynthesis),
		TriggerType:      run.TriggerType,
		TargetType:       run.TargetType,
		TargetKey:        run.TargetKey,
		TargetLabel:      targetLabel,
		ExecutiveSummary: executiveSummary,
		PrimaryScenario:  primaryScenario,
		ActionGuidance:   firstString(actionGuidance),
		ConfidenceLabel:  buildStrategyForecastL3ConfidenceLabel(roles),
		PriorityScore:    run.PriorityScore,
		GeneratedAt:      now.UTC().Format(time.RFC3339),
		ReportAvailable:  true,
	}
	report.MarkdownBody = buildStrategyForecastL3Markdown(run, pack, report)
	report.HTMLBody = buildStrategyForecastL3HTML(run, pack, report)
	return report
}

func resolveStrategyForecastL3PrimaryScenario(targetType string, roles []strategyForecastL3RoleResult, pack strategyForecastL3ResearchPack) string {
	if strings.TrimSpace(pack.L2PrimaryScenario) != "" {
		return strings.TrimSpace(pack.L2PrimaryScenario)
	}
	vetoed := false
	constructiveCount := 0
	for _, role := range roles {
		if role.Veto {
			vetoed = true
		}
		stance := strings.ToUpper(strings.TrimSpace(role.Stance))
		if strings.Contains(stance, "BULL") || strings.Contains(stance, "CONSTRUCTIVE") || strings.Contains(stance, "SUPPORT") {
			constructiveCount++
		}
	}
	if vetoed {
		if strings.EqualFold(targetType, model.StrategyForecastL3TargetTypeFutures) {
			return "reversal"
		}
		return "bear"
	}
	if constructiveCount >= 2 {
		if strings.EqualFold(targetType, model.StrategyForecastL3TargetTypeFutures) {
			return "trend_continue"
		}
		return "bull"
	}
	return "base"
}

func buildStrategyForecastL3ExecutiveSummary(targetLabel string, pack strategyForecastL3ResearchPack, primaryScenario string) string {
	base := strings.TrimSpace(pack.CoreThesis)
	if base == "" {
		base = "Current evidence supports a monitored base-case path."
	}
	if strings.TrimSpace(targetLabel) == "" {
		return fmt.Sprintf("%s Primary scenario: %s.", base, primaryScenario)
	}
	return fmt.Sprintf("%s Primary scenario for %s: %s.", base, targetLabel, primaryScenario)
}

func buildStrategyForecastL3ActionGuidance(pack strategyForecastL3ResearchPack, primaryScenario string) []string {
	guidance := make([]string, 0, 4)
	if len(pack.ActionHints) > 0 {
		guidance = append(guidance, pack.ActionHints...)
	}
	if guidance == nil {
		guidance = []string{}
	}
	if primaryScenario == "bear" || primaryScenario == "reversal" {
		guidance = append(guidance, "Reduce exposure first and wait for confirmation.")
	} else {
		guidance = append(guidance, "Keep position sizing disciplined and wait for confirmation.")
	}
	if trimmed := strings.TrimSpace(pack.RiskBoundary); trimmed != "" {
		guidance = append(guidance, "Risk boundary: "+trimmed)
	}
	return uniqueForecastL3Strings(guidance)
}

func buildStrategyForecastL3TriggerChecklist(pack strategyForecastL3ResearchPack) []model.StrategyForecastL3ChecklistItem {
	items := make([]model.StrategyForecastL3ChecklistItem, 0, len(pack.RelatedHighlights)+1)
	for _, item := range uniqueForecastL3Strings(pack.RelatedHighlights) {
		items = append(items, model.StrategyForecastL3ChecklistItem{
			Label:   item,
			Status:  "WATCH",
			Note:    "Track whether this signal remains aligned with the thesis.",
			Trigger: item,
		})
	}
	if trimmed := strings.TrimSpace(pack.EvaluationSummary); trimmed != "" {
		items = append(items, model.StrategyForecastL3ChecklistItem{
			Label:   "Evaluation feedback",
			Status:  "READY",
			Note:    trimmed,
			Trigger: "Compare post-publish performance with current setup.",
		})
	}
	if len(items) == 0 {
		items = append(items, model.StrategyForecastL3ChecklistItem{
			Label:   "Primary scenario confirmation",
			Status:  "WATCH",
			Note:    "Wait for the main evidence chain to confirm.",
			Trigger: "Confirm price, flow and event alignment.",
		})
	}
	return items
}

func buildStrategyForecastL3RoleDisagreements(roles []strategyForecastL3RoleResult) []model.StrategyForecastL3RoleDisagreement {
	items := make([]model.StrategyForecastL3RoleDisagreement, 0, len(roles))
	for _, role := range roles {
		items = append(items, model.StrategyForecastL3RoleDisagreement{
			Role:    role.Role,
			Stance:  role.Stance,
			Summary: role.Summary,
			Veto:    role.Veto,
		})
	}
	return items
}

func buildStrategyForecastL3AlternativeScenarios(targetType string, pack strategyForecastL3ResearchPack) []model.StrategyForecastL3Scenario {
	if strings.EqualFold(targetType, model.StrategyForecastL3TargetTypeFutures) {
		return []model.StrategyForecastL3Scenario{
			{Name: "trend_continue", Probability: 0.34, Thesis: pack.CoreThesis, Action: "Follow the trend with tighter risk control."},
			{Name: "base", Probability: 0.46, Thesis: "Wait for confirmation from spread, inventory and flow.", Action: "Observe and confirm."},
			{Name: "reversal", Probability: 0.20, Thesis: "Failure of the main evidence chain can force a fast reversal.", Action: "Reduce exposure quickly."},
		}
	}
	return []model.StrategyForecastL3Scenario{
		{Name: "bull", Probability: 0.32, Thesis: pack.CoreThesis, Action: "Add only after confirmation."},
		{Name: "base", Probability: 0.48, Thesis: "Main thesis still needs confirmation from flow and events.", Action: "Hold and verify."},
		{Name: "bear", Probability: 0.20, Thesis: "If the risk boundary breaks, the current setup is invalidated.", Action: "Reduce and reassess."},
	}
}

func buildStrategyForecastL3InvalidationSignals(pack strategyForecastL3ResearchPack) []string {
	signals := uniqueForecastL3Strings(pack.Invalidations)
	if len(signals) == 0 && strings.TrimSpace(pack.RiskBoundary) != "" {
		signals = append(signals, pack.RiskBoundary)
	}
	if len(signals) == 0 {
		signals = []string{"Primary evidence chain fails to confirm."}
	}
	return signals
}

func buildStrategyForecastL3ConfidenceLabel(roles []strategyForecastL3RoleResult) string {
	if len(roles) == 0 {
		return "LOW"
	}
	total := 0.0
	for _, role := range roles {
		total += role.Confidence
	}
	avg := total / float64(len(roles))
	switch {
	case avg >= 0.75:
		return "HIGH"
	case avg >= 0.55:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func buildStrategyForecastL3Markdown(
	run model.StrategyForecastL3Run,
	pack strategyForecastL3ResearchPack,
	report model.StrategyForecastL3Report,
) string {
	lines := []string{
		"# Forecast L3 Report",
		"",
		"## Executive Summary",
		report.ExecutiveSummary,
		"",
		"## Primary Scenario",
		report.PrimaryScenario,
		"",
		"## Trigger Checklist",
	}
	for _, item := range report.TriggerChecklist {
		lines = append(lines, fmt.Sprintf("- %s: %s", item.Label, firstNonEmpty(item.Note, item.Trigger)))
	}
	lines = append(lines, "", "## Invalidation Signals")
	for _, item := range report.InvalidationSignals {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "", "## Role Disagreements")
	for _, item := range report.RoleDisagreements {
		lines = append(lines, fmt.Sprintf("- %s (%s): %s", item.Role, firstNonEmpty(item.Stance, "N/A"), item.Summary))
	}
	lines = append(lines, "", "## Action Guidance")
	for _, item := range report.ActionGuidance {
		lines = append(lines, "- "+item)
	}
	if strings.TrimSpace(pack.RiskBoundary) != "" {
		lines = append(lines, "", "## Risk Boundary", pack.RiskBoundary)
	}
	if strings.TrimSpace(run.Reason) != "" {
		lines = append(lines, "", "## Trigger Reason", run.Reason)
	}
	return strings.Join(lines, "\n")
}

func buildStrategyForecastL3HTML(
	run model.StrategyForecastL3Run,
	pack strategyForecastL3ResearchPack,
	report model.StrategyForecastL3Report,
) string {
	var builder strings.Builder
	builder.WriteString("<h1>Forecast L3 Report</h1>")
	builder.WriteString("<h2>Executive Summary</h2><p>" + htmlEscape(report.ExecutiveSummary) + "</p>")
	builder.WriteString("<h2>Primary Scenario</h2><p>" + htmlEscape(report.PrimaryScenario) + "</p>")
	builder.WriteString("<h2>Trigger Checklist</h2><ul>")
	for _, item := range report.TriggerChecklist {
		builder.WriteString("<li>" + htmlEscape(item.Label) + ": " + htmlEscape(firstNonEmpty(item.Note, item.Trigger)) + "</li>")
	}
	builder.WriteString("</ul><h2>Invalidation Signals</h2><ul>")
	for _, item := range report.InvalidationSignals {
		builder.WriteString("<li>" + htmlEscape(item) + "</li>")
	}
	builder.WriteString("</ul><h2>Role Disagreements</h2><ul>")
	for _, item := range report.RoleDisagreements {
		builder.WriteString("<li>" + htmlEscape(item.Role) + " (" + htmlEscape(firstNonEmpty(item.Stance, "N/A")) + "): " + htmlEscape(item.Summary) + "</li>")
	}
	builder.WriteString("</ul><h2>Action Guidance</h2><ul>")
	for _, item := range report.ActionGuidance {
		builder.WriteString("<li>" + htmlEscape(item) + "</li>")
	}
	builder.WriteString("</ul>")
	if strings.TrimSpace(pack.RiskBoundary) != "" {
		builder.WriteString("<h2>Risk Boundary</h2><p>" + htmlEscape(pack.RiskBoundary) + "</p>")
	}
	if strings.TrimSpace(run.Reason) != "" {
		builder.WriteString("<h2>Trigger Reason</h2><p>" + htmlEscape(run.Reason) + "</p>")
	}
	return builder.String()
}

func uniqueForecastL3Strings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func htmlEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(value)
}

func firstString(items []string) string {
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
