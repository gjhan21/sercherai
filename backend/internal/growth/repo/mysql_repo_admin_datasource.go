package repo

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminListDataSources(page int, pageSize int) ([]model.DataSource, int, error) {
	if err := r.ensureBuiltinDataSources(); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM system_configs WHERE config_key LIKE 'data_source.%'").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT id, config_key, config_value, description, updated_at
FROM system_configs
WHERE config_key LIKE 'data_source.%'
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.DataSource, 0)
	for rows.Next() {
		var id, configKey, configValue string
		var desc sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&id, &configKey, &configValue, &desc, &updatedAt); err != nil {
			return nil, 0, err
		}
		item := model.DataSource{
			ID:         id,
			SourceKey:  strings.TrimPrefix(configKey, "data_source."),
			Name:       "",
			SourceType: "",
			Status:     "ACTIVE",
			UpdatedAt:  updatedAt.Format(time.RFC3339),
		}
		var payload struct {
			Name       string                 `json:"name"`
			SourceType string                 `json:"source_type"`
			Status     string                 `json:"status"`
			Config     map[string]interface{} `json:"config"`
		}
		if err := json.Unmarshal([]byte(configValue), &payload); err == nil {
			item.Name = payload.Name
			item.SourceType = payload.SourceType
			if strings.TrimSpace(payload.Status) != "" {
				item.Status = payload.Status
			}
			item.Config = payload.Config
		}
		if item.Name == "" {
			if desc.Valid {
				item.Name = desc.String
			} else {
				item.Name = item.SourceKey
			}
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateDataSource(item model.DataSource) (string, error) {
	sourceKey := strings.TrimSpace(item.SourceKey)
	if sourceKey == "" {
		return "", errors.New("source_key is required")
	}
	configKey := "data_source." + sourceKey
	var exists int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM system_configs WHERE config_key = ?", configKey).Scan(&exists); err != nil {
		return "", err
	}
	if exists > 0 {
		return "", errors.New("data source already exists")
	}

	payloadBytes, err := json.Marshal(map[string]interface{}{
		"name":        item.Name,
		"source_type": strings.ToUpper(strings.TrimSpace(item.SourceType)),
		"status":      strings.ToUpper(strings.TrimSpace(item.Status)),
		"config":      item.Config,
	})
	if err != nil {
		return "", err
	}

	id := newID("cfg")
	_, err = r.db.Exec(`
INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		id, configKey, string(payloadBytes), item.Name, "admin", time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func cloneDataSourceConfigMap(value map[string]interface{}) map[string]interface{} {
	if len(value) == 0 {
		return map[string]interface{}{}
	}
	clone := make(map[string]interface{}, len(value))
	for key, item := range value {
		clone[key] = item
	}
	return clone
}

func mergeDataSourceConfigMap(existing map[string]interface{}, incoming map[string]interface{}) map[string]interface{} {
	merged := cloneDataSourceConfigMap(existing)
	for key, value := range incoming {
		merged[key] = value
	}
	return merged
}

func (r *MySQLGrowthRepo) AdminUpdateDataSource(sourceKey string, item model.DataSource) error {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return sql.ErrNoRows
	}
	configKey := "data_source." + sourceKey
	var existingConfigValue string
	if err := r.db.QueryRow("SELECT config_value FROM system_configs WHERE config_key = ?", configKey).Scan(&existingConfigValue); err != nil {
		return err
	}
	var existingPayload struct {
		Name       string                 `json:"name"`
		SourceType string                 `json:"source_type"`
		Status     string                 `json:"status"`
		Config     map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal([]byte(existingConfigValue), &existingPayload); err != nil {
		existingPayload.Config = map[string]interface{}{}
	}
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = strings.TrimSpace(existingPayload.Name)
	}
	sourceType := strings.ToUpper(strings.TrimSpace(item.SourceType))
	if sourceType == "" {
		sourceType = strings.ToUpper(strings.TrimSpace(existingPayload.SourceType))
	}
	status := strings.ToUpper(strings.TrimSpace(item.Status))
	if status == "" {
		status = strings.ToUpper(strings.TrimSpace(existingPayload.Status))
	}
	if status == "" {
		status = "ACTIVE"
	}
	mergedConfig := mergeDataSourceConfigMap(existingPayload.Config, item.Config)
	payloadBytes, err := json.Marshal(map[string]interface{}{
		"name":        name,
		"source_type": sourceType,
		"status":      status,
		"config":      mergedConfig,
	})
	if err != nil {
		return err
	}

	result, err := r.db.Exec(`
UPDATE system_configs
SET config_value = ?, description = ?, updated_by = ?, updated_at = ?
WHERE config_key = ?`,
		string(payloadBytes), name, "admin", time.Now(), configKey,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminDeleteDataSource(sourceKey string) error {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return sql.ErrNoRows
	}
	configKey := "data_source." + sourceKey
	result, err := r.db.Exec("DELETE FROM system_configs WHERE config_key = ?", configKey)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminCheckDataSourceHealth(sourceKey string) (model.DataSourceHealthCheck, error) {
	item, err := r.getDataSourceBySourceKey(sourceKey)
	if err != nil {
		return model.DataSourceHealthCheck{}, err
	}
	checkedAt := time.Now()
	result := model.DataSourceHealthCheck{
		SourceKey: item.SourceKey,
		Status:    "UNKNOWN",
		CheckedAt: checkedAt.Format(time.RFC3339),
	}
	previousStatus, err := r.getLatestDataSourceHealthStatus(item.SourceKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.DataSourceHealthCheck{}, err
	}

	if strings.ToUpper(strings.TrimSpace(item.Status)) != "ACTIVE" {
		result.Message = "data source is disabled"
		result.FailureCategory = "DISABLED"
	} else {
		provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
		isTushare := strings.EqualFold(strings.TrimSpace(item.SourceKey), "TUSHARE") || provider == "TUSHARE"
		isAkshare := strings.EqualFold(strings.TrimSpace(item.SourceKey), "AKSHARE") || provider == "AKSHARE"
		endpoint := parseDataSourceStringConfig(item.Config, "endpoint", "quotes_endpoint")
		if isTushare && strings.TrimSpace(endpoint) == "" {
			endpoint = "https://api.tushare.pro"
		}
		timeoutMS := parseDataSourceTimeoutMS(item.Config)
		retryTimes := parseDataSourceRetryTimes(item.Config)
		retryIntervalMS := parseDataSourceRetryIntervalMS(item.Config)
		result.MaxAttempts = retryTimes + 1

		if isAkshare {
			for attempt := 1; attempt <= result.MaxAttempts; attempt++ {
				result.Attempts = attempt
				attemptResult := performAkshareDataSourceHealthCheckAttempt(item.Config, timeoutMS)
				result.Status = attemptResult.Status
				result.Reachable = attemptResult.Reachable
				result.HTTPStatus = attemptResult.HTTPStatus
				result.LatencyMS = attemptResult.LatencyMS
				result.Message = attemptResult.Message
				result.FailureCategory = attemptResult.FailureCategory
				if result.Status == "HEALTHY" || result.Status == "UNKNOWN" {
					break
				}
				if attempt < result.MaxAttempts && retryIntervalMS > 0 {
					time.Sleep(time.Duration(retryIntervalMS) * time.Millisecond)
				}
			}
		} else if endpoint == "" {
			result.Message = "endpoint not configured"
			result.FailureCategory = "CONFIG_ERROR"
		} else {
			parsed, parseErr := url.Parse(endpoint)
			if parseErr != nil || parsed.Scheme == "" || parsed.Host == "" {
				result.Message = "invalid endpoint"
				result.FailureCategory = "CONFIG_ERROR"
			} else if parsed.Scheme != "http" && parsed.Scheme != "https" {
				result.Message = "unsupported endpoint scheme"
				result.FailureCategory = "CONFIG_ERROR"
			} else {
				tushareToken := ""
				if isTushare {
					tushareToken = parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token")
					if strings.TrimSpace(tushareToken) == "" {
						tushareToken = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
					}
					if strings.TrimSpace(tushareToken) == "" {
						result.Status = "UNHEALTHY"
						result.FailureCategory = "CONFIG_ERROR"
						result.Message = "tushare token not configured"
					}
				}

				for attempt := 1; attempt <= result.MaxAttempts; attempt++ {
					if isTushare && strings.TrimSpace(tushareToken) == "" {
						break
					}
					result.Attempts = attempt
					attemptResult := model.DataSourceHealthCheck{}
					if isTushare {
						attemptResult = performTushareDataSourceHealthCheckAttempt(endpoint, tushareToken, timeoutMS)
					} else {
						attemptResult = performDataSourceHealthCheckAttempt(endpoint, timeoutMS)
					}
					result.Status = attemptResult.Status
					result.Reachable = attemptResult.Reachable
					result.HTTPStatus = attemptResult.HTTPStatus
					result.LatencyMS = attemptResult.LatencyMS
					result.Message = attemptResult.Message
					result.FailureCategory = attemptResult.FailureCategory
					if result.Status == "HEALTHY" || result.Status == "UNKNOWN" {
						break
					}
					if attempt < result.MaxAttempts && retryIntervalMS > 0 {
						time.Sleep(time.Duration(retryIntervalMS) * time.Millisecond)
					}
				}
			}
		}
	}

	if err := r.persistDataSourceHealthLog(result, checkedAt); err != nil {
		return model.DataSourceHealthCheck{}, err
	}

	if result.Status == "UNHEALTHY" {
		consecutiveFailures, err := r.countConsecutiveDataSourceFailures(item.SourceKey, 50)
		if err != nil {
			return model.DataSourceHealthCheck{}, err
		}
		result.ConsecutiveFailures = consecutiveFailures
		threshold := parseDataSourceFailThreshold(item.Config)
		result.AlertTriggered = consecutiveFailures >= threshold
		if result.AlertTriggered && consecutiveFailures == threshold {
			r.createDataSourceHealthWorkflowMessage(item, result, "UNHEALTHY")
		}
	}
	if result.Status == "HEALTHY" && previousStatus == "UNHEALTHY" {
		r.createDataSourceHealthWorkflowMessage(item, result, "RECOVERED")
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminBatchCheckDataSourceHealth(sourceKeys []string) ([]model.DataSourceHealthCheck, error) {
	targets := make([]string, 0)
	seen := make(map[string]struct{})
	if len(sourceKeys) == 0 {
		rows, err := r.db.Query(`
SELECT config_key
FROM system_configs
WHERE config_key LIKE 'data_source.%'
ORDER BY updated_at DESC`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var configKey string
			if err := rows.Scan(&configKey); err != nil {
				return nil, err
			}
			sourceKey := strings.TrimSpace(strings.TrimPrefix(configKey, "data_source."))
			if sourceKey == "" {
				continue
			}
			if _, ok := seen[sourceKey]; ok {
				continue
			}
			seen[sourceKey] = struct{}{}
			targets = append(targets, sourceKey)
		}
	} else {
		for _, sourceKey := range sourceKeys {
			trimmed := strings.TrimSpace(sourceKey)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			targets = append(targets, trimmed)
		}
	}

	items := make([]model.DataSourceHealthCheck, 0, len(targets))
	for _, sourceKey := range targets {
		item, err := r.AdminCheckDataSourceHealth(sourceKey)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				items = append(items, model.DataSourceHealthCheck{
					SourceKey: sourceKey,
					Status:    "UNKNOWN",
					Message:   "data source not found",
					CheckedAt: time.Now().Format(time.RFC3339),
				})
				continue
			}
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) AdminListDataSourceHealthLogs(sourceKey string, page int, pageSize int) ([]model.DataSourceHealthLog, int, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return nil, 0, sql.ErrNoRows
	}
	if _, err := r.getDataSourceBySourceKey(sourceKey); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM data_source_health_logs WHERE source_key = ?", sourceKey).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, source_key, status, reachable, http_status, latency_ms, message, checked_at
FROM data_source_health_logs
WHERE source_key = ?
ORDER BY checked_at DESC, id DESC
LIMIT ? OFFSET ?`, sourceKey, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.DataSourceHealthLog, 0)
	for rows.Next() {
		var item model.DataSourceHealthLog
		var reachableInt int
		var httpStatus sql.NullInt64
		var message sql.NullString
		var checkedAt time.Time
		if err := rows.Scan(&item.ID, &item.SourceKey, &item.Status, &reachableInt, &httpStatus, &item.LatencyMS, &message, &checkedAt); err != nil {
			return nil, 0, err
		}
		item.Reachable = reachableInt == 1
		if httpStatus.Valid {
			item.HTTPStatus = int(httpStatus.Int64)
		}
		if message.Valid {
			item.Message = message.String
		}
		item.CheckedAt = checkedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) getDataSourceBySourceKey(sourceKey string) (model.DataSource, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return model.DataSource{}, sql.ErrNoRows
	}
	if err := r.ensureBuiltinDataSources(); err != nil {
		return model.DataSource{}, err
	}
	candidates := buildDataSourceLookupCandidates(sourceKey)
	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		trimmedCandidate := strings.TrimSpace(candidate)
		if trimmedCandidate == "" {
			continue
		}
		lookupKey := strings.ToLower(trimmedCandidate)
		if _, exists := seen[lookupKey]; exists {
			continue
		}
		seen[lookupKey] = struct{}{}

		configKey := "data_source." + trimmedCandidate
		var id, configValue string
		var desc sql.NullString
		var updatedAt time.Time
		var storedConfigKey string
		err := r.db.QueryRow(`
SELECT id, config_key, config_value, description, updated_at
FROM system_configs
WHERE LOWER(config_key) = LOWER(?)
LIMIT 1`, configKey).Scan(&id, &storedConfigKey, &configValue, &desc, &updatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return model.DataSource{}, err
		}
		actualSourceKey := strings.TrimSpace(strings.TrimPrefix(storedConfigKey, "data_source."))
		if actualSourceKey == "" {
			actualSourceKey = trimmedCandidate
		}
		item := model.DataSource{
			ID:         id,
			SourceKey:  actualSourceKey,
			Name:       "",
			SourceType: "",
			Status:     "ACTIVE",
			UpdatedAt:  updatedAt.Format(time.RFC3339),
		}
		var payload struct {
			Name       string                 `json:"name"`
			SourceType string                 `json:"source_type"`
			Status     string                 `json:"status"`
			Config     map[string]interface{} `json:"config"`
		}
		if err := json.Unmarshal([]byte(configValue), &payload); err == nil {
			item.Name = payload.Name
			item.SourceType = payload.SourceType
			if strings.TrimSpace(payload.Status) != "" {
				item.Status = payload.Status
			}
			item.Config = payload.Config
		}
		if item.Name == "" {
			if desc.Valid {
				item.Name = desc.String
			} else {
				item.Name = actualSourceKey
			}
		}
		return item, nil
	}
	return model.DataSource{}, sql.ErrNoRows
}

