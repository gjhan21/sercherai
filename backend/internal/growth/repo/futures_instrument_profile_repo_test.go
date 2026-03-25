package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const futuresInstrumentProfileUpsertPattern = `INSERT INTO futures_instrument_profiles_v2`
const futuresInstrumentProfileSelectPattern = `SELECT asset_class, product_key, commodity_label, exchange_code, contract_chain_json, delivery_places_json, warehouses_json, brands_json, grades_json, inventory_metric_keys_json, metadata_json, source_updated_at, created_at, updated_at FROM futures_instrument_profiles_v2 WHERE asset_class = \? AND product_key = \?`

func TestAdminUpsertAndGetFuturesInstrumentProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	sourceUpdatedAt := time.Date(2026, 3, 24, 9, 30, 0, 0, time.FixedZone("CST", 8*3600))
	profile := model.FuturesInstrumentProfile{
		AssetClass:         "FUTURES",
		ProductKey:         "AU",
		CommodityLabel:     "沪金",
		ExchangeCode:       "SHF",
		ContractChain:      []string{"AU2506.SHF", "AU2508.SHF"},
		DeliveryPlaces:     []string{"上海", "深圳"},
		Warehouses:         []string{"上期所一库", "上期所二库"},
		Brands:             []string{"国标一号"},
		Grades:             []string{"标准品"},
		InventoryMetricKeys: []string{"receipt_volume", "change_volume"},
		Metadata: map[string]any{
			"contract_count": float64(2),
		},
		SourceUpdatedAt: sourceUpdatedAt.Format(time.RFC3339),
	}

	mock.ExpectExec(futuresInstrumentProfileUpsertPattern).
		WithArgs(
			"FUTURES",
			"AU",
			"沪金",
			"SHF",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	stored, err := repo.AdminUpsertFuturesInstrumentProfile(profile)
	if err != nil {
		t.Fatalf("upsert futures instrument profile: %v", err)
	}
	if stored.ProductKey != "AU" || len(stored.ContractChain) != 2 {
		t.Fatalf("unexpected stored profile: %+v", stored)
	}

	mock.ExpectQuery(futuresInstrumentProfileSelectPattern).
		WithArgs("FUTURES", "AU").
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class", "product_key", "commodity_label", "exchange_code", "contract_chain_json", "delivery_places_json",
			"warehouses_json", "brands_json", "grades_json", "inventory_metric_keys_json", "metadata_json",
			"source_updated_at", "created_at", "updated_at",
		}).AddRow(
			"FUTURES", "AU", "沪金", "SHF", `["AU2506.SHF","AU2508.SHF"]`, `["上海","深圳"]`,
			`["上期所一库","上期所二库"]`, `["国标一号"]`, `["标准品"]`, `["receipt_volume","change_volume"]`, `{"contract_count":2}`,
			sourceUpdatedAt, sourceUpdatedAt, sourceUpdatedAt,
		))

	loaded, err := repo.AdminGetFuturesInstrumentProfile("AU")
	if err != nil {
		t.Fatalf("get futures instrument profile: %v", err)
	}
	if loaded.ProductKey != "AU" || loaded.ExchangeCode != "SHF" {
		t.Fatalf("unexpected loaded profile: %+v", loaded)
	}
	if len(loaded.DeliveryPlaces) != 2 || loaded.DeliveryPlaces[0] != "上海" {
		t.Fatalf("unexpected delivery places: %+v", loaded.DeliveryPlaces)
	}
	if len(loaded.InventoryMetricKeys) != 2 {
		t.Fatalf("unexpected inventory metrics: %+v", loaded.InventoryMetricKeys)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
