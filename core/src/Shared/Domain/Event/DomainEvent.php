<?php

declare(strict_types=1);

namespace App\Shared\Domain\Event;

interface DomainEvent
{
    /** Идентификатор агрегата-источника — для проекторов, аудита, ключей кеша. */
    public function aggregateId(): string;

    public function occurredOn(): \DateTimeImmutable;
}
