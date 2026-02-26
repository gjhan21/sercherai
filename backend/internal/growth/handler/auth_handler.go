package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/platform/auth"
)

type AuthHandler struct {
	jwtSecret        string
	defaultExpires   int
	refreshExpires   int
	failThreshold    int
	ipFailThreshold  int
	ipPhoneThreshold int
	lockSeconds      int
	allowMockLogin   bool
	db               *sql.DB
	redis            *redis.Client
}

type riskConfig struct {
	phoneFailThreshold int
	ipFailThreshold    int
	ipPhoneThreshold   int
	lockSeconds        int
}

func NewAuthHandler(jwtSecret string, defaultExpires int, refreshExpires int, failThreshold int, ipFailThreshold int, ipPhoneThreshold int, lockSeconds int, allowMockLogin bool, db *sql.DB, redisClient *redis.Client) *AuthHandler {
	if defaultExpires <= 0 {
		defaultExpires = 86400
	}
	if refreshExpires <= 0 {
		refreshExpires = 604800
	}
	if failThreshold <= 0 {
		failThreshold = 5
	}
	if ipFailThreshold <= 0 {
		ipFailThreshold = 20
	}
	if ipPhoneThreshold <= 0 {
		ipPhoneThreshold = 5
	}
	if lockSeconds <= 0 {
		lockSeconds = 900
	}
	return &AuthHandler{
		jwtSecret:        jwtSecret,
		defaultExpires:   defaultExpires,
		refreshExpires:   refreshExpires,
		failThreshold:    failThreshold,
		ipFailThreshold:  ipFailThreshold,
		ipPhoneThreshold: ipPhoneThreshold,
		lockSeconds:      lockSeconds,
		allowMockLogin:   allowMockLogin,
		db:               db,
		redis:            redisClient,
	}
}

