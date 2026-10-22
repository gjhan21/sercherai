package repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) ListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM membership_products"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, name, price, status, member_level, duration_days
FROM membership_products` + filter + `
ORDER BY id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.MembershipProduct, 0)
	for rows.Next() {
		var item model.MembershipProduct
		var memberLevel sql.NullString
		var durationDays sql.NullInt64
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Status, &memberLevel, &durationDays); err != nil {
			return nil, 0, err
		}
		if memberLevel.Valid {
			item.MemberLevel = memberLevel.String
		}
		if durationDays.Valid {
			item.DurationDays = int(durationDays.Int64)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) CreateMembershipOrder(userID string, productID string, payChannel string) (model.MembershipOrderAdmin, error) {
	var price float64
	var status string
	err := r.db.QueryRow("SELECT price, status FROM membership_products WHERE id = ?", productID).Scan(&price, &status)
	if err != nil {
		return model.MembershipOrderAdmin{}, err
	}
	if strings.ToUpper(status) != "ACTIVE" {
		return model.MembershipOrderAdmin{}, errors.New("membership product not active")
	}
	id := newID("mo")
	orderNo := id
	now := time.Now()
	_, err = r.db.Exec(`
INSERT INTO membership_orders (id, order_no, user_id, product_id, amount, pay_channel, status, paid_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, 'PENDING', NULL, ?, ?)`,
		id, orderNo, userID, productID, price, strings.ToUpper(payChannel), now, now,
	)
	if err != nil {
		return model.MembershipOrderAdmin{}, err
	}
	return model.MembershipOrderAdmin{
		ID:         id,
		OrderNo:    orderNo,
		UserID:     userID,
		ProductID:  productID,
		Amount:     price,
		PayChannel: strings.ToUpper(payChannel),
		Status:     "PENDING",
		CreatedAt:  now.Format(time.RFC3339),
	}, nil
}

func (r *MySQLGrowthRepo) ListMembershipOrders(userID string, status string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{userID}
	filter := " WHERE user_id = ?"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM membership_orders"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, order_no, user_id, product_id, amount, pay_channel, status, paid_at, created_at
FROM membership_orders` + filter + `
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.MembershipOrderAdmin, 0)
	for rows.Next() {
		var item model.MembershipOrderAdmin
		var paidAt, createdAt sql.NullTime
		var orderNo, payChannel sql.NullString
		if err := rows.Scan(&item.ID, &orderNo, &item.UserID, &item.ProductID, &item.Amount, &payChannel, &item.Status, &paidAt, &createdAt); err != nil {
			return nil, 0, err
		}
		if orderNo.Valid {
			item.OrderNo = orderNo.String
		}
		if payChannel.Valid {
			item.PayChannel = payChannel.String
		}
		if paidAt.Valid {
			item.PaidAt = paidAt.Time.Format(time.RFC3339)
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) TrackExperimentEvent(item model.ExperimentEvent) error {
	metadataJSON, err := marshalExperimentMetadata(item.Metadata)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
INSERT INTO experiment_events (
	id, experiment_key, variant_key, event_type, page_key, target_key, user_stage,
	anonymous_id, session_id, pathname, referrer, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("exp"),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ExperimentKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.VariantKey)), 32),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.EventType)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.PageKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.TargetKey)), 64),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.UserStage)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.AnonymousID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.SessionID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Pathname)), 255),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Referrer)), 255),
		nullableString(metadataJSON),
		time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) BindMembershipOrderExperiment(orderNo string, item model.ExperimentOrderAttribution) error {
	metadataJSON, err := marshalExperimentMetadata(item.Metadata)
	if err != nil {
		return err
	}
	conversionType := strings.ToUpper(strings.TrimSpace(item.ConversionType))
	if conversionType == "" {
		conversionType = deriveExperimentConversionType(item.UserStage)
	}
	_, err = r.db.Exec(`
INSERT INTO experiment_order_attributions (
	order_no, experiment_key, variant_key, page_key, target_key, user_stage,
	anonymous_id, session_id, pathname, referrer, conversion_type, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	experiment_key = VALUES(experiment_key),
	variant_key = VALUES(variant_key),
	page_key = VALUES(page_key),
	target_key = VALUES(target_key),
	user_stage = VALUES(user_stage),
	anonymous_id = VALUES(anonymous_id),
	session_id = VALUES(session_id),
	pathname = VALUES(pathname),
	referrer = VALUES(referrer),
	conversion_type = VALUES(conversion_type),
	metadata_json = VALUES(metadata_json)`,
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(orderNo)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ExperimentKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.VariantKey)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.PageKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.TargetKey)), 64),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.UserStage)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.AnonymousID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.SessionID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Pathname)), 255),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Referrer)), 255),
		truncateByRunes(conversionType, 32),
		nullableString(metadataJSON),
		time.Now(),
	)
	return err
}

func marshalExperimentMetadata(metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return "", nil
	}
	payloadBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return truncateByRunes(string(payloadBytes), 4096), nil
}

func deriveExperimentConversionType(userStage string) string {
	stage := strings.ToUpper(strings.TrimSpace(userStage))
	if stage == "VIP" || stage == "EXPIRED" {
		return "RENEWAL_SUCCESS"
	}
	return "PAYMENT_SUCCESS"
}

