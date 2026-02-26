package dto

type PageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type ListBrowseHistoryQuery struct {
	PageQuery
	ContentType string `form:"content_type"`
}

type ListRechargeRecordsQuery struct {
	PageQuery
	Status string `form:"status"`
}

type CreateShareLinkRequest struct {
	Channel   string `json:"channel" binding:"required"`
	ExpiredAt string `json:"expired_at"`
}

type ReviewRewardRequest struct {
	Status string `json:"status" binding:"required,oneof=ISSUED REJECTED FROZEN"`
	Reason string `json:"reason"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentCallbackRequest struct {
	OrderNo        string      `json:"order_no" binding:"required"`
	ChannelTxnNo   string      `json:"channel_txn_no" binding:"required"`
	IdempotencyKey string      `json:"idempotency_key" binding:"required"`
	Sign           string      `json:"sign" binding:"required"`
	Payload        interface{} `json:"payload"`
}

type RiskRuleRequest struct {
	RuleCode  string `json:"rule_code"`
	RuleName  string `json:"rule_name"`
	Threshold int    `json:"threshold"`
	Status    string `json:"status"`
}

type ReviewRiskHitRequest struct {
	Status string `json:"status" binding:"required,oneof=CONFIRMED RELEASED"`
	Reason string `json:"reason"`
}

type ReviewWithdrawRequest struct {
	Status string `json:"status" binding:"required,oneof=APPROVED REJECTED PAID FAILED"`
	Reason string `json:"reason"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type NewsCategoryRequest struct {
	Name       string `json:"name" binding:"required"`
	Slug       string `json:"slug" binding:"required"`
	Sort       int    `json:"sort"`
	Visibility string `json:"visibility" binding:"required,oneof=PUBLIC VIP"`
	Status     string `json:"status" binding:"required,oneof=PUBLISHED DRAFT DISABLED"`
}

type NewsArticleRequest struct {
	CategoryID string `json:"category_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Summary    string `json:"summary"`
	Content    string `json:"content" binding:"required"`
	Visibility string `json:"visibility" binding:"required,oneof=PUBLIC VIP"`
	Status     string `json:"status" binding:"required,oneof=PUBLISHED DRAFT DISABLED"`
}

type NewsAttachmentRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileURL  string `json:"file_url" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required,gt=0"`
	MimeType string `json:"mime_type"`
}

type NewsPublishRequest struct {
	Status string `json:"status" binding:"required,oneof=PUBLISHED"`
}

type DataSourceCreateRequest struct {
	SourceKey  string                 `json:"source_key" binding:"required"`
	Name       string                 `json:"name" binding:"required"`
	SourceType string                 `json:"source_type" binding:"required"`
	Status     string                 `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	Config     map[string]interface{} `json:"config"`
}

type DataSourceUpdateRequest struct {
	Name       string                 `json:"name" binding:"required"`
	SourceType string                 `json:"source_type" binding:"required"`
	Status     string                 `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	Config     map[string]interface{} `json:"config"`
}

type DataSourceBatchHealthCheckRequest struct {
	SourceKeys []string `json:"source_keys"`
}

type StockRecommendationRequest struct {
	Symbol        string  `json:"symbol" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	Score         float64 `json:"score" binding:"required"`
	RiskLevel     string  `json:"risk_level" binding:"required"`
	PositionRange string  `json:"position_range"`
	ValidFrom     string  `json:"valid_from" binding:"required"`
	ValidTo       string  `json:"valid_to" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	ReasonSummary string  `json:"reason_summary"`
}

type StockRecommendationStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type FuturesStrategyRequest struct {
	Contract      string `json:"contract" binding:"required"`
	Name          string `json:"name"`
	Direction     string `json:"direction" binding:"required"`
	RiskLevel     string `json:"risk_level" binding:"required"`
	PositionRange string `json:"position_range"`
	ValidFrom     string `json:"valid_from" binding:"required"`
	ValidTo       string `json:"valid_to" binding:"required"`
	Status        string `json:"status" binding:"required"`
	ReasonSummary string `json:"reason_summary"`
}

type FuturesStrategyStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ACTIVE DISABLED BANNED"`
}

type UpdateUserMemberLevelRequest struct {
	MemberLevel string `json:"member_level" binding:"required"`
}

type UpdateUserKYCStatusRequest struct {
	KYCStatus string `json:"kyc_status" binding:"required,oneof=PENDING APPROVED REJECTED"`
}

type UpdateUserProfileRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type KYCSubmitRequest struct {
	RealName string `json:"real_name" binding:"required"`
	IDNumber string `json:"id_number" binding:"required"`
}

type SubscriptionCreateRequest struct {
	Type      string `json:"type" binding:"required,oneof=STOCK_RECO FUTURES_STRATEGY ARBITRAGE EVENT"`
	Scope     string `json:"scope"`
	Frequency string `json:"frequency" binding:"required,oneof=INSTANT DAILY WEEKLY"`
}

type SubscriptionUpdateRequest struct {
	Frequency string `json:"frequency" binding:"required,oneof=INSTANT DAILY WEEKLY"`
	Status    string `json:"status" binding:"required,oneof=ACTIVE PAUSED"`
}

