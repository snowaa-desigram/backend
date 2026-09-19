<?php

declare(strict_types=1);

namespace App\Shared\Infrastructure\Security;

use App\Shared\Application\Security\AuthenticatedUser;
use Symfony\Component\Security\Core\User\UserInterface;

/** Пользователь, восстановленный из claims access-JWT: в БД core не ходим. */
final readonly class JwtUser implements AuthenticatedUser, UserInterface
{
    /** @param non-empty-string $id */
    public function __construct(private string $id, private string $email)
    {
    }

    public function id(): string
    {
        return $this->id;
    }

    public function email(): string
    {
        return $this->email;
    }

    /** @return non-empty-string */
    public function getUserIdentifier(): string
    {
        return $this->id;
    }

    /** @return list<string> */
    public function getRoles(): array
    {
        return ['ROLE_USER'];
    }

    #[\Deprecated]
    public function eraseCredentials(): void
    {
    }
}
