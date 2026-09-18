<?php

declare(strict_types=1);

namespace App\Tests\Ping\Presentation;

use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;

/** Интеграция: HTTP → QueryBus (Messenger) → handler → порт (in-memory в test env). */
final class PingControllerTest extends WebTestCase
{
    public function testPing(): void
    {
        $client = self::createClient();

        $client->request('GET', '/api/ping?message=ci');

        self::assertResponseIsSuccessful();
        self::assertJsonStringEqualsJsonString('{"message":"pong: ci"}', (string) $client->getResponse()->getContent());
    }

    public function testHealth(): void
    {
        $client = self::createClient();

        $client->request('GET', '/health');

        self::assertResponseIsSuccessful();
    }
}
