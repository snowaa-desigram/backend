<?php

declare(strict_types=1);

namespace App\Shared\Presentation\Http;

use Symfony\Component\HttpFoundation\JsonResponse;

/** Тело ошибки API — схема Error из backend/openapi/common.yaml. Одна на auth (Go), core и sync. */
final readonly class ApiError
{
    /** @param array<string, string> $details */
    public function __construct(
        public string $code,
        public string $message,
        public array $details = [],
    ) {
    }

    /** @param array<string, string> $headers */
    public function toResponse(int $status, array $headers = []): JsonResponse
    {
        $body = ['code' => $this->code, 'message' => $this->message];
        if ([] !== $this->details) {
            $body['details'] = $this->details;
        }

        return new JsonResponse($body, $status, $headers);
    }
}
