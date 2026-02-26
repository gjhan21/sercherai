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
	"sercherai/backend/internal/growth/model"
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

func (h *AuthHandler) AdminGetAccessProfile(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	roleValue, _ := c.Get("role")
	role, _ := roleValue.(string)
	profile, err := h.loadAdminAccessProfile(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(profile))
}

func (h *AuthHandler) AdminListPermissions(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	page, pageSize := parsePage(c)
	offset := (page - 1) * pageSize
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	module := strings.ToUpper(strings.TrimSpace(c.Query("module")))
	keyword := strings.TrimSpace(c.Query("keyword"))

	filter := " WHERE 1=1"
	args := make([]interface{}, 0)
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	if module != "" {
		filter += " AND module = ?"
		args = append(args, module)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		filter += " AND (code LIKE ? OR name LIKE ?)"
		args = append(args, like, like)
	}

	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM rbac_permissions"+filter, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	query := `
SELECT code, name, module, action, description, status, created_at, updated_at
FROM rbac_permissions` + filter + `
ORDER BY module ASC, code ASC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer rows.Close()

	items := make([]model.AdminPermission, 0)
	for rows.Next() {
		var item model.AdminPermission
		var description sql.NullString
		var createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.Code,
			&item.Name,
			&item.Module,
			&item.Action,
			&description,
			&item.Status,
			&createdAt,
			&updatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		if description.Valid {
			item.Description = description.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}))
}

func (h *AuthHandler) AdminListRoles(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	page, pageSize := parsePage(c)
	offset := (page - 1) * pageSize
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	keyword := strings.TrimSpace(c.Query("keyword"))

	filter := " WHERE 1=1"
	args := make([]interface{}, 0)
	if status != "" {
		filter += " AND r.status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		filter += " AND (r.role_key LIKE ? OR r.role_name LIKE ?)"
		args = append(args, like, like)
	}

	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM rbac_roles r"+filter, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	query := `
SELECT
  r.id,
  r.role_key,
  r.role_name,
  r.description,
  r.status,
  r.built_in,
  r.created_at,
  r.updated_at,
  (SELECT COUNT(*) FROM rbac_user_roles ur WHERE ur.role_id = r.id) AS user_count
FROM rbac_roles r` + filter + `
ORDER BY r.built_in DESC, r.role_key ASC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer rows.Close()

	items := make([]model.AdminRole, 0)
	roleIDs := make([]string, 0)
	for rows.Next() {
		var item model.AdminRole
		var description sql.NullString
		var createdAt, updatedAt time.Time
		var builtIn int
		if err := rows.Scan(
			&item.ID,
			&item.RoleKey,
			&item.RoleName,
			&description,
			&item.Status,
			&builtIn,
			&createdAt,
			&updatedAt,
			&item.UserCount,
		); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		if description.Valid {
			item.Description = description.String
		}
		item.BuiltIn = builtIn == 1
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
		roleIDs = append(roleIDs, item.ID)
	}

	rolePermissionMap, err := h.loadRolePermissionCodes(roleIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	for i := range items {
		items[i].PermissionCodes = rolePermissionMap[items[i].ID]
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}))
}

