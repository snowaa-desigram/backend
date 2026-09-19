<?php

declare(strict_types=1);

namespace App\Shared\Application\Security;

/** Текущий пользователь из access-токена (см. Infrastructure\Security\JwtUser). */
interface AuthenticatedUser
{
    public function id(): string;

    public function email(): string;
}
