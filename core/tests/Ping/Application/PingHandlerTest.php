<?php

declare(strict_types=1);

namespace App\Tests\Ping\Application;

use App\Ping\Application\Query\Ping\PingHandler;
use App\Ping\Application\Query\Ping\PingQuery;
use App\Tests\Fake\InMemoryPingGateway;
use PHPUnit\Framework\TestCase;

final class PingHandlerTest extends TestCase
{
    public function testDelegatesToGateway(): void
    {
        $gateway = new InMemoryPingGateway();

        $result = (new PingHandler($gateway))(new PingQuery('hi'));

        self::assertSame('pong: hi', $result->message);
        self::assertSame(['hi'], $gateway->received);
    }
}
