package dto

import "sercherai/backend/internal/growth/model"

type StrategyForecastL3CreateRequest struct {
	TargetType    string  `json:"target_type" binding:"required,oneof=STOCK FUTURES"`
	TargetID      string  `json:"target_id"`
	TargetKey     string  `json:"target_key"`
	TargetLabel   string  `json:"target_label"`
	TriggerType   string  `json:"trigger_type" binding:"omitempty,oneof=ADMIN_MANUAL AUTO_PRIORITY USER_REQUEST"`
	PriorityScore float64 `json:"priority_score" binding:"omitempty,gte=0,lte=1"`
	Reason        string  `json:"reason"`
}

type StrategyForecastL3RetryRequest struct {
	Reason string `json:"reason"`
}

type StrategyForecastL3CancelRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type StrategyForecastL3RunListResponse struct {
	Items []model.StrategyForecastL3Run `json:"items"`
	Total int                           `json:"total"`
}

type StrategyForecastL3RunDetailResponse struct {
	Run    model.StrategyForecastL3Run     `json:"run"`
	Report *model.StrategyForecastL3Report `json:"report,omitempty"`
	Logs   []model.StrategyForecastL3Log   `json:"logs,omitempty"`
}

type StrategyForecastL3QualitySummaryResponse struct {
	Items []model.StrategyForecastL3QualitySummary `json:"items"`
}
