<?php

declare(strict_types=1);

namespace App\Tests\Shared\Presentation;

use App\Shared\Infrastructure\Security\JwtAccessTokenHandler;
use Firebase\JWT\JWT;
use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;

/** Интеграция: firewall api → access_token → JwtAccessTokenHandler → контроллер. */
final class MeControllerTest extends WebTestCase
{
    public function testWithoutTokenIs401(): void
    {
        $client = self::createClient();

        $client->request('GET', '/api/me');

        self::assertResponseStatusCodeSame(401);
        self::assertResponseHeaderSame('WWW-Authenticate', 'Bearer');
        self::assertJsonStringEqualsJsonString('{"code":"unauthorized","message":"unauthorized"}', (string) $client->getResponse()->getContent());
    }

    public function testWithInvalidTokenIs401(): void
    {
        $client = self::createClient();

        $client->request('GET', '/api/me', server: ['HTTP_AUTHORIZATION' => 'Bearer '.$this->token('wrong-secret-with-at-least-32-bytes')]);

        self::assertResponseStatusCodeSame(401);
        self::assertJsonStringEqualsJsonString('{"code":"unauthorized","message":"unauthorized"}', (string) $client->getResponse()->getContent());
    }

    public function testWithValidTokenReturnsUser(): void
    {
        $client = self::createClient();
        $secret = (string) ($_SERVER['JWT_SECRET'] ?? throw new \LogicException('JWT_SECRET is not set (see .env)'));

        $client->request('GET', '/api/me', server: ['HTTP_AUTHORIZATION' => 'Bearer '.$this->token($secret)]);

        self::assertResponseIsSuccessful();
        self::assertJsonStringEqualsJsonString('{"id":"u1","email":"user@example.com"}', (string) $client->getResponse()->getContent());
    }

    public function testPublicRoutesStayPublic(): void
    {
        $client = self::createClient();

        $client->request('GET', '/api/ping?message=x');

        self::assertResponseIsSuccessful();
    }

    private function token(string $secret): string
    {
        return JWT::encode([
            'iss' => JwtAccessTokenHandler::ISSUER,
            'sub' => 'u1',
            'uid' => 'u1',
            'email' => 'user@example.com',
            'iat' => time(),
            'exp' => time() + 900,
        ], $secret, JwtAccessTokenHandler::ALGORITHM);
    }
}
