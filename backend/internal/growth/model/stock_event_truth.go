package model

type StockEventCluster struct {
	ID            string             `json:"id"`
	ClusterKey    string             `json:"cluster_key"`
	EventType     string             `json:"event_type"`
	Title         string             `json:"title"`
	Summary       string             `json:"summary,omitempty"`
	Source        string             `json:"source,omitempty"`
	PrimarySymbol string             `json:"primary_symbol,omitempty"`
	SectorLabel   string             `json:"sector_label,omitempty"`
	TopicLabel    string             `json:"topic_label,omitempty"`
	Status        string             `json:"status"`
	ReviewStatus  string             `json:"review_status"`
	NewsCount     int                `json:"news_count"`
	Confidence    float64            `json:"confidence"`
	Metadata      map[string]any     `json:"metadata,omitempty"`
	PublishedAt   string             `json:"published_at,omitempty"`
	CreatedAt     string             `json:"created_at,omitempty"`
	UpdatedAt     string             `json:"updated_at,omitempty"`
	Items         []StockEventItem   `json:"items,omitempty"`
	Entities      []StockEventEntity `json:"entities,omitempty"`
	Edges         []StockEventEdge   `json:"edges,omitempty"`
	LatestReview  *StockEventReview  `json:"latest_review,omitempty"`
}

type StockEventItem struct {
	ID            string         `json:"id"`
	ClusterID     string         `json:"cluster_id"`
	SourceKey     string         `json:"source_key"`
	SourceItemID  string         `json:"source_item_id"`
	Title         string         `json:"title"`
	Summary       string         `json:"summary,omitempty"`
	PrimarySymbol string         `json:"primary_symbol,omitempty"`
	Symbols       []string       `json:"symbols,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	PublishedAt   string         `json:"published_at,omitempty"`
	CreatedAt     string         `json:"created_at,omitempty"`
	UpdatedAt     string         `json:"updated_at,omitempty"`
}

type StockEventEntity struct {
	ID          string         `json:"id"`
	ClusterID   string         `json:"cluster_id"`
	EntityType  string         `json:"entity_type"`
	EntityKey   string         `json:"entity_key"`
	Label       string         `json:"label"`
	Symbol      string         `json:"symbol,omitempty"`
	SectorLabel string         `json:"sector_label,omitempty"`
	TopicLabel  string         `json:"topic_label,omitempty"`
	Confidence  float64        `json:"confidence"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   string         `json:"created_at,omitempty"`
	UpdatedAt   string         `json:"updated_at,omitempty"`
}

type StockEventEdge struct {
	ID              string         `json:"id"`
	ClusterID       string         `json:"cluster_id"`
	SubjectEntityID string         `json:"subject_entity_id"`
	ObjectEntityID  string         `json:"object_entity_id"`
	RelationType    string         `json:"relation_type"`
	Weight          float64        `json:"weight"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	CreatedAt       string         `json:"created_at,omitempty"`
	UpdatedAt       string         `json:"updated_at,omitempty"`
}

type StockEventReview struct {
	ID             string         `json:"id"`
	ClusterID      string         `json:"cluster_id"`
	ReviewStatus   string         `json:"review_status"`
	Reviewer       string         `json:"reviewer,omitempty"`
	ReviewNote     string         `json:"review_note,omitempty"`
	ReviewMetadata map[string]any `json:"review_metadata,omitempty"`
	ReviewedAt     string         `json:"reviewed_at,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
}

type StockEventQuery struct {
	EventType      string `json:"event_type,omitempty"`
	Status         string `json:"status,omitempty"`
	ReviewStatus   string `json:"review_status,omitempty"`
	ReviewPriority string `json:"review_priority,omitempty"`
	Symbol         string `json:"symbol,omitempty"`
	Sector         string `json:"sector,omitempty"`
	Topic          string `json:"topic,omitempty"`
	Page           int    `json:"page,omitempty"`
	PageSize       int    `json:"page_size,omitempty"`
}
