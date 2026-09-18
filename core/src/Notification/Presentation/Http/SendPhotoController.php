<?php

declare(strict_types=1);

namespace App\Notification\Presentation\Http;

use App\Notification\Application\Command\SendPhoto\SendPhotoCommand;
use App\Shared\Application\Bus\Command\CommandBus;
use Symfony\Component\HttpFoundation\File\UploadedFile;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Attribute\AsController;
use Symfony\Component\Routing\Attribute\Route;

#[AsController]
final readonly class SendPhotoController
{
    public function __construct(private CommandBus $commandBus)
    {
    }

    /** multipart/form-data: chat_id, caption, photo (файл). */
    #[Route('/api/telegram/photo', name: 'api_telegram_send_photo', methods: ['POST'])]
    public function __invoke(Request $request): JsonResponse
    {
        $photo = $request->files->get('photo');
        if (!$photo instanceof UploadedFile) {
            return new JsonResponse(['error' => 'photo is required'], Response::HTTP_BAD_REQUEST);
        }

        $this->commandBus->dispatch(new SendPhotoCommand(
            chatId: $request->request->getInt('chat_id'),
            photo: (string) file_get_contents($photo->getPathname()),
            caption: $request->request->getString('caption'),
        ));

        return new JsonResponse(null, Response::HTTP_ACCEPTED);
    }
}
