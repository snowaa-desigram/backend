import pytest
from telegram_service.service import EmptyPhotoError, TelegramService


class FakePhotoSender:
    def __init__(self) -> None:
        self.calls: list[tuple[int, bytes, str]] = []

    def send_photo(self, chat_id: int, photo: bytes, caption: str) -> int:
        self.calls.append((chat_id, photo, caption))
        return 42


def test_send_photo_delegates_to_sender() -> None:
    sender = FakePhotoSender()

    message_id = TelegramService(sender).send_photo(1, b"png", "hi")

    assert message_id == 42
    assert sender.calls == [(1, b"png", "hi")]


def test_send_photo_rejects_empty_photo() -> None:
    sender = FakePhotoSender()

    with pytest.raises(EmptyPhotoError):
        TelegramService(sender).send_photo(1, b"")

    assert sender.calls == []
