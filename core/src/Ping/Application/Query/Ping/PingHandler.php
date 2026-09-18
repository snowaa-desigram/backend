<?php

declare(strict_types=1);

namespace App\Ping\Application\Query\Ping;

use App\Ping\Application\Port\PingGateway;
use App\Shared\Application\Bus\Query\QueryHandler;

final readonly class PingHandler implements QueryHandler
{
    public function __construct(private PingGateway $gateway)
    {
    }

    public function __invoke(PingQuery $query): PingResult
    {
        return new PingResult($this->gateway->ping($query->message));
    }
}
