<?php

declare(strict_types=1);

namespace App\Shared\Application\Event;

use App\Shared\Domain\Event\DomainEvent;

/**
 * Порт публикации доменных событий. Единственный способ связи между контекстами:
 * контекст A публикует событие, контекст B подписывается (EventSubscriber). Вызывать чужие хендлеры нельзя.
 */
interface EventBus
{
    public function publish(DomainEvent ...$events): void;
}
