<?php

declare(strict_types=1);

namespace App\Tests\Shared\Infrastructure\Bus;

use App\Shared\Application\Event\EventBus;
use App\Shared\Application\Event\EventSubscriber;
use App\Tests\Fake\SomethingHappened;
use Symfony\Bundle\FrameworkBundle\Test\KernelTestCase;

/** Интеграция: EventBus → Messenger event.bus → подписчик, помеченный EventSubscriber (_instanceof). */
final class MessengerEventBusTest extends KernelTestCase
{
    public function testSubscriberReceivesPublishedEvent(): void
    {
        self::bootKernel();
        $container = self::getContainer();

        /** @var RecordingSubscriber $subscriber */
        $subscriber = $container->get(RecordingSubscriber::class);
        /** @var EventBus $bus */
        $bus = $container->get(EventBus::class);

        $bus->publish(new SomethingHappened('agg-1'), new SomethingHappened('agg-2'));

        self::assertSame(['agg-1', 'agg-2'], $subscriber->seen);
    }

    public function testEventWithoutSubscribersIsFine(): void
    {
        self::bootKernel();
        /** @var EventBus $bus */
        $bus = self::getContainer()->get(EventBus::class);

        $bus->publish(new class('x') implements \App\Shared\Domain\Event\DomainEvent {
            public function __construct(private readonly string $id)
            {
            }

            public function aggregateId(): string
            {
                return $this->id;
            }

            public function occurredOn(): \DateTimeImmutable
            {
                return new \DateTimeImmutable();
            }
        });

        $this->addToAssertionCount(1); // allow_no_handlers: без подписчиков не падает
    }
}

final class RecordingSubscriber implements EventSubscriber
{
    /** @var list<string> */
    public array $seen = [];

    public function __invoke(SomethingHappened $event): void
    {
        $this->seen[] = $event->aggregateId();
    }
}
