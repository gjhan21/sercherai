package service

import (
	"strings"

	"sercherai/backend/internal/growth/model"
)

type stockEventGraphSyncer interface {
	SyncReviewedStockEventGraph(cluster model.StockEventCluster) error
}

func (s *growthService) AdminGetStockSelectionOverview() (model.AdminStockSelectionOverview, error) {
	return s.repo.AdminGetStockSelectionOverview()
}

func (s *growthService) AdminListStockSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.StockSelectionRun, int, error) {
	return s.repo.AdminListStockSelectionRuns(status, reviewStatus, profileID, page, pageSize)
}

func (s *growthService) AdminCreateStockSelectionRun(input model.StockSelectionRunCreateRequest, operator string) (model.StockSelectionRun, error) {
	return s.repo.AdminCreateStockSelectionRun(input, operator)
}

func (s *growthService) AdminGetStockSelectionRun(runID string) (model.StockSelectionRun, error) {
	return s.repo.AdminGetStockSelectionRun(runID)
}

func (s *growthService) AdminCompareStockSelectionRuns(runIDs []string) (model.StockSelectionRunCompareResult, error) {
	return s.repo.AdminCompareStockSelectionRuns(runIDs)
}

func (s *growthService) AdminListStockSelectionProfiles(status string, page int, pageSize int) ([]model.StockSelectionProfile, int, error) {
	return s.repo.AdminListStockSelectionProfiles(status, page, pageSize)
}

func (s *growthService) AdminListStockSelectionProfileVersions(profileID string) ([]model.StockSelectionProfileVersion, error) {
	return s.repo.AdminListStockSelectionProfileVersions(profileID)
}

func (s *growthService) AdminCreateStockSelectionProfile(item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error) {
	return s.repo.AdminCreateStockSelectionProfile(item, changeNote)
}

func (s *growthService) AdminUpdateStockSelectionProfile(id string, item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error) {
	return s.repo.AdminUpdateStockSelectionProfile(id, item, changeNote)
}

func (s *growthService) AdminPublishStockSelectionProfile(id string, operator string) (model.StockSelectionProfile, error) {
	return s.repo.AdminPublishStockSelectionProfile(id, operator)
}

func (s *growthService) AdminRollbackStockSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.StockSelectionProfile, error) {
	return s.repo.AdminRollbackStockSelectionProfile(id, versionNo, changeNote, operator)
}

func (s *growthService) AdminListStockSelectionProfileTemplates(status string, page int, pageSize int) ([]model.StockSelectionProfileTemplate, int, error) {
	return s.repo.AdminListStockSelectionProfileTemplates(status, page, pageSize)
}

func (s *growthService) AdminCreateStockSelectionProfileTemplate(item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error) {
	return s.repo.AdminCreateStockSelectionProfileTemplate(item)
}

func (s *growthService) AdminUpdateStockSelectionProfileTemplate(id string, item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error) {
	return s.repo.AdminUpdateStockSelectionProfileTemplate(id, item)
}

func (s *growthService) AdminSetDefaultStockSelectionProfileTemplate(id string, operator string) (model.StockSelectionProfileTemplate, error) {
	return s.repo.AdminSetDefaultStockSelectionProfileTemplate(id, operator)
}

func (s *growthService) AdminListStockSelectionRunCandidates(runID string) ([]model.StockSelectionCandidateSnapshot, error) {
	return s.repo.AdminListStockSelectionRunCandidates(runID)
}

func (s *growthService) AdminListStockSelectionRunPortfolio(runID string) ([]model.StockSelectionPortfolioEntry, error) {
	return s.repo.AdminListStockSelectionRunPortfolio(runID)
}

func (s *growthService) AdminListStockSelectionRunEvidence(runID string, symbol string) ([]model.StockSelectionRunEvidence, error) {
	return s.repo.AdminListStockSelectionRunEvidence(runID, symbol)
}

func (s *growthService) AdminListStockSelectionRunEvaluations(runID string, symbol string) ([]model.StockSelectionRunEvaluation, error) {
	return s.repo.AdminListStockSelectionRunEvaluations(runID, symbol)
}

func (s *growthService) AdminListStockSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.StockSelectionEvaluationLeaderboardItem, error) {
	return s.repo.AdminListStockSelectionEvaluationLeaderboard(templateID, profileID, marketRegime)
}

func (s *growthService) AdminListStockSelectionReviews(status string, page int, pageSize int) ([]model.StockSelectionPublishReview, int, error) {
	return s.repo.AdminListStockSelectionReviews(status, page, pageSize)
}

func (s *growthService) AdminApproveStockSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.StockSelectionPublishReview, error) {
	return s.repo.AdminApproveStockSelectionReview(runID, operator, reviewNote, force, overrideReason)
}

func (s *growthService) AdminRejectStockSelectionReview(runID string, operator string, reviewNote string) (model.StockSelectionPublishReview, error) {
	return s.repo.AdminRejectStockSelectionReview(runID, operator, reviewNote)
}

func (s *growthService) AdminListStockEventClusters(query model.StockEventQuery) ([]model.StockEventCluster, int, error) {
	return s.repo.AdminListStockEventClusters(query)
}

func (s *growthService) AdminGetStockEventCluster(id string) (model.StockEventCluster, error) {
	return s.repo.AdminGetStockEventCluster(id)
}

func (s *growthService) AdminReviewStockEventCluster(id string, review model.StockEventReview) (model.StockEventCluster, error) {
	review.ClusterID = strings.TrimSpace(id)
	if _, err := s.repo.AdminCreateStockEventReview(review); err != nil {
		return model.StockEventCluster{}, err
	}

	tasks, _, err := s.repo.AdminListReviewTasks("STOCK_EVENT", "PENDING", "", "", 1, 10000)
	if err == nil {
		for _, task := range tasks {
			if strings.TrimSpace(task.TargetID) != review.ClusterID {
				continue
			}
			_ = s.repo.AdminReviewTaskDecision(task.ID, review.ReviewStatus, review.Reviewer, review.ReviewNote)
			break
		}
	}

	cluster, err := s.repo.AdminGetStockEventCluster(review.ClusterID)
	if err != nil {
		return model.StockEventCluster{}, err
	}
	if syncer, ok := s.repo.(stockEventGraphSyncer); ok {
		if err := syncer.SyncReviewedStockEventGraph(cluster); err != nil {
			if cluster.Metadata == nil {
				cluster.Metadata = map[string]any{}
			}
			cluster.Metadata["graph_sync_warning"] = err.Error()
		} else if cluster.Metadata != nil {
			delete(cluster.Metadata, "graph_sync_warning")
		}
	}
	return cluster, nil
}
