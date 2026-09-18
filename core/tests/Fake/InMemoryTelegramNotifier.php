<?php

declare(strict_types=1);

namespace App\Tests\Fake;

use App\Notification\Application\Port\TelegramNotifier;

/** Подмена Python-сервиса в тестах. */
final class InMemoryTelegramNotifier implements TelegramNotifier
{
    /** @var list<array{chatId: int, photo: string, caption: string}> */
    public array $sent = [];

    public function sendPhoto(int $chatId, string $photo, string $caption = ''): int
    {
        $this->sent[] = ['chatId' => $chatId, 'photo' => $photo, 'caption' => $caption];

        return \count($this->sent);
    }
}
