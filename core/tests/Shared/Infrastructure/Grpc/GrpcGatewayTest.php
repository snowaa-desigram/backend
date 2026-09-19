<?php

declare(strict_types=1);

namespace App\Tests\Shared\Infrastructure\Grpc;

use App\Shared\Application\Exception\ExternalServiceUnavailable;
use App\Shared\Infrastructure\Grpc\GrpcGateway;
use PHPUnit\Framework\TestCase;

final class GrpcGatewayTest extends TestCase
{
    public function testOkReturnsReply(): void
    {
        $reply = new \stdClass();

        self::assertSame($reply, $this->gateway()->exec(static fn (): array => [$reply, (object) ['code' => 0]]));
    }

    public function testNonOkStatusIsExternalServiceUnavailable(): void
    {
        $this->expectException(ExternalServiceUnavailable::class);
        $this->expectExceptionMessage('fake: grpc status 14 connection refused');

        $this->gateway()->exec(static fn (): array => [null, (object) ['code' => 14, 'details' => 'connection refused']]);
    }

    public function testNullReplyWithOkStatusIsExternalServiceUnavailable(): void
    {
        $this->expectException(ExternalServiceUnavailable::class);

        $this->gateway()->exec(static fn (): array => [null, (object) ['code' => 0]]);
    }

    public function testTransportExceptionIsWrapped(): void
    {
        try {
            $this->gateway()->exec(static fn (): array => throw new \RuntimeException('channel closed'));
            self::fail('expected ExternalServiceUnavailable');
        } catch (ExternalServiceUnavailable $e) {
            self::assertSame('fake: channel closed', $e->getMessage());
            self::assertSame(503, $e->httpStatus());
            self::assertInstanceOf(\RuntimeException::class, $e->getPrevious());
        }
    }

    private function gateway(): FakeGateway
    {
        return new FakeGateway();
    }
}

final class FakeGateway extends GrpcGateway
{
    /** @param callable(): array{object|null, object} $rpc */
    public function exec(callable $rpc): object
    {
        return $this->call($rpc);
    }

    protected static function serviceName(): string
    {
        return 'fake';
    }
}
