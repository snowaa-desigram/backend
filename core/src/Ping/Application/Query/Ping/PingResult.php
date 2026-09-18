<?php

declare(strict_types=1);

namespace App\Ping\Application\Query\Ping;

final readonly class PingResult
{
    public function __construct(public string $message)
    {
    }
}
