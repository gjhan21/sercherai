package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	mysqlDSN      = "root:abc123@tcp(127.0.0.1:3306)/sercherai?parseTime=true&charset=utf8mb4"
	sourceBaseURL = "https://img.cloudup518.top"
	authorID      = "admin_001"
)

type sourceArticle struct {
	ID          int64
	ChannelID   int64
	Title       string
	Image       string
	Description string
	PublishUnix sql.NullInt64
	CreateUnix  sql.NullInt64
	UpdateUnix  sql.NullInt64
	Content     string
	DownloadURL string
}

type sourceAttachmentRef struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Password string `json:"password"`
}

type sourceAttachmentMeta struct {
	FileName string
	MimeType string
	FileSize int64
}

var channelToCategory = map[int64]string{
	13: "nc_1772283341970969000", // 国内研报
	14: "nc_1772283370752651000", // 海外研报
	26: "nc_1772158767497404000", // 期刊
	30: "nc_1772283434862505000", // 图书-金融类
	31: "nc_1772283434862505000", // 图书-历史
	33: "nc_1772283434862505000", // 图书-名人传记
	34: "nc_1772283434862505000", // 图书-社科文学
}

func main() {
	syncLimit := getEnvInt("SYNC_LIMIT", 10)
	syncBatchSize := getEnvInt("SYNC_BATCH_SIZE", 200)
	syncOffset := getEnvInt("SYNC_OFFSET", 0)
	if syncBatchSize <= 0 {
		syncBatchSize = 200
	}
	if syncOffset < 0 {
		syncOffset = 0
	}

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("open mysql failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping mysql failed: %v", err)
	}
	log.Printf("start sync: limit=%d batch_size=%d offset=%d", syncLimit, syncBatchSize, syncOffset)

	visibilityMap, err := loadCategoryVisibility(db)
	if err != nil {
		log.Fatalf("load category visibility failed: %v", err)
	}

	articleCount := 0
	attachmentCount := 0
	offset := syncOffset
	for {
		batchLimit := syncBatchSize
		if syncLimit > 0 {
			remaining := syncLimit - articleCount
			if remaining <= 0 {
				break
			}
			if remaining < batchLimit {
				batchLimit = remaining
			}
		}

		articles, err := loadSourceArticlesBatch(db, batchLimit, offset)
		if err != nil {
			log.Fatalf("load source articles failed: %v", err)
		}
		if len(articles) == 0 {
			break
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("begin tx failed: %v", err)
		}

		batchArticleCount := 0
		batchAttachmentCount := 0
		for _, src := range articles {
			categoryID, ok := channelToCategory[src.ChannelID]
			if !ok {
				continue
			}
			visibility := visibilityMap[categoryID]
			if visibility == "" {
				visibility = "PUBLIC"
			}

			articleID := fmt.Sprintf("na_df_%d", src.ID)
			summary := strings.TrimSpace(src.Description)
			if summary == "" {
				summary = src.Title
			}
			content := strings.TrimSpace(src.Content)
			if content == "" {
				content = src.Title
			}
			coverURL := normalizeSourceURL(src.Image)
			createdAt := unixToTimeOrNow(src.CreateUnix)
			updatedAt := unixToTimeOrNow(src.UpdateUnix)
			publishedAt := unixToTimeOrNil(src.PublishUnix)

			_, err = tx.Exec(`
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
	updated_at = VALUES(updated_at)`,
				articleID,
				categoryID,
				src.Title,
				truncate(summary, 512),
				content,
				nullIfEmpty(coverURL),
				nil,
				visibility,
				publishedAt,
				authorID,
				createdAt,
				updatedAt,
			)
			if err != nil {
				_ = tx.Rollback()
				log.Fatalf("upsert article failed source_id=%d: %v", src.ID, err)
			}
			articleCount++
			batchArticleCount++

			refs := parseAttachmentRefs(src.DownloadURL)
			for idx, ref := range refs {
				u := normalizeSourceURL(ref.URL)
				if u == "" {
					continue
				}
				originPath := strings.TrimSpace(ref.URL)
				meta, _ := loadAttachmentMeta(tx, originPath)
				fileName := strings.TrimSpace(meta.FileName)
				if fileName == "" {
					fileName = inferFileName(u, src.Title, idx+1)
				}
				attachmentID := fmt.Sprintf("nat_df_%d_%d", src.ID, idx+1)
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
					truncate(fileName, 256),
					truncate(u, 512),
					meta.FileSize,
					nullIfEmpty(meta.MimeType),
					updatedAt,
				)
				if err != nil {
					_ = tx.Rollback()
					log.Fatalf("upsert attachment failed source_id=%d idx=%d: %v", src.ID, idx+1, err)
				}
				attachmentCount++
				batchAttachmentCount++
			}
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("commit failed: %v", err)
		}
		offset += len(articles)
		log.Printf("batch done: offset=%d batch_articles=%d batch_attachments=%d total_articles=%d total_attachments=%d",
			offset, batchArticleCount, batchAttachmentCount, articleCount, attachmentCount)

		if len(articles) < batchLimit {
			break
		}
	}

	log.Printf("sync done: articles=%d attachments=%d offset=%d", articleCount, attachmentCount, offset)

	printPreview(db)
}

