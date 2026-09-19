from unittest.mock import MagicMock

import grpc
from desigram.telegram.v1 import telegram_pb2
from telegram_service.service import EmptyPhotoError, TelegramService
from telegram_service.servicer import TelegramServicer


class FakeService(TelegramService):
    def __init__(self) -> None:  # без sender: подменяем send_photo целиком
        pass

    def send_photo(self, chat_id: int, photo: bytes, caption: str = "") -> int:
        if not photo:
            raise EmptyPhotoError
        return 7


def test_send_photo_maps_response() -> None:
    servicer = TelegramServicer(FakeService())
    request = telegram_pb2.SendPhotoRequest(chat_id=1, photo=b"png", caption="hi")

    response = servicer.SendPhoto(request, MagicMock())

    assert isinstance(response, telegram_pb2.SendPhotoResponse)
    assert response.message_id == 7


def test_send_photo_empty_is_invalid_argument() -> None:
    servicer = TelegramServicer(FakeService())
    context = MagicMock()
    context.abort.side_effect = grpc.RpcError

    try:
        servicer.SendPhoto(telegram_pb2.SendPhotoRequest(chat_id=1, photo=b""), context)
    except grpc.RpcError:
        pass

    context.abort.assert_called_once_with(grpc.StatusCode.INVALID_ARGUMENT, "photo is empty")
