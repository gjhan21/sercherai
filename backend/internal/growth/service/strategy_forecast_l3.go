package service

import "sercherai/backend/internal/growth/model"

func (s *growthService) CreateStrategyForecastL3Run(input model.StrategyForecastL3RunCreateInput) (model.StrategyForecastL3Run, error) {
	return s.repo.CreateStrategyForecastL3Run(input)
}

func (s *growthService) ListStrategyForecastL3Runs(requestUserID string, status string, targetType string, triggerType string, page int, pageSize int) ([]model.StrategyForecastL3Run, int, error) {
	return s.repo.ListStrategyForecastL3Runs(requestUserID, status, targetType, triggerType, page, pageSize)
}

func (s *growthService) GetLatestStrategyForecastL3Run(targetType string, targetID string) (model.StrategyForecastL3Run, error) {
	return s.repo.GetLatestStrategyForecastL3Run(targetType, targetID)
}

func (s *growthService) GetStrategyForecastL3Run(runID string) (model.StrategyForecastL3Run, error) {
	return s.repo.GetStrategyForecastL3Run(runID)
}

func (s *growthService) GetStrategyForecastL3RunDetail(runID string) (model.StrategyForecastL3RunDetail, error) {
	return s.repo.GetStrategyForecastL3RunDetail(runID)
}

func (s *growthService) GetStrategyForecastL3RunDetailForUser(runID string, userID string) (model.StrategyForecastL3RunDetail, error) {
	return s.repo.GetStrategyForecastL3RunDetailForUser(runID, userID)
}

func (s *growthService) RetryStrategyForecastL3Run(runID string, operatorUserID string, reason string) (model.StrategyForecastL3Run, error) {
	return s.repo.RetryStrategyForecastL3Run(runID, operatorUserID, reason)
}

func (s *growthService) CancelStrategyForecastL3Run(runID string, operatorUserID string, reason string) (model.StrategyForecastL3Run, error) {
	return s.repo.CancelStrategyForecastL3Run(runID, operatorUserID, reason)
}

func (s *growthService) ListStrategyForecastL3Logs(runID string) ([]model.StrategyForecastL3Log, error) {
	return s.repo.ListStrategyForecastL3Logs(runID)
}

func (s *growthService) ListStrategyForecastL3QualitySummaries(targetType string, days int) ([]model.StrategyForecastL3QualitySummary, error) {
	return s.repo.ListStrategyForecastL3QualitySummaries(targetType, days)
}

func (s *growthService) ExecuteQueuedStrategyForecastL3Runs(limit int, operatorUserID string) (int, error) {
	return s.repo.ExecuteQueuedStrategyForecastL3Runs(limit, operatorUserID)
}

func (s *growthService) RunStrategyForecastL3QualityBackfill(limit int, operatorUserID string) (int, error) {
	return s.repo.RunStrategyForecastL3QualityBackfill(limit, operatorUserID)
}
