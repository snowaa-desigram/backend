<?php

declare(strict_types=1);

namespace App\Tests\Fake;

use App\Shared\Application\Event\EventBus;
use App\Shared\Domain\Event\DomainEvent;

final class SpyEventBus implements EventBus
{
    /** @var list<DomainEvent> */
    public array $published = [];

    public function publish(DomainEvent ...$events): void
    {
        $this->published = array_values([...$this->published, ...$events]);
    }
}
