package handler

import (
	"errors"
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestIsTushareNewsPermissionDeniedError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "tushare anns_d permission denied",
			err:  errors.New("TUSHARE: tushare error(anns_d): 抱歉，您没有接口访问权限"),
			want: true,
		},
		{
			name: "different tushare error",
			err:  errors.New("tushare error(daily): 抱歉，您每分钟最多访问该接口500次"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isTushareNewsPermissionDeniedError(tc.err)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestMarketSyncResultCount(t *testing.T) {
	if got := marketSyncResultCount(model.MarketSyncResult{TruthCount: 12, BarCount: 24}); got != 12 {
		t.Fatalf("expected truth_count to win, got %d", got)
	}
	if got := marketSyncResultCount(model.MarketSyncResult{BarCount: 24}); got != 24 {
		t.Fatalf("expected bar_count, got %d", got)
	}
	if got := marketSyncResultCount(model.MarketSyncResult{NewsCount: 9}); got != 9 {
		t.Fatalf("expected news_count, got %d", got)
	}
	if got := marketSyncResultCount(model.MarketSyncResult{InventoryCount: 5}); got != 5 {
		t.Fatalf("expected inventory_count, got %d", got)
	}
	if got := marketSyncResultCount(model.MarketSyncResult{}); got != 0 {
		t.Fatalf("expected zero count, got %d", got)
	}
}
