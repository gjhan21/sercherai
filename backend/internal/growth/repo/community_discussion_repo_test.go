package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestCommunityTopicListFiltersPublishedTopics(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	rows, total, err := repo.ListCommunityTopics("", model.CommunityTopicListQuery{
		TopicType: model.CommunityTopicTypeStock,
		Page:      1,
		PageSize:  20,
	})
	if err != nil {
		t.Fatalf("ListCommunityTopics() error = %v", err)
	}
	if total == 0 || len(rows) == 0 {
		t.Fatal("expected at least one published stock topic")
	}
	for _, item := range rows {
		if item.TopicType != string(model.CommunityTopicTypeStock) {
			t.Fatalf("unexpected topic type %q", item.TopicType)
		}
		if item.Status != string(model.CommunityTopicStatusPublished) {
			t.Fatalf("unexpected status %q", item.Status)
		}
	}
}

func TestCommunityTopicListMineFilter(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	rows, total, err := repo.ListCommunityTopics("u_demo_001", model.CommunityTopicListQuery{
		Mine:     "topics",
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListCommunityTopics() error = %v", err)
	}
	if total == 0 || len(rows) == 0 {
		t.Fatal("expected my topic rows")
	}
	for _, item := range rows {
		if item.UserID != "u_demo_001" {
			t.Fatalf("expected only user topics, got %q", item.UserID)
		}
	}
}

func TestListMyCommunityCommentsReturnsOnlyCurrentUserComments(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	rows, total, err := repo.ListMyCommunityComments("u_demo_002", 1, 20)
	if err != nil {
		t.Fatalf("ListMyCommunityComments() error = %v", err)
	}
	if total == 0 || len(rows) == 0 {
		t.Fatal("expected my comment rows")
	}
	for _, item := range rows {
		if item.UserID != "u_demo_002" {
			t.Fatalf("expected only user comments, got %q", item.UserID)
		}
	}
}
