package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv                  string
	AppPort                 string
	JWTSecret               string
	JWTExpireSeconds        int
	JWTRefreshExpireSeconds int
	LoginFailThreshold      int
	LoginIPFailThreshold    int
	LoginIPPhoneThreshold   int
	LoginLockSeconds        int

	MySQLHost string
	MySQLPort string
	MySQLUser string
	MySQLPass string
	MySQLDB   string

	RedisHost string
	RedisPort string
	RedisPass string

	PublicBaseURL              string
	AllowMockLogin             bool
	AllowJobSimulation         bool
	AttachmentSigningSecret    string
	AttachmentSigningTTLSecond int
	PaymentSigningSecret       string
}

func Load() Config {
	appEnv := getEnv("APP_ENV", "dev")
	return Config{
		AppEnv:                  appEnv,
		AppPort:                 getEnv("APP_PORT", "8080"),
		JWTSecret:               getEnv("JWT_SECRET", "sercherai_dev_secret_change_me"),
		JWTExpireSeconds:        getEnvInt("JWT_EXPIRE_SECONDS", 86400),
		JWTRefreshExpireSeconds: getEnvInt("JWT_REFRESH_EXPIRE_SECONDS", 604800),
		LoginFailThreshold:      getEnvInt("LOGIN_FAIL_THRESHOLD", 5),
		LoginIPFailThreshold:    getEnvInt("LOGIN_IP_FAIL_THRESHOLD", 20),
		LoginIPPhoneThreshold:   getEnvInt("LOGIN_IP_PHONE_THRESHOLD", 5),
		LoginLockSeconds:        getEnvInt("LOGIN_LOCK_SECONDS", 900),

		MySQLHost: getEnv("MYSQL_HOST", "127.0.0.1"),
		MySQLPort: getEnv("MYSQL_PORT", "3306"),
		MySQLUser: getEnv("MYSQL_USER", "root"),
		MySQLPass: getEnv("MYSQL_PWD", "abc123"),
		MySQLDB:   getEnv("MYSQL_DB", "sercherai"),

		RedisHost: getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		RedisPass: getEnv("REDIS_PWD", "abc123"),

		PublicBaseURL:              getEnv("PUBLIC_BASE_URL", "http://127.0.0.1:8080"),
		AllowMockLogin:             getEnvBool("ALLOW_MOCK_LOGIN", false),
		AllowJobSimulation:         getEnvBool("ALLOW_JOB_SIMULATION", appEnv != "production"),
		AttachmentSigningSecret:    getEnv("ATTACHMENT_SIGNING_SECRET", ""),
		AttachmentSigningTTLSecond: getEnvInt("ATTACHMENT_SIGNING_TTL_SECONDS", 300),
		PaymentSigningSecret:       getEnv("PAYMENT_SIGNING_SECRET", ""),
	}
}

func getEnv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "no", "NO", "off", "OFF":
		return false
	default:
		return def
	}
}
