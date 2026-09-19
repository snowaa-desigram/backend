<?php

declare(strict_types=1);

namespace App\Shared\Infrastructure\Grpc;

use App\Shared\Application\Exception\ExternalServiceUnavailable;
use Grpc\ChannelCredentials;

/**
 * Скелет gRPC-адаптера (Template Method): канал, deadline, проверка статуса, единая ошибка.
 * Наследник — это клиент из gen/ и по строке на RPC: return $this->call(fn () => $this->client->X($req)->wait())->getY();
 * Единственное место для TLS/mTLS, ретраев, трассировки, когда они понадобятся.
 */
abstract class GrpcGateway
{
    /** Таймаут unary-вызова, мс. */
    protected const int TIMEOUT_MS = 5000;

    /** @return array<string, mixed> опции канала для конструктора клиента из gen/ */
    protected static function channelOptions(): array
    {
        return ['credentials' => ChannelCredentials::createInsecure()];
    }

    /** @return array<string, mixed> опции вызова (третий аргумент RPC-метода клиента) */
    protected static function callOptions(): array
    {
        return ['timeout' => self::TIMEOUT_MS * 1000];
    }

    /**
     * @template T of object
     *
     * @param callable(): array{T|null, object} $rpc возвращает [reply, status] — результат ->wait()
     *
     * @return T
     *
     * @throws ExternalServiceUnavailable
     */
    protected function call(callable $rpc): object
    {
        try {
            [$reply, $status] = $rpc();
        } catch (\Throwable $e) {
            throw new ExternalServiceUnavailable(\sprintf('%s: %s', static::serviceName(), $e->getMessage()), previous: $e);
        }

        /** @var object{code: int, details?: string} $status */
        if (0 !== $status->code || null === $reply) {
            throw new ExternalServiceUnavailable(\sprintf('%s: grpc status %d %s', static::serviceName(), $status->code, $status->details ?? ''));
        }

        return $reply;
    }

    /** Имя внешнего сервиса для сообщений об ошибках и логов. */
    abstract protected static function serviceName(): string;
}
