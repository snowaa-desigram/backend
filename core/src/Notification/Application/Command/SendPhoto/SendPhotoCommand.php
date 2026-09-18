<?php

declare(strict_types=1);

namespace App\Notification\Application\Command\SendPhoto;

use App\Shared\Application\Bus\Command\Command;

final readonly class SendPhotoCommand implements Command
{
    public function __construct(
        public int $chatId,
        public string $photo,
        public string $caption = '',
    ) {
    }
}
