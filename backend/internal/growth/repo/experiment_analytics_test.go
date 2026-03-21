package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const (
	experimentOverviewQueryPattern = `(?s)SELECT\s+COUNT\(\*\)\s+AS total_events.*FROM experiment_events\s+WHERE created_at >= \?`
	experimentItemsQueryPattern    = `(?s)SELECT\s+experiment_key,\s+variant_key,\s+page_key,.*FROM experiment_events\s+WHERE created_at >= \?\s+GROUP BY experiment_key, variant_key, page_key, COALESCE`
	experimentPageQueryPattern     = `(?s)SELECT\s+page_key,.*FROM experiment_events\s+WHERE created_at >= \?\s+GROUP BY page_key`
	experimentTrendQueryPattern    = `(?s)SELECT\s+DATE\(created_at\)\s+AS metric_date,.*FROM experiment_events\s+WHERE created_at >= \?\s+GROUP BY DATE\(created_at\)`
	experimentPayQueryPattern      = `(?s)SELECT pay_channel, payment_success_count, renewal_success_count, last_event_at\s+FROM\s+\(\s*SELECT.*FROM experiment_events.*event_type IN \('PAYMENT_SUCCESS', 'RENEWAL_SUCCESS'\).*GROUP BY .*?\)\s+pay_summary`
	experimentDeviceQueryPattern   = `(?s)SELECT\s+experiment_key,\s+variant_key,\s+page_key,\s+device_type,.*FROM\s+\(\s*SELECT.*FROM experiment_events.*GROUP BY experiment_key, variant_key, page_key, COALESCE.*\)\s+device_summary`
	experimentVariantQueryPattern  = `(?s)SELECT\s+DATE\(created_at\)\s+AS metric_date,\s+experiment_key,\s+variant_key,\s+page_key,.*FROM experiment_events\s+WHERE created_at >= \?\s+GROUP BY DATE\(created_at\), experiment_key, variant_key, page_key`
)

func TestAdminGetExperimentAnalyticsSummaryReturnsBreakdowns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.Local)
	day := time.Date(2026, 3, 19, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(experimentOverviewQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_events",
			"total_experiments",
			"exposure_count",
			"click_count",
			"upgrade_intent_count",
			"payment_success_count",
			"renewal_success_count",
			"last_event_at",
		}).AddRow(10, 2, 8, 4, 3, 1, 1, now))

	mock.ExpectQuery(experimentItemsQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"experiment_key",
			"variant_key",
			"page_key",
			"user_stage",
			"exposure_count",
			"click_count",
			"upgrade_intent_count",
			"payment_success_count",
			"renewal_success_count",
			"last_event_at",
		}).AddRow("exp_home", "A", "HOME", "GUEST", 8, 4, 3, 1, 1, now))

	mock.ExpectQuery(experimentPageQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"page_key",
			"exposure_count",
			"click_count",
			"upgrade_intent_count",
			"payment_success_count",
			"renewal_success_count",
			"last_event_at",
		}).AddRow("HOME", 8, 4, 3, 1, 1, now))

	mock.ExpectQuery(experimentTrendQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"metric_date",
			"exposure_count",
			"click_count",
			"upgrade_intent_count",
			"payment_success_count",
			"renewal_success_count",
		}).AddRow(day, 8, 4, 3, 1, 1))

	mock.ExpectQuery(experimentPayQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"pay_channel",
			"payment_success_count",
			"renewal_success_count",
			"last_event_at",
		}).AddRow("ALIPAY", 1, 1, now))

	mock.ExpectQuery(experimentDeviceQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"experiment_key",
			"variant_key",
			"page_key",
			"device_type",
			"exposure_count",
			"click_count",
			"upgrade_intent_count",
			"payment_success_count",
			"renewal_success_count",
			"last_event_at",
		}).AddRow("exp_home", "A", "HOME", "MOBILE", 8, 4, 3, 1, 1, now))

	mock.ExpectQuery(experimentVariantQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"metric_date",
			"experiment_key",
			"variant_key",
			"page_key",
			"device_type",
			"user_stage",
			"exposure_count",
			"click_count",
			"upgrade_intent_count",
			"payment_success_count",
			"renewal_success_count",
		}).AddRow(day, "exp_home", "A", "HOME", "MOBILE", "GUEST", 8, 4, 3, 1, 1))

	summary, err := repo.AdminGetExperimentAnalyticsSummary(7)
	if err != nil {
		t.Fatalf("AdminGetExperimentAnalyticsSummary returned error: %v", err)
	}
	if summary.Overview.TotalEvents != 10 {
		t.Fatalf("expected total events 10, got %d", summary.Overview.TotalEvents)
	}
	if len(summary.Items) != 1 || len(summary.PayChannelBreakdown) != 1 {
		t.Fatalf("unexpected summary slices: items=%d pay=%d", len(summary.Items), len(summary.PayChannelBreakdown))
	}
	if len(summary.DeviceBreakdown) != 1 || len(summary.VariantDailyTrend) != 1 {
		t.Fatalf("unexpected device/variant slices: devices=%d variants=%d", len(summary.DeviceBreakdown), len(summary.VariantDailyTrend))
	}
	if summary.PayChannelBreakdown[0].PaidSuccessCount != 2 {
		t.Fatalf("expected pay success count 2, got %d", summary.PayChannelBreakdown[0].PaidSuccessCount)
	}
	if summary.DeviceBreakdown[0].PaidPerExposureRate <= 0 {
		t.Fatalf("expected positive paid per exposure rate, got %.4f", summary.DeviceBreakdown[0].PaidPerExposureRate)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
