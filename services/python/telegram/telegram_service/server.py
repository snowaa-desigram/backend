from desigram.telegram.v1 import telegram_pb2, telegram_pb2_grpc
from desigram_common import Settings, serve

from telegram_service.servicer import TelegramServicer


def main() -> None:
    settings = Settings(name="telegram")
    serve(
        settings,
        register=lambda server: telegram_pb2_grpc.add_TelegramServiceServicer_to_server(
            TelegramServicer(bot_token=Settings.env("TELEGRAM_BOT_TOKEN")),
            server,
        ),
        service_names=[telegram_pb2.DESCRIPTOR.services_by_name["TelegramService"].full_name],
    )


if __name__ == "__main__":
    main()
