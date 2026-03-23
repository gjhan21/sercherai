package repo

import (
	"database/sql"
	"strings"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminGetStrategyGraphSnapshot(snapshotID string) (model.StrategyGraphSnapshot, error) {
	if r.strategyGraph == nil {
		return model.StrategyGraphSnapshot{}, sql.ErrNoRows
	}
	return r.strategyGraph.getSnapshot(strings.TrimSpace(snapshotID))
}

func (r *MySQLGrowthRepo) AdminQueryStrategyGraphSubgraph(query model.StrategyGraphSubgraphQuery) (model.StrategyGraphSubgraph, error) {
	if r.strategyGraph == nil {
		return model.StrategyGraphSubgraph{}, sql.ErrNoRows
	}
	return r.strategyGraph.querySubgraph(query)
}
