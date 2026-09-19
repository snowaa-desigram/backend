"""Точка входа: настройки → clients → service → servicer → serve(). Только wiring."""

from desigram.telegram.v1 import telegram_pb2, telegram_pb2_grpc
from desigram_common import serve

from telegram_service.clients.telegram_api import TelegramApiPhotoSender
from telegram_service.service import TelegramService
from telegram_service.servicer import TelegramServicer
from telegram_service.settings import TelegramSettings


def main() -> None:
    settings = TelegramSettings(name="telegram")
    service = TelegramService(TelegramApiPhotoSender(settings.bot_token))
    serve(
        settings,
        register=lambda server: telegram_pb2_grpc.add_TelegramServiceServicer_to_server(
            TelegramServicer(service),
            server,
        ),
        service_names=[telegram_pb2.DESCRIPTOR.services_by_name["TelegramService"].full_name],
    )


if __name__ == "__main__":
    main()