func (h *AuthHandler) AdminCreateRole(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	var req dto.AdminRoleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	roleKey := strings.ToUpper(strings.TrimSpace(req.RoleKey))
	if roleKey == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "role_key required", Data: struct{}{}})
		return
	}
	roleName := strings.TrimSpace(req.RoleName)
	if roleName == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "role_name required", Data: struct{}{}})
		return
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = "ACTIVE"
	}
	permissionCodes := normalizeCodeList(req.PermissionCodes, false)
	if len(permissionCodes) == 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "permission_codes required", Data: struct{}{}})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer tx.Rollback()

	if err := h.assertPermissionsExistTx(tx, permissionCodes); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: err.Error(), Data: struct{}{}})
		return
	}

	roleID := newID("role")
	now := time.Now()
	_, err = tx.Exec(`
INSERT INTO rbac_roles (id, role_key, role_name, description, status, built_in, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 0, ?, ?)`,
		roleID,
		roleKey,
		roleName,
		strings.TrimSpace(req.Description),
		status,
		now,
		now,
	)
	if err != nil {
		if isDuplicateEntry(err) {
			c.JSON(http.StatusConflict, dto.APIResponse{Code: 40902, Message: "role_key already exists", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	if err := h.replaceRolePermissionsTx(tx, roleID, permissionCodes); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": roleID}))
}

func (h *AuthHandler) AdminUpdateRole(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	roleID := strings.TrimSpace(c.Param("id"))
	if roleID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "role id required", Data: struct{}{}})
		return
	}
	var req dto.AdminRoleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	roleName := strings.TrimSpace(req.RoleName)
	if roleName == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "role_name required", Data: struct{}{}})
		return
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = "ACTIVE"
	}
	permissionCodes := normalizeCodeList(req.PermissionCodes, false)
	if len(permissionCodes) == 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "permission_codes required", Data: struct{}{}})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer tx.Rollback()

	var roleKey string
	var builtIn int
	if err := tx.QueryRow("SELECT role_key, built_in FROM rbac_roles WHERE id = ? LIMIT 1", roleID).Scan(&roleKey, &builtIn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "role not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if strings.EqualFold(roleKey, "SUPER_ADMIN") && status != "ACTIVE" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "SUPER_ADMIN must stay ACTIVE", Data: struct{}{}})
		return
	}
	if err := h.assertPermissionsExistTx(tx, permissionCodes); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: err.Error(), Data: struct{}{}})
		return
	}

	_, err = tx.Exec(`
UPDATE rbac_roles
SET role_name = ?, description = ?, status = ?, updated_at = ?
WHERE id = ?`,
		roleName,
		strings.TrimSpace(req.Description),
		status,
		time.Now(),
		roleID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.replaceRolePermissionsTx(tx, roleID, permissionCodes); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AuthHandler) AdminUpdateRoleStatus(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	roleID := strings.TrimSpace(c.Param("id"))
	var req dto.AdminUpdateAccountStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	var roleKey string
	if err := h.db.QueryRow("SELECT role_key FROM rbac_roles WHERE id = ? LIMIT 1", roleID).Scan(&roleKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "role not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if strings.EqualFold(roleKey, "SUPER_ADMIN") && status != "ACTIVE" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "SUPER_ADMIN must stay ACTIVE", Data: struct{}{}})
		return
	}
	if _, err := h.db.Exec("UPDATE rbac_roles SET status = ?, updated_at = ? WHERE id = ?", status, time.Now(), roleID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AuthHandler) AdminListAdminUsers(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	page, pageSize := parsePage(c)
	offset := (page - 1) * pageSize
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	keyword := strings.TrimSpace(c.Query("keyword"))
	roleID := strings.TrimSpace(c.Query("role_id"))

	filter := " WHERE (u.id LIKE 'admin_%' OR EXISTS (SELECT 1 FROM rbac_user_roles urx WHERE urx.user_id = u.id))"
	args := make([]interface{}, 0)
	if status != "" {
		filter += " AND u.status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		filter += " AND (u.id LIKE ? OR u.phone LIKE ? OR u.email LIKE ?)"
		args = append(args, like, like, like)
	}
	if roleID != "" {
		filter += " AND EXISTS (SELECT 1 FROM rbac_user_roles urf WHERE urf.user_id = u.id AND urf.role_id = ?)"
		args = append(args, roleID)
	}

	var total int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM users u"+filter, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	query := `
SELECT u.id, u.phone, u.email, u.status, u.created_at
FROM users u` + filter + `
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer rows.Close()

	items := make([]model.AdminAccount, 0)
	userIDs := make([]string, 0)
	for rows.Next() {
		var item model.AdminAccount
		var email sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Phone, &email, &item.Status, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		if email.Valid {
			item.Email = email.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
		userIDs = append(userIDs, item.ID)
	}

	roleMap, err := h.loadUserRoleMap(userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	lastLoginMap, err := h.loadUserLastLoginMap(userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	for i := range items {
		roleEntries := roleMap[items[i].ID]
		items[i].RoleIDs = make([]string, 0, len(roleEntries))
		items[i].RoleKeys = make([]string, 0, len(roleEntries))
		items[i].RoleNames = make([]string, 0, len(roleEntries))
		for _, roleEntry := range roleEntries {
			items[i].RoleIDs = append(items[i].RoleIDs, roleEntry.ID)
			items[i].RoleKeys = append(items[i].RoleKeys, roleEntry.RoleKey)
			items[i].RoleNames = append(items[i].RoleNames, roleEntry.RoleName)
		}
		if lastLoginAt, exists := lastLoginMap[items[i].ID]; exists {
			items[i].LastLogin = lastLoginAt
		}
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}))
}

func (h *AuthHandler) AdminCreateAdminUser(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	var req dto.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "phone required", Data: struct{}{}})
		return
	}
	roleIDs := normalizeCodeList(req.RoleIDs, false)
	if len(roleIDs) == 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "role_ids required", Data: struct{}{}})
		return
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = "ACTIVE"
	}
	if exists, err := h.phoneExists(phone); err != nil {
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

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer tx.Rollback()

	if err := h.assertRolesExistTx(tx, roleIDs); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: err.Error(), Data: struct{}{}})
		return
	}

	userID := fmt.Sprintf("admin_%d", time.Now().UnixNano())
	now := time.Now()
	_, err = tx.Exec(`
INSERT INTO users (id, phone, email, password_hash, status, kyc_status, member_level, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'APPROVED', 'VIP3', ?, ?)`,
		userID,
		phone,
		strings.TrimSpace(req.Email),
		passwordHash,
		status,
		now,
		now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.replaceUserRolesTx(tx, userID, roleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": userID}))
}

func (h *AuthHandler) AdminUpdateAdminUserStatus(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	userID := strings.TrimSpace(c.Param("id"))
	if !strings.HasPrefix(strings.ToLower(userID), "admin_") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "admin user id required", Data: struct{}{}})
		return
	}
	var req dto.AdminUpdateAccountStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorID, _ := c.Get("user_id")
	operator, _ := operatorID.(string)
	if strings.EqualFold(operator, userID) && strings.ToUpper(req.Status) != "ACTIVE" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "cannot disable current account", Data: struct{}{}})
		return
	}
	result, err := h.db.Exec("UPDATE users SET status = ?, updated_at = ? WHERE id = ?", strings.ToUpper(req.Status), time.Now(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "admin user not found", Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AuthHandler) AdminAssignAdminUserRoles(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	userID := strings.TrimSpace(c.Param("id"))
	if !strings.HasPrefix(strings.ToLower(userID), "admin_") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "admin user id required", Data: struct{}{}})
		return
	}
	var req dto.AdminAssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	roleIDs := normalizeCodeList(req.RoleIDs, false)
	if len(roleIDs) == 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "role_ids required", Data: struct{}{}})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&exists); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if exists == 0 {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "admin user not found", Data: struct{}{}})
		return
	}
	if err := h.assertRolesExistTx(tx, roleIDs); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.replaceUserRolesTx(tx, userID, roleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AuthHandler) AdminResetAdminUserPassword(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "auth db unavailable", Data: struct{}{}})
		return
	}
	userID := strings.TrimSpace(c.Param("id"))
	if !strings.HasPrefix(strings.ToLower(userID), "admin_") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "admin user id required", Data: struct{}{}})
		return
	}
	var req dto.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	passwordHash, err := bcryptHash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	result, err := h.db.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", passwordHash, time.Now(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "admin user not found", Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
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

