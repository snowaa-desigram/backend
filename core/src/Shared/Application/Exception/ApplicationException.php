<?php

declare(strict_types=1);

namespace App\Shared\Application\Exception;

/**
 * Базовая ошибка приложения: сама знает свой машиночитаемый код и HTTP-статус.
 * Presentation переводит её в единый ответ Error (backend/openapi/common.yaml) — см. ApiExceptionListener.
 * Новый тип ошибки = новый наследник; listener не меняется.
 */
abstract class ApplicationException extends \RuntimeException
{
    /** @param array<string, string> $details */
    public function __construct(string $message, private readonly array $details = [], ?\Throwable $previous = null)
    {
        parent::__construct($message, 0, $previous);
    }

    /** Код из enum Error.code в common.yaml. */
    abstract public function code(): string;

    abstract public function httpStatus(): int;

    /** @return array<string, string> */
    public function details(): array
    {
        return $this->details;
    }
}
