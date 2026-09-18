<?php

declare(strict_types=1);

namespace App\Tests\Notification\Application;

use App\Notification\Application\Command\SendPhoto\SendPhotoCommand;
use App\Notification\Application\Command\SendPhoto\SendPhotoHandler;
use App\Tests\Fake\InMemoryTelegramNotifier;
use PHPUnit\Framework\TestCase;

final class SendPhotoHandlerTest extends TestCase
{
    public function testSendsPhoto(): void
    {
        $notifier = new InMemoryTelegramNotifier();

        (new SendPhotoHandler($notifier))(new SendPhotoCommand(42, 'png-bytes', 'hello'));

        self::assertSame([['chatId' => 42, 'photo' => 'png-bytes', 'caption' => 'hello']], $notifier->sent);
    }
}
