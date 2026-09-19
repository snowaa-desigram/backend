<?php

declare(strict_types=1);

namespace App\Shared\Application\Exception;

use Symfony\Component\HttpFoundation\Response;

final class Forbidden extends ApplicationException
{
    /** @param array<string, string> $details */
    public function __construct(string $message = 'forbidden', array $details = [], ?\Throwable $previous = null)
    {
        parent::__construct($message, $details, $previous);
    }

    public function code(): string
    {
        return 'forbidden';
    }

    public function httpStatus(): int
    {
        return Response::HTTP_FORBIDDEN;
    }
}
