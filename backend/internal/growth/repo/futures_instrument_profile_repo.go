package repo

import (
	"database/sql"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

type futuresInstrumentInventoryProfileRow struct {
	Symbol    string
	Place     string
	Warehouse string
	Brand     string
	Grade     string
}

func (r *MySQLGrowthRepo) AdminUpsertFuturesInstrumentProfile(profile model.FuturesInstrumentProfile) (model.FuturesInstrumentProfile, error) {
	profile = normalizeFuturesInstrumentProfile(profile)
	now := time.Now()
	sourceUpdatedAt := nullableRFC3339Time(profile.SourceUpdatedAt)
	_, err := r.db.Exec(`
INSERT INTO futures_instrument_profiles_v2
  (asset_class, product_key, commodity_label, exchange_code, contract_chain_json, delivery_places_json, warehouses_json, brands_json, grades_json, inventory_metric_keys_json, metadata_json, source_updated_at, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  commodity_label = VALUES(commodity_label),
  exchange_code = VALUES(exchange_code),
  contract_chain_json = VALUES(contract_chain_json),
  delivery_places_json = VALUES(delivery_places_json),
  warehouses_json = VALUES(warehouses_json),
  brands_json = VALUES(brands_json),
  grades_json = VALUES(grades_json),
  inventory_metric_keys_json = VALUES(inventory_metric_keys_json),
  metadata_json = VALUES(metadata_json),
  source_updated_at = VALUES(source_updated_at),
  updated_at = VALUES(updated_at)`,
		profile.AssetClass,
		profile.ProductKey,
		profile.CommodityLabel,
		nullableString(profile.ExchangeCode),
		marshalJSONSilently(profile.ContractChain),
		marshalJSONSilently(profile.DeliveryPlaces),
		marshalJSONSilently(profile.Warehouses),
		marshalJSONSilently(profile.Brands),
		marshalJSONSilently(profile.Grades),
		marshalJSONSilently(profile.InventoryMetricKeys),
		marshalJSONSilently(profile.Metadata),
		sourceUpdatedAt,
		now,
		now,
	)
	if err != nil {
		return model.FuturesInstrumentProfile{}, err
	}
	profile.CreatedAt = now.Format(time.RFC3339)
	profile.UpdatedAt = profile.CreatedAt
	return profile, nil
}

func (r *MySQLGrowthRepo) AdminGetFuturesInstrumentProfile(productKey string) (model.FuturesInstrumentProfile, error) {
	productKey = strings.ToUpper(strings.TrimSpace(productKey))
	if productKey == "" {
		return model.FuturesInstrumentProfile{}, sql.ErrNoRows
	}
	var item model.FuturesInstrumentProfile
	var (
		contractChainJSON       sql.NullString
		deliveryPlacesJSON      sql.NullString
		warehousesJSON          sql.NullString
		brandsJSON              sql.NullString
		gradesJSON              sql.NullString
		inventoryMetricKeysJSON sql.NullString
		metadataJSON            sql.NullString
		sourceUpdatedAt         sql.NullTime
		createdAt               sql.NullTime
		updatedAt               sql.NullTime
	)
	err := r.db.QueryRow(`
SELECT asset_class, product_key, commodity_label, exchange_code, contract_chain_json, delivery_places_json, warehouses_json, brands_json, grades_json, inventory_metric_keys_json, metadata_json, source_updated_at, created_at, updated_at
FROM futures_instrument_profiles_v2
WHERE asset_class = ? AND product_key = ?`,
		marketAssetClassFutures, productKey,
	).Scan(
		&item.AssetClass,
		&item.ProductKey,
		&item.CommodityLabel,
		&item.ExchangeCode,
		&contractChainJSON,
		&deliveryPlacesJSON,
		&warehousesJSON,
		&brandsJSON,
		&gradesJSON,
		&inventoryMetricKeysJSON,
		&metadataJSON,
		&sourceUpdatedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return model.FuturesInstrumentProfile{}, err
	}
	item.ContractChain = parseJSONStringList(contractChainJSON.String)
	item.DeliveryPlaces = parseJSONStringList(deliveryPlacesJSON.String)
	item.Warehouses = parseJSONStringList(warehousesJSON.String)
	item.Brands = parseJSONStringList(brandsJSON.String)
	item.Grades = parseJSONStringList(gradesJSON.String)
	item.InventoryMetricKeys = parseJSONStringList(inventoryMetricKeysJSON.String)
	item.Metadata = parseJSONMap(metadataJSON.String)
	item.SourceUpdatedAt = formatNullTime(sourceUpdatedAt)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return normalizeFuturesInstrumentProfile(item), nil
}

func normalizeFuturesInstrumentProfile(item model.FuturesInstrumentProfile) model.FuturesInstrumentProfile {
	item.AssetClass = strings.ToUpper(strings.TrimSpace(item.AssetClass))
	if item.AssetClass == "" {
		item.AssetClass = marketAssetClassFutures
	}
	item.ProductKey = strings.ToUpper(strings.TrimSpace(item.ProductKey))
	item.CommodityLabel = strings.TrimSpace(item.CommodityLabel)
	item.ExchangeCode = strings.ToUpper(strings.TrimSpace(item.ExchangeCode))
	item.ContractChain = normalizeUpperStringList(item.ContractChain)
	item.DeliveryPlaces = compactStrings(item.DeliveryPlaces)
	item.Warehouses = compactStrings(item.Warehouses)
	item.Brands = compactStrings(item.Brands)
	item.Grades = compactStrings(item.Grades)
	item.InventoryMetricKeys = normalizeUpperStringList(item.InventoryMetricKeys)
	if item.Metadata == nil {
		item.Metadata = map[string]any{}
	}
	return item
}

func buildFuturesInstrumentProfileSnapshot(productKey string, truths []marketInstrumentTruth, inventoryRows []futuresInstrumentInventoryProfileRow) model.FuturesInstrumentProfile {
	productKey = strings.ToUpper(strings.TrimSpace(productKey))
	profile := normalizeFuturesInstrumentProfile(model.FuturesInstrumentProfile{
		AssetClass:          marketAssetClassFutures,
		ProductKey:          productKey,
		InventoryMetricKeys: []string{"receipt_volume", "previous_volume", "change_volume"},
		Metadata:            map[string]any{},
	})
	latestSourceUpdatedAt := time.Time{}
	commodityLabel := ""
	exchangeCode := ""
	contractChain := make([]string, 0, len(truths))
	for _, item := range truths {
		instrumentKey := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		if instrumentKey != "" {
			contractChain = append(contractChain, instrumentKey)
		}
		if commodityLabel == "" {
			commodityLabel = strings.TrimSpace(item.DisplayName)
		}
		if exchangeCode == "" {
			exchangeCode = strings.ToUpper(strings.TrimSpace(item.ExchangeCode))
		}
		if item.SourceUpdatedAt.After(latestSourceUpdatedAt) {
			latestSourceUpdatedAt = item.SourceUpdatedAt
		}
	}
	profile.ContractChain = normalizeUpperStringList(contractChain)
	profile.CommodityLabel = firstNonEmpty(strings.TrimSpace(commodityLabel), productKey)
	profile.ExchangeCode = exchangeCode
	profile.DeliveryPlaces = collectFuturesInstrumentProfileDimension(inventoryRows, func(item futuresInstrumentInventoryProfileRow) string { return item.Place })
	profile.Warehouses = collectFuturesInstrumentProfileDimension(inventoryRows, func(item futuresInstrumentInventoryProfileRow) string { return item.Warehouse })
	profile.Brands = collectFuturesInstrumentProfileDimension(inventoryRows, func(item futuresInstrumentInventoryProfileRow) string { return item.Brand })
	profile.Grades = collectFuturesInstrumentProfileDimension(inventoryRows, func(item futuresInstrumentInventoryProfileRow) string { return item.Grade })
	if !latestSourceUpdatedAt.IsZero() {
		profile.SourceUpdatedAt = latestSourceUpdatedAt.Format(time.RFC3339)
	}
	profile.Metadata["contract_count"] = float64(len(profile.ContractChain))
	profile.Metadata["delivery_place_count"] = float64(len(profile.DeliveryPlaces))
	profile.Metadata["warehouse_count"] = float64(len(profile.Warehouses))
	profile.Metadata["brand_count"] = float64(len(profile.Brands))
	profile.Metadata["grade_count"] = float64(len(profile.Grades))
	return profile
}

func collectFuturesInstrumentProfileDimension(items []futuresInstrumentInventoryProfileRow, pick func(futuresInstrumentInventoryProfileRow) string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		value := strings.TrimSpace(pick(item))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
