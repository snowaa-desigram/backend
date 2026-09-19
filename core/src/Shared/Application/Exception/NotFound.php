<?php

declare(strict_types=1);

namespace App\Shared\Application\Exception;

use Symfony\Component\HttpFoundation\Response;

final class NotFound extends ApplicationException
{
    /** @param array<string, string> $details */
    public function __construct(string $message = 'not found', array $details = [], ?\Throwable $previous = null)
    {
        parent::__construct($message, $details, $previous);
    }

    public function code(): string
    {
        return 'not_found';
    }

    public function httpStatus(): int
    {
        return Response::HTTP_NOT_FOUND;
    }
}
