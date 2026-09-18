import os
from dataclasses import dataclass, field


@dataclass(frozen=True)
class Settings:
    """Общие настройки сервиса. Всё — из переменных окружения (единая точка: enviropment/.env)."""

    name: str
    port: int = field(default_factory=lambda: int(os.getenv("GRPC_PORT", "50051")))
    log_level: str = field(default_factory=lambda: os.getenv("LOG_LEVEL", "INFO"))
    max_workers: int = field(default_factory=lambda: int(os.getenv("GRPC_MAX_WORKERS", "10")))
    shutdown_grace: float = field(
        default_factory=lambda: float(os.getenv("GRPC_SHUTDOWN_GRACE", "10"))
    )

    @staticmethod
    def env(key: str, default: str = "") -> str:
        return os.getenv(key, default)
