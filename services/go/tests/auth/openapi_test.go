package auth_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/snowaa-desigram/backend/services/go/internal/auth"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/zeromicro/go-zero/rest"
)

const specPath = "../../../../openapi/auth.yaml"

var (
	specOnce sync.Once
	spec     *openapi3.T
	specErr  error
)

// loadSpec грузит и валидирует backend/openapi/auth.yaml (один раз на пакет).
func loadSpec(t *testing.T) *openapi3.T {
	t.Helper()
	specOnce.Do(func() {
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true // Error — из common.yaml
		spec, specErr = loader.LoadFromFile(specPath)
		if specErr == nil {
			specErr = spec.Validate(context.Background())
		}
	})
	if specErr != nil {
		t.Fatalf("openapi spec: %v", specErr)
	}
	return spec
}

func TestOpenAPISpecIsValid(t *testing.T) {
	s := loadSpec(t)
	if s.Info.Title == "" || len(s.Paths.Map()) == 0 {
		t.Fatal("spec is empty")
	}
}

// Маршруты сервиса и операции в спеке должны совпадать 1:1 (кроме /health — он не для клиентов).
func TestRoutesMatchSpec(t *testing.T) {
	s := loadSpec(t)

	inSpec := map[string]bool{}
	for path, item := range s.Paths.Map() {
		for method := range item.Operations() {
			inSpec[method+" "+path] = true
		}
	}

	public, protected := auth.Routes(auth.NewHandler(nil))
	inCode := map[string]bool{}
	for _, r := range append(public, protected...) {
		inCode[strings.ToUpper(r.Method)+" "+r.Path] = true
	}

	for k := range inSpec {
		if !inCode[k] {
			t.Errorf("in spec, not in auth.Routes(): %s", k)
		}
	}
	for k := range inCode {
		if !inSpec[k] {
			t.Errorf("in auth.Routes(), not in spec: %s", k)
		}
	}
}

// Защищённые маршруты — ровно те, у которых в спеке security: bearerAuth.
func TestProtectedRoutesMatchSpecSecurity(t *testing.T) {
	s := loadSpec(t)
	_, protected := auth.Routes(auth.NewHandler(nil))

	secured := map[string]bool{}
	for path, item := range s.Paths.Map() {
		for method, op := range item.Operations() {
			if op.Security != nil && len(*op.Security) > 0 {
				secured[method+" "+path] = true
			}
		}
	}
	got := map[string]bool{}
	for _, r := range protected {
		got[strings.ToUpper(r.Method)+" "+r.Path] = true
	}
	for k := range secured {
		if !got[k] {
			t.Errorf("spec requires bearerAuth, route is public: %s", k)
		}
	}
	for k := range got {
		if !secured[k] {
			t.Errorf("route is protected, spec has no security: %s", k)
		}
	}
}

var _ = rest.Route{} // rest используется в auth.Routes()
