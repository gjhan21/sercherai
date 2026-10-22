package repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type docFastSyncRuntimeConfig struct {
	Enabled       bool
	BatchSize     int
	SourceBaseURL string
	AuthorID      string
	APIToken      string
}

type docFastSyncCategoryTarget struct {
	CategoryID string
	Visibility string
}

type docFastSourceArticle struct {
	ID          int64
	ChannelID   int64
	Title       string
	Image       string
	Description string
	PublishUnix int64
	CreateUnix  int64
	UpdateUnix  int64
	Content     string
	DownloadURL string
}

type docFastDownloadAttachment struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Password string `json:"password"`
}

type docFastAttachmentMeta struct {
	FileName string
	MimeType string
	FileSize int64
}

type phpSyncAttachment struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	FileName string `json:"filename"`
	MimeType string `json:"mimetype"`
	FileSize int64  `json:"filesize"`
}

type phpSyncArticle struct {
	ID          int64               `json:"id"`
	ChannelID   int64               `json:"channel_id"`
	Title       string              `json:"title"`
	Image       string              `json:"image"`
	Description string              `json:"description"`
	PublishUnix int64               `json:"publishtime"`
	CreateUnix  int64               `json:"createtime"`
	UpdateUnix  int64               `json:"updatetime"`
	Content     string              `json:"content"`
	DownloadURL string              `json:"downloadurl"`
	Attachments []phpSyncAttachment `json:"attachments"`
}

type phpSyncResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data []phpSyncArticle `json:"data"`
}

func fetchArticlesFromPHPCMS(baseURL, token string, updatedAt, sourceID int64, limit int) ([]phpSyncArticle, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	urlStr := fmt.Sprintf("%s/api/sync/articles?sync_token=%s&updated_at=%d&source_id=%d&limit=%d",
		strings.TrimRight(baseURL, "/"),
		url.QueryEscape(token),
		updatedAt,
		sourceID,
		limit,
	)

	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status error: %s", resp.Status)
	}

	var apiResp phpSyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode json response: %w", err)
	}

	if apiResp.Code != 1 {
		return nil, fmt.Errorf("php api error (code %d): %s", apiResp.Code, apiResp.Msg)
	}

	return apiResp.Data, nil
}

const docFastSyncCheckpointKey = "doc_fast_news_incremental"
const tushareNewsIncrementalJobName = "tushare_news_incremental"

const (
	tushareNewsCategoryBriefSlug        = "ts-news-brief"
	tushareNewsCategoryMajorSlug        = "ts-news-major"
	tushareNewsCategoryResearchSlug     = "ts-report-research"
	tushareNewsCategoryForecastSlug     = "ts-report-forecast"
	tushareNewsCategoryAnnouncementSlug = "ts-announcement"
)

const (
	tushareSyncTypeNewsBrief    = "NEWS_BRIEF"
	tushareSyncTypeNewsMajor    = "NEWS_MAJOR"
	tushareSyncTypeResearch     = "RESEARCH_REPORT"
	tushareSyncTypeForecast     = "REPORT_RC"
	tushareSyncTypeAnnouncement = "ANNOUNCEMENT"
)

type tushareNewsSyncRuntimeConfig struct {
	Enabled                  bool
	BatchSize                int
	TimeoutMS                int
	AuthorID                 string
	SyncTypes                []string
	BriefSources             []string
	MajorSources             []string
	BriefLookbackHours       int
	MajorLookbackHours       int
	ResearchLookbackDays     int
	ForecastLookbackDays     int
	AnnouncementLookbackDays int
	BriefVisibility          string
	MajorVisibility          string
	ResearchVisibility       string
	ForecastVisibility       string
	AnnouncementVisibility   string
}

type tushareNewsCategoryTarget struct {
	CategoryID string
	Visibility string
}

type tushareNewsCategoryTargets struct {
	Brief        tushareNewsCategoryTarget
	Major        tushareNewsCategoryTarget
	Research     tushareNewsCategoryTarget
	Forecast     tushareNewsCategoryTarget
	Announcement tushareNewsCategoryTarget
}