func (h *AuthHandler) loadAdminAccessProfile(userID string, role string) (model.AdminAccessProfile, error) {
	profile := model.AdminAccessProfile{
		UserID:          userID,
		Role:            strings.ToUpper(strings.TrimSpace(role)),
		Roles:           make([]model.AdminRoleBrief, 0),
		PermissionCodes: make([]string, 0),
	}

	roleRows, err := h.db.Query(`
SELECT r.id, r.role_key, r.role_name
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
WHERE ur.user_id = ? AND r.status = 'ACTIVE'
ORDER BY r.role_key ASC`, userID)
	if err != nil {
		return model.AdminAccessProfile{}, err
	}
	defer roleRows.Close()
	for roleRows.Next() {
		var item model.AdminRoleBrief
		if err := roleRows.Scan(&item.ID, &item.RoleKey, &item.RoleName); err != nil {
			return model.AdminAccessProfile{}, err
		}
		profile.Roles = append(profile.Roles, item)
	}

	permissionRows, err := h.db.Query(`
SELECT DISTINCT p.code
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
JOIN rbac_role_permissions rp ON rp.role_id = r.id
JOIN rbac_permissions p ON p.code = rp.permission_code
WHERE ur.user_id = ? AND r.status = 'ACTIVE' AND p.status = 'ACTIVE'
ORDER BY p.code ASC`, userID)
	if err != nil {
		return model.AdminAccessProfile{}, err
	}
	defer permissionRows.Close()
	for permissionRows.Next() {
		var code string
		if err := permissionRows.Scan(&code); err != nil {
			return model.AdminAccessProfile{}, err
		}
		profile.PermissionCodes = append(profile.PermissionCodes, code)
	}
	return profile, nil
}

