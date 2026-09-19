<?php

declare(strict_types=1);

namespace App\Tests\Fake;

use App\Shared\Domain\Event\DomainEvent;

final readonly class SomethingHappened implements DomainEvent
{
    public function __construct(private string $aggregateId, private \DateTimeImmutable $occurredOn = new \DateTimeImmutable())
    {
    }

    public function aggregateId(): string
    {
        return $this->aggregateId;
    }

    public function occurredOn(): \DateTimeImmutable
    {
        return $this->occurredOn;
    }
}
