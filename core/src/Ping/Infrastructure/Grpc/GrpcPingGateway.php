<?php

declare(strict_types=1);

namespace App\Ping\Infrastructure\Grpc;

use App\Ping\Application\Port\PingGateway;
use Desigram\Ping\V1\PingRequest;
use Desigram\Ping\V1\PingResponse;
use Desigram\Ping\V1\PingServiceClient;
use Grpc\ChannelCredentials;
use Symfony\Component\DependencyInjection\Attribute\Autowire;

final class GrpcPingGateway implements PingGateway
{
    private PingServiceClient $client;

    public function __construct(#[Autowire(env: 'PING_GRPC_ADDR')] string $address)
    {
        $this->client = new PingServiceClient($address, [
            'credentials' => ChannelCredentials::createInsecure(),
        ]);
    }

    public function ping(string $message): string
    {
        /** @var PingResponse|null $reply */
        [$reply, $status] = $this->client->Ping((new PingRequest())->setMessage($message))->wait();

        if (\Grpc\STATUS_OK !== $status->code || null === $reply) {
            throw new \RuntimeException(\sprintf('ping grpc failed: [%d] %s', $status->code, $status->details));
        }

        return $reply->getMessage();
    }
}
