package model

type BrowseHistory struct {
	ID          string `json:"id"`
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	Title       string `json:"title"`
	SourcePage  string `json:"source_page"`
	ViewedAt    string `json:"viewed_at"`
}

type AdminBrowseHistory struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	UserPhone   string `json:"user_phone,omitempty"`
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	Title       string `json:"title"`
	SourcePage  string `json:"source_page"`
	ViewedAt    string `json:"viewed_at"`
}

type AdminBrowseHistorySummary struct {
	TotalViews   int `json:"total_views"`
	UniqueUsers  int `json:"unique_users"`
	NewsViews    int `json:"news_views"`
	ReportViews  int `json:"report_views"`
	JournalViews int `json:"journal_views"`
	TodayViews   int `json:"today_views"`
	Last7dViews  int `json:"last_7d_views"`
}

type AdminBrowseUserSegment struct {
	Segment         string `json:"segment"`
	UserID          string `json:"user_id"`
	UserPhone       string `json:"user_phone,omitempty"`
	ViewCount7d     int    `json:"view_count_7d"`
	LastViewedAt    string `json:"last_viewed_at,omitempty"`
	LastContentID   string `json:"last_content_id,omitempty"`
	LastContentType string `json:"last_content_type,omitempty"`
}

type AdminBrowseTrendPoint struct {
	Date         string `json:"date"`
	TotalViews   int    `json:"total_views"`
	NewsViews    int    `json:"news_views"`
	ReportViews  int    `json:"report_views"`
	JournalViews int    `json:"journal_views"`
}

type AdminMessageSendFailure struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
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

type InviteSummary struct {
	ShareLinkCount         int     `json:"share_link_count"`
	RegisteredCount        int     `json:"registered_count"`
	FirstPaidCount         int     `json:"first_paid_count"`
	ConversionRate         float64 `json:"conversion_rate"`
	Last7dRegisteredCount  int     `json:"last_7d_registered_count"`
	Last7dFirstPaidCount   int     `json:"last_7d_first_paid_count"`
	Last7dConversionRate   float64 `json:"last_7d_conversion_rate"`
	Last30dRegisteredCount int     `json:"last_30d_registered_count"`
	Last30dFirstPaidCount  int     `json:"last_30d_first_paid_count"`
	Last30dConversionRate  float64 `json:"last_30d_conversion_rate"`
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
	ID                 string `json:"id"`
	Phone              string `json:"phone"`
	Email              string `json:"email,omitempty"`
	KYCStatus          string `json:"kyc_status"`
	MemberLevel        string `json:"member_level"`
	VIPStartedAt       string `json:"vip_started_at,omitempty"`
	VIPExpireAt        string `json:"vip_expire_at,omitempty"`
	VIPStatus          string `json:"vip_status,omitempty"`
	VIPRemainingDays   int    `json:"vip_remaining_days,omitempty"`
	RegistrationSource string `json:"registration_source,omitempty"`
	InviterUserID      string `json:"inviter_user_id,omitempty"`
	InviteCode         string `json:"invite_code,omitempty"`
	InviteLinkID       string `json:"invite_link_id,omitempty"`
	InvitedAt          string `json:"invited_at,omitempty"`
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

type AdminUserMessage struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	UserPhone  string `json:"user_phone,omitempty"`
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
	VIPExpireAt            string `json:"vip_expire_at,omitempty"`
	VIPStatus              string `json:"vip_status,omitempty"`
	VIPRemainingDays       int    `json:"vip_remaining_days,omitempty"`
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
	ID              string `json:"id"`
	CategoryID      string `json:"category_id"`
	Title           string `json:"title"`
	Summary         string `json:"summary"`
	Content         string `json:"content,omitempty"`
	CoverURL        string `json:"cover_url,omitempty"`
	Visibility      string `json:"visibility"`
	Status          string `json:"status"`
	PublishedAt     string `json:"published_at,omitempty"`
	AuthorID        string `json:"author_id,omitempty"`
	AttachmentCount int    `json:"attachment_count,omitempty"`
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

type StockRecommendationFactorScore struct {
	Key          string  `json:"key"`
	Label        string  `json:"label"`
	Weight       float64 `json:"weight"`
	Score        float64 `json:"score"`
	Contribution float64 `json:"contribution"`
}

type StockRecommendationScoreFramework struct {
	Method        string                           `json:"method"`
	TotalScore    float64                          `json:"total_score"`
	WeightedScore float64                          `json:"weighted_score"`
	ScoreGap      float64                          `json:"score_gap"`
	Factors       []StockRecommendationFactorScore `json:"factors"`
}

type StockRecommendationRelatedNews struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Summary        string  `json:"summary"`
	Source         string  `json:"source"`
	Visibility     string  `json:"visibility"`
	PublishedAt    string  `json:"published_at"`
	RelevanceScore float64 `json:"relevance_score"`
}

