package config

import (
	"log"
	"strings"
)

// Development defaults compiled into this package. They exist so a fresh
// checkout runs without any setup; they must never be in effect in production.
const (
	DevJWTSecret     = "sercherai_dev_secret_change_me"
	DevDatabasePass  = "abc123"
	DevRedisPassword = "abc123"
)

// WarnInsecureDefaults logs loudly when a production deployment still runs on
// the development defaults above.
//
// It deliberately does not fail startup: a loud boot-log line is actionable,
// while refusing to boot turns one forgotten env var into an outage. The check
// is here rather than in main() so every entrypoint (API, workers, one-off
// tools) inherits it.
func (c Config) WarnInsecureDefaults() {
	if !strings.EqualFold(c.AppEnv, "production") && !strings.EqualFold(c.AppEnv, "prod") {
		return
	}
	if c.JWTSecret == DevJWTSecret {
		log.Printf("[security] JWT_SECRET is the built-in development value; anyone with this source can mint admin tokens. Set JWT_SECRET before serving traffic.")
	}
	if c.MySQLPass == DevDatabasePass {
		log.Printf("[security] MYSQL_PWD is the built-in development password.")
	}
	if c.RedisPass == DevRedisPassword {
		log.Printf("[security] REDIS_PWD is the built-in development password.")
	}
	if c.AllowMockLogin {
		log.Printf("[security] ALLOW_MOCK_LOGIN is enabled in production.")
	}
	if c.AttachmentSigningSecret == "" {
		log.Printf("[security] ATTACHMENT_SIGNING_SECRET is empty; attachment links cannot be signature-checked.")
	}
}
