package repo

import (
	"database/sql"
	"strings"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminListStockSelectionProfileTemplates(status string, page int, pageSize int) ([]model.StockSelectionProfileTemplate, int, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	offset := (page - 1) * pageSize
	args := make([]any, 0, 4)
	filter := ""
	if status != "" {
		filter = " WHERE status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_selection_profile_templates"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
  id,
  template_key,
  name,
  COALESCE(description, ''),
  COALESCE(market_regime_bias, ''),
  is_default,
  status,
  COALESCE(CAST(universe_defaults_json AS CHAR), ''),
  COALESCE(CAST(seed_defaults_json AS CHAR), ''),
  COALESCE(CAST(factor_defaults_json AS CHAR), ''),
  COALESCE(CAST(portfolio_defaults_json AS CHAR), ''),
  COALESCE(CAST(publish_defaults_json AS CHAR), ''),
  COALESCE(updated_by, ''),
  DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
FROM stock_selection_profile_templates`+filter+`
ORDER BY is_default DESC, updated_at DESC, id ASC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StockSelectionProfileTemplate, 0)
	for rows.Next() {
		var item model.StockSelectionProfileTemplate
		var isDefault bool
		var universeJSON, seedJSON, factorJSON, portfolioJSON, publishJSON string
		if err := rows.Scan(
			&item.ID,
			&item.TemplateKey,
			&item.Name,
			&item.Description,
			&item.MarketRegimeBias,
			&isDefault,
			&item.Status,
			&universeJSON,
			&seedJSON,
			&factorJSON,
			&portfolioJSON,
			&publishJSON,
			&item.UpdatedBy,
			&item.UpdatedAt,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		item.IsDefault = isDefault
		item.UniverseDefaults = parseJSONMap(universeJSON)
		item.SeedDefaults = parseJSONMap(seedJSON)
		item.FactorDefaults = parseJSONMap(factorJSON)
		item.PortfolioDefaults = parseJSONMap(portfolioJSON)
		item.PublishDefaults = parseJSONMap(publishJSON)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminCreateStockSelectionProfileTemplate(item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error) {
	item = normalizeStockSelectionProfileTemplate(item)
	item.ID = newID("sstpl")
	tx, err := r.db.Begin()
	if err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if item.IsDefault {
		if _, err = tx.Exec("UPDATE stock_selection_profile_templates SET is_default = 0, updated_at = NOW() WHERE id <> ?", item.ID); err != nil {
			return model.StockSelectionProfileTemplate{}, err
		}
	}
	_, err = tx.Exec(`
INSERT INTO stock_selection_profile_templates (
  id, template_key, name, description, market_regime_bias, is_default, status,
  universe_defaults_json, seed_defaults_json, factor_defaults_json, portfolio_defaults_json, publish_defaults_json,
  updated_by, updated_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		item.ID,
		item.TemplateKey,
		item.Name,
		item.Description,
		item.MarketRegimeBias,
		item.IsDefault,
		item.Status,
		stockSelectionMustJSON(item.UniverseDefaults),
		stockSelectionMustJSON(item.SeedDefaults),
		stockSelectionMustJSON(item.FactorDefaults),
		stockSelectionMustJSON(item.PortfolioDefaults),
		stockSelectionMustJSON(item.PublishDefaults),
		item.UpdatedBy,
	)
	if err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	return r.getStockSelectionProfileTemplate(item.ID)
}

func (r *MySQLGrowthRepo) AdminUpdateStockSelectionProfileTemplate(id string, item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error) {
	current, err := r.getStockSelectionProfileTemplate(id)
	if err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	item = normalizeStockSelectionProfileTemplate(item)
	item.ID = current.ID
	if item.TemplateKey == "" {
		item.TemplateKey = current.TemplateKey
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if item.IsDefault {
		if _, err = tx.Exec("UPDATE stock_selection_profile_templates SET is_default = 0, updated_at = NOW() WHERE id <> ?", item.ID); err != nil {
			return model.StockSelectionProfileTemplate{}, err
		}
	}
	_, err = tx.Exec(`
UPDATE stock_selection_profile_templates
SET template_key = ?,
    name = ?,
    description = ?,
    market_regime_bias = ?,
    is_default = ?,
    status = ?,
    universe_defaults_json = ?,
    seed_defaults_json = ?,
    factor_defaults_json = ?,
    portfolio_defaults_json = ?,
    publish_defaults_json = ?,
    updated_by = ?,
    updated_at = NOW()
WHERE id = ?`,
		item.TemplateKey,
		item.Name,
		item.Description,
		item.MarketRegimeBias,
		item.IsDefault,
		item.Status,
		stockSelectionMustJSON(item.UniverseDefaults),
		stockSelectionMustJSON(item.SeedDefaults),
		stockSelectionMustJSON(item.FactorDefaults),
		stockSelectionMustJSON(item.PortfolioDefaults),
		stockSelectionMustJSON(item.PublishDefaults),
		firstNonEmpty(item.UpdatedBy, current.UpdatedBy),
		item.ID,
	)
	if err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	return r.getStockSelectionProfileTemplate(item.ID)
}

func (r *MySQLGrowthRepo) AdminSetDefaultStockSelectionProfileTemplate(id string, operator string) (model.StockSelectionProfileTemplate, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.StockSelectionProfileTemplate{}, sql.ErrNoRows
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.Exec("UPDATE stock_selection_profile_templates SET is_default = 0, updated_at = NOW() WHERE id <> ?", id); err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	if _, err = tx.Exec("UPDATE stock_selection_profile_templates SET is_default = 1, updated_by = ?, updated_at = NOW() WHERE id = ?", strings.TrimSpace(operator), id); err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	return r.getStockSelectionProfileTemplate(id)
}

func (r *MySQLGrowthRepo) getDefaultStockSelectionProfileTemplate() (*model.StockSelectionProfileTemplate, error) {
	items, _, err := r.AdminListStockSelectionProfileTemplates("ACTIVE", 1, 50)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.IsDefault {
			cloned := item
			return &cloned, nil
		}
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	cloned := items[0]
	return &cloned, nil
}

func (r *MySQLGrowthRepo) getStockSelectionProfileTemplate(id string) (model.StockSelectionProfileTemplate, error) {
	items, _, err := r.AdminListStockSelectionProfileTemplates("", 1, 200)
	if err != nil {
		return model.StockSelectionProfileTemplate{}, err
	}
	for _, item := range items {
		if item.ID == strings.TrimSpace(id) {
			return item, nil
		}
	}
	return model.StockSelectionProfileTemplate{}, sql.ErrNoRows
}

func normalizeStockSelectionProfileTemplate(item model.StockSelectionProfileTemplate) model.StockSelectionProfileTemplate {
	item.TemplateKey = strings.ToUpper(strings.TrimSpace(item.TemplateKey))
	item.Name = strings.TrimSpace(item.Name)
	item.Description = strings.TrimSpace(item.Description)
	item.MarketRegimeBias = strings.ToUpper(strings.TrimSpace(item.MarketRegimeBias))
	item.Status = strings.ToUpper(strings.TrimSpace(item.Status))
	item.UpdatedBy = strings.TrimSpace(item.UpdatedBy)
	if item.Status == "" {
		item.Status = "ACTIVE"
	}
	if item.UniverseDefaults == nil {
		item.UniverseDefaults = map[string]any{}
	}
	if item.SeedDefaults == nil {
		item.SeedDefaults = map[string]any{}
	}
	if item.FactorDefaults == nil {
		item.FactorDefaults = map[string]any{}
	}
	if item.PortfolioDefaults == nil {
		item.PortfolioDefaults = map[string]any{}
	}
	if item.PublishDefaults == nil {
		item.PublishDefaults = map[string]any{}
	}
	return item
}