type StockRecommendationPerformanceSummary struct {
	SampleDays                int     `json:"sample_days"`
	WinRate                   float64 `json:"win_rate"`
	AvgDailyReturn            float64 `json:"avg_daily_return"`
	CumulativeReturn          float64 `json:"cumulative_return"`
	BenchmarkCumulativeReturn float64 `json:"benchmark_cumulative_return"`
	ExcessReturn              float64 `json:"excess_return"`
	MaxDrawdown               float64 `json:"max_drawdown"`
	BenchmarkSymbol           string  `json:"benchmark_symbol"`
	BenchmarkSource           string  `json:"benchmark_source"`
}

type StockRecommendationInsight struct {
	Recommendation   StockRecommendation                   `json:"recommendation"`
	Detail           StockRecommendationDetail             `json:"detail"`
	ScoreFramework   StockRecommendationScoreFramework     `json:"score_framework"`
	RelatedNews      []StockRecommendationRelatedNews      `json:"related_news"`
	Performance      []RecommendationPerformancePoint      `json:"performance"`
	Benchmark        []RecommendationPerformancePoint      `json:"benchmark"`
	PerformanceStats StockRecommendationPerformanceSummary `json:"performance_stats"`
	GeneratedAt      string                                `json:"generated_at"`
}

type StockMarketQuote struct {
	ID             string  `json:"id"`
	Symbol         string  `json:"symbol"`
	TradeDate      string  `json:"trade_date"`
	OpenPrice      float64 `json:"open_price"`
	HighPrice      float64 `json:"high_price"`
	LowPrice       float64 `json:"low_price"`
	ClosePrice     float64 `json:"close_price"`
	PrevClosePrice float64 `json:"prev_close_price"`
	Volume         int64   `json:"volume"`
	Turnover       float64 `json:"turnover"`
	SourceKey      string  `json:"source_key"`
	CreatedAt      string  `json:"created_at,omitempty"`
	UpdatedAt      string  `json:"updated_at,omitempty"`
}

type StockQuantScore struct {
	Rank             int      `json:"rank"`
	Symbol           string   `json:"symbol"`
	Name             string   `json:"name"`
	TradeDate        string   `json:"trade_date"`
	ClosePrice       float64  `json:"close_price"`
	Momentum5        float64  `json:"momentum_5"`
	Momentum20       float64  `json:"momentum_20"`
	Volatility20     float64  `json:"volatility_20"`
	VolumeRatio      float64  `json:"volume_ratio"`
	Drawdown20       float64  `json:"drawdown_20"`
	TrendStrength    float64  `json:"trend_strength"`
	TrendScore       float64  `json:"trend_score"`
	FlowScore        float64  `json:"flow_score"`
	ValueScore       float64  `json:"value_score"`
	NewsScore        float64  `json:"news_score"`
	NetMFAmount      float64  `json:"net_mf_amount"`
	PeTTM            float64  `json:"pe_ttm"`
	PB               float64  `json:"pb"`
	TurnoverRate     float64  `json:"turnover_rate"`
	NewsHeat         int      `json:"news_heat"`
	PositiveNewsRate float64  `json:"positive_news_rate"`
	Score            float64  `json:"score"`
	RiskLevel        string   `json:"risk_level"`
	ReasonSummary    string   `json:"reason_summary"`
	Reasons          []string `json:"reasons"`
}

