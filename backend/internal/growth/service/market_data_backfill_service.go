package service

import "sercherai/backend/internal/growth/model"

func (s *growthService) AdminBuildMarketUniverseSnapshot(sourceKey string, assetScope []string, operator string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error) {
	return s.repo.AdminBuildMarketUniverseSnapshot(sourceKey, assetScope, operator)
}

func (s *growthService) AdminCreateMarketDataBackfillRun(input model.MarketBackfillCreateInput, operator string) (model.MarketBackfillRun, error) {
	return s.repo.AdminCreateMarketDataBackfillRun(input, operator)
}

func (s *growthService) AdminListMarketDataBackfillRuns(status string, runType string, assetType string, sourceKey string, page int, pageSize int) ([]model.MarketBackfillRun, int, error) {
	return s.repo.AdminListMarketDataBackfillRuns(status, runType, assetType, sourceKey, page, pageSize)
}

func (s *growthService) AdminGetMarketDataBackfillRun(id string) (model.MarketBackfillRun, error) {
	return s.repo.AdminGetMarketDataBackfillRun(id)
}

func (s *growthService) AdminListMarketDataBackfillRunDetails(runID string, stage string, assetType string, status string, page int, pageSize int) ([]model.MarketBackfillRunDetail, int, error) {
	return s.repo.AdminListMarketDataBackfillRunDetails(runID, stage, assetType, status, page, pageSize)
}

func (s *growthService) AdminRetryMarketDataBackfillRun(runID string, input model.MarketBackfillRetryInput, operator string) (model.MarketBackfillRun, error) {
	return s.repo.AdminRetryMarketDataBackfillRun(runID, input, operator)
}

func (s *growthService) AdminListMarketUniverseSnapshots(page int, pageSize int) ([]model.MarketUniverseSnapshot, int, error) {
	return s.repo.AdminListMarketUniverseSnapshots(page, pageSize)
}

func (s *growthService) AdminGetMarketUniverseSnapshot(id string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error) {
	return s.repo.AdminGetMarketUniverseSnapshot(id)
}

func (s *growthService) AdminGetMarketCoverageSummary() (model.MarketCoverageSummary, error) {
	return s.repo.AdminGetMarketCoverageSummary()
}
