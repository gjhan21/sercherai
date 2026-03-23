package service

import "sercherai/backend/internal/growth/model"

func (s *growthService) AdminListMarketDataQualityLogs(assetClass string, dataKind string, severity string, issueCode string, hours int, page int, pageSize int) ([]model.MarketDataQualityLog, int, error) {
	return s.repo.AdminListMarketDataQualityLogs(assetClass, dataKind, severity, issueCode, hours, page, pageSize)
}

func (s *growthService) AdminGetMarketDataQualitySummary(assetClass string, hours int) (model.MarketDataQualitySummary, error) {
	return s.repo.AdminGetMarketDataQualitySummary(assetClass, hours)
}

func (s *growthService) AdminListMarketProviderCapabilities(providerKey string, assetClass string, dataKind string) ([]model.MarketProviderCapability, error) {
	return s.repo.AdminListMarketProviderCapabilities(providerKey, assetClass, dataKind)
}

func (s *growthService) AdminListMarketProviderRoutingPolicies(assetClass string, dataKind string) ([]model.MarketProviderRoutingPolicy, error) {
	return s.repo.AdminListMarketProviderRoutingPolicies(assetClass, dataKind)
}

func (s *growthService) AdminListMarketProviderQualityScores(assetClass string, dataKind string, hours int) ([]model.MarketProviderQualityScore, error) {
	return s.repo.AdminListMarketProviderQualityScores(assetClass, dataKind, hours)
}

func (s *growthService) AdminGetMarketProviderGovernanceOverview(assetClass string, dataKind string, hours int) (model.MarketProviderGovernanceOverview, error) {
	return s.repo.AdminGetMarketProviderGovernanceOverview(assetClass, dataKind, hours)
}

func (s *growthService) AdminUpsertMarketProviderRoutingPolicy(policyKey string, item model.MarketProviderRoutingPolicy) (model.MarketProviderRoutingPolicy, error) {
	return s.repo.AdminUpsertMarketProviderRoutingPolicy(policyKey, item)
}

func (s *growthService) AdminGetMarketDerivedTruthSummary(assetClass string) (*model.MarketDerivedTruthSummary, error) {
	return s.repo.AdminGetMarketDerivedTruthSummary(assetClass)
}

func (s *growthService) AdminRebuildMarketDerivedTruth(assetClass string, tradeDate string, days int) (model.MarketDerivedTruthRebuildResult, error) {
	return s.repo.AdminRebuildMarketDerivedTruth(assetClass, tradeDate, days)
}