type StockQuantEvaluationPoint struct {
	TradeDate             string  `json:"trade_date"`
	SampleCount           int     `json:"sample_count"`
	AvgReturn5            float64 `json:"avg_return_5"`
	HitRate5              float64 `json:"hit_rate_5"`
	AvgReturn10           float64 `json:"avg_return_10"`
	HitRate10             float64 `json:"hit_rate_10"`
	BenchmarkReturn       float64 `json:"benchmark_return"`
	BenchmarkReturn10     float64 `json:"benchmark_return_10"`
	CumulativeReturn5     float64 `json:"cumulative_return_5"`
	CumulativeBenchmark5  float64 `json:"cumulative_benchmark_5"`
	CumulativeExcess5     float64 `json:"cumulative_excess_5"`
	CumulativeReturn10    float64 `json:"cumulative_return_10"`
	CumulativeBenchmark10 float64 `json:"cumulative_benchmark_10"`
	CumulativeExcess10    float64 `json:"cumulative_excess_10"`
}

type StockQuantEvaluationSummary struct {
	WindowDays           int     `json:"window_days"`
	TopN                 int     `json:"top_n"`
	SampleDays           int     `json:"sample_days"`
	SampleCount          int     `json:"sample_count"`
	AvgReturn5           float64 `json:"avg_return_5"`
	HitRate5             float64 `json:"hit_rate_5"`
	MaxDrawdown5         float64 `json:"max_drawdown_5"`
	AvgReturn10          float64 `json:"avg_return_10"`
	HitRate10            float64 `json:"hit_rate_10"`
	MaxDrawdown10        float64 `json:"max_drawdown_10"`
	BenchmarkAvgReturn5  float64 `json:"benchmark_avg_return_5"`
	BenchmarkAvgReturn10 float64 `json:"benchmark_avg_return_10"`
	GeneratedAt          string  `json:"generated_at"`
}

type StockQuantRiskPerformance struct {
	RiskLevel   string  `json:"risk_level"`
	SampleCount int     `json:"sample_count"`
	AvgReturn5  float64 `json:"avg_return_5"`
	HitRate5    float64 `json:"hit_rate_5"`
	AvgReturn10 float64 `json:"avg_return_10"`
	HitRate10   float64 `json:"hit_rate_10"`
}

