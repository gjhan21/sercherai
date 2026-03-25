package repo

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/config"
)

type strategyGraphClient struct {
	baseURL    string
	httpClient *http.Client
}

func newStrategyGraphClient(cfg config.Config) *strategyGraphClient {
	baseURL := strings.TrimSpace(cfg.StrategyGraphBaseURL)
	if baseURL == "" {
		return nil
	}
	timeoutMS := cfg.StrategyGraphTimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 5000
	}
	return &strategyGraphClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMS) * time.Millisecond,
		},
	}
}

func (c *strategyGraphClient) getSnapshot(snapshotID string) (model.StrategyGraphSnapshot, error) {
	snapshotID = strings.TrimSpace(snapshotID)
	if snapshotID == "" {
		return model.StrategyGraphSnapshot{}, sql.ErrNoRows
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/graph/snapshots/"+url.PathEscape(snapshotID), nil)
	if err != nil {
		return model.StrategyGraphSnapshot{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.StrategyGraphSnapshot{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return model.StrategyGraphSnapshot{}, sql.ErrNoRows
	}
	if resp.StatusCode != http.StatusOK {
		return model.StrategyGraphSnapshot{}, fmt.Errorf("strategy-graph returned %d when fetching snapshot %s: %s", resp.StatusCode, snapshotID, readBodyText(resp.Body))
	}
	var item model.StrategyGraphSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return model.StrategyGraphSnapshot{}, err
	}
	return item, nil
}

func (c *strategyGraphClient) querySubgraph(query model.StrategyGraphSubgraphQuery) (model.StrategyGraphSubgraph, error) {
	params := url.Values{}
	params.Set("entity_type", strings.TrimSpace(query.EntityType))
	params.Set("entity_key", strings.TrimSpace(query.EntityKey))
	depth := query.Depth
	if depth <= 0 {
		depth = 1
	}
	params.Set("depth", fmt.Sprintf("%d", depth))
	if assetDomain := strings.TrimSpace(query.AssetDomain); assetDomain != "" {
		params.Set("asset_domain", assetDomain)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/graph/subgraph?"+params.Encode(), nil)
	if err != nil {
		return model.StrategyGraphSubgraph{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.StrategyGraphSubgraph{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return model.StrategyGraphSubgraph{}, sql.ErrNoRows
	}
	if resp.StatusCode != http.StatusOK {
		return model.StrategyGraphSubgraph{}, fmt.Errorf("strategy-graph returned %d when querying subgraph: %s", resp.StatusCode, readBodyText(resp.Body))
	}
	var item model.StrategyGraphSubgraph
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return model.StrategyGraphSubgraph{}, err
	}
	return item, nil
}

func (c *strategyGraphClient) writeReviewedEvent(payload model.StrategyGraphReviewedEventWriteRequest) (model.StrategyGraphReviewedEventWriteResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return model.StrategyGraphReviewedEventWriteResponse{}, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/internal/v1/graph/reviewed-events", bytes.NewReader(body))
	if err != nil {
		return model.StrategyGraphReviewedEventWriteResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.StrategyGraphReviewedEventWriteResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.StrategyGraphReviewedEventWriteResponse{}, fmt.Errorf("strategy-graph returned %d when writing reviewed event: %s", resp.StatusCode, readBodyText(resp.Body))
	}
	var item model.StrategyGraphReviewedEventWriteResponse
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return model.StrategyGraphReviewedEventWriteResponse{}, err
	}
	return item, nil
}
