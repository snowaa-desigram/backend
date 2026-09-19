<?php

declare(strict_types=1);

namespace App\Tests\Architecture;

use PHPUnit\Framework\TestCase;

/**
 * Межконтекстная связь — только через доменные события (EventBus/EventSubscriber) и Shared.
 * Контекст A не импортирует классы контекста B: ни хендлеры, ни порты, ни сущности.
 * deptrac проверяет слои внутри контекста, этот тест — границы между контекстами.
 */
final class ContextIsolationTest extends TestCase
{
    private const string SRC = __DIR__.'/../../src';
    private const array ALLOWED = ['Shared', 'Kernel.php'];

    public function testContextsDoNotImportEachOther(): void
    {
        $violations = [];

        foreach ($this->contexts() as $context) {
            foreach ($this->phpFiles(self::SRC.'/'.$context) as $file) {
                preg_match_all('/^use App\\\\([A-Za-z]+)\\\\/m', (string) file_get_contents($file), $m);
                foreach (array_unique($m[1]) as $imported) {
                    if ($imported !== $context && !\in_array($imported, self::ALLOWED, true)) {
                        $violations[] = \sprintf('%s → App\\%s', substr($file, \strlen(self::SRC) + 1), $imported);
                    }
                }
            }
        }

        self::assertSame([], $violations, "Контексты связываются только через события:\n".implode("\n", $violations));
    }

    /** @return list<string> */
    private function contexts(): array
    {
        $dirs = array_filter(scandir(self::SRC) ?: [], static fn (string $d): bool => is_dir(self::SRC.'/'.$d) && !\in_array($d, ['.', '..', ...self::ALLOWED], true));

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
