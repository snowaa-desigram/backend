from dataclasses import dataclass, field

from desigram_common import Settings


@dataclass(frozen=True)
class TelegramSettings(Settings):
    """Настройки telegram поверх общих: токен бота (TELEGRAM_BOT_TOKEN)."""

    bot_token: str = field(default_factory=lambda: Settings.env("TELEGRAM_BOT_TOKEN"))
