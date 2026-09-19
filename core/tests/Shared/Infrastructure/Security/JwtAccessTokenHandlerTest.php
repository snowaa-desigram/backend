<?php

declare(strict_types=1);

namespace App\Tests\Shared\Infrastructure\Security;

use App\Shared\Infrastructure\Security\JwtAccessTokenHandler;
use App\Shared\Infrastructure\Security\JwtUser;
use Firebase\JWT\JWT;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;
use Symfony\Component\Security\Core\Exception\BadCredentialsException;

final class JwtAccessTokenHandlerTest extends TestCase
{
    private const SECRET = 'test-secret-with-at-least-32-bytes!';

    public function testValidToken(): void
    {
        $badge = (new JwtAccessTokenHandler(self::SECRET))->getUserBadgeFrom($this->token());

        self::assertSame('u1', $badge->getUserIdentifier());
        $user = ($badge->getUserLoader() ?? throw new \LogicException('no loader'))('u1');
        self::assertInstanceOf(JwtUser::class, $user);
        self::assertSame('u1', $user->id());
        self::assertSame('user@example.com', $user->email());
        self::assertSame(['ROLE_USER'], $user->getRoles());
    }

    /** @return iterable<string, array{string}> */
    public static function invalidTokens(): iterable
    {
        yield 'garbage' => ['not-a-jwt'];
        yield 'wrong secret' => [self::encode(self::claims(), 'other-secret-with-at-least-32-bytes')];
        yield 'expired' => [self::encode(self::claims(['exp' => time() - 60]))];
        yield 'wrong issuer' => [self::encode(self::claims(['iss' => 'someone-else']))];
        yield 'no uid' => [self::encode(self::claims(['uid' => null]))];
        yield 'no email' => [self::encode(self::claims(['email' => null]))];
        yield 'alg none' => [self::encode(self::claims(), self::SECRET, 'none')];
    }

    #[DataProvider('invalidTokens')]
    public function testInvalidToken(string $token): void
    {
        $this->expectException(BadCredentialsException::class);

        (new JwtAccessTokenHandler(self::SECRET))->getUserBadgeFrom($token);
    }

    private function token(): string
    {
        return self::encode(self::claims());
    }

    /**
     * @param array<string, mixed> $override null в override удаляет claim
     *
     * @return array<string, mixed>
     */
    private static function claims(array $override = []): array
    {
        $claims = array_merge([
            'iss' => JwtAccessTokenHandler::ISSUER,
            'sub' => 'u1',
            'uid' => 'u1',
            'email' => 'user@example.com',
            'iat' => time(),
            'exp' => time() + 900,
        ], $override);

        return array_filter($claims, static fn ($v) => null !== $v);
    }

    /** @param array<string, mixed> $claims */
    private static function encode(array $claims, string $secret = self::SECRET, string $alg = 'HS256'): string
    {
        if ('none' === $alg) {
            $b64 = static fn (string $s): string => rtrim(strtr(base64_encode($s), '+/', '-_'), '=');

            return $b64((string) json_encode(['typ' => 'JWT', 'alg' => 'none'])).'.'.$b64((string) json_encode($claims)).'.';
        }

        return JWT::encode($claims, $secret, $alg);
    }
}