type normalizedTushareSyncOptions struct {
	BatchSize   int
	SourceSet   map[string]struct{}
	SymbolSet   map[string]struct{}
	SyncTypeSet map[string]struct{}
}

type tushareNewsArticlePayload struct {
	ID             string
	CategoryID     string
	Visibility     string
	SyncType       string
	Source         string
	Symbol         string
	Title          string
	Summary        string
	Content        string
	PublishedAt    time.Time
	AuthorID       string
	AttachmentURL  string
	AttachmentName string
	AttachmentMime string
}

func (r *MySQLGrowthRepo) AdminSyncDocFastNewsIncremental(batchSize int) (string, error) {
	runtime := r.resolveDocFastSyncRuntimeConfig(batchSize)
	if !runtime.Enabled {
		return "doc_fast incremental sync disabled", nil
	}
	categoryTargets, err := r.resolveDocFastSyncCategoryTargets()
	if err != nil {
		return "", err
	}

	// 1. Initialize checkpoint row if it doesn't exist yet
	if _, err := r.db.Exec(`
INSERT INTO news_sync_checkpoints
	(sync_key, cursor_updated_at, cursor_source_id, last_status, synced_articles, synced_attachments, updated_at, created_at)
VALUES
	(?, 0, 0, 'IDLE', 0, 0, NOW(), NOW())
ON DUPLICATE KEY UPDATE
	sync_key = VALUES(sync_key)`, docFastSyncCheckpointKey); err != nil {
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	cursorUpdatedAt := int64(0)
	cursorSourceID := int64(0)
	if err := r.db.QueryRow(`
SELECT cursor_updated_at, cursor_source_id
FROM news_sync_checkpoints
WHERE sync_key = ?`, docFastSyncCheckpointKey).Scan(&cursorUpdatedAt, &cursorSourceID); err != nil {
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	// 2. HTTP GET Request to PHP CMS Sync API (no DB lock held during network I/O)
	sourceItems, err := fetchArticlesFromPHPCMS(runtime.SourceBaseURL, runtime.APIToken, cursorUpdatedAt, cursorSourceID, runtime.BatchSize)
	if err != nil {
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	// 3. Database Transaction to Upsert Sync Records
	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	rollbacked := false
	defer func() {
		if !rollbacked {
			_ = tx.Rollback()
		}
	}()

	// Query with FOR UPDATE to ensure safety against race conditions
	if err := tx.QueryRow(`
SELECT cursor_updated_at, cursor_source_id
FROM news_sync_checkpoints
WHERE sync_key = ?
FOR UPDATE`, docFastSyncCheckpointKey).Scan(&cursorUpdatedAt, &cursorSourceID); err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	syncedArticles := 0
	syncedAttachments := 0
	nextCursorUpdatedAt := cursorUpdatedAt
	nextCursorSourceID := cursorSourceID

	for _, item := range sourceItems {
		// Skip if this article has already been synchronized in a concurrent/newer execution
		if item.UpdateUnix < cursorUpdatedAt || (item.UpdateUnix == cursorUpdatedAt && item.ID <= cursorSourceID) {
			continue
		}

		target, ok := categoryTargets[item.ChannelID]
		if !ok || strings.TrimSpace(target.CategoryID) == "" {
			continue
		}
		articleID := fmt.Sprintf("na_df_%d", item.ID)
		title := truncateByRunes(normalizeUTF8Text(item.Title), 256)
		if title == "" {
			title = fmt.Sprintf("doc_fast_%d", item.ID)
		}
		summary := truncateByRunes(normalizeUTF8Text(item.Description), 512)
		if summary == "" {
			summary = title
		}
		content := normalizeUTF8Text(item.Content)
		if content == "" {
			content = title
		}
		content = strings.ToValidUTF8(content, "")

		coverURL := normalizeDocFastAssetURL(runtime.SourceBaseURL, item.Image)
		createdAt := fromUnixSecondOrNow(item.CreateUnix)
		updatedAt := fromUnixSecondOrNow(item.UpdateUnix)
		publishedAt := nullableUnixSecond(item.PublishUnix)
		visibility := strings.ToUpper(strings.TrimSpace(target.Visibility))
		if visibility == "" {
			visibility = "PUBLIC"
		}

		_, err := tx.Exec(`
INSERT INTO news_articles
	(id, category_id, title, summary, content, cover_url, tags, visibility, status, published_at, author_id, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, 'PUBLISHED', ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	category_id = VALUES(category_id),
	title = VALUES(title),
	summary = VALUES(summary),
	content = VALUES(content),
	cover_url = VALUES(cover_url),
	visibility = VALUES(visibility),
	status = VALUES(status),
	published_at = VALUES(published_at),
	author_id = VALUES(author_id),
	updated_at = VALUES(updated_at)`,
			articleID,
			target.CategoryID,
			title,
			summary,
			content,
			nullableString(coverURL),
			nil,
			visibility,
			publishedAt,
			runtime.AuthorID,
			createdAt,
			updatedAt,
		)
		if err != nil {
			rollbacked = true
			_ = tx.Rollback()
			r.markDocFastSyncFailed(err.Error())
			return "", err
		}
		syncedArticles++

		// Process attachments using meta directly provided by the HTTP JSON payload
		for idx, att := range item.Attachments {
			normalizedURL := normalizeDocFastAssetURL(runtime.SourceBaseURL, att.URL)
			if normalizedURL == "" {
				continue
			}
			fileName := normalizeUTF8Text(att.FileName)
			if fileName == "" {
				fileName = normalizeUTF8Text(att.Name)
			}
			if fileName == "" {
				fileName = inferDocFastAttachmentFileName(normalizedURL, title, idx+1)
			}
			fileName = truncateByRunes(fileName, 256)
			if fileName == "" {
				fileName = fmt.Sprintf("attachment_%d", idx+1)
			}

			mimeType := truncateByRunes(normalizeUTF8Text(att.MimeType), 128)
			attachmentID := fmt.Sprintf("nat_df_%d_%d", item.ID, idx+1)
			_, err = tx.Exec(`
INSERT INTO news_attachments
	(id, article_id, file_name, file_url, file_size, mime_type, created_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	file_name = VALUES(file_name),
	file_url = VALUES(file_url),
	file_size = VALUES(file_size),
	mime_type = VALUES(mime_type)`,
				attachmentID,
				articleID,
				fileName,
				truncateByRunes(normalizedURL, 512),
				att.FileSize,
				nullableString(mimeType),
				updatedAt,
			)
			if err != nil {
				rollbacked = true
				_ = tx.Rollback()
				r.markDocFastSyncFailed(err.Error())
				return "", err
			}
			syncedAttachments++
		}

		if item.UpdateUnix > nextCursorUpdatedAt || (item.UpdateUnix == nextCursorUpdatedAt && item.ID > nextCursorSourceID) {
			nextCursorUpdatedAt = item.UpdateUnix
			nextCursorSourceID = item.ID
		}
	}

	now := time.Now()
	_, err = tx.Exec(`
UPDATE news_sync_checkpoints
SET
	cursor_updated_at = ?,
	cursor_source_id = ?,
	last_run_at = ?,
	last_success_at = ?,
	last_status = 'SUCCESS',
	last_error = NULL,
	synced_articles = synced_articles + ?,
	synced_attachments = synced_attachments + ?,
	updated_at = ?
WHERE sync_key = ?`,
		nextCursorUpdatedAt,
		nextCursorSourceID,
		now,
		now,
		syncedArticles,
		syncedAttachments,
		now,
		docFastSyncCheckpointKey,
	)
	if err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	if err := tx.Commit(); err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}
	rollbacked = true

	return fmt.Sprintf(
		"synced_articles=%d synced_attachments=%d cursor_updated_at=%d cursor_source_id=%d batch=%d",
		syncedArticles,
		syncedAttachments,
		nextCursorUpdatedAt,
		nextCursorSourceID,
		runtime.BatchSize,
	), nil
}

func (r *MySQLGrowthRepo) resolveDocFastSyncRuntimeConfig(batchSize int) docFastSyncRuntimeConfig {
	cfg := docFastSyncRuntimeConfig{
		Enabled:       true,
		BatchSize:     200,
		SourceBaseURL: "https://img.cloudup518.top",
		AuthorID:      "admin_001",
		APIToken:      "doc_fast_sync_secret_token_2026",
	}
	items, _, err := r.AdminListSystemConfigs("news.sync.doc_fast.", 1, 200)
	if err == nil {
		for _, item := range items {
			key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
			value := strings.TrimSpace(item.ConfigValue)
			switch key {
			case "news.sync.doc_fast.enabled":
				cfg.Enabled = parseRepoConfigBool(value, cfg.Enabled)
			case "news.sync.doc_fast.batch_size":
				cfg.BatchSize = parseRepoConfigInt(value, cfg.BatchSize)
			case "news.sync.doc_fast.source_base_url":
				if value != "" {
					cfg.SourceBaseURL = value
				}
			case "news.sync.doc_fast.author_id":
				if value != "" {
					cfg.AuthorID = value
				}
			case "news.sync.doc_fast.api_token":
				if value != "" {
					cfg.APIToken = value
				}
			}
		}
	}
	if batchSize > 0 {
		cfg.BatchSize = batchSize
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 200
	}
	if cfg.BatchSize > 5000 {
		cfg.BatchSize = 5000
	}
	cfg.SourceBaseURL = strings.TrimSpace(cfg.SourceBaseURL)
	if cfg.SourceBaseURL == "" {
		cfg.SourceBaseURL = "https://img.cloudup518.top"
	}
	if !strings.HasPrefix(strings.ToLower(cfg.SourceBaseURL), "http://") && !strings.HasPrefix(strings.ToLower(cfg.SourceBaseURL), "https://") {
		cfg.SourceBaseURL = "https://" + strings.TrimLeft(cfg.SourceBaseURL, "/")
	}
	cfg.SourceBaseURL = strings.TrimRight(cfg.SourceBaseURL, "/")
	cfg.AuthorID = strings.TrimSpace(cfg.AuthorID)
	if cfg.AuthorID == "" {
		cfg.AuthorID = "admin_001"
	}
	return cfg
}

func (r *MySQLGrowthRepo) resolveDocFastSyncCategoryTargets() (map[int64]docFastSyncCategoryTarget, error) {
	rows, err := r.db.Query(`
SELECT id, name, visibility
FROM news_categories
WHERE name IN ('国内研报', '海外研报', '期刊', '图书')
ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type categoryInfo struct {
		ID         string
		Visibility string
	}
	byName := make(map[string]categoryInfo)
	for rows.Next() {
		var (
			id         string
			name       string
			visibility string
		)
		if err := rows.Scan(&id, &name, &visibility); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, exists := byName[name]; exists {
			continue
		}
		byName[name] = categoryInfo{
			ID:         strings.TrimSpace(id),
			Visibility: strings.ToUpper(strings.TrimSpace(visibility)),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	required := []string{"国内研报", "海外研报", "期刊", "图书"}
	missing := make([]string, 0)
	for _, name := range required {
		if _, ok := byName[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing target news categories: %s", strings.Join(missing, ","))
	}

	targets := map[int64]docFastSyncCategoryTarget{
		13: {CategoryID: byName["国内研报"].ID, Visibility: byName["国内研报"].Visibility},
		14: {CategoryID: byName["海外研报"].ID, Visibility: byName["海外研报"].Visibility},
		26: {CategoryID: byName["期刊"].ID, Visibility: byName["期刊"].Visibility},
		30: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
		31: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
		33: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
		34: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
	}
	return targets, nil
}

func (r *MySQLGrowthRepo) loadDocFastSourceArticlesForSync(tx *sql.Tx, cursorUpdatedAt int64, cursorSourceID int64, limit int) ([]docFastSourceArticle, error) {
	rows, err := tx.Query(`
SELECT
	a.id,
	a.channel_id,
	a.title,
	a.image,
	a.description,
	COALESCE(a.publishtime, 0) AS publishtime,
	COALESCE(a.createtime, 0) AS createtime,
	COALESCE(a.updatetime, 0) AS updatetime,
	COALESCE(ad.content, '') AS content,
	COALESCE(ad.downloadurl, '') AS downloadurl
FROM doc_fast.fa_cms_archives a
LEFT JOIN doc_fast.fa_cms_addondownload ad ON ad.id = a.id
WHERE
	a.status = 'normal'
	AND a.channel_id IN (13, 14, 26, 30, 31, 33, 34)
	AND (
		COALESCE(a.updatetime, 0) > ?
		OR (COALESCE(a.updatetime, 0) = ? AND a.id > ?)
	)
ORDER BY COALESCE(a.updatetime, 0) ASC, a.id ASC
LIMIT ?`, cursorUpdatedAt, cursorUpdatedAt, cursorSourceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]docFastSourceArticle, 0, limit)
	for rows.Next() {
		var item docFastSourceArticle
		if err := rows.Scan(
			&item.ID,
			&item.ChannelID,
			&item.Title,
			&item.Image,
			&item.Description,
			&item.PublishUnix,
			&item.CreateUnix,
			&item.UpdateUnix,
			&item.Content,
			&item.DownloadURL,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) loadDocFastAttachmentMeta(tx *sql.Tx, sourceURL string) (docFastAttachmentMeta, error) {
	candidate := strings.TrimSpace(sourceURL)
	if candidate == "" {
		return docFastAttachmentMeta{}, nil
	}
	alt := candidate
	if strings.HasPrefix(candidate, "/") {
		alt = strings.TrimPrefix(candidate, "/")
	} else {
		alt = "/" + candidate
	}
	var (
		fileName sql.NullString
		mimeType sql.NullString
		fileSize sql.NullInt64
	)
	err := tx.QueryRow(`
SELECT filename, mimetype, filesize
FROM doc_fast.fa_attachment
WHERE url IN (?, ?)
ORDER BY id DESC
LIMIT 1`, candidate, alt).Scan(&fileName, &mimeType, &fileSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return docFastAttachmentMeta{}, nil
		}
		return docFastAttachmentMeta{}, err
	}
	result := docFastAttachmentMeta{}
	if fileName.Valid {
		result.FileName = fileName.String
	}
	if mimeType.Valid {
		result.MimeType = mimeType.String
	}
	if fileSize.Valid {
		result.FileSize = fileSize.Int64
	}
	return result, nil
}

func parseDocFastDownloadAttachments(raw string) []docFastDownloadAttachment {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	items := make([]docFastDownloadAttachment, 0)
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		return nil
	}
	return items
}

func normalizeDocFastAssetURL(baseURL string, raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return value
	}
	if strings.HasPrefix(value, "//") {
		return "https:" + value
	}
	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBase == "" {
		trimmedBase = "https://img.cloudup518.top"
	}
	if strings.HasPrefix(value, "/") {
		return trimmedBase + value
	}
	return trimmedBase + "/" + strings.TrimLeft(value, "/")
}

func inferDocFastAttachmentFileName(fileURL string, title string, index int) string {
	trimmedURL := strings.TrimSpace(fileURL)
	base := path.Base(trimmedURL)
	if base != "" && base != "." && base != "/" {
		return base
	}
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		cleanTitle = "attachment"
	}
	return fmt.Sprintf("%s_%d", cleanTitle, index)
}

func truncateByRunes(value string, size int) string {
	if size <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= size {
		return value
	}
	return string(runes[:size])
}

func normalizeUTF8Text(value string) string {
	return strings.TrimSpace(strings.ToValidUTF8(value, ""))
}

func fromUnixSecondOrNow(value int64) time.Time {
	if value > 0 {
		return time.Unix(value, 0)
	}
	return time.Now()
}

func nullableUnixSecond(value int64) interface{} {
	if value > 0 {
		return time.Unix(value, 0)
	}
	return nil
}

func (r *MySQLGrowthRepo) markDocFastSyncFailed(errText string) {
	now := time.Now()
	_, _ = r.db.Exec(`
UPDATE news_sync_checkpoints
SET
	last_run_at = ?,
	last_status = 'FAILED',
	last_error = ?,
	updated_at = ?
WHERE sync_key = ?`,
		now,
		truncateByRunes(normalizeUTF8Text(errText), 255),
		now,
		docFastSyncCheckpointKey,
	)
}
