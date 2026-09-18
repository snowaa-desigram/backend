<?php

declare(strict_types=1);

namespace App\Notification\Application\Command\SendPhoto;

use App\Notification\Application\Port\TelegramNotifier;
use App\Shared\Application\Bus\Command\CommandHandler;

final readonly class SendPhotoHandler implements CommandHandler
{
    public function __construct(private TelegramNotifier $notifier)
    {
    }

    public function __invoke(SendPhotoCommand $command): void
    {
        $this->notifier->sendPhoto($command->chatId, $command->photo, $command->caption);
    }
}
