// Архитектура Go-сервисов (openspec/specs/architecture-go-service): слои internal/<name>/{transport,service,store,adapter},
// импорты только вниз, сервисы не импортируют друг друга, тестов внутри internal/ нет.
// Обходит internal/ парсером без сборки — новый сервис попадает под проверку без правок.
package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const modulePrefix = "github.com/snowaa-desigram/backend/services/go/internal/"

// Что каждому слою можно импортировать из своего сервиса. Корень пакета (config.go) — композиция, ему можно всё.
var allowed = map[string][]string{
	"":          {"transport", "service", "store", "adapter"},
	"transport": {"service", "store"},
	"service":   {"store"},
	"adapter":   {"store"},
	"store":     {},
}

func TestLayers(t *testing.T) {
	root := filepath.Join("..", "..", "internal")
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		svc := parts[0]

		if strings.HasSuffix(path, "_test.go") {
			t.Errorf("%s: тесты — только в tests/%s/", rel, svc)
			return nil
		}

		// internal/pkg/<name> — общий код без зависимостей на сервисы.
		layer := ""
		if svc != "pkg" && len(parts) > 2 {
			layer = parts[1]
			if _, ok := allowed[layer]; !ok {
				t.Errorf("%s: слой %q не из спеки (transport|service|store|adapter)", rel, layer)
				return nil
			}
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if !strings.HasPrefix(p, modulePrefix) {
				continue
			}
			target := strings.Split(strings.TrimPrefix(p, modulePrefix), "/")
			switch {
			case target[0] == "pkg":
				// общий код — можно всем
			case svc == "pkg":
				t.Errorf("%s: internal/pkg не импортирует сервисы (%s)", rel, p)
			case target[0] != svc:
				t.Errorf("%s: сервис %s импортирует сервис %s (%s) — общее выносится в internal/pkg", rel, svc, target[0], p)
			case len(target) == 1:
				t.Errorf("%s: слой %q импортирует корень пакета %s — цикл", rel, layer, p)
			case !slices.Contains(allowed[layer], target[1]):
				t.Errorf("%s: слой %q не может импортировать %q (разрешено: %v)", rel, layer, target[1], allowed[layer])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestLayout: у каждого сервиса cmd/<name> есть internal/<name>/config.go и хотя бы transport/ и service/.
func TestLayout(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "cmd"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		base := filepath.Join("..", "..", "internal", name)
		for _, want := range []string{"config.go", "transport", "service"} {
			if _, err := os.Stat(filepath.Join(base, want)); err != nil {
				t.Errorf("сервис %s: нет internal/%s/%s", name, name, want)
			}
		}
		if _, err := os.Stat(filepath.Join("..", name)); err != nil {
			t.Errorf("сервис %s: нет tests/%s/", name, name)
		}
	}
}
