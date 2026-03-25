package repo

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/config"
)

func TestStrategyGraphClientGetSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/graph/snapshots/gss_demo_001" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"snapshot_id":"gss_demo_001",
			"run_id":"ssr_demo_001",
			"asset_domain":"stock",
			"trade_date":"2026-03-22",
			"summary":"图谱快照摘要",
			"related_entities":[{"entity_type":"ConceptTheme","entity_key":"ROBOTICS","label":"机器人"}],
			"entities":[{"entity_type":"Stock","entity_key":"SZ300024","label":"机器人龙头"}],
			"relations":[{"relation_type":"BELONGS_TO","source_type":"Stock","source_key":"SZ300024","target_type":"ConceptTheme","target_key":"ROBOTICS","strength":0.82}],
			"meta":{"backend":"test"},
			"created_at":"2026-03-22T00:00:00Z"
		}`))
	}))
	defer server.Close()

	client := newStrategyGraphClient(config.Config{
		StrategyGraphBaseURL:   server.URL,
		StrategyGraphTimeoutMS: 3000,
	})
	if client == nil {
		t.Fatalf("expected strategyGraphClient to be created")
	}

	item, err := client.getSnapshot("gss_demo_001")
	if err != nil {
		t.Fatalf("getSnapshot returned error: %v", err)
	}
	if item.SnapshotID != "gss_demo_001" {
		t.Fatalf("expected snapshot_id=gss_demo_001, got %s", item.SnapshotID)
	}
	if len(item.Entities) != 1 || item.Entities[0].Label != "机器人龙头" {
		t.Fatalf("expected decoded entities, got %#v", item.Entities)
	}
	if len(item.Relations) != 1 || item.Relations[0].RelationType != "BELONGS_TO" {
		t.Fatalf("expected decoded relations, got %#v", item.Relations)
	}
}

func TestStrategyGraphClientQuerySubgraph(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/graph/subgraph" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("entity_type"); got != "ConceptTheme" {
			t.Fatalf("expected entity_type=ConceptTheme, got %s", got)
		}
		if got := r.URL.Query().Get("entity_key"); got != "ROBOTICS" {
			t.Fatalf("expected entity_key=ROBOTICS, got %s", got)
		}
		if got := r.URL.Query().Get("depth"); got != "2" {
			t.Fatalf("expected depth=2, got %s", got)
		}
		if got := r.URL.Query().Get("asset_domain"); got != "stock" {
			t.Fatalf("expected asset_domain=stock, got %s", got)
		}
		_, _ = w.Write([]byte(`{
			"entity":{"entity_type":"ConceptTheme","entity_key":"ROBOTICS","label":"机器人"},
			"entities":[{"entity_type":"ConceptTheme","entity_key":"ROBOTICS","label":"机器人"}],
			"relations":[{"relation_type":"CONFIRMS_SIGNAL","source_type":"ConceptTheme","source_key":"ROBOTICS","target_type":"ConceptTheme","target_key":"AI_COMPUTE","strength":0.77}],
			"matched_snapshot_ids":["gss_demo_001"],
			"backend":"test"
		}`))
	}))
	defer server.Close()

	client := newStrategyGraphClient(config.Config{
		StrategyGraphBaseURL:   server.URL,
		StrategyGraphTimeoutMS: 3000,
	})

	item, err := client.querySubgraph(model.StrategyGraphSubgraphQuery{
		EntityType:  "ConceptTheme",
		EntityKey:   "ROBOTICS",
		Depth:       2,
		AssetDomain: "stock",
	})
	if err != nil {
		t.Fatalf("querySubgraph returned error: %v", err)
	}
	if item.Entity == nil || item.Entity.Label != "机器人" {
		t.Fatalf("expected entity label=机器人, got %#v", item.Entity)
	}
	if len(item.MatchedSnapshotIDs) != 1 || item.MatchedSnapshotIDs[0] != "gss_demo_001" {
		t.Fatalf("expected matched snapshot ids to decode, got %#v", item.MatchedSnapshotIDs)
	}
}

func TestStrategyGraphClientGetSnapshotReturnsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := newStrategyGraphClient(config.Config{
		StrategyGraphBaseURL:   server.URL,
		StrategyGraphTimeoutMS: 3000,
	})

	_, err := client.getSnapshot("gss_missing")
	if err == nil {
		t.Fatalf("expected not found error")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestStrategyGraphClientWriteReviewedEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/graph/reviewed-events" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var payload model.StrategyGraphReviewedEventWriteRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload.ClusterID != "sec_demo_001" || !payload.Approved {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		_, _ = w.Write([]byte(`{
			"snapshot_id":"reviewed-event-sec_demo_001",
			"cluster_id":"sec_demo_001",
			"node_count":3,
			"relation_count":2,
			"backend":"test"
		}`))
	}))
	defer server.Close()

	client := newStrategyGraphClient(config.Config{
		StrategyGraphBaseURL:   server.URL,
		StrategyGraphTimeoutMS: 3000,
	})

	item, err := client.writeReviewedEvent(model.StrategyGraphReviewedEventWriteRequest{
		ClusterID: "sec_demo_001",
		Approved:  true,
		Entities: []model.StrategyGraphEntity{
			{EntityType: "StockEvent", EntityKey: "sec_demo_001", Label: "白酒景气事件"},
		},
	})
	if err != nil {
		t.Fatalf("writeReviewedEvent returned error: %v", err)
	}
	if item.ClusterID != "sec_demo_001" || item.NodeCount != 3 || item.RelationCount != 2 {
		t.Fatalf("unexpected reviewed event response: %+v", item)
	}
}