func buildDataSourceLookupCandidates(sourceKey string) []string {
	trimmed := strings.TrimSpace(sourceKey)
	if trimmed == "" {
		return nil
	}
	items := []string{trimmed}
	lowerKey := strings.ToLower(trimmed)
	upperKey := strings.ToUpper(trimmed)
	if lowerKey != trimmed {
		items = append(items, lowerKey)
	}
	if upperKey != trimmed && upperKey != lowerKey {
		items = append(items, upperKey)
	}
	switch upperKey {
	case "MOCK":
		items = append(items, "mock_stock", "MOCK_STOCK")
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func (r *MySQLGrowthRepo) ensureBuiltinDataSources() error {
	now := time.Now()
	type builtinDataSource struct {
		ID          string
		SourceKey   string
		ConfigValue string
		Description string
	}
	builtins := []builtinDataSource{
		{
			ID:          "cfg_data_source_mock_stock",
			SourceKey:   "mock_stock",
			ConfigValue: `{"name":"Mock Stock Quotes","source_type":"STOCK","status":"ACTIVE","config":{"provider":"MOCK","endpoint":"http://127.0.0.1:18080/healthz","retry_times":0,"fail_threshold":5,"retry_interval_ms":200,"health_timeout_ms":3000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置模拟股票行情数据源",
		},
		{
			ID:          "cfg_data_source_tushare",
			SourceKey:   "tushare",
			ConfigValue: `{"name":"Tushare","source_type":"STOCK","status":"ACTIVE","config":{"provider":"TUSHARE","endpoint":"https://api.tushare.pro","token":"","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置Tushare股票行情数据源",
		},
		{
			ID:          "cfg_data_source_akshare",
			SourceKey:   "akshare",
			ConfigValue: `{"name":"AkShare","source_type":"MARKET","status":"ACTIVE","config":{"provider":"AKSHARE","python_bin":"../services/strategy-engine/.venv/bin/python","bridge_script":"../services/strategy-engine/app/tools/market_bridge.py","retry_times":0,"fail_threshold":3,"retry_interval_ms":300,"health_timeout_ms":10000,"sync_timeout_ms":240000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置 AkShare 多市场数据源",
		},
		{
			ID:          "cfg_data_source_tickermd",
			SourceKey:   "tickermd",
			ConfigValue: `{"name":"TickerMD","source_type":"MARKET","status":"ACTIVE","config":{"provider":"TICKERMD","endpoint":"http://39.107.99.235:1008","quotes_endpoint":"http://39.107.99.235:1008/getQuote.php","kline_endpoint":"http://39.107.99.235:1008/redis.php","ws_endpoint":"ws://39.107.99.235/ws","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置 TickerMD 行情备份数据源",
		},
		{
			ID:          "cfg_data_source_myself",
			SourceKey:   "myself",
			ConfigValue: `{"name":"Myself","source_type":"MARKET","status":"ACTIVE","config":{"provider":"MYSELF","endpoint":"https://qt.gtimg.cn/q=s_sh000001","stock_kline_endpoint_tencent":"https://web.ifzq.gtimg.cn/appstock/app/fqkline/get","stock_kline_endpoint_sina":"https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData","futures_kline_endpoint_sina":"https://stock2.finance.sina.com.cn/futures/api/jsonp.php/var%20_TEST=/InnerFuturesNewService.getDailyKLine","referer":"https://finance.sina.com.cn","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置 Myself 新浪/腾讯聚合行情数据源",
		},
	}
	for _, item := range builtins {
		configKey := "data_source." + item.SourceKey
		_, err := r.db.Exec(`
	INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
	SELECT ?, ?, ?, ?, ?, ?
	FROM DUAL
	WHERE NOT EXISTS (
	  SELECT 1 FROM system_configs WHERE LOWER(config_key) = LOWER(?)
	)`,
			item.ID,
			configKey,
			item.ConfigValue,
			item.Description,
			"system",
			now,
			configKey,
		)
		if err != nil {
			return err
		}
	}
	defaultConfigs := []struct {
		Key         string
		Value       string
		Description string
	}{
		{
			Key:         "stock.quotes.default_source_key",
			Value:       "TUSHARE",
			Description: "股票行情默认数据源",
		},
		{
			Key:         "stock.master.default_source_key",
			Value:       "AKSHARE",
			Description: "股票主数据默认数据源",
		},
		{
			Key:         "stock.daily_basic.default_source_key",
			Value:       "AKSHARE",
			Description: "股票指标默认数据源",
		},
		{
			Key:         "stock.moneyflow.default_source_key",
			Value:       "AKSHARE",
			Description: "股票资金流默认数据源",
		},
		{
			Key:         "stock.news.default_source_key",
			Value:       "AKSHARE",
			Description: "股票资讯默认数据源",
		},
		{
			Key:         "futures.quotes.default_source_key",
			Value:       "TUSHARE",
			Description: "期货行情默认数据源",
		},
		{
			Key:         marketStockPriorityConfigKey,
			Value:       "TUSHARE,AKSHARE,TICKERMD,MOCK",
			Description: "股票日线多源优先级",
		},
		{
			Key:         marketFuturesPriorityConfigKey,
			Value:       "TUSHARE,TICKERMD,AKSHARE,MOCK",
			Description: "期货日线多源优先级",
		},
		{
			Key:         marketNewsPriorityConfigKey,
			Value:       "AKSHARE,TUSHARE",
			Description: "市场资讯多源优先级",
		},
		{
			Key:         "market.stock.factor.source_priority",
			Value:       "AKSHARE,TUSHARE",
			Description: "股票指标多源优先级",
		},
		{
			Key:         "market.stock.news.source_priority",
			Value:       "AKSHARE,TUSHARE",
			Description: "股票资讯多源优先级",
		},
		{
			Key:         marketInstrumentStockPriorityConfigKey,
			Value:       "TUSHARE,AKSHARE,TICKERMD,MYSELF,MOCK",
			Description: "股票主数据多源优先级",
		},
		{
			Key:         marketInstrumentFuturesPriorityConfigKey,
			Value:       "TUSHARE,AKSHARE,TICKERMD,MYSELF,MOCK",
			Description: "期货主数据多源优先级",
		},
		{
			Key:         "market.news.default_source_key",
			Value:       "AKSHARE",
			Description: "市场资讯默认数据源",
		},
	}
	for _, item := range defaultConfigs {
		_, err := r.db.Exec(`
	INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
	SELECT ?, ?, ?, ?, ?, ?
	FROM DUAL
	WHERE NOT EXISTS (
	  SELECT 1 FROM system_configs WHERE LOWER(config_key) = LOWER(?)
	)`,
			newID("cfg"),
			item.Key,
			item.Value,
			item.Description,
			"system",
			now,
			item.Key,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) persistDataSourceHealthLog(item model.DataSourceHealthCheck, checkedAt time.Time) error {
	reachable := 0
	if item.Reachable {
		reachable = 1
	}
	httpStatus := interface{}(nil)
	if item.HTTPStatus > 0 {
		httpStatus = item.HTTPStatus
	}
	_, err := r.db.Exec(`
INSERT INTO data_source_health_logs (id, source_key, status, reachable, http_status, latency_ms, message, checked_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("dshl"),
		item.SourceKey,
		item.Status,
		reachable,
		httpStatus,
		item.LatencyMS,
		item.Message,
		checkedAt,
	)
	return err
}

func (r *MySQLGrowthRepo) getLatestDataSourceHealthStatus(sourceKey string) (string, error) {
	var status string
	if err := r.db.QueryRow(`
SELECT status
FROM data_source_health_logs
WHERE source_key = ?
ORDER BY checked_at DESC, id DESC
LIMIT 1`, sourceKey).Scan(&status); err != nil {
		return "", err
	}
	return strings.ToUpper(strings.TrimSpace(status)), nil
}

func performDataSourceHealthCheckAttempt(endpoint string, timeoutMS int) model.DataSourceHealthCheck {
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	start := time.Now()
	req, reqErr := http.NewRequest(http.MethodGet, endpoint, nil)
	if reqErr != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNKNOWN",
			Message:         reqErr.Error(),
			FailureCategory: "REQUEST_ERROR",
		}
	}
	resp, callErr := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if callErr != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       false,
			LatencyMS:       latency,
			Message:         callErr.Error(),
			FailureCategory: classifyDataSourceRequestError(callErr),
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return model.DataSourceHealthCheck{
			Status:     "HEALTHY",
			Reachable:  true,
			HTTPStatus: resp.StatusCode,
			LatencyMS:  latency,
			Message:    "ok",
		}
	}
	return model.DataSourceHealthCheck{
		Status:          "UNHEALTHY",
		Reachable:       true,
		HTTPStatus:      resp.StatusCode,
		LatencyMS:       latency,
		Message:         resp.Status,
		FailureCategory: "HTTP_ERROR",
	}
}

func performTushareDataSourceHealthCheckAttempt(endpoint string, token string, timeoutMS int) model.DataSourceHealthCheck {
	if strings.TrimSpace(token) == "" {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Message:         "tushare token not configured",
			FailureCategory: "CONFIG_ERROR",
		}
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	start := time.Now()
	now := time.Now()
	startDate := now.AddDate(0, 0, -3).Format("20060102")
	endDate := now.Format("20060102")
	payload := map[string]interface{}{
		"api_name": "trade_cal",
		"token":    token,
		"params": map[string]string{
			"exchange":   "SSE",
			"start_date": startDate,
			"end_date":   endDate,
		},
		"fields": "exchange,cal_date,is_open",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNKNOWN",
			Message:         err.Error(),
			FailureCategory: "REQUEST_ERROR",
		}
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNKNOWN",
			Message:         err.Error(),
			FailureCategory: "REQUEST_ERROR",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       false,
			LatencyMS:       latency,
			Message:         err.Error(),
			FailureCategory: classifyDataSourceRequestError(err),
		}
	}
	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         readErr.Error(),
			FailureCategory: "RESPONSE_ERROR",
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         resp.Status,
			FailureCategory: "HTTP_ERROR",
		}
	}

	parsed := struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}{}
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         "invalid tushare response",
			FailureCategory: "RESPONSE_PARSE_ERROR",
		}
	}
	if parsed.Code != 0 {
		category := "AUTH_ERROR"
		msgLower := strings.ToLower(parsed.Msg)
		if !strings.Contains(msgLower, "token") && !strings.Contains(msgLower, "auth") {
			category = "API_ERROR"
		}
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         strings.TrimSpace(parsed.Msg),
			FailureCategory: category,
		}
	}
	return model.DataSourceHealthCheck{
		Status:     "HEALTHY",
		Reachable:  true,
		HTTPStatus: resp.StatusCode,
		LatencyMS:  latency,
		Message:    "ok",
	}
}

