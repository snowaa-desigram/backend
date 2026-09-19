"""Адаптер к Telegram Bot API. Реализует service.PhotoSender."""

import logging

log = logging.getLogger(__name__)


class TelegramApiPhotoSender:
    """Заглушка: логирует вызов. TODO: Bot API sendPhoto с bot_token."""

    def __init__(self, bot_token: str) -> None:
        self._bot_token = bot_token

    def send_photo(self, chat_id: int, photo: bytes, caption: str) -> int:
        log.info("sendPhoto chat_id=%s bytes=%d caption=%r", chat_id, len(photo), caption)
        return 0
