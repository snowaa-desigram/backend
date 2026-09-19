<?php

declare(strict_types=1);

namespace App\Ping\Infrastructure\Grpc;

use App\Ping\Application\Port\PingGateway;
use App\Shared\Infrastructure\Grpc\GrpcGateway;
use Desigram\Ping\V1\PingRequest;
use Desigram\Ping\V1\PingResponse;
use Desigram\Ping\V1\PingServiceClient;
use Symfony\Component\DependencyInjection\Attribute\Autowire;

final class GrpcPingGateway extends GrpcGateway implements PingGateway
{
    private PingServiceClient $client;

    public function __construct(#[Autowire(env: 'PING_GRPC_ADDR')] string $address)
    {
        $this->client = new PingServiceClient($address, self::channelOptions());
    }

    public function ping(string $message): string
    {
        /** @var PingResponse $reply */
        $reply = $this->call(fn (): array => $this->client->Ping((new PingRequest())->setMessage($message), [], self::callOptions())->wait());

        return $reply->getMessage();
    }

    protected static function serviceName(): string
    {
        return 'ping';
    }
}
