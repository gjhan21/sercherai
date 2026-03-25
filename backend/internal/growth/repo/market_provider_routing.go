package repo

import (
	"fmt"
	"strings"

	"sercherai/backend/internal/growth/model"
)

type marketRoutingCapability string

const (
	marketRoutingCapabilitySync               marketRoutingCapability = "SYNC"
	marketRoutingCapabilityContextSeed        marketRoutingCapability = "CONTEXT_SEED"
	marketRoutingCapabilityMetadataEnrichment marketRoutingCapability = "METADATA_ENRICHMENT"
)

type marketProviderRoutingDecision struct {
	PolicyKey          string
	SelectedSource     string
	ResolvedSourceKeys []string
	FallbackChain      []string
	DecisionReason     string
}

func (r *MySQLGrowthRepo) resolveMarketProviderRoutingDecision(requestedSourceKey string, assetClass string, dataKind string, defaultPriority []string, capability marketRoutingCapability) marketProviderRoutingDecision {
	policy, capabilities := r.loadMarketProviderRoutingInputs(assetClass, dataKind)
	return buildMarketProviderRoutingDecision(requestedSourceKey, assetClass, dataKind, defaultPriority, policy, capabilities, capability)
}

func (r *MySQLGrowthRepo) loadMarketProviderRoutingInputs(assetClass string, dataKind string) (*model.MarketProviderRoutingPolicy, []model.MarketProviderCapability) {
	policies, err := r.AdminListMarketProviderRoutingPolicies(assetClass, dataKind)
	if err != nil {
		return nil, nil
	}
	var policy *model.MarketProviderRoutingPolicy
	for _, item := range policies {
		if normalizeUpper(item.PolicyKey) != buildRoutingPolicyKey(assetClass, dataKind) {
			continue
		}
		if normalizeUpper(item.Status) != "ACTIVE" {
			continue
		}
		cloned := item
		policy = &cloned
		break
	}
	capabilities, err := r.AdminListMarketProviderCapabilities("", assetClass, dataKind)
	if err != nil {
		return policy, nil
	}
	return policy, capabilities
}

func buildMarketProviderRoutingDecision(requestedSourceKey string, assetClass string, dataKind string, defaultPriority []string, policy *model.MarketProviderRoutingPolicy, capabilities []model.MarketProviderCapability, capability marketRoutingCapability) marketProviderRoutingDecision {
	policyKey := buildRoutingPolicyKey(assetClass, dataKind)
	if policy != nil && strings.TrimSpace(policy.PolicyKey) != "" {
		policyKey = normalizeUpper(policy.PolicyKey)
	}

	candidateSources, selectionMode := buildMarketProviderRoutingCandidates(requestedSourceKey, defaultPriority, policy)
	capabilityMap := buildMarketCapabilityLookup(assetClass, dataKind, capabilities)
	hasCapabilityRows := len(capabilityMap) > 0
	seen := make(map[string]struct{}, len(candidateSources))
	resolved := make([]string, 0, len(candidateSources))
	skippedReasons := make([]string, 0)

	for _, sourceKey := range candidateSources {
		normalizedSource := normalizeUpper(sourceKey)
		if normalizedSource == "" {
			continue
		}
		if _, exists := seen[normalizedSource]; exists {
			continue
		}
		seen[normalizedSource] = struct{}{}
		if normalizedSource == "MOCK" && !marketRoutingAllowsMock(requestedSourceKey, policy) {
			skippedReasons = append(skippedReasons, "MOCK blocked by routing policy")
			continue
		}
		if hasCapabilityRows {
			if item, ok := capabilityMap[normalizedSource]; ok && !marketProviderSupportsCapability(item, capability) {
				skippedReasons = append(skippedReasons, fmt.Sprintf("%s skipped by capability", normalizedSource))
				continue
			}
		}
		resolved = append(resolved, normalizedSource)
	}

	if len(resolved) == 0 {
		resolved = fallbackFirstSource(candidateSources, defaultPriority, policy)
	}

	decision := marketProviderRoutingDecision{
		PolicyKey:          policyKey,
		ResolvedSourceKeys: resolved,
	}
	if len(resolved) > 0 {
		decision.SelectedSource = resolved[0]
	}
	if len(resolved) > 1 {
		decision.FallbackChain = append([]string(nil), resolved[1:]...)
	}
	decision.DecisionReason = buildMarketRoutingDecisionReason(selectionMode, decision.SelectedSource, policy, skippedReasons)
	return decision
}

func buildMarketProviderRoutingCandidates(requestedSourceKey string, defaultPriority []string, policy *model.MarketProviderRoutingPolicy) ([]string, string) {
	normalizedRequested := normalizeUpper(requestedSourceKey)
	switch normalizedRequested {
	case "", "AUTO":
		if policy != nil && normalizeUpper(policy.PrimaryProviderKey) != "" {
			items := []string{policy.PrimaryProviderKey}
			if policy.FallbackAllowed {
				items = append(items, policy.BackupProviderKeys...)
			}
			return normalizeUpperStringSlice(items), "policy"
		}
		return normalizeUpperStringSlice(defaultPriority), "default"
	default:
		return splitSourcePriorityList(normalizedRequested), "explicit"
	}
}

