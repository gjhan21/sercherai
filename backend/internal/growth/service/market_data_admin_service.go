package service

import "sercherai/backend/internal/growth/model"

func (s *growthService) AdminListMarketDataQualityLogs(assetClass string, dataKind string, severity string, issueCode string, hours int, page int, pageSize int) ([]model.MarketDataQualityLog, int, error) {
	return s.repo.AdminListMarketDataQualityLogs(assetClass, dataKind, severity, issueCode, hours, page, pageSize)
}

func (s *growthService) AdminGetMarketDataQualitySummary(assetClass string, hours int) (model.MarketDataQualitySummary, error) {
	return s.repo.AdminGetMarketDataQualitySummary(assetClass, hours)
}

func (s *growthService) AdminGetMarketDerivedTruthSummary(assetClass string) (*model.MarketDerivedTruthSummary, error) {
	return s.repo.AdminGetMarketDerivedTruthSummary(assetClass)
}

func (s *growthService) AdminRebuildMarketDerivedTruth(assetClass string, tradeDate string, days int) (model.MarketDerivedTruthRebuildResult, error) {
	return s.repo.AdminRebuildMarketDerivedTruth(assetClass, tradeDate, days)
}
