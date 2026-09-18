<?php

declare(strict_types=1);

namespace App\Notification\Application\Port;

/** Порт к Python-микросервису telegram. Реализация — в Infrastructure. */
interface TelegramNotifier
{
    public function sendPhoto(int $chatId, string $photo, string $caption = ''): int;
}
