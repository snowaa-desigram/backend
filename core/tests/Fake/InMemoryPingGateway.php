<?php

declare(strict_types=1);

namespace App\Tests\Fake;

use App\Ping\Application\Port\PingGateway;

/** Подмена Go-сервиса в тестах: без сети и gRPC. */
final class InMemoryPingGateway implements PingGateway
{
    /** @var list<string> */
    public array $received = [];

    public function ping(string $message): string
    {
        $this->received[] = $message;

        return 'pong: '.$message;
    }
}
