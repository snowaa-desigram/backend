<?php

declare(strict_types=1);

namespace App\Tests\Shared\Infrastructure\Persistence;

use App\Shared\Domain\AggregateRoot;
use App\Shared\Infrastructure\Persistence\Doctrine\CachedRepository;
use App\Tests\Fake\SomethingHappened;
use App\Tests\Fake\SpyEventBus;
use Doctrine\ORM\EntityManagerInterface;
use PHPUnit\Framework\TestCase;
use Symfony\Component\Cache\Adapter\ArrayAdapter;

final class CachedRepositoryTest extends TestCase
{
    public function testSavePublishesAggregateEventsAfterFlushAndClearsThem(): void
    {
        $aggregate = new class extends AggregateRoot {
            public function doSomething(): void
            {
                $this->record(new SomethingHappened('agg-1'));
            }
        };
        $aggregate->doSomething();

        $em = $this->createMock(EntityManagerInterface::class);
        $em->expects(self::once())->method('persist')->with($aggregate);
        $em->expects(self::once())->method('flush');
        $bus = new SpyEventBus();

        (new FakeRepository($em, new ArrayAdapter(), $bus))->store($aggregate);

        self::assertCount(1, $bus->published);
        self::assertSame('agg-1', $bus->published[0]->aggregateId());
        self::assertSame([], $aggregate->pullEvents(), 'события отданы один раз');
    }

    public function testSaveOfPlainEntityDoesNotTouchBus(): void
    {
        $em = $this->createStub(EntityManagerInterface::class);
        $bus = new SpyEventBus();

        (new FakeRepository($em, new ArrayAdapter(), $bus))->store(new \stdClass());

        self::assertSame([], $bus->published);
    }

    public function testRememberCachesLoaderResult(): void
    {
        $repo = new FakeRepository($this->createStub(EntityManagerInterface::class), new ArrayAdapter(), new SpyEventBus());
        $calls = 0;
        $loader = static function () use (&$calls): string {
            ++$calls;

            return 'value';
        };

        self::assertSame('value', $repo->load('k', $loader));
        self::assertSame('value', $repo->load('k', $loader));
        self::assertSame(1, $calls);

        $repo->drop('k');
        $repo->load('k', $loader);
        self::assertSame(2, $calls, 'forget() инвалидирует ключ');
    }
}

final class FakeRepository extends CachedRepository
{
    public function store(object $entity): void
    {
        $this->save($entity);
    }

    /** @param callable(): string $loader */
    public function load(string $key, callable $loader): string
    {
        return $this->remember($key, $loader);
    }

    public function drop(string $key): void
    {
        $this->forget($key);
    }
}
