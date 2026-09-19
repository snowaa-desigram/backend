<?php

declare(strict_types=1);

namespace App\Tests\Architecture;

use PHPUnit\Framework\TestCase;

/**
 * Папки и имена внутри контекста — по спеке openspec/specs/architecture-core:
 *   Domain/{Model,ValueObject,Event,Repository,Exception}
 *   Application/{Command/<Name>/{<Name>Command,<Name>Handler}, Query/<Name>/{<Name>Query,<Name>Handler,<Name>Result}, Port/*(Gateway|Notifier|Client), EventSubscriber/<Do>When<Event>}
 *   Infrastructure/<Tech>/<Tech>*   (Grpc/GrpcPingGateway, Persistence/Doctrine/DoctrineUserRepository)
 *   Presentation/Http/*Controller
 * и у каждого хендлера есть tests/<Context>/Application/<Name>HandlerTest.php.
 * Shared — ядро, его состав описан в README и здесь не проверяется. deptrac проверяет зависимости слоёв, этот тест — папки и имена.
 */
final class NamingConventionTest extends TestCase
{
    private const string SRC = __DIR__.'/../../src';
    private const string TESTS = __DIR__.'/..';
    private const array SKIP = ['Shared', 'Kernel.php'];

    private const array LAYERS = ['Domain', 'Application', 'Infrastructure', 'Presentation'];
    private const array DOMAIN_DIRS = ['Model', 'ValueObject', 'Event', 'Repository', 'Exception'];
    private const array APPLICATION_DIRS = ['Command', 'Query', 'Port', 'EventSubscriber'];

    public function testFilesFollowLayout(): void
    {
        $violations = [];

        foreach ($this->contexts() as $context) {
            foreach ($this->phpFiles(self::SRC.'/'.$context) as $file) {
                $rel = substr($file, \strlen(self::SRC) + \strlen($context) + 2); // Application/Command/SendPhoto/SendPhotoHandler.php
                $parts = explode('/', $rel);
                $class = basename($rel, '.php');

                if ($violation = $this->check($context, $parts, $class)) {
                    $violations[] = \sprintf('%s/%s: %s', $context, $rel, $violation);
                }
            }
        }

        self::assertSame([], $violations, "Папки и имена — по openspec/specs/architecture-core:\n".implode("\n", $violations));
    }

    /** @param list<string> $parts */
    private function check(string $context, array $parts, string $class): ?string
    {
        $layer = $parts[0];
        if (!\in_array($layer, self::LAYERS, true)) {
            return 'слой — только '.implode('|', self::LAYERS);
        }
        if (2 === \count($parts)) {
            return 'файл прямо в слое: положи в подпапку роли';
        }
        $dir = $parts[1];

        return match ($layer) {
            'Domain' => match (true) {
                !\in_array($dir, self::DOMAIN_DIRS, true) => 'Domain/ — только '.implode('|', self::DOMAIN_DIRS),
                'Repository' === $dir && !str_ends_with($class, 'Repository') => 'интерфейс репозитория — <Aggregate>Repository',
                default => null,
            },
            'Application' => match (true) {
                !\in_array($dir, self::APPLICATION_DIRS, true) => 'Application/ — только '.implode('|', self::APPLICATION_DIRS),
                'Command' === $dir => $this->checkMessage($context, $parts, $class, ['Command', 'Handler']),
                'Query' === $dir => $this->checkMessage($context, $parts, $class, ['Query', 'Handler', 'Result']),
                'Port' === $dir && !preg_match('/(Gateway|Notifier|Client)$/', $class) => 'порт — *Gateway | *Notifier | *Client',
                'EventSubscriber' === $dir && !preg_match('/^[A-Z]\w+When[A-Z]\w+$/', $class) => 'подписчик — <Do>When<Event>',
                default => null,
            },
            'Infrastructure' => str_starts_with($class, $parts[\count($parts) - 2]) ? null : \sprintf('реализация в Infrastructure/%s/ начинается с %2$s (например %2$s%3$s)', implode('/', \array_slice($parts, 1, -1)), $parts[\count($parts) - 2], $class),
            'Presentation' => match (true) {
                'Http' !== $dir => 'Presentation/ — только Http',
                !str_ends_with($class, 'Controller') => 'HTTP-вход — *Controller',
                !file_exists(\sprintf('%s/%s/Presentation/%sTest.php', self::TESTS, $context, $class)) => \sprintf('нет tests/%s/Presentation/%sTest.php', $context, $class),
                default => null,
            },
        };
    }

    /**
     * Application/Command/<Name>/<Name>(Command|Handler).php — сообщение и его обработчик в папке по имени сообщения.
     *
     * @param list<string> $parts
     * @param list<string> $suffixes
     */
    private function checkMessage(string $context, array $parts, string $class, array $suffixes): ?string
    {
        $kind = $parts[1];
        if (4 !== \count($parts)) {
            return \sprintf('Application/%1$s/<Name>/<Name>%1$s.php + <Name>Handler.php', $kind);
        }
        $name = $parts[2];
        $allowed = array_map(static fn (string $s): string => $name.$s, $suffixes);
        if (!\in_array($class, $allowed, true)) {
            return \sprintf('в Application/%s/%s/ — только %s', $kind, $name, implode('|', $allowed));
        }
        if (str_ends_with($class, 'Handler') && !file_exists(\sprintf('%s/%s/Application/%sTest.php', self::TESTS, $context, $class))) {
            return \sprintf('нет tests/%s/Application/%sTest.php', $context, $class);
        }

        return null;
    }

    /** @return list<string> */
    private function contexts(): array
    {
        $dirs = array_filter(scandir(self::SRC) ?: [], static fn (string $d): bool => is_dir(self::SRC.'/'.$d) && !\in_array($d, ['.', '..', ...self::SKIP], true));

        return array_values($dirs);
    }

    /** @return iterable<string> */
    private function phpFiles(string $dir): iterable
    {
        /** @var iterable<\SplFileInfo> $files */
        $files = new \RecursiveIteratorIterator(new \RecursiveDirectoryIterator($dir, \FilesystemIterator::SKIP_DOTS));
        foreach ($files as $file) {
            if ('php' === $file->getExtension()) {
                yield $file->getPathname();
            }
        }
    }
}
