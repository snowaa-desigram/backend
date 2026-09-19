<?php

declare(strict_types=1);

namespace App\Shared\Infrastructure\Security;

use Firebase\JWT\JWT;
use Firebase\JWT\Key;
use Symfony\Component\DependencyInjection\Attribute\Autowire;
use Symfony\Component\Security\Core\Exception\BadCredentialsException;
use Symfony\Component\Security\Http\AccessToken\AccessTokenHandlerInterface;
use Symfony\Component\Security\Http\Authenticator\Passport\Badge\UserBadge;

/**
 * Проверяет access-JWT auth-сервиса: HS256, общий JWT_SECRET, issuer desigram-auth.
 * Claims: uid (id пользователя), email — см. services/go/internal/auth/token.go.
 */
final readonly class JwtAccessTokenHandler implements AccessTokenHandlerInterface
{
    public const ISSUER = 'desigram-auth';
    public const ALGORITHM = 'HS256';

    public function __construct(#[Autowire(env: 'JWT_SECRET')] private string $secret)
    {
    }

    public function getUserBadgeFrom(#[\SensitiveParameter] string $accessToken): UserBadge
    {
        try {
            /** @var array<string, mixed> $claims */
            $claims = (array) JWT::decode($accessToken, new Key($this->secret, self::ALGORITHM));
        } catch (\Throwable $e) {
            throw new BadCredentialsException('Invalid access token', 0, $e);
        }

        if (($claims['iss'] ?? null) !== self::ISSUER) {
            throw new BadCredentialsException('Unexpected token issuer');
        }

        $id = $claims['uid'] ?? null;
        $email = $claims['email'] ?? null;
        if (!\is_string($id) || '' === $id || !\is_string($email) || '' === $email) {
            throw new BadCredentialsException('Access token has no user claims');
        }

        return new UserBadge($id, static fn (): JwtUser => new JwtUser($id, $email));
    }
}
