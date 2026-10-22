package repo

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminSyncTushareNewsIncremental(batchSize int) (string, error) {
	summary, _, err := r.AdminSyncTushareNewsIncrementalWithOptions(model.TushareNewsSyncOptions{
		BatchSize: batchSize,
	})
	return summary, err
}

func (r *MySQLGrowthRepo) AdminSyncTushareNewsIncrementalWithOptions(opts model.TushareNewsSyncOptions) (string, []model.NewsSyncRunDetail, error) {
	normalizedOpts := normalizeTushareSyncOptions(opts)
	runtime := r.resolveTushareNewsSyncRuntimeConfig(normalizedOpts.BatchSize)
	if !runtime.Enabled {
		return "tushare news incremental sync disabled", nil, nil
	}
	if len(normalizedOpts.SyncTypeSet) == 0 {
		for _, syncType := range runtime.SyncTypes {
			key := normalizeTushareSyncType(syncType)
			if key == "" {
				continue
			}
			normalizedOpts.SyncTypeSet[key] = struct{}{}
		}
	}
	if err := r.ensureTushareNewsCategories(); err != nil {
		return "", nil, err
	}
	targets, err := r.resolveTushareNewsCategoryTargets(runtime)
	if err != nil {
		return "", nil, err
	}
	token, err := r.resolveTushareTokenForNewsSync()
	if err != nil {
		return "", nil, err
	}

	now := time.Now()
	mergedItems := make(map[string]tushareNewsArticlePayload)
	fetchWarnings := make([]string, 0)
	detailItems := make([]model.NewsSyncRunDetail, 0, 32)
	counts := map[string]int{
		"brief":        0,
		"major":        0,
		"research":     0,
		"forecast":     0,
		"announcement": 0,
	}

	mergeItems := func(counterKey string, items []tushareNewsArticlePayload, details []model.NewsSyncRunDetail) {
		counts[counterKey] += len(items)
		detailItems = append(detailItems, details...)
		for _, item := range items {
			if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.CategoryID) == "" {
				continue
			}
			existing, exists := mergedItems[item.ID]
			if !exists || item.PublishedAt.After(existing.PublishedAt) {
				mergedItems[item.ID] = item
			}
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeNewsBrief) {
		items, details, fetchErr := fetchTushareNewsBriefArticles(token, runtime, targets.Brief, now, normalizedOpts)
		mergeItems("brief", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("news: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeNewsMajor) {
		items, details, fetchErr := fetchTushareMajorNewsArticles(token, runtime, targets.Major, now, normalizedOpts)
		mergeItems("major", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("major_news: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeResearch) {
		items, details, fetchErr := fetchTushareResearchReportArticles(token, runtime, targets.Research, now, normalizedOpts)
		mergeItems("research", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("research_report: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeForecast) {
		items, details, fetchErr := fetchTushareForecastArticles(token, runtime, targets.Forecast, now, normalizedOpts)
		mergeItems("forecast", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("report_rc: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeAnnouncement) {
		items, details, fetchErr := fetchTushareAnnouncementArticles(token, runtime, targets.Announcement, now, normalizedOpts)
		mergeItems("announcement", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("anns_d: %v", fetchErr))
		}
	}

	readyItems := make([]tushareNewsArticlePayload, 0, len(mergedItems))
	for _, item := range mergedItems {
		readyItems = append(readyItems, item)
	}

	syncedArticles, syncedAttachments, err := r.upsertTushareNewsArticles(readyItems)
	if err != nil {
		return "", detailItems, err
	}
	applyTushareDetailUpsertCounts(detailItems, readyItems)

	summary := fmt.Sprintf(
		"brief=%d major=%d research=%d forecast=%d announcement=%d upserted=%d attachments=%d",
		counts["brief"],
		counts["major"],
		counts["research"],
		counts["forecast"],
		counts["announcement"],
		syncedArticles,
		syncedAttachments,
	)
	if len(normalizedOpts.SourceSet) > 0 {
		summary += fmt.Sprintf(" source_filter=%d", len(normalizedOpts.SourceSet))
	}
	if len(normalizedOpts.SymbolSet) > 0 {
		summary += fmt.Sprintf(" symbol_filter=%d", len(normalizedOpts.SymbolSet))
	}
	if len(fetchWarnings) > 0 {
		summary += " warnings=" + strings.Join(compactErrorMessages(fetchWarnings, 3), " | ")
	}
	if len(readyItems) == 0 && len(fetchWarnings) > 0 {
		if onlyRateLimitMessages(fetchWarnings) {
			summary += " rate_limited=true"
			return summary, detailItems, nil
		}
		return summary, detailItems, errors.New(strings.Join(compactErrorMessages(fetchWarnings, 2), "; "))
	}
	return summary, detailItems, nil
}

func (r *MySQLGrowthRepo) resolveTushareNewsSyncRuntimeConfig(batchSize int) tushareNewsSyncRuntimeConfig {
	cfg := tushareNewsSyncRuntimeConfig{
		Enabled:                  true,
		BatchSize:                200,
		TimeoutMS:                12000,
		AuthorID:                 "admin_001",
		SyncTypes:                []string{tushareSyncTypeNewsBrief, tushareSyncTypeNewsMajor},
		BriefSources:             []string{"cls", "yicai", "sina", "eastmoney"},
		MajorSources:             []string{"新华网", "财联社", "第一财经", "新浪财经", "华尔街见闻"},
		BriefLookbackHours:       8,
		MajorLookbackHours:       24,
		ResearchLookbackDays:     7,
		ForecastLookbackDays:     7,
		AnnouncementLookbackDays: 2,
		BriefVisibility:          "PUBLIC",
		MajorVisibility:          "PUBLIC",
		ResearchVisibility:       "VIP",
		ForecastVisibility:       "VIP",
		AnnouncementVisibility:   "PUBLIC",
	}

	items, _, err := r.AdminListSystemConfigs("news.sync.tushare.", 1, 200)
	if err == nil {
		for _, item := range items {
			key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
			value := strings.TrimSpace(item.ConfigValue)
			switch key {
			case "news.sync.tushare.enabled":
				cfg.Enabled = parseRepoConfigBool(value, cfg.Enabled)
			case "news.sync.tushare.batch_size":
				cfg.BatchSize = parseRepoConfigInt(value, cfg.BatchSize)
			case "news.sync.tushare.timeout_ms":
				cfg.TimeoutMS = parseRepoConfigInt(value, cfg.TimeoutMS)
			case "news.sync.tushare.author_id":
				if value != "" {
					cfg.AuthorID = value
				}
			case "news.sync.tushare.sync_types":
				cfg.SyncTypes = splitTushareSyncTypeList(value, cfg.SyncTypes)
			case "news.sync.tushare.sources.news_brief":
				cfg.BriefSources = splitTushareSourceList(value, cfg.BriefSources)
			case "news.sync.tushare.sources.news_major":
				cfg.MajorSources = splitTushareSourceList(value, cfg.MajorSources)
			case "news.sync.tushare.lookback_hours.news_brief":
				cfg.BriefLookbackHours = parseRepoConfigInt(value, cfg.BriefLookbackHours)
			case "news.sync.tushare.lookback_hours.news_major":
				cfg.MajorLookbackHours = parseRepoConfigInt(value, cfg.MajorLookbackHours)
			case "news.sync.tushare.lookback_days.research_report":
				cfg.ResearchLookbackDays = parseRepoConfigInt(value, cfg.ResearchLookbackDays)
			case "news.sync.tushare.lookback_days.report_rc":
				cfg.ForecastLookbackDays = parseRepoConfigInt(value, cfg.ForecastLookbackDays)
			case "news.sync.tushare.lookback_days.announcement":
				cfg.AnnouncementLookbackDays = parseRepoConfigInt(value, cfg.AnnouncementLookbackDays)
			case "news.sync.tushare.visibility.news_brief":
				cfg.BriefVisibility = value
			case "news.sync.tushare.visibility.news_major":
				cfg.MajorVisibility = value
			case "news.sync.tushare.visibility.research_report":
				cfg.ResearchVisibility = value
			case "news.sync.tushare.visibility.report_rc":
				cfg.ForecastVisibility = value
			case "news.sync.tushare.visibility.announcement":
				cfg.AnnouncementVisibility = value
			}
		}
	}

	if batchSize > 0 {
		cfg.BatchSize = batchSize
	}
	cfg.BatchSize = maxInt(20, minInt(cfg.BatchSize, 1000))
	cfg.TimeoutMS = maxInt(2000, minInt(cfg.TimeoutMS, 30000))
	cfg.BriefLookbackHours = maxInt(1, minInt(cfg.BriefLookbackHours, 72))
	cfg.MajorLookbackHours = maxInt(1, minInt(cfg.MajorLookbackHours, 168))
	cfg.ResearchLookbackDays = maxInt(1, minInt(cfg.ResearchLookbackDays, 60))
	cfg.ForecastLookbackDays = maxInt(1, minInt(cfg.ForecastLookbackDays, 60))
	cfg.AnnouncementLookbackDays = maxInt(1, minInt(cfg.AnnouncementLookbackDays, 30))
	cfg.SyncTypes = splitTushareSyncTypeList(strings.Join(cfg.SyncTypes, ","), []string{tushareSyncTypeNewsBrief, tushareSyncTypeNewsMajor})
	cfg.BriefSources = splitTushareSourceList(strings.Join(cfg.BriefSources, ","), []string{"cls", "yicai", "sina", "eastmoney"})
	cfg.MajorSources = splitTushareSourceList(strings.Join(cfg.MajorSources, ","), []string{"新华网", "财联社", "第一财经", "新浪财经", "华尔街见闻"})
	cfg.AuthorID = strings.TrimSpace(cfg.AuthorID)
	if cfg.AuthorID == "" {
		cfg.AuthorID = "admin_001"
	}
	cfg.BriefVisibility = normalizeTushareNewsVisibility(cfg.BriefVisibility)
	cfg.MajorVisibility = normalizeTushareNewsVisibility(cfg.MajorVisibility)
	cfg.ResearchVisibility = normalizeTushareNewsVisibility(cfg.ResearchVisibility)
	cfg.ForecastVisibility = normalizeTushareNewsVisibility(cfg.ForecastVisibility)
	cfg.AnnouncementVisibility = normalizeTushareNewsVisibility(cfg.AnnouncementVisibility)
	return cfg
}

func (r *MySQLGrowthRepo) ensureTushareNewsCategories() error {
	type categorySeed struct {
		ID         string
		Name       string
		Slug       string
		Sort       int
		Visibility string
	}
	seeds := []categorySeed{
		{ID: "nc_ts_news_brief", Name: "新闻快讯", Slug: tushareNewsCategoryBriefSlug, Sort: 210, Visibility: "PUBLIC"},
		{ID: "nc_ts_news_major", Name: "新闻通讯", Slug: tushareNewsCategoryMajorSlug, Sort: 220, Visibility: "PUBLIC"},
		{ID: "nc_ts_report_research", Name: "券商研报", Slug: tushareNewsCategoryResearchSlug, Sort: 230, Visibility: "VIP"},
		{ID: "nc_ts_report_forecast", Name: "盈利预测", Slug: tushareNewsCategoryForecastSlug, Sort: 240, Visibility: "VIP"},
		{ID: "nc_ts_announcement", Name: "上市公司公告", Slug: tushareNewsCategoryAnnouncementSlug, Sort: 250, Visibility: "PUBLIC"},
	}
	now := time.Now()
	for _, seed := range seeds {
		_, err := r.db.Exec(`
	INSERT INTO news_categories
		(id, name, slug, sort, visibility, status, created_at, updated_at)
	SELECT
		?, ?, ?, ?, ?, 'PUBLISHED', ?, ?
	FROM DUAL
	WHERE NOT EXISTS (
		SELECT 1 FROM news_categories WHERE slug = ?
	)`,
			seed.ID,
			seed.Name,
			seed.Slug,
			seed.Sort,
			seed.Visibility,
			now,
			now,
			seed.Slug,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) resolveTushareNewsCategoryTargets(runtime tushareNewsSyncRuntimeConfig) (tushareNewsCategoryTargets, error) {
	rows, err := r.db.Query(`
SELECT id, slug, visibility
FROM news_categories
WHERE slug IN (?, ?, ?, ?, ?)`,
		tushareNewsCategoryBriefSlug,
		tushareNewsCategoryMajorSlug,
		tushareNewsCategoryResearchSlug,
		tushareNewsCategoryForecastSlug,
		tushareNewsCategoryAnnouncementSlug,
	)
	if err != nil {
		return tushareNewsCategoryTargets{}, err
	}
	defer rows.Close()

	bySlug := make(map[string]tushareNewsCategoryTarget)
	for rows.Next() {
		var (
			id         string
			slug       string
			visibility string
		)
		if err := rows.Scan(&id, &slug, &visibility); err != nil {
			return tushareNewsCategoryTargets{}, err
		}
		bySlug[strings.TrimSpace(slug)] = tushareNewsCategoryTarget{
			CategoryID: strings.TrimSpace(id),
			Visibility: normalizeTushareNewsVisibility(visibility),
		}
	}
	if err := rows.Err(); err != nil {
		return tushareNewsCategoryTargets{}, err
	}

	missing := make([]string, 0)
	requiredSlugs := []string{
		tushareNewsCategoryBriefSlug,
		tushareNewsCategoryMajorSlug,
		tushareNewsCategoryResearchSlug,
		tushareNewsCategoryForecastSlug,
		tushareNewsCategoryAnnouncementSlug,
	}
	for _, slug := range requiredSlugs {
		if _, ok := bySlug[slug]; !ok {
			missing = append(missing, slug)
		}
	}
	if len(missing) > 0 {
		return tushareNewsCategoryTargets{}, fmt.Errorf("missing tushare news categories: %s", strings.Join(missing, ","))
	}

	targets := tushareNewsCategoryTargets{
		Brief: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryBriefSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.BriefVisibility, bySlug[tushareNewsCategoryBriefSlug].Visibility),
		},
		Major: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryMajorSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.MajorVisibility, bySlug[tushareNewsCategoryMajorSlug].Visibility),
		},
		Research: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryResearchSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.ResearchVisibility, bySlug[tushareNewsCategoryResearchSlug].Visibility),
		},
		Forecast: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryForecastSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.ForecastVisibility, bySlug[tushareNewsCategoryForecastSlug].Visibility),
		},
		Announcement: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryAnnouncementSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.AnnouncementVisibility, bySlug[tushareNewsCategoryAnnouncementSlug].Visibility),
		},
	}
	return targets, nil
}

