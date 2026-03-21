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
    allow_sample_stock_seeds: bool = False
    allow_sample_futures_seeds: bool = False

    @property
    def supported_job_types(self) -> tuple[str, ...]:
        return ("stock-selection", "futures-strategy")


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    return Settings()
