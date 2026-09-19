"""Транспорт: gRPC TelegramService → TelegramService (use-cases). Здесь — только маппинг pb ↔ python и статусы."""

import grpc
from desigram.telegram.v1 import telegram_pb2, telegram_pb2_grpc

from telegram_service.service import EmptyPhotoError, TelegramService


class TelegramServicer(telegram_pb2_grpc.TelegramServiceServicer):
    def __init__(self, service: TelegramService) -> None:
        self._service = service

    def SendPhoto(
        self,
        request: telegram_pb2.SendPhotoRequest,
        context: grpc.ServicerContext,
    ) -> telegram_pb2.SendPhotoResponse:
        try:
            message_id = self._service.send_photo(request.chat_id, request.photo, request.caption)
        except EmptyPhotoError:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "photo is empty")
        return telegram_pb2.SendPhotoResponse(message_id=message_id)