type StockQuantRotationPoint struct {
	TradeDate    string   `json:"trade_date"`
	TopSymbols   []string `json:"top_symbols"`
	Entered      []string `json:"entered"`
	Exited       []string `json:"exited"`
	StayedCount  int      `json:"stayed_count"`
	ChangedCount int      `json:"changed_count"`
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

type FuturesStrategyFactorScore struct {
	Key          string  `json:"key"`
	Label        string  `json:"label"`
	Weight       float64 `json:"weight"`
	Score        float64 `json:"score"`
	Contribution float64 `json:"contribution"`
}

type FuturesStrategyScoreFramework struct {
	Method        string                       `json:"method"`
	TotalScore    float64                      `json:"total_score"`
	WeightedScore float64                      `json:"weighted_score"`
	ScoreGap      float64                      `json:"score_gap"`
	Factors       []FuturesStrategyFactorScore `json:"factors"`
}

type FuturesStrategyPerformanceSummary struct {
	SampleDays                int     `json:"sample_days"`
	WinRate                   float64 `json:"win_rate"`
	AvgDailyReturn            float64 `json:"avg_daily_return"`
	CumulativeReturn          float64 `json:"cumulative_return"`
	BenchmarkCumulativeReturn float64 `json:"benchmark_cumulative_return"`
	ExcessReturn              float64 `json:"excess_return"`
	MaxDrawdown               float64 `json:"max_drawdown"`
	BenchmarkSymbol           string  `json:"benchmark_symbol"`
	BenchmarkSource           string  `json:"benchmark_source"`
}

type FuturesStrategyInsight struct {
	Strategy         FuturesStrategy                   `json:"strategy"`
	Guidance         FuturesGuidance                   `json:"guidance"`
	ScoreFramework   FuturesStrategyScoreFramework     `json:"score_framework"`
	RelatedNews      []StockRecommendationRelatedNews  `json:"related_news"`
	RelatedEvents    []MarketEvent                     `json:"related_events"`
	Performance      []RecommendationPerformancePoint  `json:"performance"`
	Benchmark        []RecommendationPerformancePoint  `json:"benchmark"`
	PerformanceStats FuturesStrategyPerformanceSummary `json:"performance_stats"`
	GeneratedAt      string                            `json:"generated_at"`
}

type AdminUser struct {
	ID                 string `json:"id"`
	Phone              string `json:"phone"`
	Email              string `json:"email,omitempty"`
	Status             string `json:"status"`
	KYCStatus          string `json:"kyc_status"`
	MemberLevel        string `json:"member_level"`
	RegistrationSource string `json:"registration_source,omitempty"`
	InviterUserID      string `json:"inviter_user_id,omitempty"`
	InviteCode         string `json:"invite_code,omitempty"`
	InviteRegisteredAt string `json:"invite_registered_at,omitempty"`
	CreatedAt          string `json:"created_at"`
}

type AdminPermission struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Module      string `json:"module"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type AdminRole struct {
	ID              string   `json:"id"`
	RoleKey         string   `json:"role_key"`
	RoleName        string   `json:"role_name"`
	Description     string   `json:"description"`
	Status          string   `json:"status"`
	BuiltIn         bool     `json:"built_in"`
	PermissionCodes []string `json:"permission_codes"`
	UserCount       int      `json:"user_count"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type AdminAccount struct {
	ID        string   `json:"id"`
	Phone     string   `json:"phone"`
	Email     string   `json:"email"`
	Status    string   `json:"status"`
	RoleIDs   []string `json:"role_ids"`
	RoleKeys  []string `json:"role_keys"`
	RoleNames []string `json:"role_names"`
	LastLogin string   `json:"last_login"`
	CreatedAt string   `json:"created_at"`
}

type AdminRoleBrief struct {
	ID       string `json:"id"`
	RoleKey  string `json:"role_key"`
	RoleName string `json:"role_name"`
}

type AdminAccessProfile struct {
	UserID          string           `json:"user_id"`
	Role            string           `json:"role"`
	Roles           []AdminRoleBrief `json:"roles"`
	PermissionCodes []string         `json:"permission_codes"`
}

type AdminDashboardOverview struct {
	TotalUsers              int     `json:"total_users"`
	ActiveUsers             int     `json:"active_users"`
	KYCApprovedUsers        int     `json:"kyc_approved_users"`
	VIPUsers                int     `json:"vip_users"`
	ActiveSubscriptions     int     `json:"active_subscriptions"`
	PendingMembershipOrders int     `json:"pending_membership_orders"`
	TodayNewUsers           int     `json:"today_new_users"`
	TodayPaidOrders         int     `json:"today_paid_orders"`
	TodayPaidAmount         float64 `json:"today_paid_amount"`
	TodayPublishedStocks    int     `json:"today_published_stocks"`
	TodayPublishedNews      int     `json:"today_published_news"`
}

type AdminUserSourceSummary struct {
	TotalUsers            int     `json:"total_users"`
	DirectUsers           int     `json:"direct_users"`
	InvitedUsers          int     `json:"invited_users"`
	InviteRate            float64 `json:"invite_rate"`
	TodayInvitedUsers     int     `json:"today_invited_users"`
	Last7dInvitedUsers    int     `json:"last_7d_invited_users"`
	Last7dFirstPaidUsers  int     `json:"last_7d_first_paid_users"`
	Last7dConversionRate  float64 `json:"last_7d_conversion_rate"`
	Last30dInvitedUsers   int     `json:"last_30d_invited_users"`
	Last30dFirstPaidUsers int     `json:"last_30d_first_paid_users"`
	Last30dConversionRate float64 `json:"last_30d_conversion_rate"`
	TotalFirstPaidUsers   int     `json:"total_first_paid_users"`
	TotalConversionRate   float64 `json:"total_conversion_rate"`
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

type TushareNewsSyncOptions struct {
	BatchSize int      `json:"batch_size,omitempty"`
	Sources   []string `json:"sources,omitempty"`
	Symbols   []string `json:"symbols,omitempty"`
	SyncTypes []string `json:"sync_types,omitempty"`
}

type NewsSyncRunDetail struct {
	ID            string `json:"id"`
	RunID         string `json:"run_id"`
	JobName       string `json:"job_name"`
	SyncType      string `json:"sync_type"`
	Source        string `json:"source,omitempty"`
	Symbol        string `json:"symbol,omitempty"`
	Status        string `json:"status"`
	FetchedCount  int    `json:"fetched_count"`
	UpsertedCount int    `json:"upserted_count"`
	FailedCount   int    `json:"failed_count"`
	WarningText   string `json:"warning_text,omitempty"`
	ErrorText     string `json:"error_text,omitempty"`
	StartedAt     string `json:"started_at"`
	FinishedAt    string `json:"finished_at,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
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
	TodayTotal         int                         `json:"today_total"`
	TodaySuccess       int                         `json:"today_success"`
	TodayFailed        int                         `json:"today_failed"`
	TodayRunning       int                         `json:"today_running"`
	RetryTotal         int                         `json:"retry_total"`
	RetrySuccess       int                         `json:"retry_success"`
	RetryFailed        int                         `json:"retry_failed"`
	RetryHitRate       float64                     `json:"retry_hit_rate"`
	AvgRetryCount      float64                     `json:"avg_retry_count"`
	AutoRetryTotal     int                         `json:"auto_retry_total"`
	RecoveryTotal      int                         `json:"recovery_total"`
	RecoverySuccess    int                         `json:"recovery_success"`
	RecoveryHitRate    float64                     `json:"recovery_hit_rate"`
	FailureReasons     []SchedulerJobFailureReason `json:"failure_reasons"`
	JobRetryStats      []SchedulerJobRetryStat     `json:"job_retry_stats"`
	JobFailureReasons  []SchedulerJobFailureByJob  `json:"job_failure_reasons"`
	FailureReasonScope string                      `json:"failure_reason_scope"`
}

type SchedulerJobFailureReason struct {
	Reason         string `json:"reason"`
	Count          int    `json:"count"`
	LastOccurredAt string `json:"last_occurred_at"`
}

type SchedulerJobRetryStat struct {
	JobName         string  `json:"job_name"`
	TodayTotal      int     `json:"today_total"`
	TodaySuccess    int     `json:"today_success"`
	TodayFailed     int     `json:"today_failed"`
	TodayRunning    int     `json:"today_running"`
	RetryTotal      int     `json:"retry_total"`
	RetrySuccess    int     `json:"retry_success"`
	RetryFailed     int     `json:"retry_failed"`
	RetryHitRate    float64 `json:"retry_hit_rate"`
	AvgRetryCount   float64 `json:"avg_retry_count"`
	AutoRetryTotal  int     `json:"auto_retry_total"`
	RecoveryTotal   int     `json:"recovery_total"`
	RecoverySuccess int     `json:"recovery_success"`
	RecoveryHitRate float64 `json:"recovery_hit_rate"`
}

type SchedulerJobFailureByJob struct {
	JobName        string `json:"job_name"`
	Reason         string `json:"reason"`
	Count          int    `json:"count"`
	LastOccurredAt string `json:"last_occurred_at"`
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
