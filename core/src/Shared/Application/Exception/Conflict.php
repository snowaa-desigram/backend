<?php

declare(strict_types=1);

namespace App\Shared\Application\Exception;

use Symfony\Component\HttpFoundation\Response;

final class Conflict extends ApplicationException
{
    /** @param array<string, string> $details */
    public function __construct(string $message = 'conflict', array $details = [], ?\Throwable $previous = null)
    {
        parent::__construct($message, $details, $previous);
    }

    public function code(): string
    {
        return 'conflict';
    }

    public function httpStatus(): int
    {
        return Response::HTTP_CONFLICT;
    }
}
