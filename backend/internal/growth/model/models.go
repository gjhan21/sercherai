package model

type BrowseHistory struct {
	ID          string `json:"id"`
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	Title       string `json:"title"`
	SourcePage  string `json:"source_page"`
	ViewedAt    string `json:"viewed_at"`
}

type RechargeRecord struct {
	ID         string  `json:"id"`
	OrderNo    string  `json:"order_no"`
	Amount     float64 `json:"amount"`
	PayChannel string  `json:"pay_channel"`
	Status     string  `json:"status"`
	PaidAt     string  `json:"paid_at"`
	Remark     string  `json:"remark"`
}

type ShareLink struct {
	ID         string `json:"id"`
	InviteCode string `json:"invite_code"`
	URL        string `json:"url"`
	Channel    string `json:"channel"`
	Status     string `json:"status"`
	ExpiredAt  string `json:"expired_at"`
}

type InviteRecord struct {
	ID          string `json:"id"`
	InviteeUser string `json:"invitee_user_id"`
	Status      string `json:"status"`
	RegisterAt  string `json:"register_at"`
	FirstPayAt  string `json:"first_pay_at"`
	RiskFlag    string `json:"risk_flag,omitempty"`
	InviterUser string `json:"inviter_user_id,omitempty"`
}

type RewardRecord struct {
	ID           string  `json:"id"`
	InviterUser  string  `json:"inviter_user_id,omitempty"`
	InviteeUser  string  `json:"invitee_user_id,omitempty"`
	RewardType   string  `json:"reward_type"`
	RewardValue  float64 `json:"reward_value"`
	TriggerEvent string  `json:"trigger_event"`
	Status       string  `json:"status"`
	IssuedAt     string  `json:"issued_at"`
}

type UserProfile struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	Email       string `json:"email,omitempty"`
	KYCStatus   string `json:"kyc_status"`
	MemberLevel string `json:"member_level"`
}

type Subscription struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Scope     string `json:"scope,omitempty"`
	Frequency string `json:"frequency"`
	Status    string `json:"status"`
}

type UserMessage struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content,omitempty"`
	Type       string `json:"type"`
	ReadStatus string `json:"read_status"`
	CreatedAt  string `json:"created_at"`
}

type MembershipQuota struct {
	MemberLevel            string `json:"member_level"`
	PeriodKey              string `json:"period_key"`
	DocReadLimit           int    `json:"doc_read_limit"`
	DocReadUsed            int    `json:"doc_read_used"`
	DocReadRemaining       int    `json:"doc_read_remaining"`
	NewsSubscribeLimit     int    `json:"news_subscribe_limit"`
	NewsSubscribeUsed      int    `json:"news_subscribe_used"`
	NewsSubscribeRemaining int    `json:"news_subscribe_remaining"`
	ResetCycle             string `json:"reset_cycle"`
	ResetAt                string `json:"reset_at,omitempty"`
}

type SignedURL struct {
	SignedURL string `json:"signed_url"`
	ExpiredAt string `json:"expired_at"`
}

type RewardWallet struct {
	CashBalance    float64 `json:"cash_balance"`
	CashFrozen     float64 `json:"cash_frozen"`
	CouponBalance  float64 `json:"coupon_balance"`
	VIPDaysBalance int     `json:"vip_days_balance"`
}