type FuturesAlertRequest struct {
	Contract  string  `json:"contract" binding:"required"`
	AlertType string  `json:"alert_type" binding:"required,oneof=ENTRY EXIT STOP_LOSS"`
	Threshold float64 `json:"threshold" binding:"required"`
}

type MembershipProductRequest struct {
	Name         string  `json:"name" binding:"required"`
	Price        float64 `json:"price" binding:"required,gt=0"`
	Status       string  `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	MemberLevel  string  `json:"member_level" binding:"omitempty,oneof=VIP1 VIP2"`
	DurationDays int     `json:"duration_days" binding:"omitempty,gte=1"`
}

type MembershipProductStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
}

type MembershipOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=PENDING PAID CANCELED REFUNDED FAILED"`
}

type CreateMembershipOrderRequest struct {
	ProductID  string `json:"product_id" binding:"required"`
	PayChannel string `json:"pay_channel" binding:"required,oneof=ALIPAY WECHAT CARD"`
}

type VIPQuotaConfigRequest struct {
	MemberLevel        string `json:"member_level" binding:"required"`
	DocReadLimit       int    `json:"doc_read_limit" binding:"required,gte=0"`
	NewsSubscribeLimit int    `json:"news_subscribe_limit" binding:"required,gte=0"`
	ResetCycle         string `json:"reset_cycle" binding:"required,oneof=MONTHLY WEEKLY DAILY"`
	Status             string `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	EffectiveAt        string `json:"effective_at" binding:"required"`
}

type VIPQuotaConfigUpdateRequest struct {
	DocReadLimit       int    `json:"doc_read_limit" binding:"required,gte=0"`
	NewsSubscribeLimit int    `json:"news_subscribe_limit" binding:"required,gte=0"`
	ResetCycle         string `json:"reset_cycle" binding:"required,oneof=MONTHLY WEEKLY DAILY"`
	Status             string `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	EffectiveAt        string `json:"effective_at" binding:"required"`
}

type UserQuotaAdjustRequest struct {
	PeriodKey          string `json:"period_key" binding:"required"`
	DocReadDelta       int    `json:"doc_read_delta"`
	NewsSubscribeDelta int    `json:"news_subscribe_delta"`
	Reason             string `json:"reason"`
}

type SystemConfigUpsertRequest struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
	Description string `json:"description"`
}

type ReviewSubmitRequest struct {
	Module     string `json:"module" binding:"required,oneof=NEWS STOCK FUTURES"`
	TargetID   string `json:"target_id" binding:"required"`
	ReviewerID string `json:"reviewer_id"`
	SubmitNote string `json:"submit_note"`
}

type ReviewDecisionRequest struct {
	Status     string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	ReviewNote string `json:"review_note"`
}

type ReviewAssignRequest struct {
	ReviewerID string `json:"reviewer_id" binding:"required"`
}

type SchedulerTriggerRequest struct {
	JobName        string `json:"job_name" binding:"required"`
	TriggerSource  string `json:"trigger_source" binding:"required,oneof=MANUAL SYSTEM"`
	ResultSummary  string `json:"result_summary"`
	SimulateStatus string `json:"simulate_status" binding:"omitempty,oneof=SUCCESS FAILED"`
	ErrorMessage   string `json:"error_message"`
}

type SchedulerRetryRequest struct {
	SimulateStatus string `json:"simulate_status" binding:"omitempty,oneof=SUCCESS FAILED"`
	ResultSummary  string `json:"result_summary"`
	ErrorMessage   string `json:"error_message"`
}

type SchedulerJobDefinitionRequest struct {
	JobName     string `json:"job_name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Module      string `json:"module" binding:"required,oneof=STOCK FUTURES NEWS SYSTEM"`
	CronExpr    string `json:"cron_expr" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
}

type SchedulerJobDefinitionStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
}

type WorkflowMessageReadRequest struct {
	IsRead bool `json:"is_read"`
}

type WorkflowMessageBulkReadRequest struct {
	Module     string `json:"module"`
	EventType  string `json:"event_type"`
	ReceiverID string `json:"receiver_id"`
}

type MockLoginRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	Role          string `json:"role" binding:"required,oneof=USER ADMIN"`
	ExpireSeconds int    `json:"expire_seconds"`
}

type LoginRequest struct {
	Phone         string `json:"phone" binding:"required"`
	Password      string `json:"password" binding:"required"`
	ExpireSeconds int    `json:"expire_seconds"`
}

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthRiskConfigRequest struct {
	PhoneFailThreshold int `json:"phone_fail_threshold" binding:"required,gt=0"`
	IPFailThreshold    int `json:"ip_fail_threshold" binding:"required,gt=0"`
	IPPhoneThreshold   int `json:"ip_phone_threshold" binding:"required,gt=0"`
	LockSeconds        int `json:"lock_seconds" binding:"required,gt=0"`
}

type AuthUnlockRequest struct {
	Phone  string `json:"phone"`
	IP     string `json:"ip"`
	Reason string `json:"reason"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
}

func OK(data interface{}) APIResponse {
	return APIResponse{
		Code:    0,
		Message: "ok",
		Data:    data,
	}
}