func (h *AuthHandler) loadRolePermissionCodes(roleIDs []string) (map[string][]string, error) {
	result := map[string][]string{}
	if len(roleIDs) == 0 {
		return result, nil
	}
	placeholders := sqlPlaceholders(len(roleIDs))
	args := make([]interface{}, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		args = append(args, roleID)
	}
	rows, err := h.db.Query(`
SELECT role_id, permission_code
FROM rbac_role_permissions
WHERE role_id IN (`+placeholders+`)
ORDER BY permission_code ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleID, permissionCode string
		if err := rows.Scan(&roleID, &permissionCode); err != nil {
			return nil, err
		}
		result[roleID] = append(result[roleID], permissionCode)
	}
	return result, nil
}

func (h *AuthHandler) loadUserRoleMap(userIDs []string) (map[string][]model.AdminRoleBrief, error) {
	result := map[string][]model.AdminRoleBrief{}
	if len(userIDs) == 0 {
		return result, nil
	}
	placeholders := sqlPlaceholders(len(userIDs))
	args := make([]interface{}, 0, len(userIDs))
	for _, userID := range userIDs {
		args = append(args, userID)
	}
	rows, err := h.db.Query(`
SELECT ur.user_id, r.id, r.role_key, r.role_name
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
WHERE ur.user_id IN (`+placeholders+`)
ORDER BY r.role_key ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		var role model.AdminRoleBrief
		if err := rows.Scan(&userID, &role.ID, &role.RoleKey, &role.RoleName); err != nil {
			return nil, err
		}
		result[userID] = append(result[userID], role)
	}
	return result, nil
}

func (h *AuthHandler) loadUserLastLoginMap(userIDs []string) (map[string]string, error) {
	result := map[string]string{}
	if len(userIDs) == 0 {
		return result, nil
	}
	placeholders := sqlPlaceholders(len(userIDs))
	args := make([]interface{}, 0, len(userIDs))
	for _, userID := range userIDs {
		args = append(args, userID)
	}
	rows, err := h.db.Query(`
SELECT user_id, MAX(created_at) AS last_login_at
FROM auth_login_logs
WHERE action = 'LOGIN' AND status = 'SUCCESS' AND user_id IN (`+placeholders+`)
GROUP BY user_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		var lastLogin time.Time
		if err := rows.Scan(&userID, &lastLogin); err != nil {
			return nil, err
		}
		result[userID] = lastLogin.Format(time.RFC3339)
	}
	return result, nil
}

func (h *AuthHandler) assertPermissionsExistTx(tx *sql.Tx, permissionCodes []string) error {
	if len(permissionCodes) == 0 {
		return errors.New("permission_codes required")
	}
	placeholders := sqlPlaceholders(len(permissionCodes))
	args := make([]interface{}, 0, len(permissionCodes))
	for _, code := range permissionCodes {
		args = append(args, code)
	}
	var count int
	err := tx.QueryRow(`
SELECT COUNT(*)
FROM rbac_permissions
WHERE code IN (`+placeholders+`) AND status = 'ACTIVE'`, args...).Scan(&count)
	if err != nil {
		return err
	}
	if count != len(permissionCodes) {
		return errors.New("invalid or inactive permission code")
	}
	return nil
}

func (h *AuthHandler) assertRolesExistTx(tx *sql.Tx, roleIDs []string) error {
	if len(roleIDs) == 0 {
		return errors.New("role_ids required")
	}
	placeholders := sqlPlaceholders(len(roleIDs))
	args := make([]interface{}, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		args = append(args, roleID)
	}
	var count int
	err := tx.QueryRow(`
SELECT COUNT(*)
FROM rbac_roles
WHERE id IN (`+placeholders+`) AND status = 'ACTIVE'`, args...).Scan(&count)
	if err != nil {
		return err
	}
	if count != len(roleIDs) {
		return errors.New("invalid or inactive role id")
	}
	return nil
}

func (h *AuthHandler) replaceRolePermissionsTx(tx *sql.Tx, roleID string, permissionCodes []string) error {
	if _, err := tx.Exec("DELETE FROM rbac_role_permissions WHERE role_id = ?", roleID); err != nil {
		return err
	}
	if len(permissionCodes) == 0 {
		return nil
	}
	now := time.Now()
	stmt, err := tx.Prepare(`
INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, code := range permissionCodes {
		if _, err := stmt.Exec(roleID, code, now); err != nil {
			return err
		}
	}
	return nil
}

func (h *AuthHandler) replaceUserRolesTx(tx *sql.Tx, userID string, roleIDs []string) error {
	if _, err := tx.Exec("DELETE FROM rbac_user_roles WHERE user_id = ?", userID); err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	now := time.Now()
	stmt, err := tx.Prepare(`
INSERT INTO rbac_user_roles (user_id, role_id, created_at)
VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, roleID := range roleIDs {
		if _, err := stmt.Exec(userID, roleID, now); err != nil {
			return err
		}
	}
	return nil
}

func normalizeCodeList(values []string, upper bool) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	items := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if upper {
			normalized = strings.ToUpper(normalized)
		}
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func sqlPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	placeholders := make([]string, 0, n)
	for i := 0; i < n; i++ {
		placeholders = append(placeholders, "?")
	}
	return strings.Join(placeholders, ",")
}

func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
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
