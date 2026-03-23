from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="STRATEGY_GRAPH_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    service_name: str = "strategy-graph"
    environment: str = "development"
    neo4j_uri: str = ""
    neo4j_user: str = "neo4j"
    neo4j_password: str = ""
    neo4j_database: str = "neo4j"
    query_limit: int = Field(default=80, ge=10, le=500)


@lru_cache(maxsize=1)
def get_settings() -> Settings:
    return Settings()
