package repo

import (
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestGetActiveLLMConfigFromMultipleEngines(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	enginesJSON := `[
		{"id":"qwen","name":"Qwen","base_url":"https://qwen.example.com","api_key":"qwen-key","model_name":"qwen-max","is_active":false},
		{"id":"deepseek","name":"DeepSeek","base_url":"https://deepseek.example.com","api_key":"deepseek-key","model_name":"deepseek-chat","is_active":true}
	]`

	// Mock querying llm.engines
	mock.ExpectQuery("SELECT config_value FROM system_configs WHERE config_key = 'llm.engines'").
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow(enginesJSON))

	apiKey, baseURL, modelName := repo.getActiveLLMConfig()

	if apiKey != "deepseek-key" || baseURL != "https://deepseek.example.com" || modelName != "deepseek-chat" {
		t.Fatalf("expected active deepseek config, got: apiKey=%s, baseURL=%s, modelName=%s", apiKey, baseURL, modelName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetActiveLLMConfigFallbackToLegacy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	// 1. Mock querying llm.engines fails (no row)
	mock.ExpectQuery("SELECT config_value FROM system_configs WHERE config_key = 'llm.engines'").
		WillReturnError(sql.ErrNoRows)

	// 2. Mock querying fallback legacy keys
	mock.ExpectQuery("SELECT config_value FROM system_configs WHERE config_key = 'llm.api_key'").
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow("legacy-key"))
	mock.ExpectQuery("SELECT config_value FROM system_configs WHERE config_key = 'llm.base_url'").
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow("https://legacy.example.com"))
	mock.ExpectQuery("SELECT config_value FROM system_configs WHERE config_key = 'llm.model_name'").
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow("legacy-model"))

	apiKey, baseURL, modelName := repo.getActiveLLMConfig()

	if apiKey != "legacy-key" || baseURL != "https://legacy.example.com" || modelName != "legacy-model" {
		t.Fatalf("expected fallback legacy config, got: apiKey=%s, baseURL=%s, modelName=%s", apiKey, baseURL, modelName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
