<?php

declare(strict_types=1);

namespace App\Tests\Notification\Presentation;

use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;

/** Интеграция: контроллер бросает ValidationFailed → ApiExceptionListener → единый формат Error. */
final class SendPhotoControllerTest extends WebTestCase
{
    public function testMissingPhotoIsValidationError(): void
    {
        $client = self::createClient();

        $client->request('POST', '/api/telegram/photo', ['chat_id' => 1]);

        self::assertResponseStatusCodeSame(400);
        self::assertJsonStringEqualsJsonString(
            '{"code":"validation","message":"invalid request","details":{"photo":"is required"}}',
            (string) $client->getResponse()->getContent(),
        );
    }

    public function testUnknownApiRouteIsNotFoundError(): void
    {
        $client = self::createClient();

        $client->request('GET', '/api/nope');

        self::assertResponseStatusCodeSame(404);
        self::assertSame('not_found', json_decode((string) $client->getResponse()->getContent(), true)['code'] ?? null);
    }
}
