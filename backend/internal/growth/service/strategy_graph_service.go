package service

import "sercherai/backend/internal/growth/model"

func (s *growthService) AdminGetStrategyGraphSnapshot(snapshotID string) (model.StrategyGraphSnapshot, error) {
	return s.repo.AdminGetStrategyGraphSnapshot(snapshotID)
}

func (s *growthService) AdminQueryStrategyGraphSubgraph(query model.StrategyGraphSubgraphQuery) (model.StrategyGraphSubgraph, error) {
	return s.repo.AdminQueryStrategyGraphSubgraph(query)
}