func buildMarketCapabilityLookup(assetClass string, dataKind string, capabilities []model.MarketProviderCapability) map[string]model.MarketProviderCapability {
	result := make(map[string]model.MarketProviderCapability, len(capabilities))
	for _, item := range capabilities {
		if normalizeUpper(item.AssetClass) != normalizeUpper(assetClass) || normalizeUpper(item.DataKind) != normalizeUpper(dataKind) {
			continue
		}
		result[normalizeUpper(item.ProviderKey)] = item
	}
	return result
}

func marketProviderSupportsCapability(item model.MarketProviderCapability, capability marketRoutingCapability) bool {
	switch capability {
	case marketRoutingCapabilityContextSeed:
		return item.SupportsContextSeed
	case marketRoutingCapabilityMetadataEnrichment:
		return item.SupportsMetadataEnrichment
	default:
		return item.SupportsSync
	}
}

func marketRoutingAllowsMock(requestedSourceKey string, policy *model.MarketProviderRoutingPolicy) bool {
	if strings.Contains(normalizeUpper(requestedSourceKey), "MOCK") {
		return true
	}
	return policy != nil && policy.MockAllowed
}

func fallbackFirstSource(candidateSources []string, defaultPriority []string, policy *model.MarketProviderRoutingPolicy) []string {
	if len(candidateSources) > 0 {
		for _, item := range candidateSources {
			if normalized := normalizeUpper(item); normalized != "" {
				return []string{normalized}
			}
		}
	}
	if policy != nil && normalizeUpper(policy.PrimaryProviderKey) != "" {
		return []string{normalizeUpper(policy.PrimaryProviderKey)}
	}
	for _, item := range defaultPriority {
		if normalized := normalizeUpper(item); normalized != "" {
			return []string{normalized}
		}
	}
	return nil
}

func buildMarketRoutingDecisionReason(selectionMode string, selectedSource string, policy *model.MarketProviderRoutingPolicy, skippedReasons []string) string {
	reason := ""
	switch selectionMode {
	case "explicit":
		reason = fmt.Sprintf("explicit source override selected %s", firstNonEmpty(selectedSource, "none"))
	case "policy":
		primary := ""
		if policy != nil {
			primary = normalizeUpper(policy.PrimaryProviderKey)
		}
		switch {
		case selectedSource == "":
			reason = fmt.Sprintf("policy %s yielded no eligible provider", firstNonEmpty(primary, "DEFAULT"))
		case primary != "" && primary != selectedSource:
			reason = fmt.Sprintf("policy fallback selected %s after %s was filtered", selectedSource, primary)
		default:
			reason = fmt.Sprintf("policy primary provider %s", firstNonEmpty(selectedSource, primary))
		}
	default:
		reason = fmt.Sprintf("default priority selected %s", firstNonEmpty(selectedSource, "none"))
	}
	if len(skippedReasons) > 0 {
		reason += "; " + strings.Join(skippedReasons, "; ")
	}
	return reason
}

func finalizeMarketRoutingSummary(decision marketProviderRoutingDecision, results []model.MarketSourceSyncItemResult) marketProviderRoutingDecision {
	selectedIndex := 0
	for index, item := range results {
		if normalizeUpper(item.Status) != "SUCCESS" {
			continue
		}
		sourceKey := normalizeUpper(item.SourceKey)
		if sourceKey == "" {
			continue
		}
		decision.SelectedSource = sourceKey
		selectedIndex = index
		if index > 0 {
			decision.DecisionReason = strings.TrimSpace(decision.DecisionReason + fmt.Sprintf("; runtime fallback selected %s", sourceKey))
		}
		break
	}
	if len(decision.ResolvedSourceKeys) > 0 {
		indexInRoute := indexOfString(decision.ResolvedSourceKeys, decision.SelectedSource)
		if indexInRoute >= 0 && indexInRoute+1 < len(decision.ResolvedSourceKeys) {
			decision.FallbackChain = append([]string(nil), decision.ResolvedSourceKeys[indexInRoute+1:]...)
		} else if indexInRoute >= 0 {
			decision.FallbackChain = nil
		}
	}
	_ = selectedIndex
	return decision
}

func indexOfString(items []string, target string) int {
	for index, item := range items {
		if normalizeUpper(item) == normalizeUpper(target) {
			return index
		}
	}
	return -1
}

func buildStrategyContextRoutingSummary(assetClass string, dataKind string, selectedSource string) marketProviderRoutingDecision {
	normalizedSource := normalizeUpper(selectedSource)
	reason := "context truth source unavailable"
	if normalizedSource != "" {
		reason = fmt.Sprintf("context truth source selected %s", normalizedSource)
	}
	return marketProviderRoutingDecision{
		PolicyKey:      buildRoutingPolicyKey(assetClass, dataKind),
		SelectedSource: normalizedSource,
		DecisionReason: reason,
	}
}
