from unittest.mock import MagicMock

from desigram.telegram.v1 import telegram_pb2
from telegram_service.servicer import TelegramServicer


def test_send_photo_returns_response() -> None:
    servicer = TelegramServicer(bot_token="test")
    request = telegram_pb2.SendPhotoRequest(chat_id=1, photo=b"png", caption="hi")

    response = servicer.SendPhoto(request, MagicMock())

    assert isinstance(response, telegram_pb2.SendPhotoResponse)
    assert response.message_id == 0
