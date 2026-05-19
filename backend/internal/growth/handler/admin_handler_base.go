package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/utils"
)

const (
	ossProviderConfigKey        = "oss.provider"
	ossEnabledConfigKey         = "oss.enabled"
	ossQiniuAccessKeyConfigKey  = "oss.qiniu.access_key"
	ossQiniuSecretKeyConfigKey   = "oss.qiniu.secret_key"
	ossQiniuBucketConfigKey     = "oss.qiniu.bucket"
	ossQiniuDomainConfigKey     = "oss.qiniu.domain"
	ossQiniuRegionConfigKey     = "oss.qiniu.region"
	ossQiniuPathPrefixConfigKey = "oss.qiniu.path_prefix"
	ossQiniuUseHTTPSConfigKey   = "oss.qiniu.use_https"
	ossUploadMaxSizeMBConfigKey = "oss.upload.max_size_mb"
)

type ossUploadConfig struct {
	Provider    string
	Enabled     bool
	AccessKey   string
	SecretKey   string
	Bucket      string
	Domain      string
	Region      string
	PathPrefix  string
	UseHTTPS    bool
	MaxUploadMB int
}

type AdminBaseHandler struct {
	service service.GrowthService
	cfg     config.Config
}

type AdminHandlers struct {
	Finance          *AdminFinanceHandler
	Risk             *AdminRiskHandler
	News             *AdminNewsHandler
	Community        *AdminCommunityHandler
	System           *AdminSystemHandler
	MarketData       *AdminMarketDataHandler
	Forecast         *AdminForecastHandler
	StockSelection   *AdminStockSelectionHandler
	FuturesSelection *AdminFuturesSelectionHandler
	Strategy         *AdminStrategyHandler
	User             *AdminUserHandler
	Audit            *AdminAuditHandler
	Membership       *AdminMembershipHandler
}

func NewAdminHandlers(service service.GrowthService, cfg config.Config) *AdminHandlers {
	base := &AdminBaseHandler{service: service, cfg: cfg}
	return &AdminHandlers{
		Finance:          NewAdminFinanceHandler(base),
		Risk:             NewAdminRiskHandler(base),
		News:             NewAdminNewsHandler(base),
		Community:        NewAdminCommunityHandler(base),
		System:           NewAdminSystemHandler(base),
		MarketData:       NewAdminMarketDataHandler(base),
		Forecast:         NewAdminForecastHandler(base),
		StockSelection:   NewAdminStockSelectionHandler(base),
		FuturesSelection: NewAdminFuturesSelectionHandler(base),
		Strategy:         NewAdminStrategyHandler(base),
		User:             NewAdminUserHandler(base),
		Audit:            NewAdminAuditHandler(base),
		Membership:       NewAdminMembershipHandler(base),
	}
}

func (h *AdminBaseHandler) writeOperationLog(c *gin.Context, module string, action string, targetType string, targetID string, beforeValue string, afterValue string, reason string) {
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin"
	}
	_ = h.service.AdminCreateOperationLog(module, action, targetType, targetID, operator, beforeValue, afterValue, reason)
}

func (h *AdminBaseHandler) writeAuditEvent(c *gin.Context, item model.AdminAuditEvent) {
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin_unknown"
	}
	if strings.TrimSpace(item.ActorUserID) == "" {
		item.ActorUserID = operator
	}
	_ = h.service.AdminCreateAuditEvent(item)
}

func currentAdminOperator(c *gin.Context) string {
	if value, ok := c.Get("user_id"); ok {
		if operator, ok := value.(string); ok && strings.TrimSpace(operator) != "" {
			return strings.TrimSpace(operator)
		}
	}
	return "admin"
}

func (h *AdminBaseHandler) resolveOSSUploadConfig() ossUploadConfig {
	cfg := ossUploadConfig{
		Provider:    "QINIU",
		Enabled:     false,
		Region:      "z0",
		PathPrefix:  "uploads/",
		UseHTTPS:    true,
		MaxUploadMB: 20,
	}
	items, _, err := h.service.AdminListSystemConfigs("oss.", 1, 300)
	if err != nil {
		return cfg
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case ossProviderConfigKey:
			if value != "" {
				cfg.Provider = strings.ToUpper(value)
			}
		case ossEnabledConfigKey:
			cfg.Enabled = utils.ParseConfigBool(value, cfg.Enabled)
		case ossQiniuAccessKeyConfigKey:
			cfg.AccessKey = value
		case ossQiniuSecretKeyConfigKey:
			cfg.SecretKey = value
		case ossQiniuBucketConfigKey:
			cfg.Bucket = value
		case ossQiniuDomainConfigKey:
			cfg.Domain = value
		case ossQiniuRegionConfigKey:
			if value != "" {
				cfg.Region = strings.ToLower(value)
			}
		case ossQiniuPathPrefixConfigKey:
			cfg.PathPrefix = value
		case ossQiniuUseHTTPSConfigKey:
			cfg.UseHTTPS = utils.ParseConfigBool(value, cfg.UseHTTPS)
		case ossUploadMaxSizeMBConfigKey:
			cfg.MaxUploadMB = utils.ParseConfigInt(value, cfg.MaxUploadMB)
		}
	}
	if cfg.MaxUploadMB <= 0 {
		cfg.MaxUploadMB = 20
	}
	return cfg
}

func normalizeAdminDateTime(value string) (string, error) {
	return utils.NormalizeAdminDateTime(value)
}
func uniqueNonEmptyStrings(values []string) []string {
	return utils.UniqueNonEmptyStrings(values)
}

func normalizeUpperValues(values []string) []string {
	return utils.UniqueUpperStrings(values)
}

func normalizeStockSymbols(values []string) []string {
	return utils.UniqueUpperStrings(values)
}

func normalizeMarketStageSymbols(values []string) []string {
	return utils.UniqueUpperStrings(values)
}

func parseIntOrDefault(s string, def int) int {
	return utils.ParseIntOrDefault(s, def)
}

func parsePage(c *gin.Context) (int, int) {
	return utils.ParsePage(c)
}
