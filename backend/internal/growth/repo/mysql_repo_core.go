package repo

import (
	"database/sql"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/go-redis/redis/v8"

	"sercherai/backend/internal/platform/config"
)

type MySQLGrowthRepo struct {
	db             *sql.DB
	redis          *redis.Client
	strategyEngine *strategyEngineClient
	strategyGraph  *strategyGraphClient
}

var repoIDSequence atomic.Uint64

func NewMySQLGrowthRepo(db *sql.DB, redisClient *redis.Client, cfg config.Config) *MySQLGrowthRepo {
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_email_logs (
			id varchar(64) PRIMARY KEY,
			user_id varchar(64) NOT NULL,
			email varchar(128) NOT NULL,
			subject varchar(256) NOT NULL,
			body text NOT NULL,
			rule_type varchar(64) NOT NULL,
			status varchar(32) NOT NULL,
			error_message text,
			sent_at datetime NOT NULL
		)
	`)

	return &MySQLGrowthRepo{
		db:             db,
		redis:          redisClient,
		strategyEngine: newStrategyEngineClient(cfg),
		strategyGraph:  newStrategyGraphClient(cfg),
	}
}

func normalizeInviteLinkURL(rawURL string, inviteCode string) string {
	trimmed := strings.TrimSpace(rawURL)
	code := strings.TrimSpace(inviteCode)
	if code == "" {
		return trimmed
	}
	if trimmed == "" {
		return "/invite/" + url.PathEscape(code)
	}
	lowerURL := strings.ToLower(trimmed)
	if strings.Contains(lowerURL, "example.com/invite/") {
		return "/invite/" + url.PathEscape(code)
	}
	return trimmed
}

func safeRatioValue(numerator int64, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}