func (h *AuthHandler) MockLogin(c *gin.Context) {
	if !h.allowMockLogin {
		c.JSON(http.StatusForbidden, dto.APIResponse{Code: 40302, Message: "mock login disabled", Data: struct{}{}})
		return
	}
	var req dto.MockLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	expires := req.ExpireSeconds
	if expires <= 0 {
		expires = h.defaultExpires
	}
	role := strings.ToUpper(strings.TrimSpace(req.Role))

	token, err := auth.SignToken(h.jwtSecret, req.UserID, role, "ACCESS", time.Duration(expires)*time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	c.JSON(http.StatusOK, dto.OK(dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expires,
		UserID:      req.UserID,
		Role:        role,
	}))
	h.writeAuthLog(req.UserID, "", "MOCK_LOGIN", "SUCCESS", "", c)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	if exists, err := h.phoneExists(req.Phone); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	} else if exists {
		c.JSON(http.StatusConflict, dto.APIResponse{Code: 40902, Message: "phone already exists", Data: struct{}{}})
		return
	}

	passwordHash, err := bcryptHash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	userID := newID("u")
	now := time.Now()
	_, err = h.db.Exec(`
INSERT INTO users (id, phone, email, password_hash, status, kyc_status, member_level, created_at, updated_at)
VALUES (?, ?, ?, ?, 'ACTIVE', 'PENDING', 'FREE', ?, ?)`,
		userID, req.Phone, strings.TrimSpace(req.Email), passwordHash, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	role := resolveRole(userID)
	accessToken, refreshToken, err := h.issueTokenPair(userID, role, h.defaultExpires)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.defaultExpires,
		UserID:       userID,
		Role:         role,
	}))
	h.writeAuthLog(userID, req.Phone, "REGISTER", "SUCCESS", "", c)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	cfg, cfgErr := h.loadRiskConfig()
	if cfgErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: cfgErr.Error(), Data: struct{}{}})
		return
	}
	clientIP := c.ClientIP()
	if lockedUntil, lockType, locked, lockErr := h.checkRedisLock(clientIP, req.Phone); lockErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: lockErr.Error(), Data: struct{}{}})
		return
	} else if locked {
		c.JSON(http.StatusTooManyRequests, dto.APIResponse{Code: 42901, Message: "too many failed attempts", Data: gin.H{"locked_until": lockedUntil, "lock_type": lockType}})
		h.writeAuthLog("", req.Phone, "LOGIN", "FAILED", lockType, c)
		return
	}
	if lockedUntil, locked, lockErr := h.checkLoginLock(req.Phone); lockErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: lockErr.Error(), Data: struct{}{}})
		return
	} else if locked {
		c.JSON(http.StatusTooManyRequests, dto.APIResponse{Code: 42901, Message: "too many failed attempts", Data: gin.H{"locked_until": lockedUntil}})
		h.writeAuthLog("", req.Phone, "LOGIN", "FAILED", "locked", c)
		return
	}

	userID, passHash, status, err := h.loadUserByPhone(req.Phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, _ = h.recordLoginFailure(req.Phone, cfg)
			_, _, _ = h.recordRedisFailure(clientIP, req.Phone, cfg)
			c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40104, Message: "invalid credentials", Data: struct{}{}})
			h.writeAuthLog("", req.Phone, "LOGIN", "FAILED", "user_not_found", c)
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if strings.ToUpper(status) != "ACTIVE" {
		_, _ = h.recordLoginFailure(req.Phone, cfg)
		_, _, _ = h.recordRedisFailure(clientIP, req.Phone, cfg)
		c.JSON(http.StatusForbidden, dto.APIResponse{Code: 40303, Message: "user status invalid", Data: struct{}{}})
		h.writeAuthLog(userID, req.Phone, "LOGIN", "FAILED", "status_not_active", c)
		return
	}
	passwordMatch, legacyHash := verifyPassword(req.Password, passHash)
	if !passwordMatch {
		lockedUntil, _ := h.recordLoginFailure(req.Phone, cfg)
		redisLockedUntil, lockType, _ := h.recordRedisFailure(clientIP, req.Phone, cfg)
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40104, Message: "invalid credentials", Data: struct{}{}})
		if redisLockedUntil != "" {
			h.writeAuthLog(userID, req.Phone, "LOGIN", "FAILED", lockType, c)
		} else if lockedUntil != "" {
			h.writeAuthLog(userID, req.Phone, "LOGIN", "FAILED", "bad_password_locked", c)
		} else {
			h.writeAuthLog(userID, req.Phone, "LOGIN", "FAILED", "bad_password", c)
		}
		return
	}

	expires := req.ExpireSeconds
	if expires <= 0 {
		expires = h.defaultExpires
	}

	if legacyHash {
		if bcryptPass, hErr := bcryptHash(req.Password); hErr == nil {
			_, _ = h.db.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", bcryptPass, time.Now(), userID)
		}
	}
	role := resolveRole(userID)
	accessToken, refreshToken, err := h.issueTokenPair(userID, role, expires)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	c.JSON(http.StatusOK, dto.OK(dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expires,
		UserID:       userID,
		Role:         role,
	}))
	_ = h.clearLoginFailures(req.Phone)
	_ = h.clearRedisFailures(clientIP, req.Phone)
	h.writeAuthLog(userID, req.Phone, "LOGIN", "SUCCESS", "", c)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}

	claims, err := auth.ParseToken(h.jwtSecret, req.RefreshToken)
	if err != nil || strings.ToUpper(claims.TokenType) != "REFRESH" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid token", Data: struct{}{}})
		h.writeAuthLog("", "", "REFRESH", "FAILED", "invalid_refresh_token", c)
		return
	}

	oldTokenID, userID, err := h.loadActiveRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid token", Data: struct{}{}})
			h.writeAuthLog("", "", "REFRESH", "FAILED", "refresh_token_not_found", c)
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	role := resolveRole(userID)
	accessToken, err := auth.SignToken(h.jwtSecret, userID, role, "ACCESS", time.Duration(h.defaultExpires)*time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	newRefreshToken, err := auth.SignToken(h.jwtSecret, userID, role, "REFRESH", time.Duration(h.refreshExpires)*time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	newTokenID := newID("rt")
	now := time.Now()
	expAt := now.Add(time.Duration(h.refreshExpires) * time.Second)
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
VALUES (?, ?, ?, ?, 0, ?)`,
		newTokenID, userID, sha256Hex(newRefreshToken), expAt, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	_, err = tx.Exec(`
UPDATE refresh_tokens
SET revoked = 1, revoked_at = ?, replaced_by = ?
WHERE id = ?`, now, newTokenID, oldTokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	c.JSON(http.StatusOK, dto.OK(dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.defaultExpires,
		UserID:       userID,
		Role:         role,
	}))
	h.writeAuthLog(userID, "", "REFRESH", "SUCCESS", "", c)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}

	claims, err := auth.ParseToken(h.jwtSecret, req.RefreshToken)
	if err != nil || strings.ToUpper(claims.TokenType) != "REFRESH" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid token", Data: struct{}{}})
		h.writeAuthLog("", "", "LOGOUT", "FAILED", "invalid_refresh_token", c)
		return
	}

	tokenID, userID, err := h.loadActiveRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, dto.OK(struct{}{}))
			h.writeAuthLog(claims.UserID, "", "LOGOUT", "SUCCESS", "already_revoked_or_expired", c)
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	_, err = h.db.Exec("UPDATE refresh_tokens SET revoked = 1, revoked_at = ? WHERE id = ?", time.Now(), tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
	h.writeAuthLog(userID, "", "LOGOUT", "SUCCESS", "", c)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid token", Data: struct{}{}})
		return
	}
	userID, _ := userIDVal.(string)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid token", Data: struct{}{}})
		return
	}

	_, err := h.db.Exec(`
UPDATE refresh_tokens
SET revoked = 1, revoked_at = ?
WHERE user_id = ? AND revoked = 0`, time.Now(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
	h.writeAuthLog(userID, "", "LOGOUT_ALL", "SUCCESS", "", c)
}

func (h *AuthHandler) AdminListLoginLogs(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	page, pageSize := parsePage(c)
	offset := (page - 1) * pageSize
	action := strings.ToUpper(strings.TrimSpace(c.Query("action")))
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))

	args := []interface{}{}
	filter := ""
	if action != "" {
		filter += " AND action = ?"
		args = append(args, action)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM auth_login_logs WHERE 1=1" + filter
	if err := h.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	query := `
SELECT id, user_id, phone, action, status, reason, ip, user_agent, created_at
FROM auth_login_logs
WHERE 1=1` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer rows.Close()

	items := make([]gin.H, 0)
	for rows.Next() {
		var id, act, st string
		var userID, phone sql.NullString
		var reason, ip, userAgent sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &phone, &act, &st, &reason, &ip, &userAgent, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		items = append(items, gin.H{
			"id":         id,
			"user_id":    userID.String,
			"phone":      phone.String,
			"action":     act,
			"status":     st,
			"reason":     reason.String,
			"ip":         ip.String,
			"user_agent": userAgent.String,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AuthHandler) AdminExportLoginLogsCSV(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	action := strings.ToUpper(strings.TrimSpace(c.Query("action")))
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	dateFrom := strings.TrimSpace(c.Query("date_from"))
	dateTo := strings.TrimSpace(c.Query("date_to"))
	if dateFrom != "" {
		if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40003, Message: "invalid date_from format, expect YYYY-MM-DD", Data: struct{}{}})
			return
		}
	}
	if dateTo != "" {
		if _, err := time.Parse("2006-01-02", dateTo); err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40004, Message: "invalid date_to format, expect YYYY-MM-DD", Data: struct{}{}})
			return
		}
	}

	args := []interface{}{}
	filter := " WHERE 1=1"
	if action != "" {
		filter += " AND action = ?"
		args = append(args, action)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	if dateFrom != "" {
		filter += " AND created_at >= ?"
		args = append(args, dateFrom+" 00:00:00")
	}
	if dateTo != "" {
		filter += " AND created_at < ?"
		args = append(args, dateTo+" 23:59:59")
	}

	rows, err := h.db.Query(`
SELECT id, user_id, phone, action, status, reason, ip, user_agent, created_at
FROM auth_login_logs`+filter+`
ORDER BY created_at DESC`, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "user_id", "phone", "action", "status", "reason", "ip", "user_agent", "created_at"})
	for rows.Next() {
		var id, act, st string
		var userID, phone, reason, ip, userAgent sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &phone, &act, &st, &reason, &ip, &userAgent, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		_ = writer.Write([]string{
			id, userID.String, phone.String, act, st, reason.String, ip.String, userAgent.String, createdAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	fileName := "auth_login_logs.csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.String(http.StatusOK, buf.String())
}

func (h *AuthHandler) AdminGetRiskConfig(c *gin.Context) {
	cfg, err := h.loadRiskConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"phone_fail_threshold": cfg.phoneFailThreshold,
		"ip_fail_threshold":    cfg.ipFailThreshold,
		"ip_phone_threshold":   cfg.ipPhoneThreshold,
		"lock_seconds":         cfg.lockSeconds,
	}))
}

func (h *AuthHandler) AdminUpdateRiskConfig(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	oldCfg, oldErr := h.loadRiskConfig()
	if oldErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: oldErr.Error(), Data: struct{}{}})
		return
	}
	var req dto.AuthRiskConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	_, err := h.db.Exec(`
INSERT INTO auth_risk_configs (id, phone_fail_threshold, ip_fail_threshold, ip_phone_threshold, lock_seconds, updated_at)
VALUES ('default', ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
phone_fail_threshold = VALUES(phone_fail_threshold),
ip_fail_threshold = VALUES(ip_fail_threshold),
ip_phone_threshold = VALUES(ip_phone_threshold),
lock_seconds = VALUES(lock_seconds),
updated_at = VALUES(updated_at)`,
		req.PhoneFailThreshold, req.IPFailThreshold, req.IPPhoneThreshold, req.LockSeconds, time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorID, _ := c.Get("user_id")
	operator, _ := operatorID.(string)
	if operator == "" {
		operator = "admin_unknown"
	}
	_, _ = h.db.Exec(`
INSERT INTO auth_risk_config_logs (
id, operator_user_id, old_phone_fail, old_ip_fail, old_ip_phone, old_lock_seconds,
new_phone_fail, new_ip_fail, new_ip_phone, new_lock_seconds, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("arcl"), operator,
		oldCfg.phoneFailThreshold, oldCfg.ipFailThreshold, oldCfg.ipPhoneThreshold, oldCfg.lockSeconds,
		req.PhoneFailThreshold, req.IPFailThreshold, req.IPPhoneThreshold, req.LockSeconds,
		time.Now(),
	)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AuthHandler) AdminListRiskConfigLogs(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	page, pageSize := parsePage(c)
	offset := (page - 1) * pageSize

	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM auth_risk_config_logs").Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	rows, err := h.db.Query(`
SELECT id, operator_user_id, old_phone_fail, old_ip_fail, old_ip_phone, old_lock_seconds,
       new_phone_fail, new_ip_fail, new_ip_phone, new_lock_seconds, created_at
FROM auth_risk_config_logs
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer rows.Close()

	items := make([]gin.H, 0)
	for rows.Next() {
		var id, operator string
		var oldPhone, oldIP, oldIPPhone, oldLock, newPhone, newIP, newIPPhone, newLock int
		var createdAt time.Time
		if err := rows.Scan(&id, &operator, &oldPhone, &oldIP, &oldIPPhone, &oldLock, &newPhone, &newIP, &newIPPhone, &newLock, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		items = append(items, gin.H{
			"id":               id,
			"operator_user_id": operator,
			"old_config": gin.H{
				"phone_fail_threshold": oldPhone,
				"ip_fail_threshold":    oldIP,
				"ip_phone_threshold":   oldIPPhone,
				"lock_seconds":         oldLock,
			},
			"new_config": gin.H{
				"phone_fail_threshold": newPhone,
				"ip_fail_threshold":    newIP,
				"ip_phone_threshold":   newIPPhone,
				"lock_seconds":         newLock,
			},
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AuthHandler) AdminUnlockRiskState(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	var req dto.AuthUnlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	ip := strings.TrimSpace(req.IP)
	if phone == "" && ip == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "phone or ip required", Data: struct{}{}})
		return
	}
	reason := strings.TrimSpace(req.Reason)

	if phone != "" {
		if _, err := h.db.Exec("DELETE FROM auth_login_failures WHERE phone = ?", phone); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
	}
	if h.redis != nil {
		ctx := context.Background()
		keys := make([]string, 0, 4)
		if ip != "" {
			keys = append(keys, fmt.Sprintf("auth:fail:ip:%s", ip), fmt.Sprintf("auth:lock:ip:%s", ip))
		}
		if ip != "" && phone != "" {
			keys = append(keys, fmt.Sprintf("auth:fail:ip_phone:%s:%s", ip, phone), fmt.Sprintf("auth:lock:ip_phone:%s:%s", ip, phone))
		}
		if len(keys) > 0 {
			if err := h.redis.Del(ctx, keys...).Err(); err != nil {
				c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
				return
			}
		}
	}
	operatorID, _ := c.Get("user_id")
	operator, _ := operatorID.(string)
	if operator == "" {
		operator = "admin_unknown"
	}
	_, _ = h.db.Exec(`
INSERT INTO auth_unlock_logs (id, operator_user_id, phone, ip, reason, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		newID("aul"), operator, phone, ip, reason, time.Now(),
	)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AuthHandler) AdminListUnlockLogs(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	page, pageSize := parsePage(c)
	offset := (page - 1) * pageSize

	phone := strings.TrimSpace(c.Query("phone"))
	ip := strings.TrimSpace(c.Query("ip"))
	operator := strings.TrimSpace(c.Query("operator_user_id"))

	args := make([]interface{}, 0)
	filter := ""
	if phone != "" {
		filter += " AND phone = ?"
		args = append(args, phone)
	}
	if ip != "" {
		filter += " AND ip = ?"
		args = append(args, ip)
	}
	if operator != "" {
		filter += " AND operator_user_id = ?"
		args = append(args, operator)
	}

	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM auth_unlock_logs WHERE 1=1"+filter, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	query := `
SELECT id, operator_user_id, phone, ip, reason, created_at
FROM auth_unlock_logs
WHERE 1=1` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer rows.Close()

	items := make([]gin.H, 0)
	for rows.Next() {
		var id, operatorUserID string
		var phoneVal, ipVal, reason sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&id, &operatorUserID, &phoneVal, &ipVal, &reason, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		items = append(items, gin.H{
			"id":               id,
			"operator_user_id": operatorUserID,
			"phone":            phoneVal.String,
			"ip":               ipVal.String,
			"reason":           reason.String,
			"created_at":       createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"user_id": userID,
		"role":    role,
	}))
}

func (h *AuthHandler) loadUserByPhone(phone string) (string, string, string, error) {
	var userID, passwordHash, status string
	err := h.db.QueryRow("SELECT id, password_hash, status FROM users WHERE phone = ? LIMIT 1", phone).Scan(&userID, &passwordHash, &status)
	if err != nil {
		return "", "", "", err
	}
	return userID, passwordHash, status, nil
}

func (h *AuthHandler) phoneExists(phone string) (bool, error) {
	var cnt int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE phone = ?", phone).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (h *AuthHandler) issueTokenPair(userID string, role string, accessExpires int) (string, string, error) {
	accessToken, err := auth.SignToken(h.jwtSecret, userID, role, "ACCESS", time.Duration(accessExpires)*time.Second)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := auth.SignToken(h.jwtSecret, userID, role, "REFRESH", time.Duration(h.refreshExpires)*time.Second)
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	_, err = h.db.Exec(`
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
VALUES (?, ?, ?, ?, 0, ?)`,
		newID("rt"), userID, sha256Hex(refreshToken), now.Add(time.Duration(h.refreshExpires)*time.Second), now,
	)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (h *AuthHandler) loadActiveRefreshToken(token string) (string, string, error) {
	var id, userID string
	err := h.db.QueryRow(`
SELECT id, user_id
FROM refresh_tokens
WHERE token_hash = ? AND revoked = 0 AND expires_at > ?
LIMIT 1`, sha256Hex(token), time.Now(),
	).Scan(&id, &userID)
	if err != nil {
		return "", "", err
	}
	return id, userID, nil
}

func (h *AuthHandler) checkLoginLock(phone string) (string, bool, error) {
	var lockedUntil sql.NullTime
	err := h.db.QueryRow("SELECT locked_until FROM auth_login_failures WHERE phone = ? LIMIT 1", phone).Scan(&lockedUntil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	if lockedUntil.Valid && lockedUntil.Time.After(time.Now()) {
		return lockedUntil.Time.Format(time.RFC3339), true, nil
	}
	return "", false, nil
}

func (h *AuthHandler) loadRiskConfig() (riskConfig, error) {
	cfg := riskConfig{
		phoneFailThreshold: h.failThreshold,
		ipFailThreshold:    h.ipFailThreshold,
		ipPhoneThreshold:   h.ipPhoneThreshold,
		lockSeconds:        h.lockSeconds,
	}
	if h.db == nil {
		return cfg, nil
	}
	var phoneFail, ipFail, ipPhone, lockSec int
	err := h.db.QueryRow(`
SELECT phone_fail_threshold, ip_fail_threshold, ip_phone_threshold, lock_seconds
FROM auth_risk_configs
WHERE id = 'default'
LIMIT 1`).Scan(&phoneFail, &ipFail, &ipPhone, &lockSec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return cfg, err
	}
	if phoneFail > 0 {
		cfg.phoneFailThreshold = phoneFail
	}
	if ipFail > 0 {
		cfg.ipFailThreshold = ipFail
	}
	if ipPhone > 0 {
		cfg.ipPhoneThreshold = ipPhone
	}
	if lockSec > 0 {
		cfg.lockSeconds = lockSec
	}
	return cfg, nil
}

func (h *AuthHandler) recordLoginFailure(phone string, cfg riskConfig) (string, error) {
	now := time.Now()
	var failCount int
	var lockedUntil sql.NullTime
	err := h.db.QueryRow("SELECT fail_count, locked_until FROM auth_login_failures WHERE phone = ? LIMIT 1", phone).Scan(&failCount, &lockedUntil)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if errors.Is(err, sql.ErrNoRows) {
		failCount = 0
	}
	if lockedUntil.Valid && lockedUntil.Time.After(now) {
		return lockedUntil.Time.Format(time.RFC3339), nil
	}
	failCount++
	if failCount >= cfg.phoneFailThreshold {
		lockAt := now.Add(time.Duration(cfg.lockSeconds) * time.Second)
		_, upErr := h.db.Exec(`
INSERT INTO auth_login_failures (phone, fail_count, locked_until, updated_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE fail_count = VALUES(fail_count), locked_until = VALUES(locked_until), updated_at = VALUES(updated_at)`,
			phone, failCount, lockAt, now,
		)
		if upErr != nil {
			return "", upErr
		}
		return lockAt.Format(time.RFC3339), nil
	}
	_, upErr := h.db.Exec(`
INSERT INTO auth_login_failures (phone, fail_count, locked_until, updated_at)
VALUES (?, ?, NULL, ?)
ON DUPLICATE KEY UPDATE fail_count = VALUES(fail_count), locked_until = NULL, updated_at = VALUES(updated_at)`,
		phone, failCount, now,
	)
	if upErr != nil {
		return "", upErr
	}
	return "", nil
}

func (h *AuthHandler) clearLoginFailures(phone string) error {
	_, err := h.db.Exec("DELETE FROM auth_login_failures WHERE phone = ?", phone)
	return err
}

func (h *AuthHandler) checkRedisLock(ip string, phone string) (string, string, bool, error) {
	if h.redis == nil {
		return "", "", false, nil
	}
	ctx := context.Background()
	ipPhoneLock, err := h.redis.Get(ctx, fmt.Sprintf("auth:lock:ip_phone:%s:%s", ip, phone)).Result()
	if err != nil && err != redis.Nil {
		return "", "", false, err
	}
	if err == nil && ipPhoneLock != "" {
		return ipPhoneLock, "ip_phone_locked", true, nil
	}
	ipLock, err := h.redis.Get(ctx, fmt.Sprintf("auth:lock:ip:%s", ip)).Result()
	if err != nil && err != redis.Nil {
		return "", "", false, err
	}
	if err == nil && ipLock != "" {
		return ipLock, "ip_locked", true, nil
	}
	return "", "", false, nil
}

func (h *AuthHandler) recordRedisFailure(ip string, phone string, cfg riskConfig) (string, string, error) {
	if h.redis == nil {
		return "", "", nil
	}
	ctx := context.Background()
	ttl := time.Duration(cfg.lockSeconds) * time.Second
	now := time.Now()

	ipKey := fmt.Sprintf("auth:fail:ip:%s", ip)
	ipPhoneKey := fmt.Sprintf("auth:fail:ip_phone:%s:%s", ip, phone)

	ipCount, err := h.redis.Incr(ctx, ipKey).Result()
	if err != nil {
		return "", "", err
	}
	if ipCount == 1 {
		_ = h.redis.Expire(ctx, ipKey, ttl).Err()
	}

	ipPhoneCount, err := h.redis.Incr(ctx, ipPhoneKey).Result()
	if err != nil {
		return "", "", err
	}
	if ipPhoneCount == 1 {
		_ = h.redis.Expire(ctx, ipPhoneKey, ttl).Err()
	}

	lockedUntil := ""
	lockType := ""
	if int(ipPhoneCount) >= cfg.ipPhoneThreshold {
		until := now.Add(ttl).Format(time.RFC3339)
		lockKey := fmt.Sprintf("auth:lock:ip_phone:%s:%s", ip, phone)
		if err := h.redis.Set(ctx, lockKey, until, ttl).Err(); err != nil {
			return "", "", err
		}
		lockedUntil = until
		lockType = "ip_phone_locked"
	}
	if int(ipCount) >= cfg.ipFailThreshold {
		until := now.Add(ttl).Format(time.RFC3339)
		lockKey := fmt.Sprintf("auth:lock:ip:%s", ip)
		if err := h.redis.Set(ctx, lockKey, until, ttl).Err(); err != nil {
			return "", "", err
		}
		if lockType == "" {
			lockedUntil = until
			lockType = "ip_locked"
		}
	}
	return lockedUntil, lockType, nil
}

func (h *AuthHandler) clearRedisFailures(ip string, phone string) error {
	if h.redis == nil {
		return nil
	}
	ctx := context.Background()
	keys := []string{
		fmt.Sprintf("auth:fail:ip_phone:%s:%s", ip, phone),
		fmt.Sprintf("auth:lock:ip_phone:%s:%s", ip, phone),
	}
	return h.redis.Del(ctx, keys...).Err()
}

func (h *AuthHandler) writeAuthLog(userID string, phone string, action string, status string, reason string, c *gin.Context) {
	if h.db == nil {
		return
	}
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	_, _ = h.db.Exec(`
INSERT INTO auth_login_logs (id, user_id, phone, action, status, reason, ip, user_agent, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("alog"), userID, strings.TrimSpace(phone), strings.ToUpper(action), strings.ToUpper(status), reason, ip, userAgent, time.Now(),
	)
}

func verifyPassword(plain string, storedHash string) (bool, bool) {
	if strings.HasPrefix(storedHash, "$2a$") || strings.HasPrefix(storedHash, "$2b$") || strings.HasPrefix(storedHash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plain))
		return err == nil, false
	}
	return sha256Hex(plain) == storedHash, true
}

func bcryptHash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func resolveRole(userID string) string {
	if strings.HasPrefix(strings.ToLower(userID), "admin_") {
		return "ADMIN"
	}
	return "USER"
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
