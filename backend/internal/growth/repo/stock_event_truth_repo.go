package repo

import (
	"database/sql"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminUpsertStockEventCluster(cluster model.StockEventCluster) (model.StockEventCluster, error) {
	cluster = normalizeStockEventCluster(cluster)
	now := time.Now()
	createdAt := parseRFC3339OrNow(cluster.CreatedAt, now)
	updatedAt := now
	if parsed, ok := parseRFC3339(cluster.UpdatedAt); ok {
		updatedAt = parsed
	}

	tx, err := r.db.Begin()
	if err != nil {
		return model.StockEventCluster{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`
INSERT INTO stock_event_clusters (
  id, cluster_key, event_type, title, summary, source, primary_symbol, sector_label,
  topic_label, cluster_status, review_status, news_count, confidence_score,
  metadata_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  cluster_key = VALUES(cluster_key),
  event_type = VALUES(event_type),
  title = VALUES(title),
  summary = VALUES(summary),
  source = VALUES(source),
  primary_symbol = VALUES(primary_symbol),
  sector_label = VALUES(sector_label),
  topic_label = VALUES(topic_label),
  cluster_status = VALUES(cluster_status),
  review_status = VALUES(review_status),
  news_count = VALUES(news_count),
  confidence_score = VALUES(confidence_score),
  metadata_json = VALUES(metadata_json),
  updated_at = VALUES(updated_at)`,
		cluster.ID,
		cluster.ClusterKey,
		cluster.EventType,
		cluster.Title,
		cluster.Summary,
		cluster.Source,
		cluster.PrimarySymbol,
		cluster.SectorLabel,
		cluster.TopicLabel,
		cluster.Status,
		cluster.ReviewStatus,
		cluster.NewsCount,
		cluster.Confidence,
		marshalJSONSilently(cluster.Metadata),
		createdAt,
		updatedAt,
	); err != nil {
		return model.StockEventCluster{}, err
	}

	if err = replaceStockEventItems(tx, cluster.ID, cluster.Items, now); err != nil {
		return model.StockEventCluster{}, err
	}
	if err = replaceStockEventEntities(tx, cluster.ID, cluster.Entities, now); err != nil {
		return model.StockEventCluster{}, err
	}
	if err = replaceStockEventEdges(tx, cluster.ID, cluster.Edges, now); err != nil {
		return model.StockEventCluster{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.StockEventCluster{}, err
	}

	cluster.CreatedAt = createdAt.Format(time.RFC3339)
	cluster.UpdatedAt = updatedAt.Format(time.RFC3339)
	return cluster, nil
}

func (r *MySQLGrowthRepo) AdminGetStockEventCluster(id string) (model.StockEventCluster, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.StockEventCluster{}, sql.ErrNoRows
	}

	var cluster model.StockEventCluster
	var metadataJSON sql.NullString
	var publishedAt, createdAt, updatedAt sql.NullTime
	var latestReviewID, latestReviewer, latestReviewNote, latestReviewStatus, latestReviewMetadata sql.NullString
	var latestReviewedAt, latestReviewCreatedAt, latestReviewUpdatedAt sql.NullTime

	err := r.db.QueryRow(`
SELECT
  c.id,
  c.cluster_key,
  c.event_type,
  c.title,
  COALESCE(c.summary, ''),
  COALESCE(c.source, ''),
  COALESCE(c.primary_symbol, ''),
  COALESCE(c.sector_label, ''),
  COALESCE(c.topic_label, ''),
  c.cluster_status,
  c.review_status,
  c.news_count,
  c.confidence_score,
  COALESCE(CAST(c.metadata_json AS CHAR), ''),
  c.published_at,
  c.created_at,
  c.updated_at,
  lr.id,
  lr.reviewer,
  lr.review_note,
  lr.review_status,
  COALESCE(CAST(lr.metadata_json AS CHAR), ''),
  lr.created_at,
  lr.created_at,
  lr.updated_at
FROM stock_event_clusters c
LEFT JOIN (
  SELECT r1.*
  FROM stock_event_reviews r1
  INNER JOIN (
    SELECT cluster_id, MAX(created_at) AS latest_created_at
    FROM stock_event_reviews
    GROUP BY cluster_id
  ) latest ON latest.cluster_id = r1.cluster_id AND latest.latest_created_at = r1.created_at
) lr ON lr.cluster_id = c.id
WHERE c.id = ?`, id).Scan(
		&cluster.ID,
		&cluster.ClusterKey,
		&cluster.EventType,
		&cluster.Title,
		&cluster.Summary,
		&cluster.Source,
		&cluster.PrimarySymbol,
		&cluster.SectorLabel,
		&cluster.TopicLabel,
		&cluster.Status,
		&cluster.ReviewStatus,
		&cluster.NewsCount,
		&cluster.Confidence,
		&metadataJSON,
		&publishedAt,
		&createdAt,
		&updatedAt,
		&latestReviewID,
		&latestReviewer,
		&latestReviewNote,
		&latestReviewStatus,
		&latestReviewMetadata,
		&latestReviewedAt,
		&latestReviewCreatedAt,
		&latestReviewUpdatedAt,
	)
	if err != nil {
		return model.StockEventCluster{}, err
	}

	cluster.Metadata = parseJSONMap(metadataJSON.String)
	cluster.PublishedAt = formatNullTime(publishedAt)
	cluster.CreatedAt = formatNullTime(createdAt)
	cluster.UpdatedAt = formatNullTime(updatedAt)
	if latestReviewID.Valid {
		cluster.LatestReview = &model.StockEventReview{
			ID:             latestReviewID.String,
			ClusterID:      cluster.ID,
			ReviewStatus:   latestReviewStatus.String,
			Reviewer:       latestReviewer.String,
			ReviewNote:     latestReviewNote.String,
			ReviewMetadata: parseJSONMap(latestReviewMetadata.String),
			ReviewedAt:     formatNullTime(latestReviewedAt),
			CreatedAt:      formatNullTime(latestReviewCreatedAt),
			UpdatedAt:      formatNullTime(latestReviewUpdatedAt),
		}
	}

	if cluster.Items, err = r.listStockEventItems(cluster.ID); err != nil {
		return model.StockEventCluster{}, err
	}
	if cluster.Entities, err = r.listStockEventEntities(cluster.ID); err != nil {
		return model.StockEventCluster{}, err
	}
	if cluster.Edges, err = r.listStockEventEdges(cluster.ID); err != nil {
		return model.StockEventCluster{}, err
	}
	return cluster, nil
}

func (r *MySQLGrowthRepo) AdminCreateStockEventReview(review model.StockEventReview) (model.StockEventReview, error) {
	review = normalizeStockEventReview(review)
	now := time.Now()

	tx, err := r.db.Begin()
	if err != nil {
		return model.StockEventReview{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.Exec(
		"UPDATE stock_event_clusters SET cluster_status = ?, review_status = ?, updated_at = NOW() WHERE id = ?",
		stockEventClusterStatusFromReview(review.ReviewStatus),
		review.ReviewStatus,
		review.ClusterID,
	)
	if err != nil {
		return model.StockEventReview{}, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.StockEventReview{}, err
	}
	if rowsAffected == 0 {
		_ = tx.Rollback()
		return model.StockEventReview{}, sql.ErrNoRows
	}

	if _, err = tx.Exec(`
INSERT INTO stock_event_reviews (
  id, cluster_id, review_status, reviewer, review_note, metadata_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		review.ID,
		review.ClusterID,
		review.ReviewStatus,
		review.Reviewer,
		review.ReviewNote,
		marshalJSONSilently(review.ReviewMetadata),
		now,
		now,
	); err != nil {
		return model.StockEventReview{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.StockEventReview{}, err
	}

	review.ReviewedAt = now.Format(time.RFC3339)
	review.CreatedAt = now.Format(time.RFC3339)
	review.UpdatedAt = now.Format(time.RFC3339)
	return review, nil
}

func (r *MySQLGrowthRepo) AdminListStockEventClusters(query model.StockEventQuery) ([]model.StockEventCluster, int, error) {
	query = normalizeStockEventQuery(query)
	args := make([]any, 0, 8)
	filters := make([]string, 0, 4)
	if query.ReviewStatus != "" {
		filters = append(filters, "c.review_status = ?")
		args = append(args, query.ReviewStatus)
	}
	if query.EventType != "" {
		filters = append(filters, "c.event_type = ?")
		args = append(args, query.EventType)
	}
	if query.ReviewPriority != "" {
		filters = append(filters, "JSON_UNQUOTE(JSON_EXTRACT(c.metadata_json, '$.review_priority')) = ?")
		args = append(args, query.ReviewPriority)
	}
	if query.Symbol != "" {
		filters = append(filters, `(c.primary_symbol = ? OR EXISTS (
SELECT 1 FROM stock_event_entities se
WHERE se.cluster_id = c.id AND se.symbol = ?
))`)
		args = append(args, query.Symbol, query.Symbol)
	}
	if query.Sector != "" {
		filters = append(filters, "c.sector_label = ?")
		args = append(args, query.Sector)
	}
	if query.Topic != "" {
		filters = append(filters, "c.topic_label = ?")
		args = append(args, query.Topic)
	}

	filterSQL := ""
	if len(filters) > 0 {
		filterSQL = " WHERE " + strings.Join(filters, " AND ")
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(DISTINCT c.id) FROM stock_event_clusters c"+filterSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
  c.id,
  c.cluster_key,
  c.event_type,
  c.title,
  COALESCE(c.summary, ''),
  COALESCE(c.source, ''),
  COALESCE(c.primary_symbol, ''),
  COALESCE(c.sector_label, ''),
  COALESCE(c.topic_label, ''),
  c.cluster_status,
  c.review_status,
  c.news_count,
  c.confidence_score,
  COALESCE(CAST(c.metadata_json AS CHAR), ''),
  c.published_at,
  c.created_at,
  c.updated_at,
  COALESCE(lr.id, ''),
  COALESCE(lr.reviewer, ''),
  COALESCE(lr.review_note, ''),
  COALESCE(lr.review_status, ''),
  COALESCE(CAST(lr.metadata_json AS CHAR), ''),
  lr.created_at,
  lr.created_at,
  lr.updated_at
FROM stock_event_clusters c
LEFT JOIN (
  SELECT r1.*
  FROM stock_event_reviews r1
  INNER JOIN (
    SELECT cluster_id, MAX(created_at) AS latest_created_at
    FROM stock_event_reviews
    GROUP BY cluster_id
  ) latest ON latest.cluster_id = r1.cluster_id AND latest.latest_created_at = r1.created_at
) lr ON lr.cluster_id = c.id`+filterSQL+`
ORDER BY c.updated_at DESC, c.id ASC
LIMIT ? OFFSET ?`, append(args, query.PageSize, (query.Page-1)*query.PageSize)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StockEventCluster, 0)
	for rows.Next() {
		item, err := scanStockEventCluster(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) listStockEventItems(clusterID string) ([]model.StockEventItem, error) {
	rows, err := r.db.Query("SELECT id, cluster_id, source_key, source_item_id, title, summary, primary_symbol, symbols_json, metadata_json, published_at, created_at, updated_at FROM stock_event_items WHERE cluster_id = ? ORDER BY published_at DESC, id ASC", clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StockEventItem, 0)
	for rows.Next() {
		var item model.StockEventItem
		var symbolsJSON, metadataJSON sql.NullString
		var publishedAt, createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.ClusterID,
			&item.SourceKey,
			&item.SourceItemID,
			&item.Title,
			&item.Summary,
			&item.PrimarySymbol,
			&symbolsJSON,
			&metadataJSON,
			&publishedAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.Symbols = parseJSONStringList(symbolsJSON.String)
		item.Metadata = parseJSONMap(metadataJSON.String)
		item.PublishedAt = formatNullTime(publishedAt)
		item.CreatedAt = formatNullTime(createdAt)
		item.UpdatedAt = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) listStockEventEntities(clusterID string) ([]model.StockEventEntity, error) {
	rows, err := r.db.Query("SELECT id, cluster_id, entity_type, entity_key, label, symbol, sector_label, topic_label, confidence_score, metadata_json, created_at, updated_at FROM stock_event_entities WHERE cluster_id = ? ORDER BY confidence_score DESC, id ASC", clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StockEventEntity, 0)
	for rows.Next() {
		var item model.StockEventEntity
		var metadataJSON sql.NullString
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.ClusterID,
			&item.EntityType,
			&item.EntityKey,
			&item.Label,
			&item.Symbol,
			&item.SectorLabel,
			&item.TopicLabel,
			&item.Confidence,
			&metadataJSON,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.Metadata = parseJSONMap(metadataJSON.String)
		item.CreatedAt = formatNullTime(createdAt)
		item.UpdatedAt = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) listStockEventEdges(clusterID string) ([]model.StockEventEdge, error) {
	rows, err := r.db.Query("SELECT id, cluster_id, subject_entity_id, object_entity_id, relation_type, relation_weight, metadata_json, created_at, updated_at FROM stock_event_edges WHERE cluster_id = ? ORDER BY relation_weight DESC, id ASC", clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StockEventEdge, 0)
	for rows.Next() {
		var item model.StockEventEdge
		var metadataJSON sql.NullString
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.ClusterID,
			&item.SubjectEntityID,
			&item.ObjectEntityID,
			&item.RelationType,
			&item.Weight,
			&metadataJSON,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.Metadata = parseJSONMap(metadataJSON.String)
		item.CreatedAt = formatNullTime(createdAt)
		item.UpdatedAt = formatNullTime(updatedAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func replaceStockEventItems(tx *sql.Tx, clusterID string, items []model.StockEventItem, now time.Time) error {
	if _, err := tx.Exec("DELETE FROM stock_event_items WHERE cluster_id = ?", clusterID); err != nil {
		return err
	}
	for index := range items {
		item := normalizeStockEventItem(items[index], clusterID)
		if _, err := tx.Exec(`
INSERT INTO stock_event_items (
  id, cluster_id, source_key, source_item_id, title, summary, primary_symbol,
  symbols_json, metadata_json, published_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID,
			clusterID,
			item.SourceKey,
			item.SourceItemID,
			item.Title,
			item.Summary,
			item.PrimarySymbol,
			marshalJSONSilently(item.Symbols),
			marshalJSONSilently(item.Metadata),
			nullableRFC3339Time(item.PublishedAt),
			now,
			now,
		); err != nil {
			return err
		}
	}
	return nil
}

func replaceStockEventEntities(tx *sql.Tx, clusterID string, entities []model.StockEventEntity, now time.Time) error {
	if _, err := tx.Exec("DELETE FROM stock_event_entities WHERE cluster_id = ?", clusterID); err != nil {
		return err
	}
	for index := range entities {
		entity := normalizeStockEventEntity(entities[index], clusterID)
		if _, err := tx.Exec(`
INSERT INTO stock_event_entities (
  id, cluster_id, entity_type, entity_key, label, symbol, sector_label,
  topic_label, confidence_score, metadata_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			entity.ID,
			clusterID,
			entity.EntityType,
			entity.EntityKey,
			entity.Label,
			entity.Symbol,
			entity.SectorLabel,
			entity.TopicLabel,
			entity.Confidence,
			marshalJSONSilently(entity.Metadata),
			now,
			now,
		); err != nil {
			return err
		}
	}
	return nil
}

func replaceStockEventEdges(tx *sql.Tx, clusterID string, edges []model.StockEventEdge, now time.Time) error {
	if _, err := tx.Exec("DELETE FROM stock_event_edges WHERE cluster_id = ?", clusterID); err != nil {
		return err
	}
	for index := range edges {
		edge := normalizeStockEventEdge(edges[index], clusterID)
		if _, err := tx.Exec(`
INSERT INTO stock_event_edges (
  id, cluster_id, subject_entity_id, object_entity_id, relation_type,
  relation_weight, metadata_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			edge.ID,
			clusterID,
			edge.SubjectEntityID,
			edge.ObjectEntityID,
			edge.RelationType,
			edge.Weight,
			marshalJSONSilently(edge.Metadata),
			now,
			now,
		); err != nil {
			return err
		}
	}
	return nil
}

func scanStockEventCluster(scanner interface{ Scan(dest ...any) error }) (model.StockEventCluster, error) {
	var item model.StockEventCluster
	var metadataJSON, latestReviewID, latestReviewer, latestReviewNote, latestReviewStatus, latestReviewMetadata sql.NullString
	var publishedAt, createdAt, updatedAt, latestReviewedAt, latestReviewCreatedAt, latestReviewUpdatedAt sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&item.ClusterKey,
		&item.EventType,
		&item.Title,
		&item.Summary,
		&item.Source,
		&item.PrimarySymbol,
		&item.SectorLabel,
		&item.TopicLabel,
		&item.Status,
		&item.ReviewStatus,
		&item.NewsCount,
		&item.Confidence,
		&metadataJSON,
		&publishedAt,
		&createdAt,
		&updatedAt,
		&latestReviewID,
		&latestReviewer,
		&latestReviewNote,
		&latestReviewStatus,
		&latestReviewMetadata,
		&latestReviewedAt,
		&latestReviewCreatedAt,
		&latestReviewUpdatedAt,
	); err != nil {
		return model.StockEventCluster{}, err
	}
	item.Metadata = parseJSONMap(metadataJSON.String)
	item.PublishedAt = formatNullTime(publishedAt)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	if latestReviewID.Valid && strings.TrimSpace(latestReviewID.String) != "" {
		item.LatestReview = &model.StockEventReview{
			ID:             latestReviewID.String,
			ClusterID:      item.ID,
			ReviewStatus:   latestReviewStatus.String,
			Reviewer:       latestReviewer.String,
			ReviewNote:     latestReviewNote.String,
			ReviewMetadata: parseJSONMap(latestReviewMetadata.String),
			ReviewedAt:     formatNullTime(latestReviewedAt),
			CreatedAt:      formatNullTime(latestReviewCreatedAt),
			UpdatedAt:      formatNullTime(latestReviewUpdatedAt),
		}
	}
	return item, nil
}

func normalizeStockEventCluster(cluster model.StockEventCluster) model.StockEventCluster {
	cluster.ID = strings.TrimSpace(cluster.ID)
	if cluster.ID == "" {
		cluster.ID = newID("sec")
	}
	cluster.ClusterKey = strings.TrimSpace(cluster.ClusterKey)
	if cluster.ClusterKey == "" {
		cluster.ClusterKey = cluster.ID
	}
	cluster.EventType = strings.ToUpper(strings.TrimSpace(cluster.EventType))
	cluster.Title = strings.TrimSpace(cluster.Title)
	cluster.Summary = strings.TrimSpace(cluster.Summary)
	cluster.Source = strings.TrimSpace(cluster.Source)
	cluster.PrimarySymbol = strings.ToUpper(strings.TrimSpace(cluster.PrimarySymbol))
	cluster.SectorLabel = strings.TrimSpace(cluster.SectorLabel)
	cluster.TopicLabel = strings.TrimSpace(cluster.TopicLabel)
	cluster.Status = normalizeStockEventClusterStatus(cluster.Status)
	cluster.ReviewStatus = normalizeStockEventReviewStatus(cluster.ReviewStatus)
	if cluster.Metadata == nil {
		cluster.Metadata = map[string]any{}
	}
	return cluster
}

func normalizeStockEventItem(item model.StockEventItem, clusterID string) model.StockEventItem {
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		item.ID = newID("sei")
	}
	item.ClusterID = clusterID
	item.SourceKey = strings.TrimSpace(item.SourceKey)
	item.SourceItemID = strings.TrimSpace(item.SourceItemID)
	item.Title = strings.TrimSpace(item.Title)
	item.Summary = strings.TrimSpace(item.Summary)
	item.PrimarySymbol = strings.ToUpper(strings.TrimSpace(item.PrimarySymbol))
	item.Symbols = normalizeUpperStringList(item.Symbols)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item
}

func normalizeStockEventEntity(entity model.StockEventEntity, clusterID string) model.StockEventEntity {
	entity.ID = strings.TrimSpace(entity.ID)
	if entity.ID == "" {
		entity.ID = newID("see")
	}
	entity.ClusterID = clusterID
	entity.EntityType = strings.ToUpper(strings.TrimSpace(entity.EntityType))
	entity.EntityKey = strings.TrimSpace(entity.EntityKey)
	entity.Label = strings.TrimSpace(entity.Label)
	entity.Symbol = strings.ToUpper(strings.TrimSpace(entity.Symbol))
	entity.SectorLabel = strings.TrimSpace(entity.SectorLabel)
	entity.TopicLabel = strings.TrimSpace(entity.TopicLabel)
	if entity.Metadata == nil {
		entity.Metadata = map[string]any{}
	}
	return entity
}

func normalizeStockEventEdge(edge model.StockEventEdge, clusterID string) model.StockEventEdge {
	edge.ID = strings.TrimSpace(edge.ID)
	if edge.ID == "" {
		edge.ID = newID("seg")
	}
	edge.ClusterID = clusterID
	edge.SubjectEntityID = strings.TrimSpace(edge.SubjectEntityID)
	edge.ObjectEntityID = strings.TrimSpace(edge.ObjectEntityID)
	edge.RelationType = strings.ToUpper(strings.TrimSpace(edge.RelationType))
	if edge.Metadata == nil {
		edge.Metadata = map[string]any{}
	}
	return edge
}

func normalizeStockEventReview(review model.StockEventReview) model.StockEventReview {
	review.ID = strings.TrimSpace(review.ID)
	if review.ID == "" {
		review.ID = newID("ser")
	}
	review.ClusterID = strings.TrimSpace(review.ClusterID)
	review.ReviewStatus = normalizeStockEventReviewStatus(review.ReviewStatus)
	review.Reviewer = strings.TrimSpace(review.Reviewer)
	review.ReviewNote = strings.TrimSpace(review.ReviewNote)
	if review.ReviewMetadata == nil {
		review.ReviewMetadata = map[string]any{}
	}
	return review
}

func normalizeStockEventQuery(query model.StockEventQuery) model.StockEventQuery {
	query.EventType = strings.ToUpper(strings.TrimSpace(query.EventType))
	query.Status = normalizeStockEventClusterStatus(query.Status)
	query.ReviewStatus = normalizeStockEventReviewStatus(query.ReviewStatus)
	query.ReviewPriority = strings.ToUpper(strings.TrimSpace(query.ReviewPriority))
	query.Symbol = strings.ToUpper(strings.TrimSpace(query.Symbol))
	query.Sector = strings.TrimSpace(query.Sector)
	query.Topic = strings.TrimSpace(query.Topic)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	return query
}

func normalizeStockEventClusterStatus(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "RAW":
		return "RAW"
	case "CLUSTERED":
		return "CLUSTERED"
	case "REVIEWED":
		return "REVIEWED"
	case "PUBLISHED":
		return "PUBLISHED"
	case "REJECTED":
		return "REJECTED"
	default:
		return "CLUSTERED"
	}
}

func normalizeStockEventReviewStatus(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "PENDING":
		return "PENDING"
	case "APPROVED":
		return "APPROVED"
	case "REJECTED":
		return "REJECTED"
	case "PUBLISHED":
		return "PUBLISHED"
	default:
		return "PENDING"
	}
}

func stockEventClusterStatusFromReview(reviewStatus string) string {
	switch normalizeStockEventReviewStatus(reviewStatus) {
	case "APPROVED":
		return "REVIEWED"
	case "PUBLISHED":
		return "PUBLISHED"
	case "REJECTED":
		return "REJECTED"
	default:
		return "CLUSTERED"
	}
}

func parseRFC3339(raw string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func parseRFC3339OrNow(raw string, fallback time.Time) time.Time {
	if parsed, ok := parseRFC3339(raw); ok {
		return parsed
	}
	return fallback
}

func nullableRFC3339Time(raw string) any {
	if parsed, ok := parseRFC3339(raw); ok {
		return parsed
	}
	return nil
}

func normalizeUpperStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		normalized := strings.ToUpper(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
