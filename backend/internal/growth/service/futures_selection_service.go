package service

import "sercherai/backend/internal/growth/model"

func (s *growthService) AdminCompareFuturesSelectionRuns(runIDs []string) (model.FuturesSelectionRunCompareResult, error) {
	return s.repo.AdminCompareFuturesSelectionRuns(runIDs)
}

func (s *growthService) AdminListFuturesSelectionProfiles(status string, page int, pageSize int) ([]model.FuturesSelectionProfile, int, error) {
	return s.repo.AdminListFuturesSelectionProfiles(status, page, pageSize)
}

func (s *growthService) AdminListFuturesSelectionProfileVersions(profileID string) ([]model.FuturesSelectionProfileVersion, error) {
	return s.repo.AdminListFuturesSelectionProfileVersions(profileID)
}

func (s *growthService) AdminCreateFuturesSelectionProfile(item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error) {
	return s.repo.AdminCreateFuturesSelectionProfile(item, changeNote)
}

func (s *growthService) AdminUpdateFuturesSelectionProfile(id string, item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error) {
	return s.repo.AdminUpdateFuturesSelectionProfile(id, item, changeNote)
}

func (s *growthService) AdminPublishFuturesSelectionProfile(id string, operator string) (model.FuturesSelectionProfile, error) {
	return s.repo.AdminPublishFuturesSelectionProfile(id, operator)
}

func (s *growthService) AdminRollbackFuturesSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.FuturesSelectionProfile, error) {
	return s.repo.AdminRollbackFuturesSelectionProfile(id, versionNo, changeNote, operator)
}

func (s *growthService) AdminListFuturesSelectionProfileTemplates(status string, page int, pageSize int) ([]model.FuturesSelectionProfileTemplate, int, error) {
	return s.repo.AdminListFuturesSelectionProfileTemplates(status, page, pageSize)
}

func (s *growthService) AdminCreateFuturesSelectionProfileTemplate(item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error) {
	return s.repo.AdminCreateFuturesSelectionProfileTemplate(item)
}

func (s *growthService) AdminUpdateFuturesSelectionProfileTemplate(id string, item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error) {
	return s.repo.AdminUpdateFuturesSelectionProfileTemplate(id, item)
}

func (s *growthService) AdminSetDefaultFuturesSelectionProfileTemplate(id string, operator string) (model.FuturesSelectionProfileTemplate, error) {
	return s.repo.AdminSetDefaultFuturesSelectionProfileTemplate(id, operator)
}

func (s *growthService) AdminListFuturesSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.FuturesSelectionEvaluationLeaderboardItem, error) {
	return s.repo.AdminListFuturesSelectionEvaluationLeaderboard(templateID, profileID, marketRegime)
}

func (s *growthService) AdminListFuturesSelectionRunCandidates(runID string) ([]model.FuturesSelectionCandidateSnapshot, error) {
	return s.repo.AdminListFuturesSelectionRunCandidates(runID)
}

func (s *growthService) AdminListFuturesSelectionRunPortfolio(runID string) ([]model.FuturesSelectionPortfolioEntry, error) {
	return s.repo.AdminListFuturesSelectionRunPortfolio(runID)
}

func (s *growthService) AdminListFuturesSelectionRunEvidence(runID string, contract string) ([]model.FuturesSelectionRunEvidence, error) {
	return s.repo.AdminListFuturesSelectionRunEvidence(runID, contract)
}

func (s *growthService) AdminListFuturesSelectionRunEvaluations(runID string, contract string) ([]model.FuturesSelectionRunEvaluation, error) {
	return s.repo.AdminListFuturesSelectionRunEvaluations(runID, contract)
}

func (s *growthService) AdminApproveFuturesSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.FuturesSelectionPublishReview, error) {
	return s.repo.AdminApproveFuturesSelectionReview(runID, operator, reviewNote, force, overrideReason)
}

func (s *growthService) AdminRejectFuturesSelectionReview(runID string, operator string, reviewNote string) (model.FuturesSelectionPublishReview, error) {
	return s.repo.AdminRejectFuturesSelectionReview(runID, operator, reviewNote)
}
