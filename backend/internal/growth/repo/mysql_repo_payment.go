package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (r *MySQLGrowthRepo) HandlePaymentCallback(channel string, orderNo string, channelTxnNo string, idempotencyKey string, sign string, signVerified bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	now := time.Now()
	verifiedVal := 0
	if signVerified {
		verifiedVal = 1
	}
	_, err = tx.Exec(`
INSERT INTO payment_callback_logs
(id, pay_channel, order_no, channel_txn_no, sign_verified, idempotency_key, callback_status, created_at)
VALUES (?, ?, ?, ?, ?, ?, 'RECEIVED', ?)`,
		newID("cb"), strings.ToUpper(channel), orderNo, channelTxnNo, verifiedVal, idempotencyKey, now,
	)
	if err != nil {
		_ = tx.Rollback()
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return errors.New("duplicate callback")
		}
		return err
	}

	var orderID, userID, productID, status string
	orderErr := tx.QueryRow(`
SELECT id, user_id, product_id, status
FROM membership_orders
WHERE order_no = ? OR id = ?
LIMIT 1`, orderNo, orderNo).Scan(&orderID, &userID, &productID, &status)
	if orderErr == nil {
		if strings.ToUpper(status) != "PAID" {
			if _, err := tx.Exec(`
UPDATE membership_orders
SET status = 'PAID', paid_at = ?, pay_channel = ?, updated_at = ?
WHERE id = ?`, now, strings.ToUpper(channel), now, orderID); err != nil {
				_ = tx.Rollback()
				return err
			}
			var memberLevel sql.NullString
			var durationDays sql.NullInt64
			if err := tx.QueryRow(
				"SELECT member_level, duration_days FROM membership_products WHERE id = ?",
				productID,
			).Scan(&memberLevel, &durationDays); err == nil {
				level := strings.ToUpper(strings.TrimSpace(memberLevel.String))
				days := int(durationDays.Int64)
				if level != "" && days <= 0 {
					days = 30
				}
				if level != "" {
					var currentLevel string
					var currentExpireAt sql.NullTime
					if err := tx.QueryRow(
						"SELECT member_level, vip_expire_at FROM users WHERE id = ? FOR UPDATE",
						userID,
					).Scan(&currentLevel, &currentExpireAt); err != nil {
						_ = tx.Rollback()
						return err
					}
					baseTime := now
					if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(currentLevel)), "VIP") &&
						currentExpireAt.Valid &&
						currentExpireAt.Time.After(now) {
						baseTime = currentExpireAt.Time
					}
					newExpireAt := baseTime
					if days > 0 {
						newExpireAt = baseTime.AddDate(0, 0, days)
					}
					startedAt := now
					if _, err := tx.Exec(`
UPDATE users
SET member_level = ?,
    vip_started_at = ?,
    vip_expire_at = ?,
    vip_remind_3d_at = NULL,
    vip_remind_1d_at = NULL,
    updated_at = ?
WHERE id = ?`,
						level,
						startedAt,
						newExpireAt,
						now,
						userID,
					); err != nil {
						_ = tx.Rollback()
						return err
					}
					title := "VIP会员开通成功"
					content := fmt.Sprintf(
						"订单%s支付成功，会员等级已更新为%s，到期时间：%s。",
						strings.TrimSpace(orderNo),
						level,
						newExpireAt.Format("2006-01-02 15:04:05"),
					)
					if _, err := tx.Exec(`
INSERT INTO messages (id, user_id, title, content, type, read_status, created_at)
VALUES (?, ?, ?, ?, 'SYSTEM', 'UNREAD', ?)`,
						newID("msg"),
						userID,
						title,
						content,
						now,
					); err != nil {
						_ = tx.Rollback()
						return err
					}
				}
			}
			if err := createExperimentSuccessEventTx(tx, strings.TrimSpace(orderNo), now); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		return tx.Commit()
	}
	if !errors.Is(orderErr, sql.ErrNoRows) {
		_ = tx.Rollback()
		return orderErr
	}

	var rechargeStatus string
	rechargeErr := tx.QueryRow("SELECT status FROM recharge_records WHERE order_no = ?", orderNo).Scan(&rechargeStatus)
	if rechargeErr == nil {
		if strings.ToUpper(rechargeStatus) != "PAID" {
			if _, err := tx.Exec("UPDATE recharge_records SET status = 'PAID', paid_at = ? WHERE order_no = ?", now, orderNo); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		return tx.Commit()
	}
	if !errors.Is(rechargeErr, sql.ErrNoRows) {
		_ = tx.Rollback()
		return rechargeErr
	}

	_ = tx.Commit()
	return errors.New("order not found")
}