type RewardWalletTxn struct {
	ID        string  `json:"id"`
	TxnType   string  `json:"txn_type"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type ReconciliationRecord struct {
	ID         string `json:"id"`
	PayChannel string `json:"pay_channel"`
	BatchDate  string `json:"batch_date"`
	Status     string `json:"status"`
	DiffCount  int    `json:"diff_count"`
}

type RiskRule struct {
	ID        string `json:"id"`
	RuleCode  string `json:"rule_code"`
	RuleName  string `json:"rule_name"`
	Threshold int    `json:"threshold"`
	Status    string `json:"status"`
}

type RiskHit struct {
	ID        string `json:"id"`
	RuleCode  string `json:"rule_code"`
	UserID    string `json:"user_id"`
	RiskLevel string `json:"risk_level"`
	Status    string `json:"status"`
}

type WithdrawRequestInfo struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	AppliedAt string  `json:"applied_at"`
}

type ArbitrageOpportunity struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	ContractA  string  `json:"contract_a"`
	ContractB  string  `json:"contract_b"`
	Spread     float64 `json:"spread"`
	Percentile float64 `json:"percentile"`
	ZScore     float64 `json:"z_score,omitempty"`
	HalfLife   float64 `json:"half_life,omitempty"`
	RiskLevel  string  `json:"risk_level,omitempty"`
	Status     string  `json:"status"`
}

type ArbitrageRecommendation struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	ContractA   string  `json:"contract_a"`
	ContractB   string  `json:"contract_b"`
	Spread      float64 `json:"spread"`
	Percentile  float64 `json:"percentile"`
	EntryPoint  float64 `json:"entry_point"`
	ExitPoint   float64 `json:"exit_point"`
	StopPoint   float64 `json:"stop_point"`
	TriggerRule string  `json:"trigger_rule,omitempty"`
	Status      string  `json:"status,omitempty"`
	ValidTo     string  `json:"valid_to,omitempty"`
}

type FuturesReview struct {
	ID          string  `json:"id"`
	StrategyID  string  `json:"strategy_id"`
	HitRate     float64 `json:"hit_rate"`
	PnL         float64 `json:"pnl"`
	MaxDrawdown float64 `json:"max_drawdown"`
	ReviewDate  string  `json:"review_date"`
}

type MarketEvent struct {
	ID          string `json:"id"`
	EventType   string `json:"event_type"`
	Symbol      string `json:"symbol"`
	Summary     string `json:"summary"`
	TriggerRule string `json:"trigger_rule"`
	Source      string `json:"source,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type FuturesGuidance struct {
	Contract          string `json:"contract"`
	GuidanceDirection string `json:"guidance_direction"`
	PositionLevel     string `json:"position_level"`
	EntryRange        string `json:"entry_range"`
	TakeProfitRange   string `json:"take_profit_range"`
	StopLossRange     string `json:"stop_loss_range"`
	RiskLevel         string `json:"risk_level"`
	InvalidCondition  string `json:"invalid_condition"`
	ValidTo           string `json:"valid_to"`
}

type NewsCategory struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Sort       int    `json:"sort"`
	Visibility string `json:"visibility"`
	Status     string `json:"status"`
}

type NewsArticle struct {
	ID          string `json:"id"`
	CategoryID  string `json:"category_id"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Content     string `json:"content,omitempty"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at,omitempty"`
	AuthorID    string `json:"author_id,omitempty"`
}