func classifyDataSourceRequestError(err error) string {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "TIMEOUT"
	}
	errorText := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errorText, "connection refused"),
		strings.Contains(errorText, "no such host"),
		strings.Contains(errorText, "dial tcp"),
		strings.Contains(errorText, "i/o timeout"):
		return "NETWORK_ERROR"
	default:
		return "REQUEST_ERROR"
	}
}

func (r *MySQLGrowthRepo) createDataSourceHealthWorkflowMessage(source model.DataSource, check model.DataSourceHealthCheck, event string) {
	receiverID := parseDataSourceAlertReceiverID(source.Config)
	if strings.TrimSpace(receiverID) == "" {
		return
	}
	switch strings.ToUpper(strings.TrimSpace(event)) {
	case "UNHEALTHY":
		threshold := parseDataSourceFailThreshold(source.Config)
		_ = r.AdminCreateAuditEvent(model.AdminAuditEvent{
			EventDomain: "DATA",
			EventType:   "DATA_SOURCE_UNHEALTHY",
			Level:       "WARNING",
			Module:      "SYSTEM",
			ObjectType:  "DATA_SOURCE",
			ObjectID:    source.SourceKey,
			ActorUserID: "system",
			Title:       "数据源健康告警",
			Summary:     fmt.Sprintf("数据源 %s 连续失败 %d 次（阈值 %d）", source.SourceKey, check.ConsecutiveFailures, threshold),
			Detail:      truncateByRunes(normalizeUTF8Text(check.Message), 512),
			Status:      "OPEN",
			DedupeKey:   "data_source_health:" + source.SourceKey + ":unhealthy",
			Metadata: map[string]any{
				"source_key":           source.SourceKey,
				"consecutive_failures": check.ConsecutiveFailures,
				"threshold":            threshold,
				"status":               check.Status,
				"receiver_id":          receiverID,
				"provider_governance":  true,
			},
		})
	case "RECOVERED":
		_ = r.AdminCreateAuditEvent(model.AdminAuditEvent{
			EventDomain: "DATA",
			EventType:   "DATA_SOURCE_RECOVERED",
			Level:       "INFO",
			Module:      "SYSTEM",
			ObjectType:  "DATA_SOURCE",
			ObjectID:    source.SourceKey,
			ActorUserID: "system",
			Title:       "数据源恢复通知",
			Summary:     fmt.Sprintf("数据源 %s 已恢复，最近耗时 %dms", source.SourceKey, check.LatencyMS),
			Status:      "OPEN",
			DedupeKey:   "data_source_health:" + source.SourceKey + ":recovered",
			Metadata: map[string]any{
				"source_key":  source.SourceKey,
				"latency_ms":  check.LatencyMS,
				"status":      check.Status,
				"receiver_id": receiverID,
			},
		})
	}
}

