<?php

declare(strict_types=1);

namespace App\Shared\Application\Event;

/**
 * Маркер подписчика: регистрируется в event.bus через _instanceof (config/services.yaml).
 * Реализация — __invoke(<ConcreteEvent> $event). Тяжёлые подписчики уходят в async через routing в messenger.yaml.
 */
interface EventSubscriber
{
}