type NewsAttachment struct {
	ID        string `json:"id"`
	ArticleID string `json:"article_id"`
	FileName  string `json:"file_name"`
	FileURL   string `json:"file_url"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	CreatedAt string `json:"created_at"`
}

type StockRecommendation struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Score         float64 `json:"score"`
	RiskLevel     string  `json:"risk_level"`
	PositionRange string  `json:"position_range"`
	ValidFrom     string  `json:"valid_from"`
	ValidTo       string  `json:"valid_to"`
	Status        string  `json:"status"`
	ReasonSummary string  `json:"reason_summary"`
}

type StockRecommendationDetail struct {
	RecoID         string  `json:"reco_id"`
	TechScore      float64 `json:"tech_score"`
	FundScore      float64 `json:"fund_score"`
	SentimentScore float64 `json:"sentiment_score"`
	MoneyFlowScore float64 `json:"money_flow_score"`
	TakeProfit     string  `json:"take_profit"`
	StopLoss       string  `json:"stop_loss"`
	RiskNote       string  `json:"risk_note"`
}

type RecommendationPerformancePoint struct {
	Date   string  `json:"date"`
	Return float64 `json:"return"`
}

type PublicHolding struct {
	ID          string  `json:"id"`
	Holder      string  `json:"holder"`
	Symbol      string  `json:"symbol"`
	Ratio       float64 `json:"ratio"`
	DisclosedAt string  `json:"disclosed_at"`
	Source      string  `json:"source"`
}

type PublicFuturesPosition struct {
	ID            string  `json:"id"`
	Contract      string  `json:"contract"`
	LongPosition  float64 `json:"long_position"`
	ShortPosition float64 `json:"short_position"`
	DisclosedAt   string  `json:"disclosed_at"`
	Source        string  `json:"source"`
}

type FuturesStrategy struct {
	ID            string `json:"id"`
	Contract      string `json:"contract"`
	Name          string `json:"name"`
	Direction     string `json:"direction"`
	RiskLevel     string `json:"risk_level"`
	PositionRange string `json:"position_range"`
	ValidFrom     string `json:"valid_from"`
	ValidTo       string `json:"valid_to"`
	Status        string `json:"status"`
	ReasonSummary string `json:"reason_summary"`
}

type AdminUser struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	Email       string `json:"email,omitempty"`
	Status      string `json:"status"`
	KYCStatus   string `json:"kyc_status"`
	MemberLevel string `json:"member_level"`
	CreatedAt   string `json:"created_at"`
}

type AdminDashboardOverview struct {
	TotalUsers           int `json:"total_users"`
	ActiveUsers          int `json:"active_users"`
	KYCApprovedUsers     int `json:"kyc_approved_users"`
	VIPUsers             int `json:"vip_users"`
	TodayNewUsers        int `json:"today_new_users"`
	TodayPaidOrders      int `json:"today_paid_orders"`
	TodayPublishedStocks int `json:"today_published_stocks"`
	TodayPublishedNews   int `json:"today_published_news"`
}

type AdminOperationLog struct {
	ID             string `json:"id"`
	Module         string `json:"module"`
	Action         string `json:"action"`
	TargetType     string `json:"target_type"`
	TargetID       string `json:"target_id"`
	OperatorUserID string `json:"operator_user_id"`
	BeforeValue    string `json:"before_value"`
	AfterValue     string `json:"after_value"`
	Reason         string `json:"reason"`
	CreatedAt      string `json:"created_at"`
}

type MembershipProduct struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Status       string  `json:"status"`
	MemberLevel  string  `json:"member_level,omitempty"`
	DurationDays int     `json:"duration_days,omitempty"`
}

type MembershipOrderAdmin struct {
	ID         string  `json:"id"`
	OrderNo    string  `json:"order_no,omitempty"`
	UserID     string  `json:"user_id"`
	ProductID  string  `json:"product_id"`
	Amount     float64 `json:"amount"`
	PayChannel string  `json:"pay_channel,omitempty"`
	Status     string  `json:"status"`
	PaidAt     string  `json:"paid_at,omitempty"`
	CreatedAt  string  `json:"created_at,omitempty"`
}

type UserAccessProfile struct {
	UserID      string `json:"user_id"`
	Status      string `json:"status"`
	KYCStatus   string `json:"kyc_status"`
	MemberLevel string `json:"member_level"`
}

type AttachmentFileInfo struct {
	FileURL   string `json:"file_url"`
	ArticleID string `json:"article_id"`
}

type VIPQuotaConfig struct {
	ID                 string `json:"id"`
	MemberLevel        string `json:"member_level"`
	DocReadLimit       int    `json:"doc_read_limit"`
	NewsSubscribeLimit int    `json:"news_subscribe_limit"`
	ResetCycle         string `json:"reset_cycle"`
	Status             string `json:"status"`
	EffectiveAt        string `json:"effective_at"`
	UpdatedAt          string `json:"updated_at"`
}

type UserQuotaUsage struct {
	UserID             string `json:"user_id"`
	MemberLevel        string `json:"member_level"`
	PeriodKey          string `json:"period_key"`
	DocReadLimit       int    `json:"doc_read_limit"`
	DocReadUsed        int    `json:"doc_read_used"`
	NewsSubscribeLimit int    `json:"news_subscribe_limit"`
	NewsSubscribeUsed  int    `json:"news_subscribe_used"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

type DataSource struct {
	ID         string                 `json:"id"`
	SourceKey  string                 `json:"source_key"`
	Name       string                 `json:"name"`
	SourceType string                 `json:"source_type"`
	Status     string                 `json:"status"`
	Config     map[string]interface{} `json:"config,omitempty"`
	UpdatedAt  string                 `json:"updated_at,omitempty"`
}

type DataSourceHealthCheck struct {
	SourceKey           string `json:"source_key"`
	Status              string `json:"status"`
	Reachable           bool   `json:"reachable"`
	HTTPStatus          int    `json:"http_status,omitempty"`
	LatencyMS           int64  `json:"latency_ms"`
	Message             string `json:"message,omitempty"`
	FailureCategory     string `json:"failure_category,omitempty"`
	Attempts            int    `json:"attempts,omitempty"`
	MaxAttempts         int    `json:"max_attempts,omitempty"`
	ConsecutiveFailures int    `json:"consecutive_failures,omitempty"`
	AlertTriggered      bool   `json:"alert_triggered"`
	CheckedAt           string `json:"checked_at"`
}

type DataSourceHealthLog struct {
	ID         string `json:"id"`
	SourceKey  string `json:"source_key"`
	Status     string `json:"status"`
	Reachable  bool   `json:"reachable"`
	HTTPStatus int    `json:"http_status,omitempty"`
	LatencyMS  int64  `json:"latency_ms"`
	Message    string `json:"message,omitempty"`
	CheckedAt  string `json:"checked_at"`
}

type SystemConfig struct {
	ID          string `json:"id"`
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	Description string `json:"description,omitempty"`
	UpdatedBy   string `json:"updated_by"`
	UpdatedAt   string `json:"updated_at"`
}

type ReviewTask struct {
	ID          string `json:"id"`
	Module      string `json:"module"`
	TargetID    string `json:"target_id"`
	SubmitterID string `json:"submitter_id"`
	ReviewerID  string `json:"reviewer_id,omitempty"`
	Status      string `json:"status"`
	SubmitNote  string `json:"submit_note,omitempty"`
	ReviewNote  string `json:"review_note,omitempty"`
	SubmittedAt string `json:"submitted_at"`
	ReviewedAt  string `json:"reviewed_at,omitempty"`
}

type SchedulerJobRun struct {
	ID            string `json:"id"`
	ParentRunID   string `json:"parent_run_id,omitempty"`
	RetryCount    int    `json:"retry_count"`
	JobName       string `json:"job_name"`
	TriggerSource string `json:"trigger_source"`
	Status        string `json:"status"`
	StartedAt     string `json:"started_at"`
	FinishedAt    string `json:"finished_at,omitempty"`
	ResultSummary string `json:"result_summary,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	OperatorID    string `json:"operator_id,omitempty"`
}

type WorkflowMessage struct {
	ID         string `json:"id"`
	ReviewID   string `json:"review_id,omitempty"`
	TargetID   string `json:"target_id"`
	Module     string `json:"module"`
	ReceiverID string `json:"receiver_id,omitempty"`
	SenderID   string `json:"sender_id,omitempty"`
	EventType  string `json:"event_type"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
	ReadAt     string `json:"read_at,omitempty"`
}

type WorkflowMetrics struct {
	PendingReviews int `json:"pending_reviews"`
	ApprovedToday  int `json:"approved_today"`
	RejectedToday  int `json:"rejected_today"`
	UnreadMessages int `json:"unread_messages"`
	TotalMessages  int `json:"total_messages"`
}

type SchedulerJobMetrics struct {
	TodayTotal   int `json:"today_total"`
	TodaySuccess int `json:"today_success"`
	TodayFailed  int `json:"today_failed"`
	TodayRunning int `json:"today_running"`
}

type SchedulerJobDefinition struct {
	ID          string `json:"id"`
	JobName     string `json:"job_name"`
	DisplayName string `json:"display_name"`
	Module      string `json:"module"`
	CronExpr    string `json:"cron_expr"`
	Status      string `json:"status"`
	LastRunAt   string `json:"last_run_at,omitempty"`
	UpdatedBy   string `json:"updated_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