func (r *MySQLGrowthRepo) countConsecutiveDataSourceFailures(sourceKey string, limit int) (int, error) {
	rows, err := r.db.Query(`
SELECT status
FROM data_source_health_logs
WHERE source_key = ?
ORDER BY checked_at DESC, id DESC
LIMIT ?`, sourceKey, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return 0, err
		}
		if strings.ToUpper(strings.TrimSpace(status)) != "UNHEALTHY" {
			break
		}
		count++
	}
	return count, nil
}

func parseDataSourceFailThreshold(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "fail_threshold", 3, 1, 20)
}

func parseDataSourceTimeoutMS(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "health_timeout_ms", 3000, 500, 15000)
}

func parseDataSourceRetryTimes(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "retry_times", 0, 0, 5)
}

func parseDataSourceRetryIntervalMS(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "retry_interval_ms", 200, 0, 10000)
}

func parseDataSourceAlertReceiverID(config map[string]interface{}) string {
	if receiverID := parseDataSourceStringConfig(config, "alert_receiver_id", "receiver_id"); receiverID != "" {
		return receiverID
	}
	return "admin_001"
}

func parseDataSourceStringConfig(config map[string]interface{}, keys ...string) string {
	if config == nil {
		return ""
	}
	for _, key := range keys {
		raw, ok := config[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case string:
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		default:
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func parseDataSourceIntConfig(config map[string]interface{}, key string, defaultValue int, minValue int, maxValue int) int {
	if config == nil {
		return defaultValue
	}
	raw, ok := config[key]
	if !ok {
		return defaultValue
	}
	value, ok := parseFlexibleInt(raw)
	if !ok {
		return defaultValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func parseFlexibleInt(raw interface{}) (int, bool) {
	switch value := raw.(type) {
	case int:
		return value, true
	case int8:
		return int(value), true
	case int16:
		return int(value), true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case float32:
		return int(value), true
	case float64:
		return int(value), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
