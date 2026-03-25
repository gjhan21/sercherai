CREATE TABLE IF NOT EXISTS stock_event_clusters (
  id               varchar(32) PRIMARY KEY,
  cluster_key      varchar(128) NOT NULL,
  event_type       varchar(32) NOT NULL,
  title            varchar(255) NOT NULL,
  summary          text,
  source           varchar(64),
  primary_symbol   varchar(64),
  sector_label     varchar(128),
  topic_label      varchar(128),
  cluster_status   varchar(32) NOT NULL DEFAULT 'RAW',
  review_status    varchar(32) NOT NULL DEFAULT 'PENDING',
  news_count       int NOT NULL DEFAULT 0,
  confidence_score decimal(8,4) NOT NULL DEFAULT 0,
  metadata_json    longtext,
  published_at     datetime DEFAULT NULL,
  created_at       datetime NOT NULL,
  updated_at       datetime NOT NULL,
  UNIQUE KEY uk_stock_event_clusters_key (cluster_key),
  INDEX idx_stock_event_clusters_lookup (event_type, review_status, updated_at),
  INDEX idx_stock_event_clusters_symbol (primary_symbol, updated_at),
  INDEX idx_stock_event_clusters_topic (topic_label, updated_at)
);

CREATE TABLE IF NOT EXISTS stock_event_items (
  id             varchar(32) PRIMARY KEY,
  cluster_id     varchar(32) NOT NULL,
  source_key     varchar(64) NOT NULL,
  source_item_id varchar(128) NOT NULL,
  title          varchar(255) NOT NULL,
  summary        text,
  primary_symbol varchar(64),
  symbols_json   text,
  metadata_json  longtext,
  published_at   datetime DEFAULT NULL,
  created_at     datetime NOT NULL,
  updated_at     datetime NOT NULL,
  UNIQUE KEY uk_stock_event_items_source (cluster_id, source_key, source_item_id),
  INDEX idx_stock_event_items_cluster (cluster_id, published_at)
);

CREATE TABLE IF NOT EXISTS stock_event_entities (
  id               varchar(32) PRIMARY KEY,
  cluster_id       varchar(32) NOT NULL,
  entity_type      varchar(32) NOT NULL,
  entity_key       varchar(128) NOT NULL,
  label            varchar(255) NOT NULL,
  symbol           varchar(64),
  sector_label     varchar(128),
  topic_label      varchar(128),
  confidence_score decimal(8,4) NOT NULL DEFAULT 0,
  metadata_json    longtext,
  created_at       datetime NOT NULL,
  updated_at       datetime NOT NULL,
  UNIQUE KEY uk_stock_event_entities_key (cluster_id, entity_type, entity_key),
  INDEX idx_stock_event_entities_cluster (cluster_id, entity_type),
  INDEX idx_stock_event_entities_symbol (symbol)
);

CREATE TABLE IF NOT EXISTS stock_event_edges (
  id                varchar(32) PRIMARY KEY,
  cluster_id        varchar(32) NOT NULL,
  subject_entity_id varchar(32) NOT NULL,
  object_entity_id  varchar(32) NOT NULL,
  relation_type     varchar(64) NOT NULL,
  relation_weight   decimal(8,4) NOT NULL DEFAULT 0,
  metadata_json     longtext,
  created_at        datetime NOT NULL,
  updated_at        datetime NOT NULL,
  UNIQUE KEY uk_stock_event_edges_pair (cluster_id, subject_entity_id, object_entity_id, relation_type),
  INDEX idx_stock_event_edges_cluster (cluster_id, relation_type)
);

CREATE TABLE IF NOT EXISTS stock_event_reviews (
  id            varchar(32) PRIMARY KEY,
  cluster_id    varchar(32) NOT NULL,
  review_status varchar(32) NOT NULL,
  reviewer      varchar(64),
  review_note   text,
  metadata_json longtext,
  created_at    datetime NOT NULL,
  updated_at    datetime NOT NULL,
  INDEX idx_stock_event_reviews_cluster (cluster_id, created_at),
  INDEX idx_stock_event_reviews_status (review_status, created_at)
);
