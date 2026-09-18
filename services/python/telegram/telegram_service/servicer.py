import logging

import grpc
from desigram.telegram.v1 import telegram_pb2, telegram_pb2_grpc

log = logging.getLogger(__name__)


class TelegramServicer(telegram_pb2_grpc.TelegramServiceServicer):
    """Заглушка: отправка изображений в Telegram."""

    def __init__(self, bot_token: str) -> None:
        self._bot_token = bot_token

    def SendPhoto(
        self,
        request: telegram_pb2.SendPhotoRequest,
        context: grpc.ServicerContext,
    ) -> telegram_pb2.SendPhotoResponse:
        log.info(
            "SendPhoto chat_id=%s bytes=%d caption=%r",
            request.chat_id,
            len(request.photo),
            request.caption,
        )
        # TODO: вызвать Bot API (sendPhoto) с self._bot_token
        return telegram_pb2.SendPhotoResponse(message_id=0)
