from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="STRATEGY_ENGINE_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    service_name: str = "strategy-engine"
    environment: str = "development"
    log_level: str = "INFO"
    simulate_job_delay_seconds: float = Field(default=0.05, ge=0.0, le=5.0)
    go_backend_base_url: str = ""
    go_backend_timeout_ms: int = Field(default=8000, ge=500, le=60000)
    graph_service_base_url: str = ""
    graph_service_timeout_ms: int = Field(default=3000, ge=500, le=30000)
    allow_sample_stock_seeds: bool = False
    allow_sample_futures_seeds: bool = False
    llm_api_key: str = ""
    llm_base_url: str = "https://api.openai.com/v1"
    llm_model: str = "gpt-4o-mini"
    llm_timeout: float = Field(default=3.0, ge=0.5, le=30.0)
    enable_llm_review: bool = True

    @property
    def supported_job_types(self) -> tuple[str, ...]:
        return ("stock-selection", "futures-strategy")


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    return Settings()