func insertExperimentEventTx(tx *sql.Tx, item model.ExperimentEvent, occurredAt time.Time) error {
	metadataJSON, err := marshalExperimentMetadata(item.Metadata)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
INSERT INTO experiment_events (
	id, experiment_key, variant_key, event_type, page_key, target_key, user_stage,
	anonymous_id, session_id, pathname, referrer, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("exp"),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ExperimentKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.VariantKey)), 32),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.EventType)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.PageKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.TargetKey)), 64),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.UserStage)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.AnonymousID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.SessionID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Pathname)), 255),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Referrer)), 255),
		nullableString(metadataJSON),
		occurredAt,
	)
	return err
}

func createExperimentSuccessEventTx(tx *sql.Tx, orderNo string, occurredAt time.Time) error {
	var attribution model.ExperimentOrderAttribution
	var targetKey, userStage, anonymousID, sessionID, pathname, referrer, conversionType sql.NullString
	var metadataJSON sql.NullString
	err := tx.QueryRow(`
SELECT experiment_key, variant_key, page_key, target_key, user_stage, anonymous_id, session_id, pathname, referrer, conversion_type, metadata_json
FROM experiment_order_attributions
WHERE order_no = ?
LIMIT 1`, orderNo).Scan(
		&attribution.ExperimentKey,
		&attribution.VariantKey,
		&attribution.PageKey,
		&targetKey,
		&userStage,
		&anonymousID,
		&sessionID,
		&pathname,
		&referrer,
		&conversionType,
		&metadataJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if targetKey.Valid {
		attribution.TargetKey = targetKey.String
	}
	if userStage.Valid {
		attribution.UserStage = userStage.String
	}
	if anonymousID.Valid {
		attribution.AnonymousID = anonymousID.String
	}
	if sessionID.Valid {
		attribution.SessionID = sessionID.String
	}
	if pathname.Valid {
		attribution.Pathname = pathname.String
	}
	if referrer.Valid {
		attribution.Referrer = referrer.String
	}
	eventMetadata := map[string]interface{}{
		"order_no": orderNo,
	}
	if metadataJSON.Valid && strings.TrimSpace(metadataJSON.String) != "" {
		var extra map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON.String), &extra); err == nil {
			for key, value := range extra {
				eventMetadata[key] = value
			}
		}
	}
	eventType := deriveExperimentConversionType(attribution.UserStage)
	if conversionType.Valid && strings.TrimSpace(conversionType.String) != "" {
		eventType = strings.ToUpper(strings.TrimSpace(conversionType.String))
	}
	sourceExperimentItems := []interface{}{}
	if rawItems, ok := eventMetadata["source_experiments"].([]interface{}); ok {
		sourceExperimentItems = rawItems
	}
	delete(eventMetadata, "source_experiments")
	primaryMetadata := map[string]interface{}{}
	for key, value := range eventMetadata {
		primaryMetadata[key] = value
	}
	primaryMetadata["attribution_role"] = "primary"
	if err := insertExperimentEventTx(tx, model.ExperimentEvent{
		ExperimentKey: attribution.ExperimentKey,
		VariantKey:    attribution.VariantKey,
		EventType:     eventType,
		PageKey:       attribution.PageKey,
		TargetKey:     attribution.TargetKey,
		UserStage:     attribution.UserStage,
		AnonymousID:   attribution.AnonymousID,
		SessionID:     attribution.SessionID,
		Pathname:      attribution.Pathname,
		Referrer:      attribution.Referrer,
		Metadata:      primaryMetadata,
	}, occurredAt); err != nil {
		return err
	}

	for _, rawItem := range sourceExperimentItems {
		itemMap, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		readSourceField := func(key string) string {
			value, exists := itemMap[key]
			if !exists || value == nil {
				return ""
			}
			return strings.TrimSpace(fmt.Sprint(value))
		}
		sourceExperimentKey := readSourceField("experiment_key")
		sourceVariantKey := readSourceField("variant_key")
		sourcePageKey := readSourceField("page_key")
		if sourceExperimentKey == "" || sourceVariantKey == "" || sourcePageKey == "" {
			continue
		}
		sourceMetadata := map[string]interface{}{}
		for key, value := range eventMetadata {
			sourceMetadata[key] = value
		}
		sourceMetadata["attribution_role"] = "source"
		sourceMetadata["source_experiment_key"] = sourceExperimentKey
		if err := insertExperimentEventTx(tx, model.ExperimentEvent{
			ExperimentKey: sourceExperimentKey,
			VariantKey:    sourceVariantKey,
			EventType:     eventType,
			PageKey:       sourcePageKey,
			TargetKey:     readSourceField("target_key"),
			UserStage:     readSourceField("user_stage"),
			AnonymousID:   readSourceField("anonymous_id"),
			SessionID:     readSourceField("session_id"),
			Pathname:      readSourceField("pathname"),
			Referrer:      readSourceField("referrer"),
			Metadata:      sourceMetadata,
		}, occurredAt); err != nil {
			return err
		}
	}
	return nil
}
