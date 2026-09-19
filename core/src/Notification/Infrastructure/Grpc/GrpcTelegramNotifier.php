<?php

declare(strict_types=1);

namespace App\Notification\Infrastructure\Grpc;

use App\Notification\Application\Port\TelegramNotifier;
use App\Shared\Infrastructure\Grpc\GrpcGateway;
use Desigram\Telegram\V1\SendPhotoRequest;
use Desigram\Telegram\V1\SendPhotoResponse;
use Desigram\Telegram\V1\TelegramServiceClient;
use Symfony\Component\DependencyInjection\Attribute\Autowire;

final class GrpcTelegramNotifier extends GrpcGateway implements TelegramNotifier
{
    private TelegramServiceClient $client;

    public function __construct(#[Autowire(env: 'TELEGRAM_GRPC_ADDR')] string $address)
    {
        $this->client = new TelegramServiceClient($address, self::channelOptions());
    }

    public function sendPhoto(int $chatId, string $photo, string $caption = ''): int
    {
        $request = (new SendPhotoRequest())
            ->setChatId($chatId)
            ->setPhoto($photo)
            ->setCaption($caption);

        /** @var SendPhotoResponse $reply */
        $reply = $this->call(fn (): array => $this->client->SendPhoto($request, [], self::callOptions())->wait());

        return (int) $reply->getMessageId();
    }

    protected static function serviceName(): string
    {
        return 'telegram';
    }
}
