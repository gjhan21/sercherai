package repo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *InMemoryGrowthRepo) AdminCreateAuditEvent(item model.AdminAuditEvent) error {
	item = normalizeAdminAuditEvent(item)
	now := time.Now()
	nowText := now.Format(time.RFC3339)

	var (
		stored     model.AdminAuditEvent
		shouldRoute bool
	)

	r.mu.Lock()
	if strings.TrimSpace(item.DedupeKey) != "" {
		for id, existing := range r.adminAuditEvents {
			if strings.TrimSpace(existing.DedupeKey) != strings.TrimSpace(item.DedupeKey) {
				continue
			}
			if strings.ToUpper(strings.TrimSpace(existing.Status)) != "OPEN" {
				continue
			}
			item.ID = existing.ID
			item.CreatedAt = existing.CreatedAt
			stored = mergeOpenAuditEvent(existing, item, now)
			r.adminAuditEvents[id] = stored
			r.mu.Unlock()
			return nil
		}
	}

	if strings.TrimSpace(item.ID) == "" {
		item.ID = fmt.Sprintf("ae_%03d", len(r.adminAuditEvents)+1)
	}
	if strings.TrimSpace(item.CreatedAt) == "" {
		item.CreatedAt = nowText
	}
	stored = item
	r.adminAuditEvents[item.ID] = stored
	shouldRoute = true
	r.mu.Unlock()

	if shouldRoute {
		message, ok := buildAuditEventWorkflowRoute(stored)
		if ok {
			return r.AdminCreateWorkflowMessage(
				message.ReviewID,
				message.TargetID,
				message.Module,
				message.ReceiverID,
				message.SenderID,
				message.EventType,
				message.Title,
				message.Content,
			)
		}
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminListAuditEvents(filter model.AdminAuditEventFilter, page int, pageSize int) ([]model.AdminAuditEvent, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	filter = normalizeAdminAuditEventFilter(filter)
	items := make([]model.AdminAuditEvent, 0, len(r.adminAuditEvents))
	for _, item := range r.adminAuditEvents {
		if filter.EventDomain != "" && item.EventDomain != filter.EventDomain {
			continue
		}
		if filter.EventType != "" && item.EventType != filter.EventType {
			continue
		}
		if filter.Level != "" && item.Level != filter.Level {
			continue
		}
		if filter.Module != "" && item.Module != filter.Module {
			continue
		}
		if filter.ObjectType != "" && item.ObjectType != filter.ObjectType {
			continue
		}
		if filter.ObjectID != "" && item.ObjectID != filter.ObjectID {
			continue
		}
		if filter.ActorUserID != "" && item.ActorUserID != filter.ActorUserID {
			continue
		}
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt == items[j].CreatedAt {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt > items[j].CreatedAt
	})

	total := len(items)
	if pageSize <= 0 {
		pageSize = total
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []model.AdminAuditEvent{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return append([]model.AdminAuditEvent(nil), items[start:end]...), total, nil
}

func (r *InMemoryGrowthRepo) AdminGetAuditEventSummary() (model.AdminAuditEventSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	summary := model.AdminAuditEventSummary{
		ByDomain: make(map[string]int),
	}
	for _, item := range r.adminAuditEvents {
		summary.TotalCount++
		if strings.ToUpper(strings.TrimSpace(item.Status)) == "OPEN" {
			summary.OpenCount++
		}
		switch strings.ToUpper(strings.TrimSpace(item.Level)) {
		case "CRITICAL":
			summary.CriticalCount++
		case "WARNING":
			summary.WarningCount++
		default:
			summary.InfoCount++
		}
		if item.EventDomain != "" {
			summary.ByDomain[item.EventDomain]++
		}
	}
	return summary, nil
}