func loadCategoryVisibility(db *sql.DB) (map[string]string, error) {
	result := make(map[string]string, len(channelToCategory))
	seen := make(map[string]struct{})
	args := make([]interface{}, 0, len(channelToCategory))
	holders := make([]string, 0, len(channelToCategory))
	for _, categoryID := range channelToCategory {
		if _, ok := seen[categoryID]; ok {
			continue
		}
		seen[categoryID] = struct{}{}
		holders = append(holders, "?")
		args = append(args, categoryID)
	}
	query := "SELECT id, visibility FROM news_categories WHERE id IN (" + strings.Join(holders, ",") + ")"
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var visibility string
		if err := rows.Scan(&id, &visibility); err != nil {
			return nil, err
		}
		result[id] = strings.ToUpper(strings.TrimSpace(visibility))
	}
	return result, rows.Err()
}

func loadSourceArticlesBatch(db *sql.DB, limit int, offset int) ([]sourceArticle, error) {
	query := `
SELECT
	a.id,
	a.channel_id,
	a.title,
	a.image,
	a.description,
	a.publishtime,
	a.createtime,
	a.updatetime,
	COALESCE(ad.content, '') AS content,
	COALESCE(ad.downloadurl, '') AS downloadurl
FROM doc_fast.fa_cms_archives a
LEFT JOIN doc_fast.fa_cms_addondownload ad ON ad.id = a.id
WHERE
	a.status = 'normal'
	AND a.channel_id IN (13, 14, 26, 30, 31, 33, 34)
ORDER BY a.publishtime DESC, a.id DESC
LIMIT ? OFFSET ?`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]sourceArticle, 0, limit)
	for rows.Next() {
		var item sourceArticle
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

func getEnvInt(key string, def int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return number
}

func normalizeSourceURL(raw string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	lower := strings.ToLower(v)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return v
	}
	if strings.HasPrefix(v, "//") {
		return "https:" + v
	}
	if strings.HasPrefix(v, "/") {
		return strings.TrimRight(sourceBaseURL, "/") + v
	}
	return strings.TrimRight(sourceBaseURL, "/") + "/" + strings.TrimLeft(v, "/")
}

func parseAttachmentRefs(raw string) []sourceAttachmentRef {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	refs := make([]sourceAttachmentRef, 0)
	if err := json.Unmarshal([]byte(text), &refs); err != nil {
		return nil
	}
	return refs
}

func loadAttachmentMeta(tx *sql.Tx, srcURL string) (sourceAttachmentMeta, error) {
	meta := sourceAttachmentMeta{}
	candidate := strings.TrimSpace(srcURL)
	if candidate == "" {
		return meta, nil
	}
	err := tx.QueryRow(`
SELECT filename, mimetype, filesize
FROM doc_fast.fa_attachment
WHERE url = ?
ORDER BY id DESC
LIMIT 1`, candidate).Scan(&meta.FileName, &meta.MimeType, &meta.FileSize)
	if err == nil {
		return meta, nil
	}
	if err != sql.ErrNoRows {
		return meta, err
	}
	return sourceAttachmentMeta{}, nil
}

func unixToTimeOrNow(v sql.NullInt64) time.Time {
	if v.Valid && v.Int64 > 0 {
		return time.Unix(v.Int64, 0)
	}
	return time.Now()
}

func unixToTimeOrNil(v sql.NullInt64) interface{} {
	if v.Valid && v.Int64 > 0 {
		return time.Unix(v.Int64, 0)
	}
	return nil
}

func truncate(v string, size int) string {
	if size <= 0 {
		return ""
	}
	runes := []rune(v)
	if len(runes) <= size {
		return v
	}
	return string(runes[:size])
}

func nullIfEmpty(v string) interface{} {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func inferFileName(fileURL string, title string, seq int) string {
	trimmed := strings.TrimSpace(fileURL)
	base := path.Base(trimmed)
	if base != "" && base != "." && base != "/" {
		return base
	}
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		cleanTitle = "attachment"
	}
	return fmt.Sprintf("%s_%d", cleanTitle, seq)
}

func printPreview(db *sql.DB) {
	rows, err := db.Query(`
SELECT id, category_id, title, visibility, status, published_at
FROM news_articles
WHERE id LIKE 'na_df_%'
ORDER BY published_at DESC, id DESC
LIMIT 10`)
	if err != nil {
		log.Printf("preview articles failed: %v", err)
		return
	}
	defer rows.Close()

	log.Println("latest synced articles:")
	for rows.Next() {
		var (
			id         string
			categoryID string
			title      string
			visibility string
			status     string
			published  sql.NullTime
		)
		if err := rows.Scan(&id, &categoryID, &title, &visibility, &status, &published); err != nil {
			log.Printf("scan preview article failed: %v", err)
			return
		}
		pub := ""
		if published.Valid {
			pub = published.Time.Format("2006-01-02 15:04:05")
		}
		log.Printf("- %s | %s | %s | %s | %s | %s", id, categoryID, visibility, status, pub, title)
	}
}
