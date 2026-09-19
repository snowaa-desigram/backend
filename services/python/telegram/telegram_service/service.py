"""Use-cases telegram. Чистый Python: без grpc и pb (проверяет lint-imports)."""

from typing import Protocol


class PhotoSender(Protocol):
    """Порт к внешней доставке фото. Реализация — clients/telegram_api.py."""

    def send_photo(self, chat_id: int, photo: bytes, caption: str) -> int:
        """Возвращает id отправленного сообщения."""
        ...


class EmptyPhotoError(Exception):
    """Фото пустое — отправлять нечего."""


class TelegramService:
    def __init__(self, sender: PhotoSender) -> None:
        self._sender = sender

    def send_photo(self, chat_id: int, photo: bytes, caption: str = "") -> int:
        if not photo:
            raise EmptyPhotoError
        return self._sender.send_photo(chat_id, photo, caption)
