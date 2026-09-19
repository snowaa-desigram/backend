<?php

declare(strict_types=1);

namespace App\Tests\Shared\Presentation;

use App\Shared\Application\Exception\ExternalServiceUnavailable;
use App\Shared\Application\Exception\NotFound;
use App\Shared\Application\Exception\ValidationFailed;
use App\Shared\Presentation\Http\ApiExceptionListener;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;
use Psr\Log\NullLogger;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpKernel\Event\ExceptionEvent;
use Symfony\Component\HttpKernel\Exception\MethodNotAllowedHttpException;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;
use Symfony\Component\HttpKernel\HttpKernelInterface;

final class ApiExceptionListenerTest extends TestCase
{
    /** @return iterable<string, array{\Throwable, int, string, array<string, string>}> */
    public static function cases(): iterable
    {
        yield 'application: validation' => [new ValidationFailed(details: ['photo' => 'is required']), 400, 'validation', ['photo' => 'is required']];
        yield 'application: not found' => [new NotFound('project not found'), 404, 'not_found', []];
        yield 'application: external' => [new ExternalServiceUnavailable('ping: [14] unavailable'), 503, 'service_unavailable', []];
        yield 'symfony: 404' => [new NotFoundHttpException('No route'), 404, 'not_found', []];
        yield 'symfony: 405' => [new MethodNotAllowedHttpException(['GET']), 405, 'method_not_allowed', []];
        yield 'unknown → 500 без деталей' => [new \RuntimeException('db password is hunter2'), 500, 'internal', []];
    }

    /** @param array<string, string> $details */
    #[DataProvider('cases')]
    public function testApiPath(\Throwable $e, int $status, string $code, array $details): void
    {
        $event = $this->dispatch('/api/anything', $e);

        $response = $event->getResponse();
        self::assertNotNull($response);
        self::assertSame($status, $response->getStatusCode());

        /** @var array{code: string, message: string, details?: array<string, string>} $body */
        $body = json_decode((string) $response->getContent(), true, flags: \JSON_THROW_ON_ERROR);
        self::assertSame($code, $body['code']);
        self::assertSame($details, $body['details'] ?? []);
        self::assertStringNotContainsString('hunter2', (string) $response->getContent());
    }

    public function testNonApiPathIsIgnored(): void
    {
        $event = $this->dispatch('/health', new NotFound());

        self::assertNull($event->getResponse());
    }

    private function dispatch(string $path, \Throwable $e): ExceptionEvent
    {
        $event = new ExceptionEvent($this->createStub(HttpKernelInterface::class), Request::create($path), HttpKernelInterface::MAIN_REQUEST, $e);
        (new ApiExceptionListener(new NullLogger()))($event);

        return $event;
    }
}
