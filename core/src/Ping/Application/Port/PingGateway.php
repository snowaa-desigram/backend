<?php

declare(strict_types=1);

namespace App\Ping\Application\Port;

/** Порт к Go-микросервису ping. Реализация — в Infrastructure. */
interface PingGateway
{
    public function ping(string $message): string;
}
