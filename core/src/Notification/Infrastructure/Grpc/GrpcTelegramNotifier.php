<?php

declare(strict_types=1);

namespace App\Notification\Infrastructure\Grpc;

use App\Notification\Application\Port\TelegramNotifier;
use Desigram\Telegram\V1\SendPhotoRequest;
use Desigram\Telegram\V1\SendPhotoResponse;
use Desigram\Telegram\V1\TelegramServiceClient;
use Grpc\ChannelCredentials;
use Symfony\Component\DependencyInjection\Attribute\Autowire;

final class GrpcTelegramNotifier implements TelegramNotifier
{
    private TelegramServiceClient $client;

    public function __construct(#[Autowire(env: 'TELEGRAM_GRPC_ADDR')] string $address)
    {
        $this->client = new TelegramServiceClient($address, [
            'credentials' => ChannelCredentials::createInsecure(),
        ]);
    }

    public function sendPhoto(int $chatId, string $photo, string $caption = ''): int
    {
        $request = (new SendPhotoRequest())
            ->setChatId($chatId)
            ->setPhoto($photo)
            ->setCaption($caption);

        /** @var SendPhotoResponse|null $reply */
        [$reply, $status] = $this->client->SendPhoto($request)->wait();

        if (\Grpc\STATUS_OK !== $status->code || null === $reply) {
            throw new \RuntimeException(\sprintf('telegram grpc failed: [%d] %s', $status->code, $status->details));
        }

        return (int) $reply->getMessageId();
    }
}
