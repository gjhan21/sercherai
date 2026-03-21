package service

import "sercherai/backend/internal/growth/model"

func (s *growthService) AdminListStrategySeedSets(targetType string, status string, page int, pageSize int) ([]model.StrategySeedSet, int, error) {
	return s.repo.AdminListStrategySeedSets(targetType, status, page, pageSize)
}

func (s *growthService) AdminCreateStrategySeedSet(item model.StrategySeedSet) (string, error) {
	return s.repo.AdminCreateStrategySeedSet(item)
}

func (s *growthService) AdminUpdateStrategySeedSet(id string, item model.StrategySeedSet) error {
	return s.repo.AdminUpdateStrategySeedSet(id, item)
}

func (s *growthService) AdminListStrategyAgentProfiles(targetType string, status string, page int, pageSize int) ([]model.StrategyAgentProfile, int, error) {
	return s.repo.AdminListStrategyAgentProfiles(targetType, status, page, pageSize)
}

func (s *growthService) AdminCreateStrategyAgentProfile(item model.StrategyAgentProfile) (string, error) {
	return s.repo.AdminCreateStrategyAgentProfile(item)
}

func (s *growthService) AdminUpdateStrategyAgentProfile(id string, item model.StrategyAgentProfile) error {
	return s.repo.AdminUpdateStrategyAgentProfile(id, item)
}

func (s *growthService) AdminListStrategyScenarioTemplates(targetType string, status string, page int, pageSize int) ([]model.StrategyScenarioTemplate, int, error) {
	return s.repo.AdminListStrategyScenarioTemplates(targetType, status, page, pageSize)
}

func (s *growthService) AdminCreateStrategyScenarioTemplate(item model.StrategyScenarioTemplate) (string, error) {
	return s.repo.AdminCreateStrategyScenarioTemplate(item)
}

func (s *growthService) AdminUpdateStrategyScenarioTemplate(id string, item model.StrategyScenarioTemplate) error {
	return s.repo.AdminUpdateStrategyScenarioTemplate(id, item)
}

func (s *growthService) AdminListStrategyPublishPolicies(targetType string, status string, page int, pageSize int) ([]model.StrategyPublishPolicy, int, error) {
	return s.repo.AdminListStrategyPublishPolicies(targetType, status, page, pageSize)
}

func (s *growthService) AdminCreateStrategyPublishPolicy(item model.StrategyPublishPolicy) (string, error) {
	return s.repo.AdminCreateStrategyPublishPolicy(item)
}

func (s *growthService) AdminUpdateStrategyPublishPolicy(id string, item model.StrategyPublishPolicy) error {
	return s.repo.AdminUpdateStrategyPublishPolicy(id, item)
}

func (s *growthService) AdminListStrategyEngineJobs(jobType string, status string, page int, pageSize int) ([]model.StrategyEngineJobRecord, int, error) {
	return s.repo.AdminListStrategyEngineJobs(jobType, status, page, pageSize)
}

func (s *growthService) AdminGetStrategyEngineJob(jobID string) (model.StrategyEngineJobRecord, error) {
	return s.repo.AdminGetStrategyEngineJob(jobID)
}

func (s *growthService) AdminPublishStrategyEngineJob(jobID string, operator string, force bool, overrideReason string) (model.StrategyEnginePublishRecord, error) {
	return s.repo.AdminPublishStrategyEngineJob(jobID, operator, force, overrideReason)
}
