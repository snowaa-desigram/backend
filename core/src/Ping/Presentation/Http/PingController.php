<?php

declare(strict_types=1);

namespace App\Ping\Presentation\Http;

use App\Ping\Application\Query\Ping\PingQuery;
use App\Ping\Application\Query\Ping\PingResult;
use App\Shared\Application\Bus\Query\QueryBus;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Attribute\AsController;
use Symfony\Component\Routing\Attribute\Route;

#[AsController]
final readonly class PingController
{
    public function __construct(private QueryBus $queryBus)
    {
    }

    #[Route('/api/ping', name: 'api_ping', methods: ['GET'])]
    public function __invoke(Request $request): JsonResponse
    {
        /** @var PingResult $result */
        $result = $this->queryBus->ask(new PingQuery($request->query->getString('message', 'hello')));

        return new JsonResponse(['message' => $result->message]);
    }
}
