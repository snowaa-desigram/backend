<?php

declare(strict_types=1);

namespace App\Ping\Application\Query\Ping;

use App\Shared\Application\Bus\Query\Query;

final readonly class PingQuery implements Query
{
    public function __construct(public string $message)
    {
    }
}
