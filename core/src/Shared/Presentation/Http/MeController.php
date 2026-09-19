<?php

declare(strict_types=1);

namespace App\Shared\Presentation\Http;

use App\Shared\Application\Security\AuthenticatedUser;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpKernel\Attribute\AsController;
use Symfony\Component\Routing\Attribute\Route;
use Symfony\Component\Security\Http\Attribute\CurrentUser;

/** Текущий пользователь по access-JWT от auth-сервиса (proof, что core доверяет токену). */
#[AsController]
final class MeController
{
    #[Route('/api/me', name: 'api_me', methods: ['GET'])]
    public function __invoke(#[CurrentUser] AuthenticatedUser $user): JsonResponse
    {
        return new JsonResponse(['id' => $user->id(), 'email' => $user->email()]);
    }
}