func (r *MySQLGrowthRepo) resolveTushareTokenForNewsSync() (string, error) {
	item, err := r.getDataSourceBySourceKey("TUSHARE")
	if err == nil {
		if token := parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token"); strings.TrimSpace(token) != "" {
			return strings.TrimSpace(token), nil
		}
	}
	if token := strings.TrimSpace(os.Getenv("TUSHARE_TOKEN")); token != "" {
		return token, nil
	}
	return "", errors.New("tushare token not configured")
}

func (r *MySQLGrowthRepo) upsertTushareNewsArticles(items []tushareNewsArticlePayload) (int, int, error) {
	if len(items) == 0 {
		return 0, 0, nil
	}
	syncedArticles := 0
	syncedAttachments := 0
	now := time.Now()
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		categoryID := strings.TrimSpace(item.CategoryID)
		title := truncateByRunes(normalizeUTF8Text(item.Title), 256)
		if id == "" || categoryID == "" || title == "" {
			continue
		}
		summary := truncateByRunes(normalizeUTF8Text(item.Summary), 512)
		if summary == "" {
			summary = title
		}
		content := normalizeUTF8Text(item.Content)
		if content == "" {
			content = summary
		}
		visibility := normalizeTushareNewsVisibility(item.Visibility)
		if visibility == "" {
			visibility = "PUBLIC"
		}
		publishedAt := item.PublishedAt
		if publishedAt.IsZero() {
			publishedAt = now
		}
		authorID := strings.TrimSpace(item.AuthorID)
		if authorID == "" {
			authorID = "admin_001"
		}

		_, err := r.db.Exec(`
INSERT INTO news_articles
	(id, category_id, title, summary, content, cover_url, tags, visibility, status, published_at, author_id, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, NULL, NULL, ?, 'PUBLISHED', ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	category_id = VALUES(category_id),
	title = VALUES(title),
	summary = VALUES(summary),
	content = VALUES(content),
	visibility = VALUES(visibility),
	status = VALUES(status),
	published_at = VALUES(published_at),
	author_id = VALUES(author_id),
	updated_at = VALUES(updated_at)`,
			id,
			categoryID,
			title,
			summary,
			content,
			visibility,
			publishedAt,
			authorID,
			now,
			now,
		)
		if err != nil {
			return syncedArticles, syncedAttachments, err
		}
		syncedArticles++

		attachmentURL := strings.TrimSpace(item.AttachmentURL)
		if attachmentURL == "" {
			continue
		}
		attachmentID := buildTushareNewsAttachmentID("att", id+"|"+attachmentURL)
		fileName := inferTushareAttachmentFileName(attachmentURL, item.AttachmentName, title)
		mimeType := strings.TrimSpace(item.AttachmentMime)
		if mimeType == "" {
			mimeType = inferAttachmentMimeTypeFromFileName(fileName)
		}
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		_, err = r.db.Exec(`
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
			id,
			fileName,
			truncateByRunes(attachmentURL, 512),
			1,
			mimeType,
			now,
		)
		if err != nil {
			return syncedArticles, syncedAttachments, err
		}
		syncedAttachments++
	}
	return syncedArticles, syncedAttachments, nil
}

func fetchTushareNewsBriefArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startAt := now.Add(-time.Duration(runtime.BriefLookbackHours) * time.Hour)
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	limit := maxInt(50, minInt(runtime.BatchSize, 1000))
	items := make([]tushareNewsArticlePayload, 0, limit)
	details := make([]model.NewsSyncRunDetail, 0, maxInt(1, len(runtime.BriefSources)))
	seen := make(map[string]struct{})
	errs := make([]string, 0)

	sources := filterTushareSources(runtime.BriefSources, opts.SourceSet)
	if len(sources) == 0 {
		return items, details, nil
	}
	for _, source := range sources {
		source = strings.ToLower(strings.TrimSpace(source))
		if source == "" {
			continue
		}
		detail := newTushareSyncDetailDraft(tushareSyncTypeNewsBrief, source, "")
		callStartedAt := time.Now()
		params := map[string]string{
			"src":        source,
			"start_date": startAt.Format("2006-01-02 15:04:05"),
			"end_date":   now.Format("2006-01-02 15:04:05"),
			"limit":      strconv.Itoa(limit),
		}
		parsed, err := callTushareAPI(client, token, "news", params, "datetime,title,content,channels")
		if err != nil && shouldRetryTushareWithoutFields(err) {
			parsed, err = callTushareAPI(client, token, "news", params, "")
		}
		if err != nil {
			rateLimited := isTushareRateLimitErrorText(err.Error())
			detail.Status = "FAILED"
			detail.FailedCount = 1
			detail.ErrorText = truncateByRunes(err.Error(), 512)
			if rateLimited {
				detail.WarningText = "触发Tushare限频，已停止后续来源抓取"
			}
			detail.StartedAt = callStartedAt.Format(time.RFC3339)
			detail.FinishedAt = time.Now().Format(time.RFC3339)
			details = append(details, detail)
			errs = append(errs, fmt.Sprintf("%s: %v", source, err))
			if rateLimited {
				break
			}
			continue
		}
		beforeCount := len(items)
		fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
		for _, row := range parsed.Data.Items {
			datetimeText := firstTushareString(row, fieldIndex, "datetime")
			contentText := firstTushareString(row, fieldIndex, "content")
			title := firstTushareString(row, fieldIndex, "title")
			channels := firstTushareString(row, fieldIndex, "channels")
			if title == "" {
				title = truncateByRunes(normalizeUTF8Text(contentText), 80)
			}
			title = normalizeUTF8Text(title)
			contentText = normalizeUTF8Text(contentText)
			if title == "" || contentText == "" {
				continue
			}
			publishedAt := parseTushareDateTimeOr(datetimeText, now)
			uniqueKey := fmt.Sprintf("brief|%s|%s|%s|%s", source, publishedAt.Format(time.RFC3339), title, channels)
			id := buildTushareNewsArticleID("nb", uniqueKey)
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			if channels != "" {
				contentText = fmt.Sprintf("频道：%s\n\n%s", channels, contentText)
			}
			items = append(items, tushareNewsArticlePayload{
				ID:          id,
				CategoryID:  target.CategoryID,
				Visibility:  target.Visibility,
				SyncType:    tushareSyncTypeNewsBrief,
				Source:      source,
				Title:       title,
				Summary:     buildTushareNewsSummary(contentText, title),
				Content:     contentText,
				PublishedAt: publishedAt,
				AuthorID:    runtime.AuthorID,
			})
		}
		fetchedCount := len(items) - beforeCount
		detail.Status = "SUCCESS"
		detail.FetchedCount = fetchedCount
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		details = append(details, detail)
	}
	if len(items) == 0 && len(errs) > 0 {
		return nil, details, errors.New(strings.Join(compactErrorMessages(errs, 3), " | "))
	}
	return items, details, nil
}

func fetchTushareMajorNewsArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startAt := now.Add(-time.Duration(runtime.MajorLookbackHours) * time.Hour)
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	limit := maxInt(50, minInt(runtime.BatchSize, 1000))
	items := make([]tushareNewsArticlePayload, 0, limit)
	details := make([]model.NewsSyncRunDetail, 0, maxInt(1, len(runtime.MajorSources)))
	seen := make(map[string]struct{})
	errs := make([]string, 0)

	sources := runtime.MajorSources
	if len(sources) == 0 {
		sources = []string{""}
	}
	sources = filterTushareSources(sources, opts.SourceSet)
	if len(sources) == 0 {
		return items, details, nil
	}

	for _, source := range sources {
		source = strings.TrimSpace(source)
		detailSource := source
		if detailSource == "" {
			detailSource = "all"
		}
		detail := newTushareSyncDetailDraft(tushareSyncTypeNewsMajor, detailSource, "")
		callStartedAt := time.Now()
		params := map[string]string{
			"start_date": startAt.Format("2006-01-02 15:04:05"),
			"end_date":   now.Format("2006-01-02 15:04:05"),
			"limit":      strconv.Itoa(limit),
		}
		if source != "" {
			params["src"] = source
		}
		parsed, err := callTushareAPI(client, token, "major_news", params, "pub_time,src,title,content")
		if err != nil && shouldRetryTushareWithoutFields(err) {
			parsed, err = callTushareAPI(client, token, "major_news", params, "")
		}
		if err != nil {
			rateLimited := isTushareRateLimitErrorText(err.Error())
			detail.Status = "FAILED"
			detail.FailedCount = 1
			detail.ErrorText = truncateByRunes(err.Error(), 512)
			if rateLimited {
				detail.WarningText = "触发Tushare限频，已停止后续来源抓取"
			}
			detail.StartedAt = callStartedAt.Format(time.RFC3339)
			detail.FinishedAt = time.Now().Format(time.RFC3339)
			details = append(details, detail)
			errs = append(errs, fmt.Sprintf("%s: %v", detailSource, err))
			if rateLimited {
				break
			}
			continue
		}
		beforeCount := len(items)
		fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
		for _, row := range parsed.Data.Items {
			pubTimeText := firstTushareString(row, fieldIndex, "pub_time")
			rowSource := firstTushareString(row, fieldIndex, "src")
			title := firstTushareString(row, fieldIndex, "title")
			contentText := firstTushareString(row, fieldIndex, "content")
			title = normalizeUTF8Text(title)
			contentText = normalizeUTF8Text(contentText)
			if title == "" {
				title = truncateByRunes(contentText, 80)
			}
			if title == "" || contentText == "" {
				continue
			}
			publishedAt := parseTushareDateTimeOr(pubTimeText, now)
			uniqueKey := fmt.Sprintf("major|%s|%s|%s", rowSource, publishedAt.Format(time.RFC3339), title)
			id := buildTushareNewsArticleID("nm", uniqueKey)
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			resolvedSource := normalizeUTF8Text(rowSource)
			if resolvedSource == "" {
				resolvedSource = detailSource
			}
			if rowSource != "" {
				contentText = fmt.Sprintf("来源：%s\n\n%s", rowSource, contentText)
			}
			items = append(items, tushareNewsArticlePayload{
				ID:          id,
				CategoryID:  target.CategoryID,
				Visibility:  target.Visibility,
				SyncType:    tushareSyncTypeNewsMajor,
				Source:      resolvedSource,
				Title:       title,
				Summary:     buildTushareNewsSummary(contentText, title),
				Content:     contentText,
				PublishedAt: publishedAt,
				AuthorID:    runtime.AuthorID,
			})
		}
		fetchedCount := len(items) - beforeCount
		detail.Status = "SUCCESS"
		detail.FetchedCount = fetchedCount
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		details = append(details, detail)
	}
	if len(items) == 0 && len(errs) > 0 {
		return nil, details, errors.New(strings.Join(compactErrorMessages(errs, 3), " | "))
	}
	return items, details, nil
}

func fetchTushareResearchReportArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startDate := now.AddDate(0, 0, -runtime.ResearchLookbackDays).Format("20060102")
	endDate := now.Format("20060102")
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	callStartedAt := time.Now()
	parsed, err := callTushareAPI(client, token, "research_report", map[string]string{
		"start_date": startDate,
		"end_date":   endDate,
	}, "ts_code,name,report_date,title,report_type,classify,org_name,author_name,url,summary")
	if err != nil && shouldRetryTushareWithoutFields(err) {
		parsed, err = callTushareAPI(client, token, "research_report", map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		}, "")
	}
	if err != nil {
		detail := newTushareSyncDetailDraft(tushareSyncTypeResearch, "research_report", "")
		detail.Status = "FAILED"
		detail.FailedCount = 1
		detail.ErrorText = truncateByRunes(err.Error(), 512)
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		return nil, []model.NewsSyncRunDetail{detail}, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	items := make([]tushareNewsArticlePayload, 0, len(parsed.Data.Items))
	groupCount := make(map[string]int)
	groupDetail := make(map[string]model.NewsSyncRunDetail)
	seen := make(map[string]struct{})
	for _, row := range parsed.Data.Items {
		tsCode := firstTushareString(row, fieldIndex, "ts_code")
		name := firstTushareString(row, fieldIndex, "name")
		reportDateText := firstTushareString(row, fieldIndex, "report_date")
		title := firstTushareString(row, fieldIndex, "title")
		reportType := firstTushareString(row, fieldIndex, "report_type")
		classify := firstTushareString(row, fieldIndex, "classify")
		orgName := firstTushareString(row, fieldIndex, "org_name", "org")
		authorName := firstTushareString(row, fieldIndex, "author_name", "author")
		sourceURL := firstTushareString(row, fieldIndex, "url")
		summary := firstTushareString(row, fieldIndex, "summary", "abstr")
		resolvedSource := normalizeUTF8Text(orgName)
		resolvedSymbol := normalizeTushareSymbol(tsCode)
		if !matchTushareSourceFilter(opts.SourceSet, resolvedSource) {
			continue
		}
		if !matchTushareSymbolFilter(opts.SymbolSet, resolvedSymbol) {
			continue
		}
		title = normalizeUTF8Text(title)
		if title == "" {
			fallbackName := strings.TrimSpace(name)
			if fallbackName == "" {
				fallbackName = strings.TrimSpace(tsCode)
			}
			title = strings.TrimSpace(fallbackName + " 券商研报")
		}
		if title == "" {
			continue
		}
		publishedAt := parseTushareDateTimeOr(reportDateText, now)
		uniqueKey := fmt.Sprintf("research|%s|%s|%s|%s|%s", reportDateText, tsCode, title, orgName, sourceURL)
		id := buildTushareNewsArticleID("rr", uniqueKey)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		detailKey := buildNewsSyncDetailKey(tushareSyncTypeResearch, resolvedSource, resolvedSymbol)
		groupCount[detailKey]++
		if _, exists := groupDetail[detailKey]; !exists {
			groupDetail[detailKey] = newTushareSyncDetailDraft(tushareSyncTypeResearch, resolvedSource, resolvedSymbol)
		}
		contentBuilder := make([]string, 0, 8)
		contentBuilder = append(contentBuilder, fmt.Sprintf("股票：%s %s", strings.TrimSpace(tsCode), strings.TrimSpace(name)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("机构：%s", strings.TrimSpace(orgName)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("作者：%s", strings.TrimSpace(authorName)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("报告类型：%s", strings.TrimSpace(reportType)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("分类：%s", strings.TrimSpace(classify)))
		if strings.TrimSpace(summary) != "" {
			contentBuilder = append(contentBuilder, "", "摘要：", strings.TrimSpace(summary))
		}
		if strings.TrimSpace(sourceURL) != "" {
			contentBuilder = append(contentBuilder, "", "原文链接：", strings.TrimSpace(sourceURL))
		}
		contentText := strings.Join(contentBuilder, "\n")
		items = append(items, tushareNewsArticlePayload{
			ID:             id,
			CategoryID:     target.CategoryID,
			Visibility:     target.Visibility,
			SyncType:       tushareSyncTypeResearch,
			Source:         resolvedSource,
			Symbol:         resolvedSymbol,
			Title:          title,
			Summary:        buildTushareNewsSummary(summary, title),
			Content:        contentText,
			PublishedAt:    publishedAt,
			AuthorID:       runtime.AuthorID,
			AttachmentURL:  strings.TrimSpace(sourceURL),
			AttachmentName: strings.TrimSpace(title) + ".pdf",
			AttachmentMime: "application/pdf",
		})
	}
	finishedAt := time.Now()
	details := make([]model.NewsSyncRunDetail, 0, len(groupDetail))
	for detailKey, draft := range groupDetail {
		draft.Status = "SUCCESS"
		draft.FetchedCount = groupCount[detailKey]
		draft.StartedAt = callStartedAt.Format(time.RFC3339)
		draft.FinishedAt = finishedAt.Format(time.RFC3339)
		details = append(details, draft)
	}
	return items, details, nil
}

func fetchTushareForecastArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startDate := now.AddDate(0, 0, -runtime.ForecastLookbackDays).Format("20060102")
	endDate := now.Format("20060102")
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	callStartedAt := time.Now()
	parsed, err := callTushareAPI(client, token, "report_rc", map[string]string{
		"start_date": startDate,
		"end_date":   endDate,
	}, "report_date,ts_code,name,report_title,org_name,author_name,quarter,op_rt,eps,pe,max_price,min_price,create_time")
	if err != nil && shouldRetryTushareWithoutFields(err) {
		parsed, err = callTushareAPI(client, token, "report_rc", map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		}, "")
	}
	if err != nil {
		detail := newTushareSyncDetailDraft(tushareSyncTypeForecast, "report_rc", "")
		detail.Status = "FAILED"
		detail.FailedCount = 1
		detail.ErrorText = truncateByRunes(err.Error(), 512)
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		return nil, []model.NewsSyncRunDetail{detail}, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	items := make([]tushareNewsArticlePayload, 0, len(parsed.Data.Items))
	groupCount := make(map[string]int)
	groupDetail := make(map[string]model.NewsSyncRunDetail)
	seen := make(map[string]struct{})
	for _, row := range parsed.Data.Items {
		reportDateText := firstTushareString(row, fieldIndex, "report_date")
		tsCode := firstTushareString(row, fieldIndex, "ts_code")
		name := firstTushareString(row, fieldIndex, "name")
		reportTitle := firstTushareString(row, fieldIndex, "report_title", "title")
		orgName := firstTushareString(row, fieldIndex, "org_name")
		authorName := firstTushareString(row, fieldIndex, "author_name")
		quarter := firstTushareString(row, fieldIndex, "quarter")
		opRT := firstTushareString(row, fieldIndex, "op_rt")
		eps := firstTushareString(row, fieldIndex, "eps")
		pe := firstTushareString(row, fieldIndex, "pe")
		maxPrice := firstTushareString(row, fieldIndex, "max_price")
		minPrice := firstTushareString(row, fieldIndex, "min_price")
		createTime := firstTushareString(row, fieldIndex, "create_time")
		resolvedSource := normalizeUTF8Text(orgName)
		resolvedSymbol := normalizeTushareSymbol(tsCode)
		if !matchTushareSourceFilter(opts.SourceSet, resolvedSource) {
			continue
		}
		if !matchTushareSymbolFilter(opts.SymbolSet, resolvedSymbol) {
			continue
		}

		reportTitle = normalizeUTF8Text(reportTitle)
		if reportTitle == "" {
			fallbackName := strings.TrimSpace(name)
			if fallbackName == "" {
				fallbackName = strings.TrimSpace(tsCode)
			}
			reportTitle = strings.TrimSpace(fallbackName + " 盈利预测")
		}
		if reportTitle == "" {
			continue
		}
		publishedAt := parseTushareDateTimeOr(createTime, parseTushareDateTimeOr(reportDateText, now))
		uniqueKey := fmt.Sprintf("forecast|%s|%s|%s|%s|%s", reportDateText, tsCode, orgName, authorName, reportTitle)
		id := buildTushareNewsArticleID("rc", uniqueKey)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		detailKey := buildNewsSyncDetailKey(tushareSyncTypeForecast, resolvedSource, resolvedSymbol)
		groupCount[detailKey]++
		if _, exists := groupDetail[detailKey]; !exists {
			groupDetail[detailKey] = newTushareSyncDetailDraft(tushareSyncTypeForecast, resolvedSource, resolvedSymbol)
		}
		summaryParts := make([]string, 0, 4)
		if strings.TrimSpace(orgName) != "" {
			summaryParts = append(summaryParts, strings.TrimSpace(orgName))
		}
		if strings.TrimSpace(opRT) != "" {
			summaryParts = append(summaryParts, "评级:"+strings.TrimSpace(opRT))
		}
		if strings.TrimSpace(eps) != "" {
			summaryParts = append(summaryParts, "EPS:"+strings.TrimSpace(eps))
		}
		if strings.TrimSpace(pe) != "" {
			summaryParts = append(summaryParts, "PE:"+strings.TrimSpace(pe))
		}
		summary := strings.Join(summaryParts, " ")
		contentBuilder := []string{
			fmt.Sprintf("股票：%s %s", strings.TrimSpace(tsCode), strings.TrimSpace(name)),
			fmt.Sprintf("机构：%s", strings.TrimSpace(orgName)),
			fmt.Sprintf("作者：%s", strings.TrimSpace(authorName)),
			fmt.Sprintf("季度：%s", strings.TrimSpace(quarter)),
			fmt.Sprintf("评级：%s", strings.TrimSpace(opRT)),
			fmt.Sprintf("EPS：%s", strings.TrimSpace(eps)),
			fmt.Sprintf("PE：%s", strings.TrimSpace(pe)),
			fmt.Sprintf("目标价区间：%s - %s", strings.TrimSpace(minPrice), strings.TrimSpace(maxPrice)),
		}
		contentText := strings.Join(contentBuilder, "\n")
		items = append(items, tushareNewsArticlePayload{
			ID:          id,
			CategoryID:  target.CategoryID,
			Visibility:  target.Visibility,
			SyncType:    tushareSyncTypeForecast,
			Source:      resolvedSource,
			Symbol:      resolvedSymbol,
			Title:       reportTitle,
			Summary:     buildTushareNewsSummary(summary, reportTitle),
			Content:     contentText,
			PublishedAt: publishedAt,
			AuthorID:    runtime.AuthorID,
		})
	}
	finishedAt := time.Now()
	details := make([]model.NewsSyncRunDetail, 0, len(groupDetail))
	for detailKey, draft := range groupDetail {
		draft.Status = "SUCCESS"
		draft.FetchedCount = groupCount[detailKey]
		draft.StartedAt = callStartedAt.Format(time.RFC3339)
		draft.FinishedAt = finishedAt.Format(time.RFC3339)
		details = append(details, draft)
	}
	return items, details, nil
}

func fetchTushareAnnouncementArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startDate := now.AddDate(0, 0, -runtime.AnnouncementLookbackDays).Format("20060102")
	endDate := now.Format("20060102")
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	callStartedAt := time.Now()
	parsed, err := callTushareAPI(client, token, "anns_d", map[string]string{
		"start_date": startDate,
		"end_date":   endDate,
		"limit":      strconv.Itoa(maxInt(50, minInt(runtime.BatchSize, 2000))),
	}, "ann_date,ts_code,name,title,url,rec_time")
	if err != nil && shouldRetryTushareWithoutFields(err) {
		parsed, err = callTushareAPI(client, token, "anns_d", map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
			"limit":      strconv.Itoa(maxInt(50, minInt(runtime.BatchSize, 2000))),
		}, "")
	}
	if err != nil {
		detail := newTushareSyncDetailDraft(tushareSyncTypeAnnouncement, "anns_d", "")
		detail.Status = "FAILED"
		detail.FailedCount = 1
		detail.ErrorText = truncateByRunes(err.Error(), 512)
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		return nil, []model.NewsSyncRunDetail{detail}, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	items := make([]tushareNewsArticlePayload, 0, len(parsed.Data.Items))
	groupCount := make(map[string]int)
	groupDetail := make(map[string]model.NewsSyncRunDetail)
	seen := make(map[string]struct{})
	for _, row := range parsed.Data.Items {
		annDate := firstTushareString(row, fieldIndex, "ann_date")
		tsCode := firstTushareString(row, fieldIndex, "ts_code")
		name := firstTushareString(row, fieldIndex, "name")
		title := firstTushareString(row, fieldIndex, "title")
		sourceURL := firstTushareString(row, fieldIndex, "url")
		recTime := firstTushareString(row, fieldIndex, "rec_time")
		resolvedSource := "交易所公告"
		resolvedSymbol := normalizeTushareSymbol(tsCode)
		if !matchTushareSourceFilter(opts.SourceSet, resolvedSource) {
			continue
		}
		if !matchTushareSymbolFilter(opts.SymbolSet, resolvedSymbol) {
			continue
		}
		title = normalizeUTF8Text(title)
		if title == "" {
			title = strings.TrimSpace(tsCode + " 上市公司公告")
		}
		if title == "" {
			continue
		}
		publishedAt := parseTushareDateTimeOr(recTime, parseTushareDateTimeOr(annDate, now))
		uniqueKey := fmt.Sprintf("announcement|%s|%s|%s|%s", annDate, tsCode, title, sourceURL)
		id := buildTushareNewsArticleID("an", uniqueKey)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		detailKey := buildNewsSyncDetailKey(tushareSyncTypeAnnouncement, resolvedSource, resolvedSymbol)
		groupCount[detailKey]++
		if _, exists := groupDetail[detailKey]; !exists {
			groupDetail[detailKey] = newTushareSyncDetailDraft(tushareSyncTypeAnnouncement, resolvedSource, resolvedSymbol)
		}
		contentBuilder := []string{
			fmt.Sprintf("股票：%s %s", strings.TrimSpace(tsCode), strings.TrimSpace(name)),
			fmt.Sprintf("公告日期：%s", strings.TrimSpace(annDate)),
			fmt.Sprintf("公告标题：%s", title),
		}
		if strings.TrimSpace(sourceURL) != "" {
			contentBuilder = append(contentBuilder, "公告链接："+strings.TrimSpace(sourceURL))
		}
		contentText := strings.Join(contentBuilder, "\n")
		items = append(items, tushareNewsArticlePayload{
			ID:             id,
			CategoryID:     target.CategoryID,
			Visibility:     target.Visibility,
			SyncType:       tushareSyncTypeAnnouncement,
			Source:         resolvedSource,
			Symbol:         resolvedSymbol,
			Title:          title,
			Summary:        buildTushareNewsSummary(contentText, title),
			Content:        contentText,
			PublishedAt:    publishedAt,
			AuthorID:       runtime.AuthorID,
			AttachmentURL:  strings.TrimSpace(sourceURL),
			AttachmentName: title + ".html",
			AttachmentMime: "text/html",
		})
	}
	finishedAt := time.Now()
	details := make([]model.NewsSyncRunDetail, 0, len(groupDetail))
	for detailKey, draft := range groupDetail {
		draft.Status = "SUCCESS"
		draft.FetchedCount = groupCount[detailKey]
		draft.StartedAt = callStartedAt.Format(time.RFC3339)
		draft.FinishedAt = finishedAt.Format(time.RFC3339)
		details = append(details, draft)
	}
	return items, details, nil
}

func buildTushareFieldIndex(fields []string) map[string]int {
	fieldIndex := make(map[string]int, len(fields))
	for index, field := range fields {
		key := strings.ToLower(strings.TrimSpace(field))
		if key == "" {
			continue
		}
		fieldIndex[key] = index
	}
	return fieldIndex
}

func firstTushareString(row []interface{}, fieldIndex map[string]int, keys ...string) string {
	for _, key := range keys {
		lookupKey := strings.ToLower(strings.TrimSpace(key))
		index, ok := fieldIndex[lookupKey]
		if !ok || index < 0 || index >= len(row) {
			continue
		}
		text := strings.TrimSpace(fmt.Sprintf("%v", row[index]))
		if text == "" || text == "<nil>" {
			continue
		}
		return text
	}
	return ""
}

func parseTushareDateTimeOr(raw string, fallback time.Time) time.Time {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"20060102150405",
		"2006-01-02",
		"20060102",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return parsed
		}
	}
	return fallback
}

func parseDateTimeOrDefault(raw string, fallback time.Time) time.Time {
	return parseTushareDateTimeOr(raw, fallback)
}

func buildTushareNewsSummary(content string, title string) string {
	text := normalizeUTF8Text(content)
	if text == "" {
		return truncateByRunes(normalizeUTF8Text(title), 200)
	}
	return truncateByRunes(text, 200)
}

func splitTushareSourceList(raw string, defaults []string) []string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return append([]string(nil), defaults...)
	}
	normalized := strings.NewReplacer("，", ",", "；", ",", ";", ",", "\n", ",", "\t", ",").Replace(text)
	parts := strings.Split(normalized, ",")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		lookupKey := strings.ToLower(item)
		if _, exists := seen[lookupKey]; exists {
			continue
		}
		seen[lookupKey] = struct{}{}
		result = append(result, item)
	}
	if len(result) == 0 {
		return append([]string(nil), defaults...)
	}
	return result
}

func splitTushareSyncTypeList(raw string, defaults []string) []string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return append([]string(nil), defaults...)
	}
	normalized := strings.NewReplacer("，", ",", "；", ",", ";", ",", "\n", ",", "\t", ",").Replace(text)
	parts := strings.Split(normalized, ",")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		syncType := normalizeTushareSyncType(part)
		if syncType == "" {
			continue
		}
		if _, exists := seen[syncType]; exists {
			continue
		}
		seen[syncType] = struct{}{}
		result = append(result, syncType)
	}
	if len(result) == 0 {
		return append([]string(nil), defaults...)
	}
	return result
}

func normalizeTushareSyncOptions(opts model.TushareNewsSyncOptions) normalizedTushareSyncOptions {
	normalized := normalizedTushareSyncOptions{
		BatchSize:   opts.BatchSize,
		SourceSet:   make(map[string]struct{}),
		SymbolSet:   make(map[string]struct{}),
		SyncTypeSet: make(map[string]struct{}),
	}
	if normalized.BatchSize < 0 {
		normalized.BatchSize = 0
	}
	for _, source := range opts.Sources {
		key := normalizeTushareSourceKey(source)
		if key == "" {
			continue
		}
		normalized.SourceSet[key] = struct{}{}
	}
	for _, symbol := range opts.Symbols {
		key := normalizeTushareSymbol(symbol)
		if key == "" {
			continue
		}
		normalized.SymbolSet[key] = struct{}{}
	}
	for _, syncType := range opts.SyncTypes {
		key := normalizeTushareSyncType(syncType)
		if key == "" {
			continue
		}
		normalized.SyncTypeSet[key] = struct{}{}
	}
	return normalized
}

func normalizeTushareSyncType(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	switch normalized {
	case "NEWS", "NEWS_BRIEF", "BRIEF":
		return tushareSyncTypeNewsBrief
	case "MAJOR", "NEWS_MAJOR", "MAJOR_NEWS":
		return tushareSyncTypeNewsMajor
	case "RESEARCH", "RESEARCH_REPORT":
		return tushareSyncTypeResearch
	case "FORECAST", "REPORT_RC":
		return tushareSyncTypeForecast
	case "ANN", "ANNS", "ANNOUNCEMENT", "ANNS_D":
		return tushareSyncTypeAnnouncement
	default:
		return ""
	}
}

func shouldSyncTushareType(syncTypeSet map[string]struct{}, targetType string) bool {
	if len(syncTypeSet) == 0 {
		return true
	}
	_, ok := syncTypeSet[targetType]
	return ok
}

func filterTushareSources(allSources []string, sourceSet map[string]struct{}) []string {
	if len(allSources) == 0 {
		return nil
	}
	if len(sourceSet) == 0 {
		return append([]string(nil), allSources...)
	}
	result := make([]string, 0, len(allSources))
	seen := make(map[string]struct{})
	for _, source := range allSources {
		normalized := normalizeTushareSourceKey(source)
		if normalized == "" {
			continue
		}
		if _, ok := sourceSet[normalized]; !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, source)
	}
	return result
}

func normalizeTushareSourceKey(raw string) string {
	return strings.ToLower(normalizeUTF8Text(raw))
}

func normalizeTushareSymbol(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

func matchTushareSourceFilter(sourceSet map[string]struct{}, source string) bool {
	if len(sourceSet) == 0 {
		return true
	}
	_, ok := sourceSet[normalizeTushareSourceKey(source)]
	return ok
}

func matchTushareSymbolFilter(symbolSet map[string]struct{}, symbol string) bool {
	if len(symbolSet) == 0 {
		return true
	}
	_, ok := symbolSet[normalizeTushareSymbol(symbol)]
	return ok
}

func newTushareSyncDetailDraft(syncType string, source string, symbol string) model.NewsSyncRunDetail {
	return model.NewsSyncRunDetail{
		JobName:   tushareNewsIncrementalJobName,
		SyncType:  strings.ToUpper(strings.TrimSpace(syncType)),
		Source:    truncateByRunes(normalizeUTF8Text(source), 128),
		Symbol:    truncateByRunes(normalizeTushareSymbol(symbol), 32),
		Status:    "SUCCESS",
		StartedAt: time.Now().Format(time.RFC3339),
	}
}

func buildNewsSyncDetailKey(syncType string, source string, symbol string) string {
	return strings.ToUpper(strings.TrimSpace(syncType)) + "|" +
		normalizeTushareSourceKey(source) + "|" +
		normalizeTushareSymbol(symbol)
}

func applyTushareDetailUpsertCounts(details []model.NewsSyncRunDetail, items []tushareNewsArticlePayload) {
	if len(details) == 0 || len(items) == 0 {
		return
	}
	groupCounts := make(map[string]int, len(items))
	for _, item := range items {
		key := buildNewsSyncDetailKey(item.SyncType, item.Source, item.Symbol)
		groupCounts[key]++
	}
	for idx := range details {
		key := buildNewsSyncDetailKey(details[idx].SyncType, details[idx].Source, details[idx].Symbol)
		if strings.EqualFold(details[idx].Status, "FAILED") {
			if details[idx].FailedCount <= 0 {
				details[idx].FailedCount = 1
			}
			continue
		}
		if upserted, ok := groupCounts[key]; ok {
			details[idx].UpsertedCount = upserted
			if details[idx].FetchedCount <= 0 {
				details[idx].FetchedCount = upserted
			}
		}
		if details[idx].FailedCount < 0 {
			details[idx].FailedCount = 0
		}
		if details[idx].Status == "" {
			details[idx].Status = "SUCCESS"
		}
	}
}

func normalizeTushareNewsVisibility(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	switch value {
	case "PUBLIC", "VIP":
		return value
	default:
		return ""
	}
}

func resolveTushareTargetVisibility(preferred string, fallback string) string {
	if value := normalizeTushareNewsVisibility(preferred); value != "" {
		return value
	}
	if value := normalizeTushareNewsVisibility(fallback); value != "" {
		return value
	}
	return "PUBLIC"
}

func buildTushareNewsArticleID(kind string, uniqueKey string) string {
	return buildTushareStableID("na_ts_"+strings.ToLower(strings.TrimSpace(kind)), uniqueKey, 32)
}

func buildTushareNewsAttachmentID(kind string, uniqueKey string) string {
	return buildTushareStableID("nat_ts_"+strings.ToLower(strings.TrimSpace(kind)), uniqueKey, 32)
}

func buildTushareStableID(prefix string, uniqueKey string, maxLen int) string {
	normalizedPrefix := strings.ToLower(strings.TrimSpace(prefix))
	if normalizedPrefix == "" {
		normalizedPrefix = "ts"
	}
	if maxLen <= 0 {
		maxLen = 32
	}
	if len(normalizedPrefix) >= maxLen-1 {
		normalizedPrefix = normalizedPrefix[:maxLen-1]
	}
	hash := sha1.Sum([]byte(strings.TrimSpace(uniqueKey)))
	hashText := hex.EncodeToString(hash[:])
	available := maxLen - len(normalizedPrefix) - 1
	if available <= 0 {
		return normalizedPrefix
	}
	if available > len(hashText) {
		available = len(hashText)
	}
	return normalizedPrefix + "_" + hashText[:available]
}

func inferTushareAttachmentFileName(fileURL string, preferredName string, title string) string {
	if preferred := strings.TrimSpace(preferredName); preferred != "" {
		return truncateByRunes(preferred, 256)
	}
	baseName := strings.TrimSpace(path.Base(strings.TrimSpace(fileURL)))
	if baseName != "" && baseName != "." && baseName != "/" {
		return truncateByRunes(baseName, 256)
	}
	cleanTitle := normalizeUTF8Text(title)
	if cleanTitle == "" {
		cleanTitle = "attachment"
	}
	return truncateByRunes(cleanTitle, 256)
}

func inferAttachmentMimeTypeFromFileName(fileName string) string {
	ext := strings.ToLower(strings.TrimSpace(path.Ext(fileName)))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".zip":
		return "application/zip"
	default:
		return ""
	}
}

func isTushareRateLimitErrorText(message string) bool {
	text := strings.ToLower(normalizeUTF8Text(message))
	if text == "" {
		return false
	}
	keywords := []string{
		"最多访问该接口",
		"触发访问频次",
		"请求过于频繁",
		"rate limit",
		"too many requests",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	if strings.Contains(text, "每分钟") && strings.Contains(text, "访问") {
		return true
	}
	if strings.Contains(text, "每小时") && strings.Contains(text, "访问") {
		return true
	}
	if strings.Contains(text, "每天") && strings.Contains(text, "访问") {
		return true
	}
	return false
}

func shouldRetryTushareWithoutFields(err error) bool {
	if err == nil {
		return false
	}
	return !isTushareRateLimitErrorText(err.Error())
}

func onlyRateLimitMessages(messages []string) bool {
	hasMessage := false
	for _, message := range messages {
		text := strings.TrimSpace(message)
		if text == "" {
			continue
		}
		hasMessage = true
		if !isTushareRateLimitErrorText(text) {
			return false
		}
	}
	return hasMessage
}

func compactErrorMessages(messages []string, limit int) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(messages))
	for _, message := range messages {
		text := truncateByRunes(normalizeUTF8Text(message), 160)
		if text == "" {
			continue
		}
		if _, exists := seen[text]; exists {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}
