<?php

declare(strict_types=1);

namespace App\Shared\Infrastructure\Persistence\Doctrine;

use App\Shared\Application\Event\EventBus;
use App\Shared\Domain\AggregateRoot;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\ORM\QueryBuilder;
use Symfony\Component\DependencyInjection\Attribute\Autowire;
use Symfony\Contracts\Cache\CacheInterface;
use Symfony\Contracts\Cache\ItemInterface;

/**
 * База для Doctrine-репозиториев: в SQL ходим только через Redis (cache-aside).
 *
 *  - чтение:  remember($key, fn () => ...запрос к БД...)  — результат кладётся в Redis
 *  - запись:  save()/remove() пишут в БД и инвалидируют ключи через forget()
 *  - события: save() после flush публикует события агрегата в EventBus — хендлер об этом не думает.
 *             Async-подписчики получат сообщение только после коммита (dispatch_after_current_bus снаружи doctrine_transaction).
 *  - произвольные DQL: cachedQuery($qb, $key) включает doctrine result cache
 *
 * Прямой вызов $em->createQuery()->getResult() в наследниках — запрещён (deptrac это не ловит, ревью — ловит).
 */
abstract class CachedRepository
{
    protected const int TTL = 3600;

    public function __construct(
        protected readonly EntityManagerInterface $em,
        #[Autowire(service: 'doctrine.result_cache_pool')]
        private readonly CacheInterface $cache,
        private readonly EventBus $events,
    ) {
    }

    /**
     * @template T
     *
     * @param callable(): T $loader
     *
     * @return T
     */
    protected function remember(string $key, callable $loader, int $ttl = self::TTL): mixed
    {
        return $this->cache->get($this->key($key), static function (ItemInterface $item) use ($loader, $ttl): mixed {
            $item->expiresAfter($ttl);

            return $loader();
        });
    }

    protected function forget(string ...$keys): void
    {
        foreach ($keys as $key) {
            $this->cache->delete($this->key($key));
        }
    }

    /** Запрос с doctrine result cache (Redis) — для списков/выборок. */
    protected function cachedQuery(QueryBuilder $qb, string $key, int $ttl = self::TTL): mixed
    {
        return $qb->getQuery()
            ->enableResultCache($ttl, $this->key($key))
            ->getResult();
    }

    protected function save(object $entity): void
    {
        $this->em->persist($entity);
        $this->em->flush();

        if ($entity instanceof AggregateRoot) {
            $this->events->publish(...$entity->pullEvents());
        }
    }

    protected function remove(object $entity): void
    {
        $this->em->remove($entity);
        $this->em->flush();
    }

    /** Ключи разделены по репозиторию, чтобы не пересекаться между контекстами. */
    private function key(string $key): string
    {
        return str_replace('\\', '.', static::class).'.'.$key;
    }
}
