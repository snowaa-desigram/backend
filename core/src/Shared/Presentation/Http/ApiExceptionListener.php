<?php

declare(strict_types=1);

namespace App\Shared\Presentation\Http;

use App\Shared\Application\Exception\ApplicationException;
use Psr\Log\LoggerInterface;
use Symfony\Component\EventDispatcher\Attribute\AsEventListener;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Event\ExceptionEvent;
use Symfony\Component\HttpKernel\Exception\HttpExceptionInterface;
use Symfony\Component\HttpKernel\KernelEvents;
use Symfony\Component\Security\Core\Exception\AccessDeniedException;
use Symfony\Component\Security\Core\Exception\AuthenticationException;

/**
 * Любое исключение под /api → ответ Error (common.yaml) с правильным статусом.
 * Цепочка из трёх звеньев: ApplicationException → HTTP-исключения Symfony → всё остальное = 500 без деталей.
 */
#[AsEventListener(event: KernelEvents::EXCEPTION, priority: 10)]
final readonly class ApiExceptionListener
{
    private const array HTTP_CODES = [
        Response::HTTP_BAD_REQUEST => 'validation',
        Response::HTTP_UNAUTHORIZED => 'unauthorized',
        Response::HTTP_FORBIDDEN => 'forbidden',
        Response::HTTP_NOT_FOUND => 'not_found',
        Response::HTTP_METHOD_NOT_ALLOWED => 'method_not_allowed',
        Response::HTTP_CONFLICT => 'conflict',
        Response::HTTP_TOO_MANY_REQUESTS => 'too_many_requests',
        Response::HTTP_SERVICE_UNAVAILABLE => 'service_unavailable',
    ];

    public function __construct(private LoggerInterface $logger)
    {
    }

    public function __invoke(ExceptionEvent $event): void
    {
        if (!str_starts_with($event->getRequest()->getPathInfo(), '/api')) {
            return;
        }

        $e = $event->getThrowable();

        // Security сам ведёт их в entry point / failure handler (JsonUnauthorizedHandler) — тот же формат ответа
        if ($e instanceof AccessDeniedException || $e instanceof AuthenticationException) {
            return;
        }

        if ($e instanceof ApplicationException) {
            $event->setResponse((new ApiError($e->code(), $e->getMessage(), $e->details()))->toResponse($e->httpStatus()));

            return;
        }

        if ($e instanceof HttpExceptionInterface) {
            $status = $e->getStatusCode();
            $code = self::HTTP_CODES[$status] ?? ($status >= 500 ? 'internal' : 'validation');
            $event->setResponse((new ApiError($code, $e->getMessage() ?: $code))->toResponse($status, $e->getHeaders()));

            return;
        }

        $this->logger->error('api: unhandled exception', ['exception' => $e]);
        $event->setResponse((new ApiError('internal', 'internal error'))->toResponse(Response::HTTP_INTERNAL_SERVER_ERROR));
    }
}
